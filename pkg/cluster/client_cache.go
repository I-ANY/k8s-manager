package cluster

import (
	"context"
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"fmt"
	"sync"
	"time"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	metricsclient "k8s.io/metrics/pkg/client/clientset/versioned"
	"k8soperation/pkg/app"
	"k8soperation/pkg/k8s"
)

// ClientBuilder is a function that builds a Kubernetes client for a given cluster ID.
// It returns the client, a fingerprint string that can be used to detect config changes,
// and any error encountered.
type ClientBuilder func(context.Context, uint32) (*k8s.Client, string, error)

// ClientCache caches Kubernetes clients with TTL expiration and concurrent build serialization.
type ClientCache struct {
	mu      sync.RWMutex
	ttl     time.Duration
	clients map[uint32]*cachedClient
	flights map[uint32]*flight // tracks ongoing builds per cluster ID
}

type cachedClient struct {
	client      *k8s.Client
	fingerprint string
	createdAt   time.Time
}

type flight struct {
	done   chan struct{}
	client *k8s.Client
	fp     string
	err    error
}

// DefaultClientCache is the default shared cache with 30-minute TTL.
var DefaultClientCache = NewClientCache(30 * time.Minute)

// NewClientCache creates a new ClientCache with the specified TTL.
func NewClientCache(ttl time.Duration) *ClientCache {
	return &ClientCache{
		ttl:     ttl,
		clients: make(map[uint32]*cachedClient),
		flights: make(map[uint32]*flight),
	}
}

