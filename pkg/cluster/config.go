package cluster

import (
	"context"
	"fmt"

	"k8s.io/client-go/rest"
	"k8soperation/global"
	"k8soperation/internal/app/requests"
	"k8soperation/internal/app/services" // 注意：这是允许的！因为 pkg/cluster 是偏上层
)

func GetRestConfig(ctx context.Context, clusterID uint32) (*rest.Config, error) {
	if clusterID == 0 {
		clusterID = global.AppSetting.DefaultClusterID
	}

	// 使用后台服务初始化 cluster（数据库取 kubeconfig）
	bg := services.NewServices(ctx)
	cli, err := bg.K8sClusterInit(&requests.K8sClusterInitRequest{
		ID: clusterID,
	})
	if err != nil {
		return nil, fmt.Errorf("K8sClusterInit failed: %w", err)
	}
	return cli.Config, nil
}
