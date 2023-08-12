package cmd

import (
	"github.com/spf13/cobra"
	"log"
)

func init() {
	RootCmd.AddCommand(deviceCmd)
}

var deviceCmd = &cobra.Command{
	Use:   "device",
	Short: "Type device add",
	Long:  `Type device add to add a device to the database.`,
	Run: func(cmd *cobra.Command, args []string) {
		log.Println("Type device add to add a device to the database.")
	},
}
