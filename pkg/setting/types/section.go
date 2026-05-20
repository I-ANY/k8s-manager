package types

import (
	"time"
)

type AppConfig struct {
	ServerSetting   *ServerSettingS
	AppSetting      *AppSettingS
	DatabaseSetting *DatabaseSettingS
	CacheSetting    *CacheSettingS
	PodLogSetting   *PodLogSetting
	NodeSetting     *NodeConfig
}

type ServerSettingS struct {
	RunMode         string        `mapstructure:"runMode"`
	Port            string        `mapstructure:"port"`
	ReadTimeout     time.Duration `mapstructure:"readTimeout"`
	WriteTimeout    time.Duration `mapstructure:"writeTimeout"`
	IdleTimeout     time.Duration `mapstructure:"idleTimeout"`
	ShutdownTimeout time.Duration `mapstructure:"shutdownTimeout"`
}

// DatabaseSettingS 定义数据库配置结构体，用于存储数据库连接相关的各项参数
type DatabaseSettingS struct {
	DBType         string        `mapstructure:"dbType"`         // 数据库类型，如：mysql、postgres等
	Username       string        `mapstructure:"username"`       // 数据库用户名
	Password       string        `mapstructure:"password"`       // 数据库密码
	Host           string        `mapstructure:"host"`           // 数据库主机地址
	Port           string        `mapstructure:"port"`           // 数据库端口号
	DBName         string        `mapstructure:"dbname"`         // 数据库名称
	Charset        string        `mapstructure:"charset"`        // 数据库字符集，如：utf8mb4
	ParseTime      bool          `mapstructure:"parseTime"`      // 是否解析时间，true表示自动解析时间类型
	MaxIdleConns   int           `mapstructure:"maxIdleConns"`   // 数据库连接池最大空闲连接数
	MaxOpenConns   int           `mapstructure:"maxOpenConns"`   // MaxOpenConns 表示数据库连接池中最大打开的连接数
	MaxLifeSeconds time.Duration `mapstructure:"maxLifeSeconds"` // 数据库连接池中连接的最大生命周期，单位为秒
}

type AppSettingS struct {
	LogLevel               string `mapstructure:"logLevel"`               // 日志级别
	LogType                string `mapstructure:"logType"`                // 日志输出类型
	LogFileName            string `mapstructure:"logFileName"`            // 日志文件名称
	LogMaxSize             int    `mapstructure:"logMaxSize"`             // 日志文件最大大小（MB）
	LogMaxBackup           int    `mapstructure:"logMaxBackup"`           // 日志备份数量
	LogMaxAge              int    `mapstructure:"logMaxAge"`              // 日志保留天数
	LogCompress            bool   `mapstructure:"logCompress"`            // 是否压缩日志文件
	BusinessLogFileName    string `mapstructure:"businessLogFileName"`    // 业务日志文件名
	MirrorBusinessToSystem bool   `mapstructure:"mirrorBusinessToSystem"` // 是否镜像业务日志到系统日志
	JWTExpireTime          int    `mapstructure:"jwtExpireTime"`          // JWT过期时间（秒）
	JWTSigningKey          string `mapstructure:"jwtSigningKey"`          // JWT签名密钥
	JWTMaxRefreshTime      int    `mapstructure:"jwtMaxRefreshTime"`      // JWT最大刷新时间（秒）
	TIMEZONE               string `mapstructure:"timezone"`               // 时区设置
	AppName                string `mapstructure:"appName"`                // 应用名称
	GlobalKubeConfigPath   string `mapstructure:"globalKubeConfigPath"`   // KubeConfig全局路径
	EnableLogStreaming     bool   `mapstructure:"enableLogStreaming"`     // 是否启用日志流式传输
	LogTailDefault         int64  `mapstructure:"logTailDefault"`         // 默认日志行数
	LogTailMax             int64  `mapstructure:"logTailMax"`             // 最大日志行数
	LogLimitBytes          int64  `mapstructure:"limitBytes"`             // 日志字节限制
	DefaultClusterID       uint32 `mapstructure:"defaultClusterId"`       // 默认集群ID
	AutoInitK8s            bool   `mapstructure:"autoInitK8s"`            // 开机自启初始化k8s集群
}

type ErrorCodeSettingS struct {
	AllowOverride bool `mapstructure:"allowOverride"` // 是否允许覆盖错误码
}

type PodLogSetting struct {
	EnableStreaming bool  `mapstructure:"enableStreaming"` // 是否启用流式传输
	TailDefault     int64 `mapstructure:"tailDefault"`     // 默认显示行数
	TailMax         int64 `mapstructure:"tailMax"`         // 最大显示行数
	LimitBytes      int64 `mapstructure:"limitBytes"`      // 字节限制
	Timestamps      bool  `mapstructure:"timestamps"`      // 是否显示时间戳
	Previous        bool  `mapstructure:"previous"`        // 是否查看之前的日志
}

// CacheSettingS 缓存配置
// CacheSettingS 定义了缓存配置的结构体，包含缓存服务器的各项设置参数
type CacheSettingS struct {
	Type       string `mapstructure:"type"`       // 缓存类型，如 redis、memcached 等
	Name       string `mapstructure:"name"`       // 缓存名称
	Address    string `mapstructure:"address"`    // 缓存服务器地址，格式如 "host:port"
	Username   string `mapstructure:"username"`   // 缓存服务器用户名（如果需要认证）
	Password   string `mapstructure:"password"`   // 缓存服务器密码（如果需要认证）
	MaxConnect int    `mapstructure:"maxConnect"` // 最大连接数，控制与缓存服务器的并发连接数量
	Network    string `mapstructure:"network"`    // 网络类型，如 "tcp"、"tcp4"、"tcp6" 等
	Secret     string `mapstructure:"secret"`     // 加密密钥，用于加密缓存数据
}

type NodeEvictionConfig struct {
	DefaultGraceSeconds   int64 `mapstructure:"defaultGraceSeconds"`
	MaxGraceSeconds       int64 `mapstructure:"maxGraceSeconds"`
	DefaultTimeoutSeconds int   `mapstructure:"defaultTimeoutSeconds"`
	IgnoreDaemonSets      bool  `mapstructure:"ignoreDaemonSets"`
	DeleteEmptyDir        bool  `mapstructure:"deleteEmptyDir"`
}

type NodeConfig struct {
	Eviction NodeEvictionConfig `mapstructure:"eviction"`
}
