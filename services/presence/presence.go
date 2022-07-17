package presence

import (
	"context"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/rueian/rueidis"
	"gitlab.com/george/shoya-go/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormLogger "gorm.io/gorm/logger"
	"log"
)

var RedisClient rueidis.Client
var RedisCtx = context.Background()

func Main() {
	if config.RuntimeConfig.Presence == nil {
		log.Fatalf("error reading config: RuntimeConfig.Presence was nil")
	}

	initializeDB()
	initializeRedis()
	app := fiber.New(fiber.Config{
		ProxyHeader: config.RuntimeConfig.Presence.Fiber.ProxyHeader,
		Prefork:     false,
	})

	initRoutes(app)

	app.Use(logger.New())
	log.Fatal(app.Listen(config.RuntimeConfig.Presence.Fiber.ListenAddress))
}

func initializeRedis() {
	redisClient, err := rueidis.NewClient(rueidis.ClientOption{
		Username:    "default",
		Password:    config.RuntimeConfig.Presence.Redis.Password,
		InitAddress: []string{config.RuntimeConfig.Presence.Redis.Host},
		SelectDB:    1,
	})

	if err != nil {
		panic(err)
	}

	RedisClient = redisClient
}

// initializeDB initializes the database connection (and runs migrations)
func initializeDB() {
	var err error
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=disable TimeZone=Etc/GMT",
		config.RuntimeConfig.Presence.Postgres.Host,
		config.RuntimeConfig.Presence.Postgres.User,
		config.RuntimeConfig.Presence.Postgres.Password,
		config.RuntimeConfig.Presence.Postgres.Database,
		config.RuntimeConfig.Presence.Postgres.Port)
	config.DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: gormLogger.Default.LogMode(gormLogger.Silent),
	})
	if err != nil {
		panic(err)
	}
}
