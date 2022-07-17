package cmd

import (
	"github.com/spf13/cobra"
	"gitlab.com/george/shoya-go/services/presence"
)

func init() {
	presenceCmd.AddCommand(presenceServe)

	rootCmd.AddCommand(presenceCmd)
}

var presenceCmd = &cobra.Command{
	Use:   "presence",
	Short: "commands relating to the management of the presence service",
}

var presenceServe = &cobra.Command{
	Use:     "serve",
	Aliases: []string{"server", "start"},
	Short:   "start the presence service",
	Run: func(cmd *cobra.Command, args []string) {
		presence.Main()
	},
}
