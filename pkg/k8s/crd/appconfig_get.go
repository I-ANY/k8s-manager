package crd

import (
	"context"
	appv1alpha1 "gitee.com/jay-kim/appconfig-operator/api/v1alpha1"
	"k8soperation/internal/app/requests"
)

func GetAppConfig(ctx context.Context, req *requests.KubeAppConfigNameRequest) (*appv1alpha1.AppConfig, error) {
	d, err := BuildAppConfig(ctx, req.ClusterID)
	if err != nil {
		return nil, err
	}
	return d.Get(ctx, req.Namespace, req.AppName)
}
