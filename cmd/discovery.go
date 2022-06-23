package cmd

import (
	"github.com/spf13/cobra"
	"gitlab.com/george/shoya-go/services/discovery"
)

func init() {
	discoveryCmd.AddCommand(discoveryServe)

	rootCmd.AddCommand(discoveryCmd)
}

var discoveryCmd = &cobra.Command{
	Use:   "discovery",
	Short: "commands relating to the management of the discovery service",
}

var discoveryServe = &cobra.Command{
	Use:     "serve",
	Aliases: []string{"server", "start"},
	Short:   "start the discovery service",
	Run: func(cmd *cobra.Command, args []string) {
		discovery.Main()
	},
}
