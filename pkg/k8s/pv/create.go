package pv

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

func CreatePersistentVolume(client *k8sclient.Client, ctx context.Context, req *requests.KubePVCreateRequest) (*corev1.PersistentVolume, error) {
	// 1) 构造 PV 对象
	pv := BuildPersistentVolumeFromReq(req)

	// 2) 调用 Kubernetes API 创建
	created, err := global.KubeClient.CoreV1().
		PersistentVolumes().
		Create(ctx, pv, metav1.CreateOptions{})
	if err != nil {
		if apierrors.IsAlreadyExists(err) {
			return nil, fmt.Errorf("PersistentVolume %q already exists", pv.Name)
		}
		client.Log().Errorf("create PersistentVolume failed: %v", err)
		return nil, err
	}

	client.Log().Infof("PersistentVolume %q created successfully", created.Name)
	return created, nil
}
