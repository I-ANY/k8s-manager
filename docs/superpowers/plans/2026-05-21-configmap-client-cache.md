# ConfigMap Per-Request K8s Client Cache Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Make ConfigMap APIs select Kubernetes clients by `cluster_id` per request and reuse clients through a safe in-process cache instead of depending on global `app.App.KubeClient`.

**Architecture:** Add a focused client cache under `pkg/cluster` that caches `*pkg/k8s.Client` by cluster ID and kubeconfig fingerprint. `Services.K8sClient(ctx, clusterID)` becomes the service-layer entry point for Kubernetes clients, and ConfigMap service methods use it before calling existing `pkg/k8s/configmap` helpers. ConfigMap request DTOs carry `cluster_id`, with `0` preserving the existing default-cluster fallback through `cluster.GetRestConfig`.

**Tech Stack:** Go 1.25, client-go, Kubernetes `rest.Config`, GORM-backed cluster metadata, standard library `sync`, `crypto/sha256`, `time`, `go test`.

---

## File Structure

- Create `pkg/cluster/client_cache.go`: in-process cache for `*k8s.Client`, keyed by cluster ID plus a kubeconfig/default-source fingerprint; exposes `DefaultClientCache` and `GetClient(ctx, a, clusterID)`.
- Create `pkg/cluster/client_cache_test.go`: unit tests for cache hit reuse, cache invalidation when kubeconfig changes, and singleflight-like concurrent initialization behavior.
- Modify `internal/app/services/services.go`: add `K8sClient(ctx, clusterID)` wrapper so services do not import cache internals everywhere.
- Modify `internal/app/services/kube_configmap.go`: replace `s.App().K8sClient()` with `s.K8sClient(ctx, req.ClusterID)` for ConfigMap APIs.
- Modify `internal/app/requests/common.go`: add `ClusterID uint32` to `KubeCommonRequest` so list/detail/delete and other common request shapes can carry `cluster_id`.
- Modify `internal/app/requests/kube_configmap.go`: add `ClusterID` to create/update request types that do not embed `KubeCommonRequest`; validate required ConfigMap fields as before.
- Modify `internal/app/controllers/api/v1/configmap/configmap.go`: update Swagger annotations to document `cluster_id` for ConfigMap endpoints.

## Task 1: Add Cluster Client Cache

**Files:**
- Create: `pkg/cluster/client_cache.go`
- Test: `pkg/cluster/client_cache_test.go`

- [ ] **Step 1: Write failing cache tests**

Create `pkg/cluster/client_cache_test.go` with these tests. They use a fake builder and do not require a real Kubernetes cluster.

