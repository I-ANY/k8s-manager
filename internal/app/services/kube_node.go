package services

import (
	"context"
	"fmt"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8soperation/global"
	"k8soperation/internal/app/models"
	"k8soperation/internal/app/requests"
	"k8soperation/pkg/k8s/node"
	"time"
)

// services/node_service.go
func (s *Services) KubeNodeList(ctx context.Context, param *requests.KubeNodeListRequest) ([]corev1.Node, int, error) {
	c, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	items, total, err := node.GetNodeList(c, param.Name, param.Page, param.Limit)
	if err != nil {
		global.Logger.Errorf("list Node failed: %v", err)
		return nil, 0, err
	}
	return items, total, nil
}

// KubeNodeDetail 获取 Node 详情
func (s *Services) KubeNodeDetail(ctx context.Context, param *requests.KubeNodeDetailRequest) (*corev1.Node, error) {
	c, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	nodeObj, err := node.GetNodeDetail(c, param.Name)
	if err != nil {
		if apierrors.IsNotFound(err) {
			global.Logger.Warnf("Node %s not found", param.Name)
			return nil, fmt.Errorf("Node %q not found", param.Name)
		}
		global.Logger.Error("get Node detail failed", zap.Error(err))
		return nil, err
	}
	return nodeObj, nil
}

// KubeNodePods 列出指定 Node 上的 Pod
func (s *Services) KubeNodePods(ctx context.Context, param *requests.KubeNodePodsRequest) ([]corev1.Pod, error) {
	c, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	pods, err := node.ListPodsByNode(c, param.Name)
	if err != nil {
		global.Logger.Error("list Pods by node failed",
			zap.String("nodeName", param.Name),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to list Pods on node %q: %v", param.Name, err)
	}

	return pods, nil
}

// KubeNodeMetricsList 获取 Node 指标列表（支持单节点或全量）
func (s *Services) KubeNodeMetricsList(ctx context.Context, param *requests.KubeNodeMetricsRequest) ([]models.NodeMetricItem, int, error) {
	c, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	items, err := node.GetNodeMetrics(c, param.Name)
	if err != nil {
		global.Logger.Errorf("list Node metrics failed: %v", err)
		return nil, 0, err
	}
	return items, len(items), nil
}

// KubeNodeCordon 标记 Node 是否可调度
func (s *Services) KubeNodeCordon(ctx context.Context, param *requests.KubeNodeCordonRequest) error {
	c, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	err := node.CordonNode(c, param.NodeName, param.Unschedulable)
	if err != nil {
		if apierrors.IsNotFound(err) {
			global.Logger.Warnf("Node %s not found", param.NodeName)
			return fmt.Errorf("Node %q not found", param.NodeName)
		}
		global.Logger.Error("cordon Node failed",
			zap.String("name", param.NodeName),
			zap.Bool("unscheduled", param.Unschedulable),
			zap.Error(err),
		)
		return err
	}

	return nil
}

// KubeNodeDrain 节点 drain：cordon + 驱逐 Pod
func (s *Services) KubeNodeDrain(ctx context.Context, param *requests.KubeNodeUncordonRequest) error {
	c, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	if err := node.DrainNode(c, param.NodeName); err != nil {
		if apierrors.IsNotFound(err) {
			global.Logger.Warnf("Node %s not found when drain", param.NodeName)
			return fmt.Errorf("Node %q not found", param.NodeName)
		}
		global.Logger.Error("drain Node failed",
			zap.String("name", param.NodeName),
			zap.Error(err),
		)
		return err
	}
	return nil
}

// KubePodEvict 驱逐单个 Pod
func (s *Services) KubePodEvict(ctx context.Context, param *requests.KubePodEvictRequest) error {
	// 给整个驱逐操作一个总超时（这里还是 10 秒，可以以后也做成配置）
	c, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	// 从全局配置读取默认优雅退出时间（秒）
	graceSeconds := global.NodeSetting.Eviction.DefaultGraceSeconds

	// 如果配置成 0 或负数，可以按约定退回 “使用 Pod 自己的 terminationGracePeriodSeconds”
	// 例如约定：0 或 -1 表示不显式指定，让 K8s 用 Pod spec 的 terminationGracePeriodSeconds
	if graceSeconds <= 0 {
		graceSeconds = -1 // 具体含义看你 EvictOnePod 里怎么约定
	}

	// 这里就不再写死 30 了
	return node.EvictOnePod(c, param.Namespace, param.PodName, graceSeconds)
}
