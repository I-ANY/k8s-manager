package cluster

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"k8soperation/pkg/k8s"
)

func TestClientCache_NewClientCache(t *testing.T) {
	cache := NewClientCache(30 * time.Minute)
	if cache == nil {
		t.Fatal("NewClientCache returned nil")
	}
}

func TestClientCache_Get_ReuseCachedClient(t *testing.T) {
	cache := NewClientCache(30 * time.Minute)

	buildCount := 0
	builder := func(ctx context.Context, clusterID uint32) (*k8s.Client, string, error) {
		buildCount++
		// Return a dummy client with fingerprint
		return &k8s.Client{}, "fingerprint-" + string(rune(clusterID)), nil
	}

	ctx := context.Background()
	clusterID := uint32(1)

	// First call should build
	client1, err := cache.Get(ctx, clusterID, builder)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if buildCount != 1 {
		t.Errorf("Expected 1 build, got %d", buildCount)
	}

	// Second call should reuse cached client
	client2, err := cache.Get(ctx, clusterID, builder)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if buildCount != 1 {
		t.Errorf("Expected still 1 build after reuse, got %d", buildCount)
	}
	if client1 != client2 {
		t.Error("Expected same client instance, got different")
	}
}

func TestClientCache_Get_RebuildAfterTTL(t *testing.T) {
	// Use very short TTL for test
	cache := NewClientCache(50 * time.Millisecond)

	buildCount := 0
	builder := func(ctx context.Context, clusterID uint32) (*k8s.Client, string, error) {
		buildCount++
		return &k8s.Client{}, "fp", nil
	}

	ctx := context.Background()
	clusterID := uint32(1)

	// First call
	_, err := cache.Get(ctx, clusterID, builder)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if buildCount != 1 {
		t.Errorf("Expected 1 build, got %d", buildCount)
	}

	// Wait for TTL to expire
	time.Sleep(100 * time.Millisecond)

	// Should rebuild
	_, err = cache.Get(ctx, clusterID, builder)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if buildCount != 2 {
		t.Errorf("Expected 2 builds after TTL, got %d", buildCount)
	}
}

func TestClientCache_Get_ConcurrentBuildSerialization(t *testing.T) {
	cache := NewClientCache(30 * time.Minute)

	buildStarted := make(chan struct{})
	buildContinue := make(chan struct{})
	buildCount := 0
	builder := func(ctx context.Context, clusterID uint32) (*k8s.Client, string, error) {
		buildCount++
		if buildCount == 1 {
			close(buildStarted)
			<-buildContinue // Wait for signal
		}
		return &k8s.Client{}, "fp", nil
	}

	ctx := context.Background()
	clusterID := uint32(1)

	var wg sync.WaitGroup
	results := make(chan struct {
		client *k8s.Client
		err    error
	}, 2)

	// First goroutine - will block in builder
	wg.Add(1)
	go func() {
		defer wg.Done()
		client, err := cache.Get(ctx, clusterID, builder)
		results <- struct {
			client *k8s.Client
			err    error
		}{client, err}
	}()

	// Wait for first builder to start
	<-buildStarted

	// Second goroutine - should wait for first to complete
	wg.Add(1)
	go func() {
		defer wg.Done()
		client, err := cache.Get(ctx, clusterID, builder)
		results <- struct {
			client *k8s.Client
			err    error
		}{client, err}
	}()

	// Let first builder complete
	close(buildContinue)
	wg.Wait()
	close(results)

	// Collect results
	var clients []*k8s.Client
	for res := range results {
		if res.err != nil {
			t.Errorf("Get failed: %v", res.err)
		}
		clients = append(clients, res.client)
	}

	// Should have exactly 1 build despite 2 concurrent calls
	if buildCount != 1 {
		t.Errorf("Expected 1 build with concurrent serialization, got %d", buildCount)
	}

	// Both should get same client instance
	if len(clients) != 2 || clients[0] != clients[1] {
		t.Error("Concurrent calls did not get same client instance")
	}
}

func TestClientCache_Get_BuilderError(t *testing.T) {
	cache := NewClientCache(30 * time.Minute)

	expectedErr := errors.New("builder error")
	builder := func(ctx context.Context, clusterID uint32) (*k8s.Client, string, error) {
		return nil, "", expectedErr
	}

	ctx := context.Background()
	client, err := cache.Get(ctx, 1, builder)

	if client != nil {
		t.Error("Expected nil client on builder error")
	}
	if err != expectedErr {
		t.Errorf("Expected error %v, got %v", expectedErr, err)
	}
}

