package svc

import (
	"context"
	"fmt"
	"k8soperation/pkg/k8s"
	"time"

	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// GetServiceDetail 获取指定命名空间下的 Service 详情
func GetServiceDetail(client *k8s.Client, ctx context.Context, name, namespace string) (*corev1.Service, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	svc, err := client.Interface.CoreV1().
		Services(namespace).
		Get(ctx, name, metav1.GetOptions{})

	if err != nil {
		if apierrors.IsNotFound(err) {
			client.Logger.Error("service not found",
				zap.String("namespace", namespace),
				zap.String("name", name),
			)
			return nil, fmt.Errorf("service %s/%s not found", namespace, name)
		}

		client.Logger.Error("get service failed",
			zap.String("namespace", namespace),
			zap.String("name", name),
			zap.Error(err),
		)
		return nil, err
	}

	return svc, nil
}
