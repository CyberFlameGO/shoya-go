package cmd

import (
	"github.com/spf13/cobra"
	"gitlab.com/george/shoya-go/services/discovery"
)

func init() {
	wsCmd.AddCommand(wsServe)

	rootCmd.AddCommand(wsCmd)
}

var wsCmd = &cobra.Command{
	Use:   "ws",
	Short: "commands relating to the management of the websocket server",
}

var wsServe = &cobra.Command{
	Use:     "serve",
	Aliases: []string{"server", "start"},
	Short:   "start the websocket server",
	Run: func(cmd *cobra.Command, args []string) {
		discovery.Main()
	},
}
