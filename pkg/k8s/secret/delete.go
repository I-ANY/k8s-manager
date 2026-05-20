package secret

import (
	"context"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8soperation/pkg/k8s"
	"time"
)

func DeleteSecret(client *k8s.Client, ctx context.Context, name, namespace string) error {
	// 统一超时/取消，避免阻塞
	c, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// 删除策略（前台级联，语义明确）
	fg := metav1.DeletePropagationForeground
	opts := metav1.DeleteOptions{PropagationPolicy: &fg}

	// 发起删除
	if err := client.Interface.CoreV1().
		Secrets(namespace).
		Delete(c, name, opts); err != nil {
		if apierrors.IsNotFound(err) {
			return nil // 幂等处理：已不存在视为成功
		}
		return err
	}

	// 轮询确认 Secret 已被真正删除
	return wait.PollUntilContextTimeout(
		c,
		2*time.Second,  // interval
		30*time.Second, // timeout（通常与 context 超时一致）
		true,           // immediate：立即检查一次
		func(ctx context.Context) (done bool, err error) {
			_, err = client.Interface.CoreV1().
				Secrets(namespace).
				Get(ctx, name, metav1.GetOptions{})
			if apierrors.IsNotFound(err) {
				return true, nil // 确认已删除
			}
			return false, err
		},
	)
}
