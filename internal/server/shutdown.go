package server

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"
	"k8soperation/pkg/app"
	"k8soperation/pkg/logger"
)

func ListenAndServeAsync(l *logger.Logger, srv *http.Server) {
	go func() {
		logServerStart(l, srv, "release")
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			l.Error("http k8soperation err", zap.Error(err))
		}
	}()
}

func GracefulShutdown(a *app.App, srv *http.Server, timeout time.Duration) {
	stopCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt,
		syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	defer stop()
	<-stopCtx.Done()

	a.Logger.Info("shutting down k8soperation...")
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		a.Logger.Error("k8soperation shutdown err", zap.Error(err))
	} else {
		a.Logger.Info("k8soperation exiting", zap.Duration("timeout", timeout))
	}

	if a.BizLogger != nil {
		a.BizLogger.Info("service.stop")
	}
}
