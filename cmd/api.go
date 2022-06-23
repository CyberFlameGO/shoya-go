package cmd

import (
	"github.com/spf13/cobra"
	"gitlab.com/george/shoya-go/services/api"
)

func init() {
	apiCmd.AddCommand(apiServe)

	rootCmd.AddCommand(apiCmd)
}

var apiCmd = &cobra.Command{
	Use:   "api",
	Short: "commands relating to the management of the api server",
}

var apiServe = &cobra.Command{
	Use:     "serve",
	Aliases: []string{"server", "start"},
	Short:   "start the api server",
	Run: func(cmd *cobra.Command, args []string) {
		api.Main()
	},
}
