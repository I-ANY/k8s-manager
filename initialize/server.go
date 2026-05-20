package initialize

import (
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"k8soperation/internal/health"
	"k8soperation/middlewares"
	"k8soperation/pkg/app"
)

type Engine struct {
	*gin.Engine
	Mode string
}

func NewEngine(a *app.App) *Engine {
	g := &Engine{
		Mode: gin.ReleaseMode,
	}
	gin.SetMode(g.Mode)
	g.injectMiddlewares(a)
	g.injectRouters(a)
	return g
}

func (s *Engine) injectMiddlewares(a *app.App) {
	s.Engine = gin.New()

	if s.Mode == gin.TestMode {
		return
	}

	s.Use(middlewares.RequestMetadata())
	s.Use(middlewares.Logger(a.Logger))
	s.Use(middlewares.Recovery(a.Logger))
	s.Use(middlewares.K8sError())

	RegisterSession(s.Engine, a)
}

func (s *Engine) injectRouters(a *app.App) {
	r := s.Engine

	// App middleware must be first so handlers can retrieve deps
	r.Use(a.Middleware())

	health.Register(r, health.Checks{DB: a.SQLDB})

	apiRouter := r.Group("")
	s.injectRouterGroup(apiRouter, a)
}

func RegisterSession(r *gin.Engine, a *app.App) {
	if a.SessionStore == nil {
		a.Logger.Warn("session store is nil, session middleware not installed")
		return
	}
	name := a.CacheSetting.Name
	if name == "" {
		name = "k8soperation_sid"
	}

	a.Logger.Info("install session middleware",
		zap.String("cookie_name", name),
	)

	r.Use(sessions.Sessions(name, a.SessionStore))
}
