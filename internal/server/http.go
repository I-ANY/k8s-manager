package server

import (
	"net/http"
	"time"

	"go.uber.org/zap"
	"k8soperation/initialize"
	"k8soperation/pkg/app"
	"k8soperation/pkg/logger"
)

func NewHTTPServer(a *app.App) *http.Server {
	engine := initialize.NewEngine(a)

	shutdownTimeout := time.Duration(a.ServerSetting.ShutdownTimeout) * time.Second
	if shutdownTimeout <= 0 {
		shutdownTimeout = 5 * time.Second
	}

	srv := &http.Server{
		Addr:              ":" + a.ServerSetting.Port,
		Handler:           engine,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       a.ServerSetting.ReadTimeout * time.Second,
		WriteTimeout:      a.ServerSetting.WriteTimeout * time.Second,
		IdleTimeout:       a.ServerSetting.IdleTimeout * time.Second,
		MaxHeaderBytes:    1 << 20,
		ErrorLog:          a.Logger.StdLogger(),
	}

	srv.RegisterOnShutdown(func() {
		a.Logger.Info("http k8soperation shutdown")
		if a.SQLDB != nil {
			_ = a.SQLDB.Close()
		}
	})

	return srv
}

func logServerStart(l *logger.Logger, srv *http.Server, runMode string) {
	l.Info("http k8soperation starting",
		zap.String("addr", srv.Addr),
		zap.String("mode", runMode))
}
