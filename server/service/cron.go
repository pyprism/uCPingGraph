package service

import (
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/pyprism/uCPingGraph/logger"
	"github.com/pyprism/uCPingGraph/models"
	"go.uber.org/zap"
)

func CleanDB() {
	done := make(chan os.Signal, 1)
	signal.Notify(done, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)

	go func() {
		ticker := time.NewTicker(24 * time.Hour)
		defer ticker.Stop()

		for {
			stats := models.Stat{}
			if err := stats.Cleanup(); err != nil {
				logger.CaptureError(err, "cleanup job failed")
			} else {
				logger.Get().Info("cleanup job completed successfully")
			}

			<-ticker.C
		}
	}()

	logger.Get().Info("cleanup cron started, waiting for signal")
	sig := <-done
	logger.Get().Info("received signal, shutting down", zap.String("signal", sig.String()))
}
