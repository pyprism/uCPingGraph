package cmd

import (
	"github.com/pyprism/uCPingGraph/models"
	"github.com/pyprism/uCPingGraph/prompts"
	"github.com/spf13/cobra"
	"log"
)

func init() {
	RootCmd.AddCommand(generateData)
}

var generateData = &cobra.Command{
	Use:   "generate",
	Short: "Generate dummy data",
	Long:  `Generate dummy data for development`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := models.ConnectDb(); err != nil {
			log.Fatalf("failed to connect database: %v", err)
		}
		log.Println("generating dummy data")
		prompts.GenerateDummyData()
	},
}
