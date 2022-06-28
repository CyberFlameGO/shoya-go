package api

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gtsatsis/harvester"
	"gitlab.com/george/shoya-go/config"
	pb "gitlab.com/george/shoya-go/gen/v1/proto"
	"gitlab.com/george/shoya-go/models"
	"gitlab.com/george/shoya-go/services/discovery/discovery_client"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormLogger "gorm.io/gorm/logger"
)

var DiscoveryService *discovery_client.Discovery
var FilesService pb.FileClient

func Main() {
	shoyaInit()

	app := fiber.New(fiber.Config{
		ProxyHeader:   config.RuntimeConfig.Api.Fiber.ProxyHeader,
		Prefork:       config.RuntimeConfig.Api.Fiber.Prefork,
		CaseSensitive: false,
	})
	app.Use(recover.New())
	app.Use(logger.New())
	app.Use(AddXPoweredByHeader, IsGameRequestMiddleware)

	initializeRoutes(app)

	log.Fatal(app.Listen(config.RuntimeConfig.Api.Fiber.ListenAddress))
}

func shoyaInit() {
	if config.RuntimeConfig.Api == nil {
		log.Fatalf("error reading config: RuntimeConfig.Api was nil")
	}

	initializeDB()
	initializeRedis()
	initializeApiConfig()

	if config.ApiConfiguration.DiscoveryServiceEnabled.Get() {
		DiscoveryService = discovery_client.NewDiscovery(config.ApiConfiguration.DiscoveryServiceUrl.Get(), config.ApiConfiguration.DiscoveryServiceApiKey.Get())
	}
	initializeFilesClient()

	initializeHealthChecks()
}

func initializeRoutes(app *fiber.App) {
	systemRoutes(app)
	authRoutes(app)
	usersRoutes(app)
	worldsRoutes(app)
	photonRoutes(app)
	instanceRoutes(app)
	avatarsRoutes(app)
	favoriteRoutes(app)
	fileRoutes(app)
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
		fmt.Printf("failed to migrate User model: %s\n", err)
	}
	err = config.DB.AutoMigrate(&models.Avatar{})
	if err != nil {
		fmt.Printf("failed to migrate Avatar model: %s\n", err)
	}

	err = config.DB.AutoMigrate(&models.FavoriteGroup{})
	if err != nil {
		fmt.Printf("failed to migrate FavoriteGroup model: %s\n", err)
	}
	err = config.DB.AutoMigrate(&models.FavoriteItem{})
	if err != nil {
		fmt.Printf("failed to migrate FavoriteItem model: %s\n", err)
	}
	err = config.DB.AutoMigrate(&models.Moderation{})
	if err != nil {
		fmt.Printf("failed to migrate Moderation model: %s\n", err)
	}
	err = config.DB.AutoMigrate(&models.Permission{})
	if err != nil {
		fmt.Printf("failed to migrate Permission model: %s\n", err)
	}
	err = config.DB.AutoMigrate(models.WorldUnityPackage{})
	if err != nil {
		fmt.Printf("failed to migrate WorldUnityPackage model: %s\n", err)
	}
	err = config.DB.AutoMigrate(&models.AvatarUnityPackage{})
	if err != nil {
		fmt.Printf("failed to migrate AvatarUnityPackage model: %s\n", err)
	}
	err = config.DB.AutoMigrate(&models.PlayerModeration{})
	if err != nil {
		fmt.Printf("failed to migrate PlayerModeration model: %s\n", err)
	}

	err = config.DB.AutoMigrate(&models.File{})
	if err != nil {
		fmt.Printf("failed to migrate File model: %s\n", err)
	}

	err = config.DB.AutoMigrate(&models.FileVersion{})
	if err != nil {
		fmt.Printf("failed to migrate FileVersion model: %s\n", err)
	}

	err = config.DB.AutoMigrate(&models.FileDescriptor{})
	if err != nil {
		fmt.Printf("failed to migrate FileDescriptor model: %s\n", err)
	}

	err = config.DB.AutoMigrate(&models.FriendRequest{})
	if err != nil {
		fmt.Printf("failed to migrate FriendRequest model: %s\n", err)
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
	if err != nil {
		panic(fmt.Errorf("failed to set up configuration harvester: %v", err))
	}

	err = h.Harvest(context.Background())
	if err != nil {
		panic(fmt.Errorf("failed to harvest configuration: %v", err))
	}
}

func initializeHealthChecks() {
	go redisHealthCheck()
	go harvestRedisHealthCheck()
	go postgresHealthCheck()
	go filesHealthCheck()
}

func initializeFilesClient() {
	conn, err := grpc.Dial(config.ApiConfiguration.FilesEndpoint.Get(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	FilesService = pb.NewFileClient(conn)
}
