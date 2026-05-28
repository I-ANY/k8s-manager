package services

import (
	"context"
	"k8soperation/internal/app/requests"
	"k8soperation/pkg/k8s/configmap"

	corev1 "k8s.io/api/core/v1"
)

func (s *Services) KubeCreateConfigMap(ctx context.Context,
	req *requests.KubeConfigMapCreateRequest,
) (*corev1.ConfigMap, error) {
	client, err := s.K8sClient(ctx, req.ClusterID)
	if err != nil {
		return nil, err
	}
	return configmap.CreateConfigMap(client, ctx, req)
}

// KubeConfigMapList 获取 ConfigMap 列表（支持名称过滤 + 分页）
func (s *Services) KubeConfigMapList(ctx context.Context, param *requests.KubeConfigMapListRequest) ([]corev1.ConfigMap, int, error) {
	client, err := s.K8sClient(ctx, param.ClusterID)
	if err != nil {
		return nil, 0, err
	}
	return configmap.GetConfigMapList(client, ctx, param.Name, param.Namespace, param.Page, param.Limit)
}

func (s *Services) KubeConfigMapDetail(ctx context.Context, param *requests.KubeConfigMapDetailRequest) (*corev1.ConfigMap, error) {
	client, err := s.K8sClient(ctx, param.ClusterID)
	if err != nil {
		return nil, err
	}
	return configmap.GetConfigMapDetail(client, ctx, param.Name, param.Namespace)
}

func (s *Services) KubeConfigMapDelete(ctx context.Context, param *requests.KubeConfigMapDeleteRequest) error {
	client, err := s.K8sClient(ctx, param.ClusterID)
	if err != nil {
		return err
	}
	return configmap.DeleteConfigMap(client, ctx, param.Name, param.Namespace)
}
