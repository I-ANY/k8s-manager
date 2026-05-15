package services

import (
	"context"
	"fmt"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8soperation/global"
	"k8soperation/internal/app/requests"
	"k8soperation/pkg/k8s/namespace"
	"time"
)

// KubeCreateNamespace 封装 Namespace 创建逻辑（调用资源层）
func (s *Services) KubeCreateNamespace(ctx context.Context, req *requests.KubeNamespaceCreateRequest) (*corev1.Namespace, error) {
	// 这里可以再包一层服务级别的超时（也可以直接把 ctx 透传，下看你项目习惯）
	c, cancel := context.WithTimeout(ctx, 20*time.Second)
	defer cancel()

	return namespace.CreateNamespace(c, req)
}

func (s *Services) KubeNamespaceList(ctx context.Context, param *requests.KubeNamespaceListRequest) ([]corev1.Namespace, int, error) {
	c, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	items, total, err := namespace.GetNamespaceList(c, param.Name, param.Page, param.Limit)
	if err != nil {
		global.Logger.Errorf("list Namespace failed: %v", err)
		return nil, 0, err
	}

	return items, total, nil
}

func (s *Services) KubeNamespaceDetail(ctx context.Context, param *requests.KubeNamespaceDetailRequest) (*corev1.Namespace, error) {
	c, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	nsDetail, err := namespace.GetNamespaceDetail(c, param.Name)
	if err != nil {
		if apierrors.IsNotFound(err) {
			global.Logger.Warnf("Namespace %s not found", param.Name)
			return nil, fmt.Errorf("Namespace %q not found", param.Name)
		}

		global.Logger.Error("get Namespace detail failed", zap.Error(err))
		return nil, err
	}

	return nsDetail, nil
}

func (s *Services) KubeNamespaceDelete(ctx context.Context, param *requests.KubeNamespaceDeleteRequest) error {
	// 设置整个删除过程的超时时间（Namespace 删除较慢，建议给 60 秒）
	c, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()

	// 调用内部逻辑删除 Namespace
	if err := namespace.DeleteNamespace(c, param.Name); err != nil {
		// 不存在视为删除成功（幂等）
		if apierrors.IsNotFound(err) {
			global.Logger.Warnf("Namespace %s not found", param.Name)
			return nil
		}

		global.Logger.Errorf("delete Namespace %s failed: %v", param.Name, err)
		return err
	}

	global.Logger.Infof("Namespace %s deleted successfully", param.Name)
	return nil
}

func (s *Services) KubeNamespaceUpdate(ctx context.Context, param *requests.KubeNamespaceUpdateRequest) (*corev1.Namespace, error) {
	c, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	updated, err := namespace.PatchNamespace(c, param.Name, param.Content)
	if err != nil {
		global.Logger.Errorf("update Namespace %s failed: %v", param.Name, err)
		return nil, err
	}

	global.Logger.Infof("Namespace %s updated successfully", param.Name)
	return updated, nil
}
