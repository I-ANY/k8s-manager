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
	"k8soperation/pkg/k8s/statefulset"
	"time"
)

func (s *Services) KubeStatefulSetCreate(ctx context.Context, req *requests.KubeStatefulSetCreateRequest) (*appv1.StatefulSet, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	sts, err := statefulset.CreateStatefulSet(ctx, req)
	if err != nil {
		if apierrors.IsAlreadyExists(err) {
			return nil, fmt.Errorf("statefulset %q already exists in namespace %q", req.Name, req.Namespace)
		}
		if apierrors.IsNotFound(err) {
			return nil, fmt.Errorf("namespace %q not found: %w", req.Namespace, err)
		}
		// 其他错误：原样抛出（可保留一条 warn 日志，避免重复打 warn+error）
		global.Logger.Warnf("[StatefulSet] create failed %s/%s: %v", req.Namespace, req.Name, err)
		return nil, err
	}

	global.Logger.Infof("[StatefulSet] %s/%s created successfully", req.Namespace, req.Name)
	return sts, nil
}

func (s *Services) KubeStatefulSetCreateService(ctx context.Context,
	req *requests.KubeStatefulSetCreateRequest,
) (*appv1.StatefulSet, *corev1.Service, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	// 创建 StatefulSet
	sts, err := statefulset.CreateStatefulSet(ctx, req)
	if err != nil {
		return nil, nil, fmt.Errorf("create statefulset failed: %w", err)
	}
	global.Logger.Infof("statefulset %s/%s created successfully", req.Namespace, req.Name)

	// 按需创建 Service
	var svcObj *corev1.Service
	if req.IsCreateService {
		svcObj, err = statefulset.CreateServiceFromStatefulSet(ctx, req)
		if err != nil {
			// 已存在直接复用
			if apierrors.IsAlreadyExists(err) {
				exist, err := global.KubeClient.CoreV1().
					Services(req.Namespace).
					Get(ctx, req.ServiceName, metav1.GetOptions{})
				if err == nil {
					global.Logger.Infof("service %s/%s already exists, reuse it", req.Namespace, req.ServiceName)
					return sts, exist, nil
				}
			}

			// 真正失败 → 回滚删除 STS
			pol := metav1.DeletePropagationForeground
			_ = global.KubeClient.AppsV1().
				StatefulSets(req.Namespace).
				Delete(ctx, sts.Name, metav1.DeleteOptions{PropagationPolicy: &pol})
			global.Logger.Errorf("rollback delete statefulset %s/%s after service failed: %v", req.Namespace, req.Name, err)
			return nil, nil, fmt.Errorf("create headless service failed: %w", err)
		}
		global.Logger.Infof("headless service %s/%s created successfully", req.Namespace, req.ServiceName)
	}

	return sts, svcObj, nil
}

func (s *Services) KubeStatefulSetList(ctx context.Context, param *requests.KubeStatefulSetListRequest) ([]appv1.StatefulSet, int, error) {
	return statefulset.GetSatefulSetList(ctx, param.Name, param.Namespace, param.Page, param.Limit)
}

func (s *Services) KubeStatefulSetDetail(ctx context.Context, param *requests.KubeStatefulSetDetailRequest) (*appv1.StatefulSet, error) {
	return statefulset.GetStatefulSetDetail(ctx, param.Namespace, param.Name)
}

func (s *Services) KubeStatefulSetPatchReplicas(ctx context.Context, param *requests.KubeStatefulSetScaleRequest) (*appv1.StatefulSet, error) {
	return statefulset.PatchScaleReplicasStatefulSet(ctx, param.Namespace, param.Name, param.ScaleNum)
}

func (s *Services) KubeStatefulSetPatchImage(ctx context.Context, param *requests.KubeStatefulSetUpdateImageRequest) (*appv1.StatefulSet, error) {
	return statefulset.PatchImageStatefulSet(ctx, param.Namespace, param.Name, param.Container, param.Image)
}

func (s *Services) KubeStatefulSetRestart(ctx context.Context, param *requests.KubeStatefulSetRestartRequest) (*appv1.StatefulSet, error) {
	return statefulset.RestartStatefulSet(ctx, param.Namespace, param.Name)
}

func (s *Services) KubeStatefulSetDelete(ctx context.Context, param *requests.KubeStatefulSetDeleteRequest) error {
	timeout := 30 * time.Second
	return statefulset.DeleteStatefulSet(ctx, param.Namespace, param.Name, timeout)
}

func (s *Services) KubeStatefulSetDeleteService(ctx context.Context, param *requests.KubeStatefulSetDeleteRequest) error {
	return statefulset.DeleteStatefulSetService(ctx, param.Namespace, param.Name)
}