// Get returns a cached Kubernetes client for the given cluster ID, or builds a new one.
// Concurrent calls for the same cluster ID are serialized - only one builder runs.
func (c *ClientCache) Get(ctx context.Context, clusterID uint32, build ClientBuilder) (*k8s.Client, error) {
	// Nil cache guard
	if c == nil {
		return nil, errors.New("client cache is nil")
	}

	// Check cache first with read lock
	c.mu.RLock()
	if cc, ok := c.clients[clusterID]; ok && (c.ttl == 0 || time.Since(cc.createdAt) < c.ttl) {
		client := cc.client
		c.mu.RUnlock()
		return client, nil
	}
	c.mu.RUnlock()

	// Need to build - acquire write lock
	c.mu.Lock()

	// Double-check after acquiring write lock
	if cc, ok := c.clients[clusterID]; ok && (c.ttl == 0 || time.Since(cc.createdAt) < c.ttl) {
		client := cc.client
		c.mu.Unlock()
		return client, nil
	}

	// Check for ongoing flight
	if f, ok := c.flights[clusterID]; ok {
		// Wait for existing flight to complete
		c.mu.Unlock()
		select {
		case <-f.done:
			// Flight completed, return its results (channel close ensures visibility)
			return f.client, f.err
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}

	// Start new flight
	f := &flight{done: make(chan struct{})}
	c.flights[clusterID] = f

	// Launch builder goroutine
	go func() {
		// Build outside the lock
		client, fp, err := build(ctx, clusterID)

		// Acquire lock to write results and close channel
		c.mu.Lock()
		f.client, f.fp, f.err = client, fp, err
		close(f.done)

		// Update cache if successful
		if err == nil {
			c.clients[clusterID] = &cachedClient{
				client:      client,
				fingerprint: fp,
				createdAt:   time.Now(),
			}
		}
		// Clean up flight entry regardless of success/failure
		delete(c.flights, clusterID)
		c.mu.Unlock()
	}()

	// Release lock and wait for flight to complete
	c.mu.Unlock()

	// Wait for flight to complete
	select {
	case <-f.done:
		return f.client, f.err
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

// GetClient returns a cached Kubernetes client for the given cluster ID using the standard builder.
// This is the public API that doesn't expose the fingerprint.
func (c *ClientCache) GetClient(ctx context.Context, a *app.App, clusterID uint32) (*k8s.Client, error) {
	client, err := c.Get(ctx, clusterID, func(ctx context.Context, id uint32) (*k8s.Client, string, error) {
		return buildClient(ctx, a, id)
	})
	return client, err
}

// buildClient builds a Kubernetes client using the standard pipeline:
// 1. Get rest.Config via GetRestConfig
// 2. Create kubernetes.Interface from config
// 3. Optionally create metrics client
// 4. Detect events v1 support via discovery API
// 5. Generate fingerprint from rest.Config
// Returns the client and fingerprint for internal use.
func buildClient(ctx context.Context, a *app.App, clusterID uint32) (*k8s.Client, string, error) {
	restConfig, err := GetRestConfig(ctx, a, clusterID)
	if err != nil {
		return nil, "", err
	}

	// Create standard Kubernetes client
	k8sClient, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return nil, "", err
	}

	// Create metrics client (log warning on failure but don't fail the request)
	var metricsClient *metricsclient.Clientset
	if mc, err := metricsclient.NewForConfig(restConfig); err != nil {
		if a.Logger != nil {
			a.Logger.Warnf("failed to create metrics client, metrics will be unavailable: clusterID=%d, error=%v",
				clusterID,
				err)
		}
	} else {
		metricsClient = mc
	}

	// Detect events.k8s.io/v1 support using discovery API
	supportsEventsV1 := detectEventsV1Support(k8sClient)

	// Create fingerprint from rest.Config
	fingerprint := generateFingerprint(restConfig)

	client := &k8s.Client{
		Interface:        k8sClient,
		Logger:           a.Logger,
		PodLogSetting:    a.PodLogSetting,
		RestConfig:       restConfig,
		MetricsClient:    metricsClient,
		SupportsEventsV1: supportsEventsV1,
	}

	return client, fingerprint, nil
}

// GetClient returns a client from the default cache for the given cluster ID.
// This is the primary public API for getting cached Kubernetes clients.
func GetClient(ctx context.Context, a *app.App, clusterID uint32) (*k8s.Client, error) {
	return DefaultClientCache.GetClient(ctx, a, clusterID)
}

// detectEventsV1Support checks if the cluster supports events.k8s.io/v1 API
// Uses discovery API instead of listing Events to avoid requiring Event list RBAC.
func detectEventsV1Support(client kubernetes.Interface) bool {
	_, err := client.Discovery().ServerResourcesForGroupVersion("events.k8s.io/v1")
	return err == nil
}

// generateFingerprint creates a stable fingerprint from rest.Config to detect config changes
func generateFingerprint(cfg *rest.Config) string {
	h := sha256.New()

	// Helper to write string safely
	writeString := func(s string) {
		binary.Write(h, binary.LittleEndian, uint64(len(s)))
		h.Write([]byte(s))
	}

	// Helper to write bytes safely
	writeBytes := func(b []byte) {
		binary.Write(h, binary.LittleEndian, uint64(len(b)))
		if len(b) > 0 {
			h.Write(b)
		}
	}

	// Helper to write bool safely
	writeBool := func(b bool) {
		if b {
			h.Write([]byte{1})
		} else {
			h.Write([]byte{0})
		}
	}

	// Write string fields
	writeString(cfg.Host)
	writeString(cfg.APIPath)
	writeString(cfg.ServerName)
	writeString(cfg.BearerToken)
	writeString(cfg.BearerTokenFile)
	writeString(cfg.Username)
	writeString(cfg.Password)
	writeString(cfg.CertFile)
	writeString(cfg.KeyFile)
	writeString(cfg.CAFile)

	// Write byte fields
	writeBytes(cfg.CertData)
	writeBytes(cfg.KeyData)
	writeBytes(cfg.CAData)

	// Write bool fields
	writeBool(cfg.Insecure)

	// Write Proxy function - skip as it's a function pointer not suitable for fingerprinting
	// cfg.Proxy is func(*http.Request) (*url.URL, error) which can't be meaningfully serialized
	writeString("")

	hash := h.Sum(nil)
	return fmt.Sprintf("%x", hash)
}
