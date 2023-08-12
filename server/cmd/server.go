package cmd

import (
	"github.com/pyprism/uCPingGraph/models"
	"github.com/pyprism/uCPingGraph/routers"
	"github.com/spf13/cobra"
)

func init() {
	RootCmd.AddCommand(serverCmd)
}

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Start the server",
	Long:  `Start the server to serve the web app.`,
	Run: func(cmd *cobra.Command, args []string) {
		models.ConnectDb()
		routers.Init()
	},
}
