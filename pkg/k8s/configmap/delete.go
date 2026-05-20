package configmap

import (
	"context"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8soperation/pkg/k8s"
	"time"
)

func DeleteConfigMap(client *k8s.Client, ctx context.Context, name, namespace string) error {
	// 统一超时/取消
	c, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// 删除（ConfigMap 通常不需要级联，前台/后台都可，这里保留与 Secret 一致写法）
	fg := metav1.DeletePropagationForeground
	opts := metav1.DeleteOptions{PropagationPolicy: &fg}

	if err := client.Interface.CoreV1().
		ConfigMaps(namespace).
		Delete(c, name, opts); err != nil {
		if apierrors.IsNotFound(err) {
			return nil // 幂等：已不存在视为成功
		}
		return err
	}

	// 轮询确认已删除
	return wait.PollUntilContextTimeout(
		c,
		2*time.Second,  // interval
		30*time.Second, // timeout
		true,           // immediate
		func(ctx context.Context) (bool, error) {
			_, err := client.Interface.CoreV1().
				ConfigMaps(namespace).
				Get(ctx, name, metav1.GetOptions{})
			if apierrors.IsNotFound(err) {
				return true, nil
			}
			return false, err
		},
	)
}
