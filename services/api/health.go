package main

import (
	"context"
	"database/sql"
	"gitlab.com/george/shoya-go/config"
	"time"
)

type HealthStatus struct {
	Redis    HealthStatusDetails `json:"redis"`
	Config   HealthStatusDetails `json:"config"`
	Postgres HealthStatusDetails `json:"postgres"`
}

type HealthStatusDetails struct {
	Ok    bool   `json:"ok"`
	Error string `json:"error,omitempty"`
}

var healthStatus HealthStatus
var healthCheckFrequency = time.Second * 10

func redisHealthCheck() {
	var err error
	for {
		_, err = config.RedisClient.Ping(context.TODO()).Result()
		if err != nil {
			healthStatus.Redis.Ok = false
			healthStatus.Redis.Error = "ping failed with error: " + err.Error()
		} else if !healthStatus.Redis.Ok {
			healthStatus.Redis.Ok = true
			healthStatus.Redis.Error = ""
		}
		time.Sleep(healthCheckFrequency)
	}
}

func harvestRedisHealthCheck() {
	var err error
	for {
		_, err = config.HarvestRedisClient.Ping(context.TODO()).Result()
		if err != nil {
			healthStatus.Config.Ok = false
			healthStatus.Config.Error = "ping failed with error: " + err.Error()
		} else if !healthStatus.Config.Ok {
			healthStatus.Config.Ok = true
			healthStatus.Config.Error = ""
		}
		time.Sleep(healthCheckFrequency)
	}
}

func postgresHealthCheck() {
	var db *sql.DB
	var err error
	for {
		db, err = config.DB.DB()
		if err != nil {
			healthStatus.Postgres.Ok = false
			healthStatus.Postgres.Error = "assigning sqlDB failed with error: " + err.Error()
			time.Sleep(healthCheckFrequency)
			continue
		}

		err = db.Ping()
		if err != nil {
			healthStatus.Postgres.Ok = false
			healthStatus.Postgres.Error = "ping failed with error: " + err.Error()
		} else if !healthStatus.Postgres.Ok {
			healthStatus.Postgres.Ok = true
			healthStatus.Postgres.Error = ""
		}

		time.Sleep(healthCheckFrequency)
	}
}
