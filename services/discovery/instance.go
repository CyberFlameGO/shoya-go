package discovery

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/rueian/rueidis"
	"gitlab.com/george/shoya-go/models"
	"time"
)

var NotFoundErr = errors.New("instance not found")

func getInstance(id string) (*models.WorldInstance, error) {
	var i *models.WorldInstance
	err := RedisClient.Do(RedisCtx, RedisClient.B().JsonGet().Key("instances:"+id).Build()).DecodeJSON(&i)
	if err != nil {
		if rueidis.IsRedisNil(err) {
			return nil, NotFoundErr
		}
		fmt.Println(err)
		return nil, err
	}

	return i, nil
}

func findInstancesForWorldId(worldId, privacy string, includeOverCapacity bool) ([]*models.WorldInstance, error) {
	var c string
	if includeOverCapacity {
		c = "(false|~true)"
	} else {
		c = "{false}"
	}
	arr, err := RedisClient.Do(RedisCtx, RedisClient.B().FtSearch().Index("instanceWorldIdIdx").Query(fmt.Sprintf("@worldId:{%s} @instanceType:{%s} @overCapacity:%s", worldId, privacy, c)).Build()).ToArray()
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	var n int64
	var p []FtSearchResult
	n, p, err = parseFtSearch(arr)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	r := make([]*models.WorldInstance, n)
	for idx, p := range p {
		i := &models.WorldInstance{}
		err = json.Unmarshal([]byte(p.Results["$"]), &i)
		if err != nil {
			fmt.Println(err)
			return nil, err
		}

		r[idx] = i
	}

	return r, nil
}

func findInstancesPlayerIsIn(playerId string) ([]*models.WorldInstance, error) {
	arr, err := RedisClient.Do(RedisCtx, RedisClient.B().FtSearch().Index("instancePlayersIdx").Query(fmt.Sprintf("@players:{%s}", escapeId(playerId))).Build()).ToArray()
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	var n int64
	var p []FtSearchResult
	n, p, err = parseFtSearch(arr)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	r := make([]*models.WorldInstance, n)
	for idx, p := range p {
		i := &models.WorldInstance{}
		err = json.Unmarshal([]byte(p.Results["$"]), &i)
		if err != nil {
			fmt.Println(err)
			return nil, err
		}

		r[idx] = i
	}

	return r, nil
}

// registerInstance registers a WorldInstance into Redis
func registerInstance(id, locationString, worldId, instanceType, ownerId string, capacity int) (*models.WorldInstance, error) {
	i := &models.WorldInstance{
		ID:              id,
		LastPing:        time.Now().Unix(),
		InstanceID:      locationString,
		WorldID:         worldId,
		InstanceType:    instanceType,
		InstanceOwnerId: ownerId,
		Capacity:        capacity,
		Players:         []string{},
		BlockedPlayers:  []models.WorldInstanceBlockedPlayers{},
	}
	j, _ := json.Marshal(i)
	err := RedisClient.Do(RedisCtx, RedisClient.B().JsonSet().Key("instances:"+id).Path(".").Value(string(j)).Build()).Error()
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	return i, nil
}

func pingInstance(instanceId string) error {
	err := RedisClient.Do(RedisCtx, RedisClient.B().JsonSet().Key("instances:"+instanceId).Path(".lastPing").Value(fmt.Sprintf("%d", time.Now().Unix())).Build()).Error()
	return err
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
		fmt.Println(err)
		return err
	}

	err = RedisClient.Do(RedisCtx, RedisClient.B().JsonNumincrby().Key("instances:"+instanceId).Path(".playerCount.total").Value(1).Build()).Error()
	if err != nil {
		fmt.Println(err)
		return err
	}

	err = RedisClient.Do(RedisCtx, RedisClient.B().JsonSet().Key("instances:"+instanceId).Path(".lastPing").Value(fmt.Sprintf("%d", time.Now().Unix())).Build()).Error()

	return err
}

// removePlayer removes a player from a WorldInstance in Redis
func removePlayer(instanceId, playerId string) error {
	playerId = fmt.Sprintf("\"%s\"", playerId)
	i, err := RedisClient.Do(RedisCtx, RedisClient.B().JsonArrindex().Key("instances:"+instanceId).Path(".players").Value(playerId).Build()).ToInt64()
	if err != nil {
		fmt.Println(err)
		return err
	}

	err = RedisClient.Do(RedisCtx, RedisClient.B().JsonArrpop().Key("instances:"+instanceId).Path(".players").Index(i).Build()).Error()
	if err != nil {
		fmt.Println(err)
		return err
	}

	err = RedisClient.Do(RedisCtx, RedisClient.B().JsonNumincrby().Key("instances:"+instanceId).Path(".playerCount.total").Value(-1).Build()).Error()
	if err != nil {
		fmt.Println(err)
		return err
	}

	err = RedisClient.Do(RedisCtx, RedisClient.B().JsonSet().Key("instances:"+instanceId).Path(".lastPing").Value(fmt.Sprintf("%d", time.Now().Unix())).Build()).Error()

	return err
}
