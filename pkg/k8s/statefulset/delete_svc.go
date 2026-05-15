package statefulset

import (
	"context"
	"fmt"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8soperation/global"
	"time"
)

func DeleteStatefulSetService(ctx context.Context, name, namespace string) error {
	c, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	err := global.KubeClient.CoreV1().
		Services(namespace).
		Delete(c, name, metav1.DeleteOptions{})

	if err != nil {
		switch {
		case apierrors.IsNotFound(err):
			// 幂等：已不存在视为成功
			return nil
		case apierrors.IsForbidden(err):
			return fmt.Errorf("当前 ServiceAccount 没有权限删除 Service %s/%s: %w", namespace, name, err)
		default:
			// ❗ 这里要返回，否则错误被吞掉
			return fmt.Errorf("删除 Service %s/%s 失败: %w", namespace, name, err)
		}
	}

	// 等待删除完成（可选）
	waitErr := wait.PollUntilContextTimeout(
		c,
		1*time.Second,
		10*time.Second,
		true,
		func(ctx context.Context) (bool, error) {
			_, err := global.KubeClient.CoreV1().
				Services(namespace).
				Get(ctx, name, metav1.GetOptions{})
			if apierrors.IsNotFound(err) {
				return true, nil
			}
			return false, err
		},
	)

	return waitErr
}
