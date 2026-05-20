package node

import (
	"context"
	"fmt"
	"go.uber.org/zap"
	policyv1 "k8s.io/api/policy/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8soperation/global"
	k8sclient "k8soperation/pkg/k8s"
)

// EvictOnePod 驱逐单个 Pod（可被 /drain 和 /pod/evict 复用）
func EvictOnePod(client *k8sclient.Client, ctx context.Context, namespace, podName string, graceSeconds int64) error {

	var deleteOptions *metav1.DeleteOptions

	// graceSeconds == -1 表示：使用 Pod 自己的 terminationGracePeriodSeconds
	if graceSeconds >= 0 {
		g := graceSeconds
		deleteOptions = &metav1.DeleteOptions{
			GracePeriodSeconds: &g,
		}
	}

	eviction := &policyv1.Eviction{
		ObjectMeta: metav1.ObjectMeta{
			Name:      podName,
			Namespace: namespace,
		},
		DeleteOptions: deleteOptions,
	}

	err := global.KubeClient.PolicyV1().
		Evictions(namespace).
		Evict(ctx, eviction)

	if err != nil {

		// Pod 不存在了：可认为驱逐成功，向上抛可控错误
		if apierrors.IsNotFound(err) {
			client.Log().Warn("pod already gone when evict",
				zap.String("pod", podName),
				zap.String("ns", namespace),
			)
			return fmt.Errorf("pod %s/%s not found", namespace, podName)
		}

		// PDB 限制，客户端可选择重试
		if apierrors.IsTooManyRequests(err) {
			client.Log().Warn("evict pod blocked by PDB",
				zap.String("pod", podName),
				zap.String("ns", namespace),
				zap.Error(err),
			)
			return fmt.Errorf("evict pod %s/%s blocked by PDB: %w", namespace, podName, err)
		}

		// 其他错误
		client.Log().Error("evict pod failed",
			zap.String("pod", podName),
			zap.String("ns", namespace),
			zap.Error(err),
		)
		return fmt.Errorf("evict pod %s/%s failed: %w", namespace, podName, err)
	}

	return nil
}
