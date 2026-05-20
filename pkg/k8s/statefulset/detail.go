package statefulset

import (
	"context"
	"fmt"
	"go.uber.org/zap"
	appv1 "k8s.io/api/apps/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8soperation/pkg/k8s"
	"time"
)

func GetStatefulSetDetail(client *k8s.Client, ctx context.Context, namespace, name string) (*appv1.StatefulSet, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	sts, err := client.Interface.AppsV1().
		StatefulSets(namespace).
		Get(ctx, name, metav1.GetOptions{})

	if err != nil {
		// 仅当有错误时再细分类型
		if apierrors.IsNotFound(err) { // 别名：apierrors 来自 k8s.io/apimachinery/pkg/api/errors
			client.Logger.Error("deployment not found",
				zap.String("namespace", namespace),
				zap.String("name", name),
			)
			return nil, fmt.Errorf("deployment %s/%s not found", namespace, name)
		}

		// 其它错误，直接返回并记录
		client.Logger.Error("get deployment failed",
			zap.String("namespace", namespace),
			zap.String("name", name),
			zap.Error(err),
		)
		return nil, err
	}

	// 正常返回
	return sts, nil
}
