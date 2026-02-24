package routers

import (
	"github.com/pyprism/uCPingGraph/logger"
	"github.com/pyprism/uCPingGraph/utils"
	"go.uber.org/zap"
)

func Init() {
	r := NewRouter()
	serverPort := utils.GetEnv("SERVER_PORT", "8080")
	logger.Get().Info("server starting", zap.String("address", "http://127.0.0.1:"+serverPort))
	if err := r.Run(":" + serverPort); err != nil {
		logger.Get().Fatal("server failed to start", zap.Error(err))
	}
}
