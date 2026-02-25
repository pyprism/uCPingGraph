package cmd

import (
	"github.com/pyprism/uCPingGraph/models"
	"github.com/pyprism/uCPingGraph/prompts"
	"github.com/spf13/cobra"
	"log"
)

func init() {
	deviceCmd.AddCommand(deviceAddCmd)
}

var deviceAddCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a network to the database",
	Long:  `Add a network to the database.`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := models.ConnectDb(); err != nil {
			log.Fatalf("failed to connect database: %v", err)
		}
		prompts.AddNewDevice()
	},
}
