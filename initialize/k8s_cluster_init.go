package initialize

import (
	"fmt"
	"k8s.io/client-go/kubernetes"
	"k8soperation/global"
	"k8soperation/internal/app/requests"
	"k8soperation/internal/app/services"
)

func SetupK8sBootstrap() error {
	// 1) 创建后台服务实例
	svc := services.NewBackgroundServices()

	// 2) 初始化 K8s 集群（返回聚合客户端）
	cli, err := svc.K8sClusterInit(&requests.K8sClusterInitRequest{
		ID: global.AppSetting.DefaultClusterID,
	})
	if err != nil {
		return fmt.Errorf("K8sClusterInit failed: %w", err)
	}

	// 3) 落到全局（同一份 cfg 复用最稳）
	global.KubeConfig = cli.Config
	global.KubeClient = cli.Kube
	if cli.Metrics != nil {
		global.MetricsClient = cli.Metrics
	} else {
		global.Logger.Warn("metrics client not initialized (metrics-server not installed?)")
	}
	global.SupportsEventsV1 = cli.SupportsEvV1

	// 4) 友好输出
	if global.SupportsEventsV1 {
		fmt.Println("当前集群支持新版事件 API：events.k8s.io/v1")
	} else {
		fmt.Println("当前集群不支持新版事件 API，自动回退至 core/v1")
	}
	return nil
}

func DetectEventAPIVersion(client *kubernetes.Clientset) bool {
	// 检查Kubernetes集群是否支持events.k8s.io API组
	// 通过查询服务器的API组信息来判断
	// 返回true表示支持，false表示不支持
	// 获取Kubernetes服务器的API组信息
	// global.KubeClient.Discovery() 提供集群发现能力
	// ServerGroups() 方法返回服务器支持的API组列表
	groups, err := client.Discovery().ServerGroups()
	if err != nil {
		return false
	}

	for _, g := range groups.Groups {
		if g.Name == "events.k8s.io" {
			return true
		}
	}
	return false
}
