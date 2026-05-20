package svc

import (
	"context"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8soperation/pkg/k8s"
	"time"
)

// 通用 Patch：传入 patch bytes，返回最新的 Service
func PatchService(client *k8s.Client, ctx context.Context, namespace, name string, patch []byte) (*corev1.Service, error) {
	c, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	svc, err := client.Interface.CoreV1().
		Services(namespace).
		Patch(c, name, types.StrategicMergePatchType, patch, metav1.PatchOptions{})
	if err != nil {
		return nil, err
	}
	return svc, nil
}

func PatchJsonService(client *k8s.Client, ctx context.Context, namespace, name string, patch []byte) (*corev1.Service, error) {
	c, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	svc, err := client.Interface.CoreV1().
		Services(namespace).
		Patch(c, name, types.MergePatchType, patch, metav1.PatchOptions{})
	if err != nil {
		return nil, err
	}
	return svc, nil
}
