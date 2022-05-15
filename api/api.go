package main

import (
	"context"
	"fmt"
	"github.com/go-redis/redis/v8"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gtsatsis/harvester"
	"github.com/tkanos/gonfig"
	"gitlab.com/george/shoya-go/config"
	"gitlab.com/george/shoya-go/discovery/discovery_client"
	"gitlab.com/george/shoya-go/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormLogger "gorm.io/gorm/logger"
	"log"
	"strings"
	"time"
)

var DiscoveryService *discovery_client.Discovery

func main() {
	vrcpsInit()

	app := fiber.New(fiber.Config{
		ProxyHeader:   config.RuntimeConfig.Api.Fiber.ProxyHeader,
		Prefork:       config.RuntimeConfig.Api.Fiber.Prefork,
		CaseSensitive: false,
	})
	app.Use(recover.New())
	app.Use(logger.New())
	app.Use(IsGameRequestMiddleware)

	systemRoutes(app)
	authRoutes(app)
	usersRoutes(app)
	worldsRoutes(app)
	photonRoutes(app)
	instanceRoutes(app)
	avatarsRoutes(app)

	log.Fatal(app.Listen(config.RuntimeConfig.Api.Fiber.ListenAddress))
}

func vrcpsInit() {
	initializeConfig()
	initializeDB()
	initializeRedis()
	initializeApiConfig()

	if config.ApiConfiguration.DiscoveryServiceEnabled.Get() {
		DiscoveryService = discovery_client.NewDiscovery(config.ApiConfiguration.DiscoveryServiceUrl.Get(), config.ApiConfiguration.DiscoveryServiceApiKey.Get())
	}
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
		config.RuntimeConfig.Api.Postgres.Host,
		config.RuntimeConfig.Api.Postgres.User,
		config.RuntimeConfig.Api.Postgres.Password,
		config.RuntimeConfig.Api.Postgres.Database,
		config.RuntimeConfig.Api.Postgres.Port)
	config.DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: gormLogger.Default.LogMode(gormLogger.Silent),
	})
	if err != nil {
		panic(err)
	}

	err = config.DB.AutoMigrate(&models.User{})
	if err != nil {
		fmt.Println(err)
	}
	err = config.DB.AutoMigrate(&models.Avatar{})
	if err != nil {
		fmt.Println(err)
	}
	err = config.DB.AutoMigrate(&models.File{})
	if err != nil {
		fmt.Println(err)
	}
	err = config.DB.AutoMigrate(&models.FavoriteGroup{})
	if err != nil {
		fmt.Println(err)
	}
	err = config.DB.AutoMigrate(&models.FavoriteItem{})
	if err != nil {
		fmt.Println(err)
	}
	err = config.DB.AutoMigrate(&models.Moderation{})
	if err != nil {
		fmt.Println(err)
	}
	err = config.DB.AutoMigrate(&models.Permission{})
	if err != nil {
		fmt.Println(err)
	}
	err = config.DB.AutoMigrate(models.WorldUnityPackage{})
	if err != nil {
		fmt.Println(err)
	}
	err = config.DB.AutoMigrate(&models.AvatarUnityPackage{})
	if err != nil {
		fmt.Println(err)
	}
	err = config.DB.AutoMigrate(&models.PlayerModeration{})
	if err != nil {
		fmt.Println(err)
	}

}

// initializeRedis initializes the redis clients
func initializeRedis() {
	config.RedisClient = redis.NewClient(&redis.Options{
		Addr:     config.RuntimeConfig.Api.Redis.Host,
		Password: config.RuntimeConfig.Api.Redis.Password,
		DB:       config.RuntimeConfig.Api.Redis.Database,
	})
	config.HarvestRedisClient = redis.NewClient(&redis.Options{
		Addr:     config.RuntimeConfig.Api.Redis.Host,
		Password: config.RuntimeConfig.Api.Redis.Password,
		DB:       config.RuntimeConfig.Api.Redis.Database,
	})

	_, err := config.RedisClient.Ping(context.Background()).Result()
	_, err2 := config.HarvestRedisClient.Ping(context.Background()).Result()
	if err != nil || err2 != nil {
		panic(err)
	}
}

// initializeApiConfig initializes harvester client used to configure the API
func initializeApiConfig() {
	h, err := harvester.New(&config.ApiConfiguration).
		WithRedisSeed(config.HarvestRedisClient).
		WithRedisMonitor(config.HarvestRedisClient, time.Duration(config.RuntimeConfig.Api.ApiConfigRefreshRateMs)*time.Millisecond).
		Create()
	err = h.Harvest(context.Background())
	if err != nil {
		panic(fmt.Errorf("failed to harvest configuration: %v", err))
	}
}

func boolConvert(s string) bool {
	s = strings.ToLower(s)
	if s == "true" {
		return true
	}

	return false
}
