package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"
)

var rootCmd = &cobra.Command{
	Use:   "shoya",
	Short: "An API emulator for VRChat",
	Long:  `Shoya is an API emulator ("private server") for the popular VR social game, VRChat.`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
