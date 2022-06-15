package cmd

import (
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/spf13/cobra"
	"gitlab.com/george/shoya-go/config"
	"log"
	"os"
	"reflect"
)

func init() {
	configCmd.AddCommand(configLsCmd)
	configCmd.AddCommand(configGetCmd)

	rootCmd.AddCommand(configCmd)
}

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "configure & manage a shoya installation",
}

var configLsCmd = &cobra.Command{
	Use:   "ls",
	Short: "lists all configuration values",
	Run: func(cmd *cobra.Command, args []string) {
		configLs()
	},
}

var configGetCmd = &cobra.Command{
	Use:   "get",
	Short: "retrieves a configuration value from Redis",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		initializeConfig()
		initializeRedis()
		initializeApiConfig()
		configGet(args)
	},
}

func configLs() {
	tb := table.NewWriter()
	tb.SetOutputMirror(os.Stdout)
	tb.AppendHeader(table.Row{"Config Key", "Redis Key"})

	t := reflect.ValueOf(config.ApiConfig{})
	for i := 0; i < t.NumField(); i++ {
		typeField := t.Type().Field(i)

		tag := typeField.Tag
		if tagg, ok := tag.Lookup("redis"); ok {
			tb.AppendRow(table.Row{typeField.Name, tagg})
		} else {
			continue
		}

	}
	tb.Render()
}

func configGet(args []string) {
	var apiConfig interface{}
	apiConfig = &config.ApiConfiguration
	tb := table.NewWriter()
	tb.SetOutputMirror(os.Stdout)
	tb.AppendHeader(table.Row{"Key", "Value"})

	t := reflect.ValueOf(apiConfig)

	for _, arg := range args {
		valueField := t.Elem().FieldByName(arg)

		if !valueField.IsValid() {
			log.Fatalf("Invalid key: %s", arg)
		}

		valueFieldTypePtr := reflect.PtrTo(valueField.Type())
		getMethod, _ := valueFieldTypePtr.MethodByName("Get")

		tb.AppendRow(table.Row{arg, getMethod.Func.Call([]reflect.Value{valueField.Addr()})})
	}
	tb.Render()
}
