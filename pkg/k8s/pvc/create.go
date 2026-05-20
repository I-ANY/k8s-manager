package pvc

import (
	"context"
	"fmt"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8soperation/global"
	"k8soperation/internal/app/requests"
	k8sclient "k8soperation/pkg/k8s"
)

// CreatePersistentVolumeClaim 创建 PVC（命名空间级）
func CreatePersistentVolumeClaim(client *k8sclient.Client, ctx context.Context, req *requests.KubePVCCreateRequest) (*corev1.PersistentVolumeClaim, error) {
	// 1) 构造 PVC 对象
	pvc := BuildPVCFromReq(req)

	// 2) 调用 Kubernetes API 创建
	created, err := global.KubeClient.CoreV1().
		PersistentVolumeClaims(req.Namespace).
		Create(ctx, pvc, metav1.CreateOptions{})
	if err != nil {
		if apierrors.IsAlreadyExists(err) {
			return nil, fmt.Errorf("PersistentVolumeClaim %q already exists in namespace %q", pvc.Name, pvc.Namespace)
		}
		client.Log().Errorf("create PersistentVolumeClaim failed: %v", err)
		return nil, err
	}

	client.Log().Infof("PersistentVolumeClaim %q created successfully in namespace %q", created.Name, created.Namespace)
	return created, nil
}
