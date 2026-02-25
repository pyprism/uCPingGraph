package cmd

import (
	"github.com/pyprism/uCPingGraph/logger"
	"github.com/pyprism/uCPingGraph/models"
	"github.com/pyprism/uCPingGraph/routers"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

func init() {
	RootCmd.AddCommand(serverCmd)
}

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Start the server",
	Long:  `Start the server to serve the web app.`,
	Run: func(cmd *cobra.Command, args []string) {
		logger.Init()
		defer logger.Shutdown()

		if err := models.ConnectDb(); err != nil {
			logger.Get().Fatal("failed to connect database", zap.Error(err))
		}
		routers.Init()
	},
}
