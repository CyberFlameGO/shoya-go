package cmd

import (
	"context"
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
	configCmd.AddCommand(configSetCmd)

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
		initializeRedis()
		initializeApiConfig()
		configGet(args)
	},
}

var configSetCmd = &cobra.Command{
	Use:   "set",
	Short: "sets a configuration value in Redis",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		initializeRedis()
		initializeApiConfig()
		configSet(args)
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

func configSet(args []string) {
	t := reflect.ValueOf(config.ApiConfiguration)

	var typeField reflect.StructField
	var ok bool
	if typeField, ok = t.Type().FieldByName(args[0]); !ok {
		log.Fatalf("Invalid key: %s", args[0])
	}

	var redisTag string
	if redisTag, ok = typeField.Tag.Lookup("redis"); !ok {
		log.Fatalf("Key %s cannot be set as it has no redis tag", args[0])
	}

	do := config.RedisClient.Set(context.Background(), redisTag, args[1], 0)
	if do.Err() != nil {
		log.Fatalf(do.Err().Error())
	}

	log.Printf("Key %s has been set with value %s\n", args[0], args[1])
}
