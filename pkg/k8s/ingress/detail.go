package ingress

import (
	"context"
	"fmt"
	"go.uber.org/zap"
	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8soperation/pkg/k8s"
	"time"
)

// GetIngressDetail 获取指定命名空间下的 Ingress 详情
func GetIngressDetail(client *k8s.Client, ctx context.Context, name, namespace string) (*networkingv1.Ingress, error) {
	// 设置超时，防止长时间阻塞
	c, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// 直接通过 K8s Client 调用
	ing, err := client.Interface.NetworkingV1().
		Ingresses(namespace).
		Get(c, name, metav1.GetOptions{})

	if err != nil {
		if errors.IsNotFound(err) {
			client.Logger.Error("ingress not found",
				zap.String("namespace", namespace),
				zap.String("name", name),
			)
			return nil, fmt.Errorf("ingress %s/%s not found", namespace, name)
		}

		client.Logger.Error("get ingress failed",
			zap.String("namespace", namespace),
			zap.String("name", name),
			zap.Error(err),
		)
		return nil, err
	}

	return ing, nil
}
