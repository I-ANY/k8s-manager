package global

import (
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	metricsclient "k8s.io/metrics/pkg/client/clientset/versioned"
)

var (
	KubeClient       *kubernetes.Clientset // k8s 客户端
	SupportsEventsV1 bool                  // 是否支持新版 events.k8s.io/v1

	// K8s metrics.k8s.io 客户端（操作 NodeMetrics / PodMetrics）
	MetricsClient *metricsclient.Clientset

	// 当前连接的集群配置（可供重用）
	KubeConfig *rest.Config
)
