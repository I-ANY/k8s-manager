package app

import (
	"context"
	"database/sql"
	k8sclient "k8soperation/pkg/k8s"
	"k8soperation/pkg/logger"
	"k8soperation/pkg/setting/types"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	metricsclient "k8s.io/metrics/pkg/client/clientset/versioned"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// App holds all application dependencies, replacing the old global package.
type App struct {
	AppConfig *types.AppConfig

	ServerSetting   *types.ServerSettingS
	AppSetting      *types.AppSettingS
	DatabaseSetting *types.DatabaseSettingS
	CacheSetting    *types.CacheSettingS
	PodLogSetting   *types.PodLogSetting
	NodeSetting     *types.NodeConfig

	DB           *gorm.DB
	SQLDB        *sql.DB
	Logger       *logger.Logger
	BizLogger    *logger.Logger
	SessionStore sessions.Store

	KubeClient       *kubernetes.Clientset
	KubeConfig       *rest.Config
	MetricsClient    *metricsclient.Clientset
	SupportsEventsV1 bool

	AppConfigClient client.Client

	DefaultPageSize int
	MaxPageSize     int
}

func NewApp() *App {
	return &App{
		DefaultPageSize: 10,
		MaxPageSize:     100,
	}
}

func (a *App) BusinessLog(ctx context.Context, action, operator string, target, details map[string]any) {
	if a == nil {
		return
	}
	mirror := false
	if a.AppSetting != nil {
		mirror = a.AppSetting.MirrorBusinessToSystem
	}
	logger.LogBusiness(ctx, a.BizLogger, a.Logger, mirror, action, operator, target, details)
}

// K8sClient returns a *k8s.Client built from this App.
func (a *App) K8sClient() *k8sclient.Client {
	return &k8sclient.Client{
		Interface:        a.KubeClient,
		Logger:           a.Logger,
		PodLogSetting:    a.PodLogSetting,
		RestConfig:       a.KubeConfig,
		MetricsClient:    a.MetricsClient,
		SupportsEventsV1: a.SupportsEventsV1,
	}
}

// Middleware stores this App in every gin.Context so handlers can retrieve it.
func (a *App) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("app", a)
		c.Next()
	}
}
