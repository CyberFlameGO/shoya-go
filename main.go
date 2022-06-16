package main

import (
	"gitlab.com/george/shoya-go/cmd"
	"gitlab.com/george/shoya-go/config"
	"log"
)

func main() {
	err := config.LoadConfig()
	if err != nil {
		log.Fatalf("%+v", err)
	}
	cmd.Execute()
}