```go
package cluster

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"k8soperation/pkg/k8s"
)

func TestClientCacheReusesClientForSameFingerprint(t *testing.T) {
	cache := NewClientCache(time.Hour)
	var builds int32
	builder := func(context.Context, uint32) (*k8s.Client, string, error) {
		atomic.AddInt32(&builds, 1)
		return &k8s.Client{}, "fingerprint-a", nil
	}

	first, err := cache.Get(context.Background(), 1, builder)
	if err != nil {
		t.Fatalf("first get failed: %v", err)
	}
	second, err := cache.Get(context.Background(), 1, builder)
	if err != nil {
		t.Fatalf("second get failed: %v", err)
	}
	if first != second {
		t.Fatalf("expected cached client pointer to be reused")
	}
	if builds != 1 {
		t.Fatalf("expected 1 build, got %d", builds)
	}
}

func TestClientCacheRebuildsAfterTTLAndFingerprintChange(t *testing.T) {
	cache := NewClientCache(time.Nanosecond)
	var builds int32
	fingerprint := "fingerprint-a"
	builder := func(context.Context, uint32) (*k8s.Client, string, error) {
		atomic.AddInt32(&builds, 1)
		return &k8s.Client{}, fingerprint, nil
	}

	first, err := cache.Get(context.Background(), 1, builder)
	if err != nil {
		t.Fatalf("first get failed: %v", err)
	}
	time.Sleep(time.Millisecond)
	fingerprint = "fingerprint-b"
	second, err := cache.Get(context.Background(), 1, builder)
	if err != nil {
		t.Fatalf("second get failed: %v", err)
	}
	if first == second {
		t.Fatalf("expected client to be rebuilt after fingerprint changed")
	}
	if builds != 2 {
		t.Fatalf("expected 2 builds, got %d", builds)
	}
}

func TestClientCacheSerializesConcurrentBuilds(t *testing.T) {
	cache := NewClientCache(time.Hour)
	var builds int32
	builder := func(context.Context, uint32) (*k8s.Client, string, error) {
		atomic.AddInt32(&builds, 1)
		time.Sleep(10 * time.Millisecond)
		return &k8s.Client{}, "fingerprint-a", nil
	}

	const workers = 10
	clients := make([]*k8s.Client, workers)
	var wg sync.WaitGroup
	wg.Add(workers)
	for i := 0; i < workers; i++ {
		go func(i int) {
			defer wg.Done()
			client, err := cache.Get(context.Background(), 1, builder)
			if err != nil {
				t.Errorf("get failed: %v", err)
				return
			}
			clients[i] = client
		}(i)
	}
	wg.Wait()

	for i := 1; i < workers; i++ {
		if clients[i] != clients[0] {
			t.Fatalf("expected all workers to receive same cached client")
		}
	}
	if builds != 1 {
		t.Fatalf("expected 1 build, got %d", builds)
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run:

```bash
go test ./pkg/cluster -run 'TestClientCache' -v
```

Expected: FAIL with errors like `undefined: NewClientCache`.

- [ ] **Step 3: Implement the cache**

Create `pkg/cluster/client_cache.go`:

```go
package cluster

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sync"
	"time"

	"k8soperation/pkg/app"
	"k8soperation/pkg/k8s"

	"k8s.io/client-go/kubernetes"
	metricsclient "k8s.io/metrics/pkg/client/clientset/versioned"
)

type ClientBuilder func(context.Context, uint32) (*k8s.Client, string, error)

type ClientCache struct {
	mu      sync.Mutex
	ttl     time.Duration
	clients map[uint32]*cachedClient
	flights map[uint32]*clientFlight
}

type cachedClient struct {
	client      *k8s.Client
	fingerprint string
	createdAt   time.Time
}

type clientFlight struct {
	ready  chan struct{}
	client *k8s.Client
	err    error
}

var DefaultClientCache = NewClientCache(30 * time.Minute)

func NewClientCache(ttl time.Duration) *ClientCache {
	return &ClientCache{
		ttl:     ttl,
		clients: make(map[uint32]*cachedClient),
		flights: make(map[uint32]*clientFlight),
	}
}

