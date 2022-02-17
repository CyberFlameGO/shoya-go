package main

import (
	"context"
	"fmt"
	"github.com/go-redis/redis/v8"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/websocket/v2"
	"github.com/gtsatsis/harvester"
	"github.com/tkanos/gonfig"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormLogger "gorm.io/gorm/logger"
	"log"
	"time"
)

var RuntimeConfig Config
var ApiConfiguration = ApiConfig{}

var RedisClient *redis.Client
var HarvestRedisClient *redis.Client
var DB *gorm.DB

func main() {
	vrcpsInit()

	app := fiber.New(fiber.Config{
		Prefork: false,
	})
	app.Use(recover.New())
	app.Use(logger.New())
	app.Use(IsGameRequestMiddleware)

	app.Use("/ws", func(c *fiber.Ctx) error {
		if websocket.IsWebSocketUpgrade(c) {
			c.Locals("allowed", true)
			return c.Next()
		}
		return fiber.ErrUpgradeRequired
	})
	app.Get("/ws", websocket.New(func(c *websocket.Conn) {
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
	systemRoutes(app)
	authRoutes(app)
	usersRoutes(app)
	worldsRoutes(app)
	photonRoutes(app)
	instanceRoutes(app)
	avatarsRoutes(app)

	log.Fatal(app.Listen(RuntimeConfig.Server.Address))
}

func vrcpsInit() {
	initializeConfig()
	initializeDB()
	initializeRedis()
	initializeApiConfig()
}

// initializeConfig reads the config.json file and initializes the runtime config
func initializeConfig() {
	err := gonfig.GetConf("config.json", &RuntimeConfig)
	if err != nil {
		panic("error reading config file")
	}
}

// initializeDB initializes the database connection (and runs migrations)
func initializeDB() {
	var err error
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=disable TimeZone=Etc/GMT",
		RuntimeConfig.Database.Host,
		RuntimeConfig.Database.User,
		RuntimeConfig.Database.Password,
		RuntimeConfig.Database.Database,
		RuntimeConfig.Database.Port)
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: gormLogger.Default.LogMode(gormLogger.Silent),
	})
	if err != nil {
		panic(err)
	}

	_ = DB.AutoMigrate(&User{}, &Avatar{}, &File{}, &FavoriteGroup{}, &FavoriteItem{}, &Moderation{}, &Permission{},
		&WorldUnityPackage{}, &AvatarUnityPackage{})

}

// initializeRedis initializes the redis clients
func initializeRedis() {
	RedisClient = redis.NewClient(&redis.Options{
		Addr:     RuntimeConfig.Redis.Host,
		Password: RuntimeConfig.Redis.Password,
		DB:       RuntimeConfig.Redis.Database,
	})
	HarvestRedisClient = redis.NewClient(&redis.Options{
		Addr:     RuntimeConfig.Redis.Host,
		Password: RuntimeConfig.Redis.Password,
		DB:       RuntimeConfig.Redis.Database,
	})

	_, err := RedisClient.Ping(context.Background()).Result()
	_, err2 := HarvestRedisClient.Ping(context.Background()).Result()
	if err != nil || err2 != nil {
		panic(err)
	}
}

// initializeApiConfig initializes harvester client used to configure the API
func initializeApiConfig() {
	h, err := harvester.New(&ApiConfiguration).
		WithRedisSeed(HarvestRedisClient).
		WithRedisMonitor(HarvestRedisClient, time.Duration(RuntimeConfig.ApiConfigRefreshRateMs)*time.Millisecond).
		Create()
	err = h.Harvest(context.Background())
	if err != nil {
		panic(fmt.Errorf("failed to harvest configuration: %v", err))
	}
}
