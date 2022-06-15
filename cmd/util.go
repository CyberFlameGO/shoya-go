package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-redis/redis/v8"
	"github.com/gtsatsis/harvester"
	"github.com/tkanos/gonfig"
	"gitlab.com/george/shoya-go/config"
	"os"
	"time"
)

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
