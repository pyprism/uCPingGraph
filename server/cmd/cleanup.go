package cmd

import (
	"github.com/pyprism/uCPingGraph/logger"
	"github.com/pyprism/uCPingGraph/models"
	"github.com/pyprism/uCPingGraph/service"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

func init() {
	RootCmd.AddCommand(cleanCmd)
}

var cleanCmd = &cobra.Command{
	Use:   "cleanup",
	Short: "Clean old data",
	Long:  `Clean old data from the database.`,
	Run: func(cmd *cobra.Command, args []string) {
		logger.Init()
		defer logger.Shutdown()

		if err := models.ConnectDb(); err != nil {
			logger.Get().Fatal("failed to connect database", zap.Error(err))
		}
		service.CleanDB()
	},
}
