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
	"gitlab.com/george/shoya-go/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormLogger "gorm.io/gorm/logger"
	"log"
	"time"
)

func main() {
	vrcpsInit()

	app := fiber.New(fiber.Config{
		ProxyHeader: config.RuntimeConfig.Server.ProxyHeader,
		Prefork:     false,
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

	log.Fatal(app.Listen(config.RuntimeConfig.Server.Address))
}

func vrcpsInit() {
	initializeConfig()
	initializeDB()
	initializeRedis()
	initializeApiConfig()
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
		config.RuntimeConfig.Database.Host,
		config.RuntimeConfig.Database.User,
		config.RuntimeConfig.Database.Password,
		config.RuntimeConfig.Database.Database,
		config.RuntimeConfig.Database.Port)
	config.DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: gormLogger.Default.LogMode(gormLogger.Silent),
	})
	if err != nil {
		panic(err)
	}

	_ = config.DB.AutoMigrate(&models.User{}, &models.Avatar{}, &models.File{}, &models.FavoriteGroup{}, &models.FavoriteItem{}, &models.Moderation{}, &models.Permission{},
		&models.WorldUnityPackage{}, &models.AvatarUnityPackage{})

}

// initializeRedis initializes the redis clients
func initializeRedis() {
	config.RedisClient = redis.NewClient(&redis.Options{
		Addr:     config.RuntimeConfig.Redis.Host,
		Password: config.RuntimeConfig.Redis.Password,
		DB:       config.RuntimeConfig.Redis.Database,
	})
	config.HarvestRedisClient = redis.NewClient(&redis.Options{
		Addr:     config.RuntimeConfig.Redis.Host,
		Password: config.RuntimeConfig.Redis.Password,
		DB:       config.RuntimeConfig.Redis.Database,
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
		WithRedisMonitor(config.HarvestRedisClient, time.Duration(config.RuntimeConfig.ApiConfigRefreshRateMs)*time.Millisecond).
		Create()
	err = h.Harvest(context.Background())
	if err != nil {
		panic(fmt.Errorf("failed to harvest configuration: %v", err))
	}
}
