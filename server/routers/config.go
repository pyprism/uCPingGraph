package routers

import (
	"context"
	"errors"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/pyprism/uCPingGraph/logger"
	"github.com/pyprism/uCPingGraph/utils"
	"go.uber.org/zap"
)

func Init() {
	r := NewRouter()
	serverPort := utils.GetEnv("SERVER_PORT", "8080")

	srv := &http.Server{
		Addr:    ":" + serverPort,
		Handler: r,
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		logger.Get().Info("server starting", zap.String("address", "http://127.0.0.1:"+serverPort))
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Get().Fatal("server failed to start", zap.Error(err))
		}
	}()

	<-ctx.Done()
	stop()
	logger.Get().Info("server shutting down")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Get().Error("server forced to shut down", zap.Error(err))
	}
}
