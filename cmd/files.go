package cmd

import (
	"github.com/spf13/cobra"
	"gitlab.com/george/shoya-go/services/files"
)

func init() {
	filesCmd.AddCommand(filesServe)

	rootCmd.AddCommand(filesCmd)
}

var filesCmd = &cobra.Command{
	Use:   "files",
	Short: "commands relating to the management of the files service",
}

var filesServe = &cobra.Command{
	Use:     "serve",
	Aliases: []string{"server", "start"},
	Short:   "start the files service",
	Run: func(cmd *cobra.Command, args []string) {
		files.Main()
	},
}
