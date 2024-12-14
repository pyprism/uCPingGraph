package cmd

import (
	"github.com/pyprism/uCPingGraph/models"
	"github.com/spf13/cobra"
)

func init() {
	RootCmd.AddCommand(cleanCmd)
}

var cleanCmd = &cobra.Command{
	Use:   "cleanup",
	Short: "Clean old data",
	Long:  `Clean old data from the database.`,
	Run: func(cmd *cobra.Command, args []string) {
		models.ConnectDb()
		stats := models.Stat{}
		err := stats.Cleanup()
		if err != nil {
			panic(err)
		}
	},
}
