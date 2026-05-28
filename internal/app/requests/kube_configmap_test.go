package requests

import (
	"strings"
	"testing"
)

func TestConfigMapCreateAllowsDefaultClusterID(t *testing.T) {
	// Test with ClusterID omitted (zero value)
	req := &KubeConfigMapCreateRequest{
		Namespace: "default",
		Name:      "app-config",
		Data:      map[string]string{"key": "value"},
	}
	errs := ValidKubeConfigMapCreateRequest(req, nil)
	if len(errs) > 0 {
		t.Fatalf("expected valid request with default cluster ID, got %v", errs)
	}

	// Test with explicit ClusterID = 0
	req2 := &KubeConfigMapCreateRequest{
		ClusterID: 0,
		Namespace: "default",
		Name:      "app-config-2",
		Data:      map[string]string{"key2": "value2"},
	}
	errs2 := ValidKubeConfigMapCreateRequest(req2, nil)
	if len(errs2) > 0 {
		t.Fatalf("expected valid request with explicit cluster ID 0, got %v", errs2)
	}
}

func TestConfigMapUpdateAllowsDefaultClusterID(t *testing.T) {
	// Test with ClusterID omitted (zero value)
	req := &KubeConfigMapUpdateRequest{
		Namespace: "default",
		Name:      "app-config",
		Content:   `{"data":{"key":"value"}}`,
	}
	errs := ValidKubeConfigMapUpdateRequest(req, nil)
	if len(errs) > 0 {
		t.Fatalf("expected valid update request with default cluster ID, got %v", errs)
	}

	// Test with explicit ClusterID = 0
	req2 := &KubeConfigMapUpdateRequest{
		ClusterID: 0,
		Namespace: "default",
		Name:      "app-config-2",
		Content:   `{"data":{"key2":"value2"}}`,
	}
	errs2 := ValidKubeConfigMapUpdateRequest(req2, nil)
	if len(errs2) > 0 {
		t.Fatalf("expected valid update request with explicit cluster ID 0, got %v", errs2)
	}
}

func TestConfigMapUpdateRequiresContent(t *testing.T) {
	// Test with non-zero ClusterID but empty content should fail
	req := &KubeConfigMapUpdateRequest{
		ClusterID: 1,
		Namespace: "default",
		Name:      "app-config",
		Content:   "", // Empty content
	}
	errs := ValidKubeConfigMapUpdateRequest(req, nil)
	if errs == nil {
		t.Fatal("expected validation error for empty content, got none")
	}

	// Check that the error is about content
	if contentErrs, ok := errs["content"]; !ok || len(contentErrs) == 0 {
		t.Fatalf("expected content validation error, got errors: %v", errs)
	}
	// Also verify the error message matches
	if !strings.Contains(errs["content"][0], "content 不能为空") {
		t.Fatalf("expected content required error message, got: %v", errs["content"][0])
	}

	// Test with non-zero ClusterID and non-empty content should pass
	req2 := &KubeConfigMapUpdateRequest{
		ClusterID: 1,
		Namespace: "default",
		Name:      "app-config-2",
		Content:   `{"data":{"key":"value"}}`,
	}
	errs2 := ValidKubeConfigMapUpdateRequest(req2, nil)
	if len(errs2) > 0 {
		t.Fatalf("expected valid update request with non-zero cluster ID and content, got %v", errs2)
	}

	// Test with content that's just spaces should fail
	req3 := &KubeConfigMapUpdateRequest{
		ClusterID: 1,
		Namespace: "default",
		Name:      "app-config-3",
		Content:   "   ", // Just spaces
	}
	errs3 := ValidKubeConfigMapUpdateRequest(req3, nil)
	if errs3 == nil {
		t.Fatal("expected validation error for spaces-only content, got none")
	}
}

func TestConfigMapListAllowsDefaultClusterID(t *testing.T) {
	// Test with ClusterID omitted (zero value)
	req := &KubeConfigMapListRequest{
		KubeCommonRequest: KubeCommonRequest{
			Namespace: "default",
		},
		Page:  1,
		Limit: 10,
	}
	errs := ValidKubeConfigMapListRequest(req, nil)
	if len(errs) > 0 {
		t.Fatalf("expected valid list request with default cluster ID, got %v", errs)
	}

	// Test with explicit ClusterID = 0
	req2 := &KubeConfigMapListRequest{
		KubeCommonRequest: KubeCommonRequest{
			ClusterID: 0,
			Namespace: "default",
		},
		Page:  1,
		Limit: 10,
	}
	errs2 := ValidKubeConfigMapListRequest(req2, nil)
	if len(errs2) > 0 {
		t.Fatalf("expected valid list request with explicit cluster ID 0, got %v", errs2)
	}
}

