package bootstrap

import (
	"k8soperation/initialize"
	"k8soperation/pkg/app"

	"go.uber.org/zap"
)

func InitAll(configFile string) (*app.App, error) {
	a := app.NewApp()

	if err := initialize.SetupSetting(a, configFile); err != nil {
		return nil, err
	}
	if err := initialize.SetupValidator(a); err != nil {
		return nil, err
	}
	if err := initialize.SetupLogger(a); err != nil {
		return nil, err
	}

	// DB init failure is non-fatal (logs error but continues)
	if err := initialize.SetupDB(a); err != nil {
		a.Logger.Error("init db failed", zap.Error(err))
	}

	if err := initialize.SetupSession(a); err != nil {
		return nil, err
	}
	if err := initialize.SetupK8sBootstrap(a); err != nil {
		return nil, err
	}
	if err := initialize.SetupAppConfigClient(a); err != nil {
		return nil, err
	}

	initialize.LogDocsReady(a)

	return a, nil
}

func FlushLoggers(a *app.App) {
	if a.Logger != nil {
		_ = a.Logger.Sync()
	}
	if a.BizLogger != nil {
		_ = a.BizLogger.Sync()
	}
}
