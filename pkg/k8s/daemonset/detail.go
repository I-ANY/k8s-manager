package daemonset

import (
	"context"
	"fmt"
	"go.uber.org/zap"
	appv1 "k8s.io/api/apps/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8soperation/pkg/k8s"
)

func GetDaemonSetDetail(client *k8s.Client, ctx context.Context, name, namespace string) (*appv1.DaemonSet, error) {
	ds, err := client.Interface.AppsV1().
		DaemonSets(namespace).
		Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			client.Logger.Error("daemonset not found",
				zap.String("namespace", namespace),
				zap.String("name", name),
			)
			return nil, fmt.Errorf("daemonset %s/%s not found", namespace, name)
		}

		client.Logger.Error("get daemonset failed",
			zap.String("namespace", namespace),
			zap.String("name", name),
			zap.Error(err),
		)
		return nil, err
	}

	return ds, nil
}