func TestClientCache_Get_DifferentClusterIDs(t *testing.T) {
	cache := NewClientCache(30 * time.Minute)

	buildCount1 := 0
	buildCount2 := 0
	builder1 := func(ctx context.Context, clusterID uint32) (*k8s.Client, string, error) {
		buildCount1++
		return &k8s.Client{}, "fp1", nil
	}
	builder2 := func(ctx context.Context, clusterID uint32) (*k8s.Client, string, error) {
		buildCount2++
		return &k8s.Client{}, "fp2", nil
	}

	ctx := context.Background()

	// Build for cluster 1
	client1, err := cache.Get(ctx, 1, builder1)
	if err != nil {
		t.Fatalf("Get cluster 1 failed: %v", err)
	}
	if buildCount1 != 1 {
		t.Errorf("Expected 1 build for cluster 1, got %d", buildCount1)
	}

	// Build for cluster 2
	client2, err := cache.Get(ctx, 2, builder2)
	if err != nil {
		t.Fatalf("Get cluster 2 failed: %v", err)
	}
	if buildCount2 != 1 {
		t.Errorf("Expected 1 build for cluster 2, got %d", buildCount2)
	}

	if client1 == client2 {
		t.Error("Expected different client instances for different cluster IDs")
	}
}

func TestClientCache_Get_ContextCancelWhileWaiting(t *testing.T) {
	cache := NewClientCache(30 * time.Minute)

	buildStarted := make(chan struct{})
	buildBlock := make(chan struct{})
	builder := func(ctx context.Context, clusterID uint32) (*k8s.Client, string, error) {
		close(buildStarted)
		<-buildBlock // Block indefinitely
		return &k8s.Client{}, "fp", nil
	}

	ctx, cancel := context.WithCancel(context.Background())
	clusterID := uint32(1)

	var wg sync.WaitGroup
	resultCh := make(chan struct {
		client *k8s.Client
		err    error
	}, 1)

	// First goroutine - will block in builder
	wg.Add(1)
	go func() {
		defer wg.Done()
		client, err := cache.Get(ctx, clusterID, builder)
		resultCh <- struct {
			client *k8s.Client
			err    error
		}{client, err}
	}()

	// Wait for builder to start
	<-buildStarted

	// Second goroutine - should wait for flight
	ctx2, cancel2 := context.WithCancel(context.Background())
	wg.Add(1)
	go func() {
		defer wg.Done()
		client, err := cache.Get(ctx2, clusterID, builder)
		resultCh <- struct {
			client *k8s.Client
			err    error
		}{client, err}
	}()

	// Give second goroutine time to start waiting
	time.Sleep(10 * time.Millisecond)

	// Cancel second goroutine's context
	cancel2()

	// Collect second result (should be context error)
	select {
	case res := <-resultCh:
		if res.err != context.Canceled {
			t.Errorf("Expected context.Canceled error, got %v", res.err)
		}
		if res.client != nil {
			t.Error("Expected nil client on context cancellation")
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("Timeout waiting for cancelled context result")
	}

	// Cancel first context and unblock builder
	cancel()
	close(buildBlock)

	// Wait for first goroutine to finish
	wg.Wait()

	// Drain remaining results
	close(resultCh)
	for range resultCh {
		// Just drain
	}
}

func TestClientCache_Get_NilCache(t *testing.T) {
	var cache *ClientCache
	builder := func(ctx context.Context, clusterID uint32) (*k8s.Client, string, error) {
		return &k8s.Client{}, "fp", nil
	}

	ctx := context.Background()
	client, err := cache.Get(ctx, 1, builder)

	if client != nil {
		t.Error("Expected nil client on nil cache")
	}
	if err == nil || err.Error() != "client cache is nil" {
		t.Errorf("Expected 'client cache is nil' error, got %v", err)
	}
}

func TestClientCache_Get_NoExpiration(t *testing.T) {
	// TTL=0 means no expiration
	cache := NewClientCache(0)

	buildCount := 0
	builder := func(ctx context.Context, clusterID uint32) (*k8s.Client, string, error) {
		buildCount++
		return &k8s.Client{}, "fp", nil
	}

	ctx := context.Background()
	clusterID := uint32(1)

	// First call
	_, err := cache.Get(ctx, clusterID, builder)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if buildCount != 1 {
		t.Errorf("Expected 1 build, got %d", buildCount)
	}

	// Wait a bit
	time.Sleep(50 * time.Millisecond)

	// Should still reuse cached client (no expiration)
	_, err = cache.Get(ctx, clusterID, builder)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if buildCount != 1 {
		t.Errorf("Expected still 1 build with TTL=0, got %d", buildCount)
	}
}
