package cluster

import (
	"context"
	"fmt"

	"k8s.io/client-go/rest"
	"k8soperation/internal/app/requests"
	"k8soperation/internal/app/services"
	"k8soperation/pkg/app"
)

func GetRestConfig(ctx context.Context, a *app.App, clusterID uint32) (*rest.Config, error) {
	if clusterID == 0 {
		clusterID = a.AppSetting.DefaultClusterID
	}

	bg := services.NewServicesWithApp(a)
	cli, err := bg.K8sClusterInit(&requests.K8sClusterInitRequest{
		ID: clusterID,
	})
	if err != nil {
		return nil, fmt.Errorf("K8sClusterInit failed: %w", err)
	}
	return cli.Config, nil
}
