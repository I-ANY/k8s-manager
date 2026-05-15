package services

import (
	"context"
	"fmt"
	appv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8soperation/global"
	"k8soperation/internal/app/requests"
	"k8soperation/pkg/k8s/daemonset"
	"strings"
	"time"
)

// 创建 DaemonSet（可选同时创建 Service）
func (s *Services) KubeDaemonSetCreate(ctx context.Context, req *requests.KubeDaemonSetCreateRequest) (*appv1.DaemonSet, *corev1.Service, error) {
	// DaemonSet 创建一般给更宽裕些的超时
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// 1) 创建 DaemonSet
	ds, err := daemonset.CreateDaemonSet(ctx, req)
	if err != nil {
		return nil, nil, fmt.Errorf("create daemonset failed: %w", err)
	}

	// 2) 按需创建 Service
	var svcObj *corev1.Service
	if req.IsCreateService {
		svcObj, err = daemonset.CreateServiceFromDaemonSet(ctx, req)
		if err != nil {
			if apierrors.IsAlreadyExists(err) {
				svcName := strings.TrimSpace(req.ServiceName)
				if svcName == "" {
					svcName = req.Name
				}
				if exist, gerr := global.KubeClient.CoreV1().
					Services(req.Namespace).
					Get(ctx, svcName, metav1.GetOptions{}); gerr == nil {
					global.Logger.Infof("service %s/%s already exists, reuse it", req.Namespace, svcName)
					return ds, exist, nil
				}
				// Get 失败才继续回滚
			}

			// Service 真失败 → 回滚删除 DaemonSet（带传播策略，清理 Pods）
			pol := metav1.DeletePropagationForeground // 或 Background
			if delErr := global.KubeClient.AppsV1().
				DaemonSets(req.Namespace).
				Delete(ctx, ds.Name, metav1.DeleteOptions{PropagationPolicy: &pol}); delErr != nil {
				global.Logger.Errorf("rollback delete daemonset %s/%s failed: %v", req.Namespace, ds.Name, delErr)
			}
			return nil, nil, fmt.Errorf("create service failed: %w", err)
		}
	}

	return ds, svcObj, nil
}

func (s *Services) KubeDaemonSetList(ctx context.Context, param *requests.KubeDaemonSetListRequest) ([]appv1.DaemonSet, int, error) {
	return daemonset.GetDaemonSetList(ctx, param.Name, param.Namespace, param.Page, param.Limit)
}

func (s *Services) KubeDaemonSetDetail(ctx context.Context, param *requests.KubeDaemonSetDetailRequest) (*appv1.DaemonSet, error) {
	return daemonset.GetDaemonSetDetail(ctx, param.Name, param.Namespace)
}

// 删除 DaemonSet
func (s *Services) KubeDaemonSetDelete(ctx context.Context, param *requests.KubeDaemonSetDeleteRequest) error {
	return daemonset.DeleteDaemonSet(ctx, param.Name, param.Namespace)
}

// 删除 DaemonSet 对应的 Service（如果有）
func (s *Services) KubeDaemonSetDeleteService(ctx context.Context, param *requests.KubeDaemonSetDeleteRequest) error {
	return daemonset.DeleteDaemonSetService(ctx, param.Name, param.Namespace)
}

// 更新 DaemonSet 的镜像
func (s *Services) KubeDaemonSetUpdateImage(ctx context.Context, param *requests.KubeDaemonSetUpdateImageRequest) (*appv1.DaemonSet, error) {
	return daemonset.PatchUpdateDaemonSetImage(ctx, param.Namespace, param.Name, param.Container, param.Image)
}

// 重启 DaemonSet
func (s *Services) KubeDaemonSetRestart(ctx context.Context, param *requests.KubeDaemonSetRestartRequest) error {
	return daemonset.RestartDaemonSet(ctx, param.Namespace, param.Name)
}

// 回滚到指定版本
func (s *Services) KubeDaemonSetRollback(ctx context.Context, param *requests.KubeDaemonSetRollbackRequest) (*appv1.DaemonSet, error) {
	return daemonset.RollbackDaemonSet(ctx, param.Name, param.Namespace, param.RevisionName)
}
