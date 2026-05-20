package statefulset

import (
	"context"
	"fmt"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8soperation/pkg/k8s"
	"time"
)

//func DeleteStatefulSet(ctx context.Context, name, namespace string) error {
//
//	// 前台级联：先删 Pods/ControllerRevisions 再删 StatefulSet
//	fg := metav1.DeletePropagationForeground
//	opts := metav1.DeleteOptions{PropagationPolicy: &fg}
//
//	// 发起删除
//	if err := client.Interface.AppsV1().
//		StatefulSets(namespace).
//		Delete(c, name, opts); err != nil {
//		if apierrors.IsNotFound(err) {
//			return nil // 幂等：已不存在视为成功
//		}
//		return err
//	}
//
//	// 轮询直到 StatefulSet 真正消失
//	err := wait.PollUntilContextTimeout(
//		c,
//		2*time.Second,   // interval
//		300*time.Second, // timeout（与上面 ctx 共同生效）
//		true,            // immediate
//		func(ctx context.Context) (done bool, err error) {
//			_, err = client.Interface.AppsV1().
//				StatefulSets(namespace).
//				Get(ctx, name, metav1.GetOptions{})
//			if apierrors.IsNotFound(err) {
//				return true, nil
//			}
//			return false, err
//		},
//	)
//
//	return err
//}

func DeleteStatefulSet(client *k8s.Client, ctx context.Context, name, ns string, timeout time.Duration) error {
	fg := metav1.DeletePropagationForeground
	if err := client.Interface.AppsV1().StatefulSets(ns).Delete(ctx, name, metav1.DeleteOptions{PropagationPolicy: &fg}); err != nil {
		if apierrors.IsNotFound(err) {
			return nil
		}
		if apierrors.IsForbidden(err) {
			return fmt.Errorf("没有权限删除 StatefulSet %s/%s", ns, name)
		}
		return err
	}
	if timeout <= 0 {
		timeout = 10 * time.Minute
	}
	c, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	return wait.PollUntilContextTimeout(c, 2*time.Second, timeout, true, func(ctx context.Context) (bool, error) {
		_, err := client.Interface.AppsV1().StatefulSets(ns).Get(ctx, name, metav1.GetOptions{})
		if apierrors.IsNotFound(err) {
			return true, nil
		}
		if err != nil && (apierrors.IsTimeout(err) || apierrors.IsServerTimeout(err) || apierrors.IsTooManyRequests(err)) {
			return false, nil
		}
		return false, err
	})
}
