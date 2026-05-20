package namespace

import (
	"context"
	"fmt"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8soperation/global"
	k8sclient "k8soperation/pkg/k8s"
)

func GetNamespaceDetail(client *k8sclient.Client, ctx context.Context, name string) (*corev1.Namespace, error) {
	ns, err := global.KubeClient.CoreV1().
		Namespaces().
		Get(ctx, name, metav1.GetOptions{})

	if err != nil {
		if apierrors.IsNotFound(err) {
			client.Log().Error("Namespace not found", zap.String("name", name))
			return nil, fmt.Errorf("Namespace %q not found", name)
		}

		client.Log().Error("get Namespace failed",
			zap.String("name", name),
			zap.Error(err),
		)
		return nil, err
	}

	return ns, nil
}
