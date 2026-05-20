package initialize

import (
	"fmt"
	"os"
	"path/filepath"

	"k8soperation/pkg/app"
	logger2 "k8soperation/pkg/logger"
)

func ensureDir(filePath string) error {
	dir := filepath.Dir(filePath)
	info, err := os.Stat(dir)
	if err != nil {
		if os.IsNotExist(err) {
			if err := os.MkdirAll(dir, 0o755); err != nil {
				return fmt.Errorf("create log dir %q: %w", dir, err)
			}
			return nil
		}
		return fmt.Errorf("stat log dir %q: %w", dir, err)
	}
	if !info.IsDir() {
		return fmt.Errorf("log path %q exists but is not a directory", dir)
	}
	return nil
}

func SetupLogger(a *app.App) error {
	if a.AppSetting == nil {
		return fmt.Errorf("AppSetting is nil")
	}

	if err := ensureDir(a.AppSetting.LogFileName); err != nil {
		return err
	}
	if a.AppSetting.BusinessLogFileName == "" {
		a.AppSetting.BusinessLogFileName = "logs/app.log"
	}
	if err := ensureDir(a.AppSetting.BusinessLogFileName); err != nil {
		return err
	}

	lvl := logger2.WithLevel(a.AppSetting.LogLevel)
	sysOpts := []logger2.Option{
		logger2.AddCaller(),
		logger2.AddCallerSkip(1),
		logger2.AddStacktrace(logger2.ErrorLevel),
	}
	bizOpts := []logger2.Option{
		logger2.AddCaller(),
		logger2.AddCallerSkip(1),
	}

	a.Logger = logger2.NewLogger(
		lvl,
		logger2.RotateOptions{
			FileName:   a.AppSetting.LogFileName,
			MaxSize:    a.AppSetting.LogMaxSize,
			MaxBackups: a.AppSetting.LogMaxBackup,
			MaxAge:     a.AppSetting.LogMaxAge,
			Compress:   a.AppSetting.LogCompress,
		},
		sysOpts...,
	)

	a.BizLogger = logger2.NewBusinessLogger(
		logger2.RotateOptions{
			FileName:   a.AppSetting.BusinessLogFileName,
			MaxSize:    a.AppSetting.LogMaxSize,
			MaxBackups: a.AppSetting.LogMaxBackup,
			MaxAge:     a.AppSetting.LogMaxAge,
			Compress:   a.AppSetting.LogCompress,
		},
		bizOpts...,
	)

	return nil
}
