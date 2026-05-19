package bootstrap

import (
	"k8soperation/global"
	"k8soperation/initialize"

	"go.uber.org/zap"
)

func InitAll(configFile ...string) error {
	// 初始化配置
	if err := initialize.SetupSetting(configFile...); err != nil {
		return err
	}
	// 初始校验规则
	if err := initialize.SetupValidator(); err != nil {
		return err
	}

	// 初始化日志
	if err := initialize.SetupLogger(); err != nil {
		return err
	}

	// 初始化数据库
	if err := initialize.SetupDB(); err != nil {
		global.Logger.Error("init db failed", zap.Error(err))
	}

	// 初始化Redis
	if err := initialize.SetupSession(); err != nil {
		return err
	}

	// 初始化K8s
	if err := initialize.SetupK8sBootstrap(); err != nil {
		return err
	}

	// 初始化 AppConfig CRD 客户端
	if err := initialize.SetupAppConfigClient(); err != nil {
		return err
	}

	// 加载 swagger 接口文档
	initialize.LogDocsReady()

	return nil
}

// Sync() 会做两件事：
// 调用底层 WriteSyncer 的 Sync()（例如 os.File.Sync()）；
// 把缓冲日志强制写到文件。
func FlushLoggers() {
	// 系统日志落盘
	_ = global.Logger.Sync()
	if global.BizLogger != nil {
		// 业务日志落盘
		_ = global.BizLogger.Sync()
	}
}
