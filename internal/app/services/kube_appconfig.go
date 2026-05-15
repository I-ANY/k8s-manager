package services

import (
	"context"
	"k8soperation/pkg/k8s/crd"

	appv1alpha1 "gitee.com/jay-kim/appconfig-operator/api/v1alpha1"
	"k8soperation/global"
	"k8soperation/internal/app/requests"
)

// 创建 AppConfig
func (s *Services) KubeAppConfigCreate(ctx context.Context, req *requests.KubeAppConfigCreateRequest) (*appv1alpha1.AppConfig, error) {
	app, err := crd.CreateAppConfig(ctx, req)
	if err != nil {
		global.Logger.Errorf("CreateAppConfig error: %v", err)
		return nil, err
	}
	global.Logger.Infof("CreateAppConfig success, ns=%s, name=%s", req.Namespace, req.AppName)
	return app, nil
}

// 更新 AppConfig
func (s *Services) KubeAppConfigUpdate(ctx context.Context, req *requests.KubeAppConfigUpdateRequest) (*appv1alpha1.AppConfig, error) {
	app, err := crd.UpdateAppConfig(ctx, req)
	if err != nil {
		global.Logger.Errorf("UpdateAppConfig error: %v", err)
		return nil, err
	}
	global.Logger.Infof("UpdateAppConfig success, ns=%s, name=%s", req.Namespace, req.AppName)
	return app, nil
}

// 获取单个 AppConfig
func (s *Services) KubeAppConfigGet(ctx context.Context, req *requests.KubeAppConfigNameRequest) (*appv1alpha1.AppConfig, error) {
	app, err := crd.GetAppConfig(ctx, req)
	if err != nil {
		global.Logger.Errorf("GetAppConfig error: %v", err)
		return nil, err
	}
	global.Logger.Infof("GetAppConfig success, ns=%s, name=%s", req.Namespace, req.AppName)
	return app, nil
}

// 删除 AppConfig
func (s *Services) KubeAppConfigDelete(ctx context.Context, req *requests.KubeAppConfigNameRequest) error {
	if err := crd.DeleteAppConfig(ctx, req); err != nil {
		global.Logger.Errorf("DeleteAppConfig error: %v", err)
		return err
	}
	global.Logger.Infof("DeleteAppConfig success, ns=%s, name=%s", req.Namespace, req.AppName)
	return nil
}

// 列表 AppConfig
func (s *Services) KubeAppConfigList(ctx context.Context, req *requests.KubeAppConfigListRequest) ([]appv1alpha1.AppConfig, error) {
	items, err := crd.List(ctx, req)
	if err != nil {
		global.Logger.Errorf("ListAppConfig error: %v", err)
		return nil, err
	}
	global.Logger.Infof("ListAppConfig success, ns=%s", req.Namespace)
	return items, nil
}
