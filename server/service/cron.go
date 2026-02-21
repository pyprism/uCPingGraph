package service

import (
	"github.com/pyprism/uCPingGraph/models"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
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
				log.Printf("cleanup job failed: %v", err)
			}

			<-ticker.C
		}
	}()

	<-done
}
