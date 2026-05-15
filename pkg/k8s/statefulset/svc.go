package statefulset

import (
	"context"
	"fmt"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8soperation/global"
	"k8soperation/internal/app/requests"
	"time"
)

func CreateServiceFromStatefulSet(
	ctx context.Context,
	req *requests.KubeStatefulSetCreateRequest,
) (*corev1.Service, error) {
	// 子上下文，避免无限阻塞
	cctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	// 1) 构造 Headless Service
	svc := BuildServiceFromStatefulSetReq(req) // name 里已做了回退

	// 2) 创建
	createdSvc, err := global.KubeClient.CoreV1().
		Services(req.Namespace).
		Create(cctx, svc, metav1.CreateOptions{FieldManager: "k8soperation"})
	if err != nil {
		if apierrors.IsForbidden(err) {
			return nil, fmt.Errorf("forbidden to create service %s/%s: %w", req.Namespace, svc.Name, err)
		}
		if apierrors.IsAlreadyExists(err) {
			return nil, fmt.Errorf("service %s/%s already exists: %w", req.Namespace, svc.Name, err)
		}
		global.Logger.Errorf("create headless service %s/%s failed: %v", req.Namespace, svc.Name, err)
		return nil, fmt.Errorf("failed to create service %s/%s: %w", req.Namespace, svc.Name, err)
	}

	global.Logger.Infof("headless service %s/%s created successfully", createdSvc.Namespace, createdSvc.Name)
	return createdSvc, nil
}
