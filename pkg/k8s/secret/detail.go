package secret

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

func GetSecretDetail(client *k8s.Client, ctx context.Context, name, namespace string) (*corev1.Secret, error) {
	// 设置超时，防止长时间阻塞
	c, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// 通过 K8s Client 获取 Secret
	sec, err := client.Interface.CoreV1().
		Secrets(namespace).
		Get(c, name, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			client.Logger.Error("secret not found",
				zap.String("namespace", namespace),
				zap.String("name", name),
			)
			return nil, fmt.Errorf("secret %s/%s not found", namespace, name)
		}
		client.Logger.Error("get secret failed",
			zap.String("namespace", namespace),
			zap.String("name", name),
			zap.Error(err),
		)
		return nil, err
	}

	return sec, nil
}
