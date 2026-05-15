package server

import (
	"go.uber.org/zap"
	"k8soperation/global"
	"k8soperation/initialize"
	"net/http"
	"time"
)

func NewHTTPServer() *http.Server {
	// 初始化引擎
	engine := initialize.NewEngine()

	// 兜底超时
	shutdownTimeout := time.Duration(global.ServerSetting.ShutdownTimeout) * time.Second

	if shutdownTimeout <= 0 {
		shutdownTimeout = 5 * time.Second
	}

	srv := &http.Server{
		Addr:              ":" + global.ServerSetting.Port,
		Handler:           engine,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       global.ServerSetting.ReadTimeout * time.Second,
		WriteTimeout:      global.ServerSetting.WriteTimeout * time.Second,
		IdleTimeout:       global.ServerSetting.IdleTimeout * time.Second,
		MaxHeaderBytes:    1 << 20,
		ErrorLog:          global.Logger.StdLogger(),
	}

	srv.RegisterOnShutdown(func() {
		global.Logger.Info("http k8soperation shutdown")
		if global.SQLDB != nil {
			_ = global.SQLDB.Close()
		}
	})

	return srv
}

// 记录服务器启动日志
func logServerStart(srv *http.Server) {
	global.Logger.Info("http k8soperation starting",
		zap.String("addr", srv.Addr),
		zap.String("mode", global.ServerSetting.RunMode))
}