func (c *ClientCache) Get(ctx context.Context, clusterID uint32, build ClientBuilder) (*k8s.Client, error) {
	if c == nil {
		return nil, fmt.Errorf("client cache is nil")
	}
	if client := c.cached(clusterID); client != nil {
		return client, nil
	}

	flight, owner := c.startFlight(clusterID)
	if !owner {
		select {
		case <-flight.ready:
			return flight.client, flight.err
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}

	client, fingerprint, err := build(ctx, clusterID)
	c.finishFlight(clusterID, flight, client, fingerprint, err)
	return client, err
}

func (c *ClientCache) cached(clusterID uint32) *k8s.Client {
	c.mu.Lock()
	defer c.mu.Unlock()

	entry := c.clients[clusterID]
	if entry == nil {
		return nil
	}
	if c.ttl > 0 && time.Since(entry.createdAt) > c.ttl {
		delete(c.clients, clusterID)
		return nil
	}
	return entry.client
}

func (c *ClientCache) startFlight(clusterID uint32) (*clientFlight, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if entry := c.clients[clusterID]; entry != nil {
		return &clientFlight{ready: closedReady(), client: entry.client}, false
	}
	if flight := c.flights[clusterID]; flight != nil {
		return flight, false
	}
	flight := &clientFlight{ready: make(chan struct{})}
	c.flights[clusterID] = flight
	return flight, true
}

func (c *ClientCache) finishFlight(clusterID uint32, flight *clientFlight, client *k8s.Client, fingerprint string, err error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if err == nil {
		if entry := c.clients[clusterID]; entry == nil || entry.fingerprint != fingerprint {
			c.clients[clusterID] = &cachedClient{client: client, fingerprint: fingerprint, createdAt: time.Now()}
		} else {
			client = entry.client
		}
	}
	flight.client = client
	flight.err = err
	delete(c.flights, clusterID)
	close(flight.ready)
}

func closedReady() chan struct{} {
	ch := make(chan struct{})
	close(ch)
	return ch
}

func GetClient(ctx context.Context, a *app.App, clusterID uint32) (*k8s.Client, error) {
	return DefaultClientCache.Get(ctx, clusterID, func(ctx context.Context, clusterID uint32) (*k8s.Client, string, error) {
		cfg, err := GetRestConfig(ctx, a, clusterID)
		if err != nil {
			return nil, "", err
		}

		kube, err := kubernetes.NewForConfig(cfg)
		if err != nil {
			return nil, "", fmt.Errorf("create k8s client failed: %w", err)
		}

		var metrics *metricsclient.Clientset
		if m, err := metricsclient.NewForConfig(cfg); err == nil {
			metrics = m
		}

		fingerprint := restConfigFingerprint(cfg)
		client := &k8s.Client{
			Interface:        kube,
			Logger:           a.Logger,
			PodLogSetting:    a.PodLogSetting,
			RestConfig:       cfg,
			MetricsClient:    metrics,
			SupportsEventsV1: detectEventsV1(kube),
		}
		return client, fingerprint, nil
	})
}

func restConfigFingerprint(cfg interface{ String() string }) string {
	sum := sha256.Sum256([]byte(cfg.String()))
	return hex.EncodeToString(sum[:])
}

func detectEventsV1(kube *kubernetes.Clientset) bool {
	_, err := kube.Discovery().ServerResourcesForGroupVersion("events.k8s.io/v1")
	return err == nil
}
```

- [ ] **Step 4: Run cache tests**

Run:

```bash
go test ./pkg/cluster -run 'TestClientCache' -v
```

Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add pkg/cluster/client_cache.go pkg/cluster/client_cache_test.go
git commit -m "添加 Kubernetes 客户端缓存"
```

## Task 2: Add Service-Level Client Entry Point

**Files:**
- Modify: `internal/app/services/services.go`
- Test: `pkg/cluster/client_cache_test.go`

- [ ] **Step 1: Write failing test for nil cache guard**

Append this test to `pkg/cluster/client_cache_test.go`:

```go
func TestNilClientCacheReturnsError(t *testing.T) {
	var cache *ClientCache
	_, err := cache.Get(context.Background(), 1, func(context.Context, uint32) (*k8s.Client, string, error) {
		return &k8s.Client{}, "fingerprint-a", nil
	})
	if err == nil {
		t.Fatalf("expected error for nil cache")
	}
}
```

- [ ] **Step 2: Run test**

Run:

```bash
go test ./pkg/cluster -run 'TestNilClientCacheReturnsError' -v
```

Expected: PASS because Task 1 included the nil guard. If it fails, add this guard at the top of `(*ClientCache).Get`:

```go
if c == nil {
	return nil, fmt.Errorf("client cache is nil")
}
```

- [ ] **Step 3: Add service method**

Modify `internal/app/services/services.go` imports and add the method below `App()`:

```go
import (
	"context"

	"k8soperation/internal/app/dao"
	"k8soperation/pkg/app"
	"k8soperation/pkg/cluster"
	"k8soperation/pkg/k8s"
)
```

```go
func (s *Services) K8sClient(ctx context.Context, clusterID uint32) (*k8s.Client, error) {
	return cluster.GetClient(ctx, s.app, clusterID)
}
```

- [ ] **Step 4: Run package tests**

Run:

```bash
go test ./internal/app/services ./pkg/cluster -v
```

Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/app/services/services.go pkg/cluster/client_cache_test.go
git commit -m "添加服务层 Kubernetes 客户端入口"
```

## Task 3: Add cluster_id to ConfigMap Requests

**Files:**
- Modify: `internal/app/requests/common.go`
- Modify: `internal/app/requests/kube_configmap.go`
- Test: `internal/app/requests/kube_configmap_test.go`

- [ ] **Step 1: Write request validation tests**

Create `internal/app/requests/kube_configmap_test.go`:

```go
package requests

import "testing"

func TestConfigMapCreateAllowsClusterID(t *testing.T) {
	req := &KubeConfigMapCreateRequest{
		ClusterID: 1,
		Namespace: "default",
		Name:      "app-config",
		Data:      map[string]string{"key": "value"},
	}
	if errs := ValidKubeConfigMapCreateRequest(req, nil); errs != nil {
		t.Fatalf("expected valid request, got %v", errs)
	}
}

func TestConfigMapUpdateAllowsClusterID(t *testing.T) {
	req := &KubeConfigMapUpdateRequest{
		ClusterID: 1,
		Namespace: "default",
		Name:      "app-config",
		Content:   `{"data":{"key":"value"}}`,
	}
	if errs := ValidKubeConfigMapUpdateRequest(req, nil); errs != nil {
		t.Fatalf("expected valid request, got %v", errs)
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run:

```bash
go test ./internal/app/requests -run 'TestConfigMap.*ClusterID' -v
```

Expected: FAIL with `unknown field ClusterID`.

- [ ] **Step 3: Add ClusterID to common and ConfigMap request types**

Modify `internal/app/requests/common.go`:

```go
type KubeCommonRequest struct {
	ClusterID  uint32 `json:"cluster_id" form:"cluster_id" valid:"cluster_id"`
	Name       string `json:"name" form:"name" valid:"name"`
	Namespace  string `json:"namespace" form:"namespace" valid:"namespace"`
}
```

Modify `internal/app/requests/kube_configmap.go` create request:

```go
type KubeConfigMapCreateRequest struct {
	ClusterID   uint32            `json:"cluster_id" form:"cluster_id" valid:"cluster_id"`
	Namespace   string            `json:"namespace" valid:"namespace"`
	Name        string            `json:"name"      valid:"name"`
	Labels      map[string]string `json:"labels"       swaggertype:"string" valid:"-"`
	Annotations map[string]string `json:"annotations"  swaggertype:"string" valid:"-"`
	Data        map[string]string `json:"data,omitempty"       swaggertype:"string" valid:"-"`
	BinaryData  map[string][]byte `json:"binaryData,omitempty"                      valid:"-"`
}
```

Modify `internal/app/requests/kube_configmap.go` update request:

```go
type KubeConfigMapUpdateRequest struct {
	ClusterID uint32 `json:"cluster_id" form:"cluster_id" valid:"cluster_id"`
	Namespace string `json:"namespace" valid:"namespace"`
	Name      string `json:"name"      valid:"name"`
	Content   string `json:"content"   valid:"required"`
}
```

- [ ] **Step 4: Run request tests**

Run:

```bash
go test ./internal/app/requests -run 'TestConfigMap.*ClusterID' -v
```

Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/app/requests/common.go internal/app/requests/kube_configmap.go internal/app/requests/kube_configmap_test.go
git commit -m "为 ConfigMap 请求添加集群 ID"
```

## Task 4: Switch ConfigMap Service to Dynamic Cached Client

**Files:**
- Modify: `internal/app/services/kube_configmap.go`
- Test: `internal/app/services` compile test through `go test`

- [ ] **Step 1: Replace Create with cached client lookup**

Modify `internal/app/services/kube_configmap.go`:

```go
func (s *Services) KubeCreateConfigMap(ctx context.Context,
	req *requests.KubeConfigMapCreateRequest,
) (*corev1.ConfigMap, error) {
	client, err := s.K8sClient(ctx, req.ClusterID)
	if err != nil {
		return nil, err
	}
	return configmap.CreateConfigMap(client, ctx, req)
}
```

- [ ] **Step 2: Replace List with cached client lookup**

Modify `KubeConfigMapList`:

```go
func (s *Services) KubeConfigMapList(ctx context.Context, param *requests.KubeConfigMapListRequest) ([]corev1.ConfigMap, int, error) {
	client, err := s.K8sClient(ctx, param.ClusterID)
	if err != nil {
		return nil, 0, err
	}
	return configmap.GetConfigMapList(client, ctx, param.Name, param.Namespace, param.Page, param.Limit)
}
```

- [ ] **Step 3: Replace Detail with cached client lookup**

Modify `KubeConfigMapDetail`:

```go
func (s *Services) KubeConfigMapDetail(ctx context.Context, param *requests.KubeConfigMapDetailRequest) (*corev1.ConfigMap, error) {
	client, err := s.K8sClient(ctx, param.ClusterID)
	if err != nil {
		return nil, err
	}
	return configmap.GetConfigMapDetail(client, ctx, param.Name, param.Namespace)
}
```

- [ ] **Step 4: Replace Delete with cached client lookup**

Modify `KubeConfigMapDelete`:

```go
func (s *Services) KubeConfigMapDelete(ctx context.Context, param *requests.KubeConfigMapDeleteRequest) error {
	client, err := s.K8sClient(ctx, param.ClusterID)
	if err != nil {
		return err
	}
	return configmap.DeleteConfigMap(client, ctx, param.Name, param.Namespace)
}
```

- [ ] **Step 5: Update patch and update methods in same file**

If `kube_configmap.go` contains `KubeConfigMapPatch` and `KubeConfigMapUpdate` below the initially read section, update them to this shape:

```go
func (s *Services) KubeConfigMapPatch(ctx context.Context, param *requests.KubeConfigMapUpdateRequest) (*corev1.ConfigMap, error) {
	client, err := s.K8sClient(ctx, param.ClusterID)
	if err != nil {
		return nil, err
	}
	return configmap.PatchConfigMap(client, ctx, param.Namespace, param.Name, []byte(param.Content))
}

func (s *Services) KubeConfigMapUpdate(ctx context.Context, param *requests.KubeConfigMapUpdateRequest) (*corev1.ConfigMap, error) {
	client, err := s.K8sClient(ctx, param.ClusterID)
	if err != nil {
		return nil, err
	}
	return configmap.PatchConfigMapJson(client, ctx, param.Namespace, param.Content)
}
```

- [ ] **Step 6: Run service compile test**

Run:

```bash
go test ./internal/app/services -v
```

Expected: PASS.

- [ ] **Step 7: Commit**

```bash
git add internal/app/services/kube_configmap.go
git commit -m "ConfigMap 接口使用动态 Kubernetes 客户端"
```

## Task 5: Document ConfigMap cluster_id API Parameters

**Files:**
- Modify: `internal/app/controllers/api/v1/configmap/configmap.go`

- [ ] **Step 1: Update Swagger annotations**

Modify ConfigMap controller annotations so each endpoint documents `cluster_id`:

For create endpoint, add before the body param:

```go
// @Param       cluster_id  body  uint32  false  "集群ID，不传则使用默认集群"
```

For list/detail/delete endpoints, add:

```go
// @Param cluster_id query int false "集群ID，不传则使用默认集群"
```

For patch and patch-json endpoints, add:

```go
// @Param cluster_id query int false "集群ID，不传则使用默认集群"
```

- [ ] **Step 2: Run Swagger generation**

Run:

```bash
make swag
```

Expected: command exits 0 and updates `docs/` Swagger artifacts if annotations are valid.

- [ ] **Step 3: Commit**

```bash
git add internal/app/controllers/api/v1/configmap/configmap.go docs/docs.go docs/swagger.json docs/swagger.yaml
git commit -m "更新 ConfigMap 集群参数文档"
```

## Task 6: Final Verification

**Files:**
- No source changes expected unless verification exposes a defect.

- [ ] **Step 1: Format code**

Run:

```bash
make fmt
```

Expected: exits 0.

- [ ] **Step 2: Run tests**

Run:

```bash
make test
```

Expected: exits 0.

- [ ] **Step 3: Run vet**

Run:

```bash
make lint
```

Expected: exits 0.

- [ ] **Step 4: Build**

Run:

```bash
make build
```

Expected: exits 0 and produces `./bin/k8soperation`.

- [ ] **Step 5: Commit verification fixes if any**

If formatting or verification changed files, commit them:

```bash
git add <changed-files>
git commit -m "修复 ConfigMap 客户端缓存验证问题"
```

If no files changed, do not create an empty commit.

---

## Self-Review

- Spec coverage: The plan covers request-level `cluster_id`, in-process client caching, ConfigMap-only rollout, cache reuse, cache invalidation by TTL/fingerprint, and verification.
- Placeholder scan: No `TBD`, `TODO`, or unspecified implementation steps remain.
- Type consistency: `ClientCache`, `GetClient`, `Services.K8sClient`, request `ClusterID`, and ConfigMap service call signatures are defined before use.
