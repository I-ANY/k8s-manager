package services

import (
	"context"
	"encoding/json"
	"fmt"
	corev1 "k8s.io/api/core/v1"
	"k8soperation/internal/app/requests"
	"k8soperation/pkg/k8s/svc"
)

func (s *Services) KubeCreateService(ctx context.Context, req *requests.KubeServiceCreateRequest) (*corev1.Service, error) {
	return svc.CreateService(ctx, req)
}

func (s *Services) KubeServiceList(ctx context.Context, param *requests.KubeServiceListRequest) ([]corev1.Service, int, error) {
	return svc.GetServiceList(ctx, param.Name, param.Namespace, param.Page, param.Limit)
}

func (s *Services) KubeServiceDetail(ctx context.Context, param *requests.KubeServiceDetailRequest) (*corev1.Service, error) {
	return svc.GetServiceDetail(ctx, param.Name, param.Namespace)
}

func (s *Services) KubeServiceDelete(ctx context.Context, param *requests.KubeServiceDeleteRequest) error {
	return svc.DeleteService(ctx, param.Name, param.Namespace)
}

func (s *Services) KubeUpdateServiceTemplate(ctx context.Context, param *requests.KubeServiceUpdateRequest) (*corev1.Service, error) {
	return svc.PatchService(ctx, param.Namespace, param.Name, []byte(param.Content))
}

// Strategic Merge Patch（结构合并）
func (s *Services) KubeServicePatch(ctx context.Context, param *requests.KubeServiceUpdateRequest) (*corev1.Service, error) {
	return svc.PatchService(ctx, param.Namespace, param.Name, []byte(param.Content))
}

// JSON Merge Patch（覆盖式更新）
func (s *Services) KubeServicePatchJSON(ctx context.Context, param *requests.KubeServiceUpdateRequest) (*corev1.Service, error) {
	// 建议在这里做 JSON 合法性校验，返回更友好的错误
	if !json.Valid([]byte(param.Content)) {
		return nil, fmt.Errorf("invalid json")
	}
	return svc.PatchJsonService(ctx, param.Namespace, param.Name, []byte(param.Content))
}

// 获取 ENPINITPOINTS列表
func (s *Services) KubeServiceEndpoints(ctx context.Context, param *requests.KubeServiceEndpointsRequest) (*corev1.Endpoints, error) {
	return svc.GetServiceEndpoints(ctx, param.Name, param.Namespace)
}
