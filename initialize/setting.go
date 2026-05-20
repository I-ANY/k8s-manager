package initialize

import (
	"k8soperation/internal/errorcode"
	"k8soperation/pkg/app"
	"k8soperation/pkg/setting"
	"k8soperation/pkg/setting/types"
)

func SetupSetting(a *app.App, configFile string) error {
	setting.SetUpAppConfig(configFile)
	ApplySetting(a, setting.AppConfig)
	errorcode.Register()
	return nil
}

func ApplySetting(a *app.App, config *types.AppConfig) {
	a.AppConfig = config
	if a.AppConfig != nil {
		a.ServerSetting = a.AppConfig.ServerSetting
		a.AppSetting = a.AppConfig.AppSetting
		a.DatabaseSetting = a.AppConfig.DatabaseSetting
		a.CacheSetting = a.AppConfig.CacheSetting
		a.PodLogSetting = a.AppConfig.PodLogSetting
		a.NodeSetting = a.AppConfig.NodeSetting
	}
}
