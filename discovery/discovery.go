package main

import (
	"context"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/rueian/rueidis"
	"github.com/tkanos/gonfig"
	"gitlab.com/george/shoya-go/config"
	"gitlab.com/george/shoya-go/models"
	"log"
	"strconv"
)

var RedisClient rueidis.Client
var RedisCtx = context.Background()

func main() {
	initializeConfig()
	initializeRedis()

	app := fiber.New(fiber.Config{
		ProxyHeader: config.RuntimeConfig.Server.ProxyHeader,
		Prefork:     false,
	})
	//app.Use(recover.New())
	app.Use(logger.New())
	app.Use(func(c *fiber.Ctx) error {
		k := c.Query("apiKey")
		if k == "" {
			k = c.Get("Authorization")
			if k == "" {
				return c.SendStatus(401)
			}
		}

		if k != config.RuntimeConfig.DiscoveryApiKey {
			return c.SendStatus(401)
		}

		return c.Next()
	})

	app.Get("/:instanceId", func(c *fiber.Ctx) error {
		id := c.Params("instanceId")
		i, err := getInstance(id)
		if err != nil {
			if err == NotFoundErr {
				return c.SendStatus(404)
			}

			return c.Status(500).JSON(fiber.Map{
				"error":      err.Error(),
				"instanceId": id,
			})
		}

		return c.JSON(i)
	})

	app.Get("/world/:worldId", func(c *fiber.Ctx) error {
		i, err := findInstancesForWorldId(escapeId(c.Params("worldId")), "public", false)
		if err != nil {
			if err == NotFoundErr {
				return c.SendStatus(404)
			}

			return c.Status(500).JSON(fiber.Map{
				"error": err.Error(),
			})
		}

		return c.JSON(i)
	})

	app.Post("/register/:instanceId", func(c *fiber.Ctx) error {
		var capacity int
		id := c.Params("instanceId")
		if _cap := c.Query("capacity"); _cap == "" {
			return c.Status(500).JSON(fiber.Map{
				"error":      "capacity query parameter is required",
				"instanceId": id,
			})
		} else {
			var err error
			capacity, err = strconv.Atoi(_cap)
			if err != nil {
				return c.Status(500).JSON(fiber.Map{
					"error":      err.Error(),
					"instanceId": id,
				})
			}
		}

		l, err := models.ParseLocationString(id)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{
				"error":      err.Error(),
				"instanceId": id,
			})
		}

		i, err := registerInstance(l.ID, l.LocationString, l.WorldID, l.InstanceType, l.OwnerID, capacity)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{
				"error":      err.Error(),
				"instanceId": id,
			})
		}

		return c.JSON(i)
	})

	app.Post("/unregister/:instanceId", func(c *fiber.Ctx) error {
		i := c.Params("instanceId")
		err := unregisterInstance(i)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{
				"error":      err.Error(),
				"instanceId": i,
			})
		}

		return c.SendStatus(200)
	})

	app.Get("/player/:playerId", func(c *fiber.Ctx) error {
		p := c.Params("playerId")
		i, err := findInstancesPlayerIsIn(p)
		if err != nil {
			if err == NotFoundErr {
				return c.SendStatus(404)
			}

			return c.Status(500).JSON(fiber.Map{
				"error":    err.Error(),
				"playerId": p,
			})
		}
		return c.JSON(i)
	})

	app.Put("/player/:instanceId/:playerId", func(c *fiber.Ctx) error {
		i := c.Params("instanceId")
		p := c.Params("playerId")

		err := addPlayer(i, p)

		if err != nil {
			fmt.Println(err)
			return c.Status(500).JSON(fiber.Map{
				"error":      err.Error(),
				"instanceId": i,
			})
		}

		return c.SendStatus(200)
	})

	app.Delete("/player/:instanceId/:playerId", func(c *fiber.Ctx) error {
		i := c.Params("instanceId")
		p := c.Params("playerId")

		err := removePlayer(i, p)

		if err != nil {
			return c.Status(500).JSON(fiber.Map{
				"error":      err.Error(),
				"instanceId": i,
			})
		}

		return c.SendStatus(200)
	})

	log.Fatal(app.Listen(config.RuntimeConfig.Server.Address))
}

// initializeConfig reads the config.json file and initializes the runtime config
func initializeConfig() {
	err := gonfig.GetConf("config.json", &config.RuntimeConfig)
	if err != nil {
		panic("error reading config file")
	}
}

func initializeRedis() {
	redisClient, err := rueidis.NewClient(rueidis.ClientOption{
		Username:    "default",
		Password:    config.RuntimeConfig.Redis.Password,
		InitAddress: []string{config.RuntimeConfig.Redis.Host},
	})

	if err != nil {
		panic(err)
	}

	RedisClient = redisClient
}