package ws

import (
	"encoding/json"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/websocket/v2"
	"github.com/tkanos/gonfig"
	"gitlab.com/george/shoya-go/config"
	"log"
	"os"
)

func Main() {
	initializeConfig()
	app := fiber.New(fiber.Config{
		ProxyHeader: config.RuntimeConfig.Ws.Fiber.ProxyHeader,
		Prefork:     false,
	})

	app.Use(recover.New())
	app.Use(logger.New())

	app.Use("/", func(c *fiber.Ctx) error {
		if websocket.IsWebSocketUpgrade(c) {
			c.Locals("allowed", true)
			return c.Next()
		}
		return fiber.ErrUpgradeRequired
	})
	app.Get("/", websocket.New(func(c *websocket.Conn) {
		var (
			mt  int
			msg []byte
			err error
		)
		for {
			fmt.Println("We got a ws conn")
			if mt, msg, err = c.ReadMessage(); err != nil {
				log.Println("read:", err)
				break
			}
			log.Printf("recv: %s", msg)

			if err = c.WriteMessage(mt, msg); err != nil {
				log.Println("write:", err)
				break
			}
		}
	}))

	log.Fatal(app.Listen(config.RuntimeConfig.Ws.Fiber.ListenAddress))
}

// initializeConfig reads the config.json file and initializes the runtime config
func initializeConfig() {
	err := gonfig.GetConf("config.json", &config.RuntimeConfig)
	if err != nil {
		envJson := os.Getenv("SHOYA_CONFIG_JSON")
		if envJson == "" {
			panic("error reading config file or environment variable")
		}

		err = json.Unmarshal([]byte(envJson), &config.RuntimeConfig)
		if err != nil {
			panic("could not unmarshal config from environment")
		}
	}
}
