package node

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

// GetNodeDetail 获取 Node 详情（集群级资源）
func GetNodeDetail(client *k8sclient.Client, ctx context.Context, name string) (*corev1.Node, error) {
	nodeObj, err := global.KubeClient.CoreV1().
		Nodes().
		Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			client.Log().Error("Node not found", zap.String("name", name))
			return nil, fmt.Errorf("Node %q not found", name)
		}
		client.Log().Error("get Node failed", zap.String("name", name), zap.Error(err))
		return nil, err
	}
	return nodeObj, nil
}
