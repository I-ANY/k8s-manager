package initialize

import (
	"fmt"

	"k8s.io/client-go/kubernetes"
	"k8soperation/internal/app/requests"
	"k8soperation/internal/app/services"
	"k8soperation/pkg/app"
)

func SetupK8sBootstrap(a *app.App) error {
	svc := services.NewServicesWithApp(a)

	cli, err := svc.K8sClusterInit(&requests.K8sClusterInitRequest{
		ID: a.AppSetting.DefaultClusterID,
	})
	if err != nil {
		return fmt.Errorf("K8sClusterInit failed: %w", err)
	}

	a.KubeConfig = cli.Config
	a.KubeClient = cli.Kube
	if cli.Metrics != nil {
		a.MetricsClient = cli.Metrics
	} else {
		a.Logger.Warn("metrics client not initialized (metrics-server not installed?)")
	}
	a.SupportsEventsV1 = cli.SupportsEvV1

	if a.SupportsEventsV1 {
		fmt.Println("当前集群支持新版事件 API：events.k8s.io/v1")
	} else {
		fmt.Println("当前集群不支持新版事件 API，自动回退至 core/v1")
	}
	return nil
}

func DetectEventAPIVersion(client *kubernetes.Clientset) bool {
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
