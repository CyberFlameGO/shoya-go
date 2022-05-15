package main

import (
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/tkanos/gonfig"
	"gitlab.com/george/shoya-go/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormLogger "gorm.io/gorm/logger"
	"log"
	"strconv"
)

func main() {
	initializeConfig()
	initializeDB()

	app := fiber.New(fiber.Config{
		ProxyHeader: config.RuntimeConfig.Analytics.Fiber.ProxyHeader,
		Prefork:     false,
	})
	app.Use(recover.New())
	app.Use(logger.New())

	app.Get("/", func(c *fiber.Ctx) error {
		var lim = 100
		var off = 0
		var e []Event
		var ae []*ApiAnalyticsEvent

		if _lim := c.Query("limit"); _lim != "" {
			__lim, err := strconv.Atoi(_lim)
			if err != nil {
				return c.Status(400).JSON(nil)
			}

			if __lim > 1250 {
				return c.Status(400).JSON("Requested too many")
			}
			lim = __lim
		}

		if _off := c.Query("offset"); _off != "" {
			__off, err := strconv.Atoi(_off)
			if err != nil {
				return c.Status(400).JSON(nil)
			}

			off = __off
		}

		config.DB.Model(&Event{}).Where("type = ?", "game").Where("data->>'event_type' = ?", EventTypeWorldEnterWorld).
			Limit(lim).Offset(off).Find(&e)

		for _, _e := range e {
			ae = append(ae, _e.ToApi())
		}
		return c.JSON(ae)
	})

	log.Fatal(app.Listen(config.RuntimeConfig.Analytics.Fiber.ListenAddress))
}

// initializeConfig reads the config.json file and initializes the runtime config
func initializeConfig() {
	err := gonfig.GetConf("config.json", &config.RuntimeConfig)
	if err != nil {
		panic("error reading config file")
	}
}

// initializeDB initializes the database connection (and runs migrations)
func initializeDB() {
	var err error
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=disable TimeZone=Etc/GMT",
		config.RuntimeConfig.Analytics.Postgres.Host,
		config.RuntimeConfig.Analytics.Postgres.User,
		config.RuntimeConfig.Analytics.Postgres.Password,
		config.RuntimeConfig.Analytics.Postgres.Database,
		config.RuntimeConfig.Analytics.Postgres.Port)
	config.DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: gormLogger.Default.LogMode(gormLogger.Info),
	})
	if err != nil {
		panic(err)
	}

	_ = config.DB.AutoMigrate(&Event{})

}
