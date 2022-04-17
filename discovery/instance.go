package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/rueian/rueidis"
	"gitlab.com/george/shoya-go/models"
)

var NotFoundErr = errors.New("instance not found")

func getInstance(id string) (*models.WorldInstance, error) {
	var i *models.WorldInstance
	err := RedisClient.Do(RedisCtx, RedisClient.B().JsonGet().Key("instances:"+id).Build()).DecodeJSON(&i)
	if err != nil {
		if rueidis.IsRedisNil(err) {
			return nil, NotFoundErr
		}
		return nil, err
	}

	return i, nil
}

func findInstancesPlayerIsIn(playerId string) ([]*models.WorldInstance, error) {
	arr, err := RedisClient.Do(RedisCtx, RedisClient.B().FtSearch().Index("instancePlayersIdx").Query(fmt.Sprintf("@players:{%s}", playerId)).Build()).ToArray()
	if err != nil {
		return nil, err
	}

	var n int64
	var p []FtSearchResult
	n, p, err = parseFtSearch(arr)
	if err != nil {
		return nil, err
	}

	r := make([]*models.WorldInstance, n)
	for idx, p := range p {
		i := &models.WorldInstance{}
		err = json.Unmarshal([]byte(p.Results["$"]), &i)
		if err != nil {
			return nil, err
		}

		r[idx] = i

	}

	return r, nil
}

// registerInstance registers a WorldInstance into Redis
func registerInstance(id, worldId, instanceType, ownerId string, capacity int) error {
	i, _ := json.Marshal(&models.WorldInstance{
		InstanceID:      id,
		WorldID:         worldId,
		InstanceType:    instanceType,
		InstanceOwnerId: ownerId,
		Capacity:        capacity,
		Players:         []string{},
		BlockedPlayers:  []models.WorldInstanceBlockedPlayers{},
	})
	return RedisClient.Do(RedisCtx, RedisClient.B().JsonSet().Key("instances:"+id).Path(".").Value(string(i)).Build()).Error()
}

// unregisterInstance removes a WorldInstance from Redis
func unregisterInstance(id string) error {
	return RedisClient.Do(RedisCtx, RedisClient.B().JsonDel().Key("instances:"+id).Build()).Error()
}

// addPlayer adds a player into a WorldInstance in Redis
func addPlayer(instanceId, playerId string) error {
	playerId = fmt.Sprintf("\"%s\"", playerId)
	err := RedisClient.Do(RedisCtx, RedisClient.B().JsonArrappend().Key("instances:"+instanceId).Path(".players").Value(playerId).Build()).Error()
	if err != nil {
		return err
	}

	err = RedisClient.Do(RedisCtx, RedisClient.B().JsonNumincrby().Key("instances:"+instanceId).Path(".playerCount.total").Value(1).Build()).Error()

	return err
}

// removePlayer removes a player from a WorldInstance in Redis
func removePlayer(instanceId, playerId string) error {
	playerId = fmt.Sprintf("\"%s\"", playerId)
	i, err := RedisClient.Do(RedisCtx, RedisClient.B().JsonArrindex().Key("instances:"+instanceId).Path(".players").Value(playerId).Build()).ToInt64()
	if err != nil {
		return err
	}

	err = RedisClient.Do(RedisCtx, RedisClient.B().JsonArrpop().Key("instances:"+instanceId).Path(".players").Index(i).Build()).Error()
	if err != nil {
		return err
	}

	err = RedisClient.Do(RedisCtx, RedisClient.B().JsonNumincrby().Key("instances:"+instanceId).Path(".playerCount.total").Value(-1).Build()).Error()

	return err
}
