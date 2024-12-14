package service

import (
	"github.com/pyprism/uCPingGraph/models"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func CleanDB() {
	done := make(chan os.Signal, 1)

	go func() {
		for {
			stats := models.Stat{}
			err := stats.Cleanup()
			if err != nil {
				panic(err)
			}
			<-time.After(time.Duration(24) * time.Hour)
		}
	}()

	<-done
	signal.Notify(done, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
}
