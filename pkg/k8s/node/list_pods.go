package node

import (
	"context"
	"fmt"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8soperation/global"
	k8sclient "k8soperation/pkg/k8s"
)

// ListPodsByNode 列出指定 Node 上运行的所有 Pod
func ListPodsByNode(client *k8sclient.Client, ctx context.Context, nodeName string) ([]corev1.Pod, error) {
	// Pod 是 namespace 级资源，这里需要跨所有命名空间查
	podList, err := global.KubeClient.CoreV1().
		Pods(""). //空字符串代表“所有命名空间”
		List(ctx, metav1.ListOptions{
			FieldSelector: fmt.Sprintf("spec.nodeName=%s", nodeName),
		})
	if err != nil {
		client.Log().Error("list Pods by node failed",
			zap.String("nodeName", nodeName),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to list Pods on node %q: %v", nodeName, err)
	}

	return podList.Items, nil
}
