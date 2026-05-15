package crd

import (
	"context"
	appv1alpha1 "gitee.com/jay-kim/appconfig-operator/api/v1alpha1"
	"k8soperation/internal/app/requests"
	"k8soperation/pkg/cluster"
)

func List(ctx context.Context, req *requests.KubeAppConfigListRequest) ([]appv1alpha1.AppConfig, error) {
	cfg, err := cluster.GetRestConfig(ctx, req.ClusterID)
	if err != nil {
		return nil, err
	}

	cli, err := cluster.NewAppConfigClient(cfg)
	if err != nil {
		return nil, err
	}

	var list appv1alpha1.AppConfigList
	if err := cli.List(ctx, &list); err != nil {
		return nil, err
	}
	return list.Items, nil
}
