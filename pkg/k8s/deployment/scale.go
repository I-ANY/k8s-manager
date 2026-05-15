package deployment

import (
	"context"
	"fmt"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/util/retry"
	"k8soperation/global"
	"time"
)

func ScaleDeployment(namespace, name string, replicas int32) error {
	if replicas < 0 {
		return fmt.Errorf("replicas must be greater than 0")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return retry.RetryOnConflict(retry.DefaultRetry, func() error {
		scale, getErr := global.KubeClient.AppsV1().
			Deployments(namespace).
			GetScale(ctx, name, metav1.GetOptions{})
		if getErr != nil {
			if apierrors.IsNotFound(getErr) {
				return fmt.Errorf("deployment %s/%s not found", namespace, name)
			}
			if apierrors.IsForbidden(getErr) {
				return fmt.Errorf("no permission to get scale of deployment %s/%s", namespace, name)
			}
			return getErr
		}

		scale.Spec.Replicas = replicas

		_, updateErr := global.KubeClient.AppsV1().
			Deployments(namespace).
			UpdateScale(ctx, name, scale, metav1.UpdateOptions{})
		if updateErr != nil {
			if apierrors.IsForbidden(updateErr) {
				return fmt.Errorf("no permission to update scale of deployment %s/%s", namespace, name)
			}
			return updateErr
		}

		// 成功就返回 nil
		return nil
	})
}
