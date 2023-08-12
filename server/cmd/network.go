package cmd

import (
	"github.com/spf13/cobra"
	"log"
)

func init() {
	RootCmd.AddCommand(networkCmd)
}

var networkCmd = &cobra.Command{
	Use:   "network",
	Short: "Type network add",
	Long:  `Type network add to add a network to the database.`,
	Run: func(cmd *cobra.Command, args []string) {
		log.Println("Type network add to add a network to the database.")
	},
}