func TestConfigMapDetailAllowsDefaultClusterID(t *testing.T) {
	// Test with ClusterID omitted (zero value)
	req := &KubeConfigMapDetailRequest{
		KubeCommonRequest: KubeCommonRequest{
			Namespace: "default",
			Name:      "test-config",
		},
	}
	errs := ValidKubeConfigMapDetailRequest(req, nil)
	if len(errs) > 0 {
		t.Fatalf("expected valid detail request with default cluster ID, got %v", errs)
	}

	// Test with explicit ClusterID = 0
	req2 := &KubeConfigMapDetailRequest{
		KubeCommonRequest: KubeCommonRequest{
			ClusterID: 0,
			Namespace: "default",
			Name:      "test-config-2",
		},
	}
	errs2 := ValidKubeConfigMapDetailRequest(req2, nil)
	if len(errs2) > 0 {
		t.Fatalf("expected valid detail request with explicit cluster ID 0, got %v", errs2)
	}
}

func TestConfigMapDeleteAllowsDefaultClusterID(t *testing.T) {
	// Test with ClusterID omitted (zero value)
	req := &KubeConfigMapDeleteRequest{
		KubeCommonRequest: KubeCommonRequest{
			Namespace: "default",
			Name:      "test-config",
		},
	}
	errs := ValidKubeConfigMapDeleteRequest(req, nil)
	if len(errs) > 0 {
		t.Fatalf("expected valid delete request with default cluster ID, got %v", errs)
	}

	// Test with explicit ClusterID = 0
	req2 := &KubeConfigMapDeleteRequest{
		KubeCommonRequest: KubeCommonRequest{
			ClusterID: 0,
			Namespace: "default",
			Name:      "test-config-2",
		},
	}
	errs2 := ValidKubeConfigMapDeleteRequest(req2, nil)
	if len(errs2) > 0 {
		t.Fatalf("expected valid delete request with explicit cluster ID 0, got %v", errs2)
	}
}

func TestConfigMapUpdateDataAllowsDefaultClusterID(t *testing.T) {
	// Test with ClusterID omitted (zero value)
	req := &KubeConfigMapUpdateDataRequest{
		KubeCommonRequest: KubeCommonRequest{
			Namespace: "default",
			Name:      "test-config",
		},
		Data: map[string]string{"key": "value"},
	}
	errs := ValidKubeConfigMapUpdateDataRequest(req, nil)
	if len(errs) > 0 {
		t.Fatalf("expected valid update data request with default cluster ID, got %v", errs)
	}

	// Test with explicit ClusterID = 0
	req2 := &KubeConfigMapUpdateDataRequest{
		KubeCommonRequest: KubeCommonRequest{
			ClusterID: 0,
			Namespace: "default",
			Name:      "test-config-2",
		},
		Data: map[string]string{"key2": "value2"},
	}
	errs2 := ValidKubeConfigMapUpdateDataRequest(req2, nil)
	if len(errs2) > 0 {
		t.Fatalf("expected valid update data request with explicit cluster ID 0, got %v", errs2)
	}
}

func TestConfigMapNamesAllowsDefaultClusterID(t *testing.T) {
	// Test with ClusterID omitted (zero value)
	req := &KubeConfigMapNamesRequest{
		Namespace: "default",
	}
	errs := ValidKubeConfigMapNamesRequest(req, nil)
	if len(errs) > 0 {
		t.Fatalf("expected valid names request with default cluster ID, got %v", errs)
	}

	// Test with explicit ClusterID = 0
	req2 := &KubeConfigMapNamesRequest{
		ClusterID: 0,
		Namespace: "default",
	}
	errs2 := ValidKubeConfigMapNamesRequest(req2, nil)
	if len(errs2) > 0 {
		t.Fatalf("expected valid names request with explicit cluster ID 0, got %v", errs2)
	}
}
