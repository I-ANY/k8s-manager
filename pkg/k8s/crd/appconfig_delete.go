package crd

import (
	"context"
	"k8soperation/internal/app/requests"
)

func DeleteAppConfig(ctx context.Context, req *requests.KubeAppConfigNameRequest) error {
	d, err := BuildAppConfig(ctx, req.ClusterID)
	if err != nil {
		return err
	}
	return d.Delete(ctx, req.Namespace, req.AppName)
}
