package svc

import (
	"context"
	"fmt"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8soperation/pkg/k8s"
	"time"
)

func GetServiceEndpoints(client *k8s.Client, ctx context.Context, name, namespace string) (*corev1.Endpoints, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	ep, err := client.Interface.CoreV1().
		Endpoints(namespace).
		Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			client.Logger.Error("endpoints not found",
				zap.String("namespace", namespace),
				zap.String("name", name),
			)
			return nil, fmt.Errorf("endpoints %s/%s not found", namespace, name)
		}
		client.Logger.Error("get endpoints failed",
			zap.String("namespace", namespace),
			zap.String("name", name),
			zap.Error(err),
		)
		return nil, err
	}
	return ep, nil
}
