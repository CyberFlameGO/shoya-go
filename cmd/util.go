package cmd

import (
	"context"
	"fmt"
	"github.com/go-redis/redis/v8"
	"github.com/gtsatsis/harvester"
	"gitlab.com/george/shoya-go/config"
	"time"
)

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
