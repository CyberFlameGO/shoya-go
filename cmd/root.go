package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"gitlab.com/george/shoya-go/config"
	"log"
	"os"
)

var ConfigLocation string

func init() {
	rootCmd.PersistentFlags().StringVarP(&ConfigLocation, "config", "c", "config.json", "The location of the config file")
}

var rootCmd = &cobra.Command{
	Use:   "shoya",
	Short: "An API emulator for VRChat",
	Long:  `Shoya is an API emulator ("private server") for the popular VR social game, VRChat.`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		ConfigLocation = cmd.Flag("config").Value.String()
		err := config.LoadConfig(ConfigLocation)
		if err != nil {
			log.Fatalf("%+v", err)
		}
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
