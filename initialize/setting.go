package initialize

import (
	"k8soperation/global"
	"k8soperation/internal/errorcode"
	"k8soperation/pkg/setting"
)

// SetupSetting 初始化全局配置
// 1. 创建 viper 实例（读取配置文件）
// 2. 将配置文件中 "Server" 部分映射到 global.Setting
func SetupSetting() error {
	// 1. 创建一个 Setting 对象（内部封装了 viper 配置读取）
	s, err := setting.NewSetting()
	if err != nil {
		// 如果读取配置文件失败，直接返回错误
		return err
	}

	// 2. 从配置文件中读取 "Server" 这一节的数据
	//    并将解析结果反序列化到 global.Setting 结构体指针中
	//    - 第一个参数 "Server" 是配置文件中的 key
	//    - 第二个参数 global.ServerSetting 是接收配置的结构体指针
	if err = s.ReadSection("Server", &global.ServerSetting); err != nil {
		return err
	}

	// 加载应用配置
	if err := s.ReadSection("App", &global.AppSetting); err != nil {
		return err
	}

	// 读取 Database 配置
	if err = s.ReadSection("Database", &global.DatabaseSetting); err != nil {
		return err
	}

	// 读取 Redis 配置
	if err = s.ReadSection("Cache", &global.CacheSetting); err != nil {
		return err
	}

	// 读取 kube_pod 配置
	if err = s.ReadSection("Pod", &global.PodLogSetting); err != nil {
		return err
	}

	// 读取 kube_node 配置
	if err = s.ReadSection("Node", &global.NodeSetting); err != nil {
		return err
	}

	// ★ 新增：读取错误码配置
	if err = s.ReadSection("ErrorCode", &global.ErrorCodeSetting); err != nil {
		return err
	}

	// ★ 在这里把开关注入 errorcode 包
	errorcode.SetAllowOverride(global.ErrorCodeSetting.AllowOverride)

	// ★ 在这里统一注册所有错误码
	errorcode.Register()

	// 成功则返回 nil，表示初始化完成
	return nil
}
