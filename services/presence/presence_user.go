package presence

import (
	"encoding/json"
	"fmt"
	"github.com/rueian/rueidis"
	"gitlab.com/george/shoya-go/models"
	"gitlab.com/george/shoya-go/models/service_types"
	"time"
)

func getPresenceForUser(id string) (*service_types.UserPresence, error) {
	var tx rueidis.RedisResult
	var p *service_types.UserPresence
	var err error
	if tx = RedisClient.Do(RedisCtx, RedisClient.B().JsonGet().Key("presence:"+id).Build()); tx.RedisError() != nil {
		if tx.RedisError().IsNil() {
			p, err = createDefaultPresenceForUser(id)
			if err != nil {
				return nil, err
			}
		} else {
			return nil, tx.Error()
		}
	} else {
		err = tx.DecodeJSON(&p)
		if err != nil {
			return nil, err
		}
	}

	// While the following are saved as a side effect of JSON serialization,
	// it's easier to compute them on delivery, than to store them individually
	// every time there's an update.
	if p.State == service_types.UserStateOffline || p.Status == service_types.UserStatusOffline {
		p.IsOnline = false
		p.State = service_types.UserStateOffline
	} else {
		p.IsOnline = true
	}

	if p.Status == service_types.UserStatusBusy || p.Status == service_types.UserStatusAskMe || p.Status == service_types.UserStatusOffline {
		p.ShouldDisclose = false
	} else if i, err := models.ParseLocationString(p.Location); err == nil {
		switch models.InstanceType(i.InstanceType) {
		case models.InstanceTypePublic:
			p.ShouldDisclose = true
		case models.InstanceTypeHidden:
			p.ShouldDisclose = true
		case models.InstanceTypePrivate:
			p.CanRequestInvite = i.CanRequestInvite
			p.ShouldDisclose = false
		default:
			fmt.Printf("Instance type %+v is unknown\n", i.InstanceType)
			p.ShouldDisclose = false
		}
	} else {
		p.ShouldDisclose = false // If we had an error, we should hide it, just in case.
	}

	return p, nil
}

func updateStatusForUser(userId string, status service_types.UserStatus) (*service_types.UserPresence, error) {
	p, err := getPresenceForUser(userId)
	p.Status = status
	if err = RedisClient.Do(RedisCtx, RedisClient.B().JsonSet().Key("presence:"+userId).Path("$.status").Value(fmt.Sprintf("\"%s\"", status)).Build()).Error(); err != nil {
		return nil, err
	}

	if err = bumpTTLForUser(userId); err != nil {
		return nil, err
	}

	return p, nil
}

func updateStateForUser(userId string, state service_types.UserState) (*service_types.UserPresence, error) {
	p, err := getPresenceForUser(userId)
	p.State = state
	if err = RedisClient.Do(RedisCtx, RedisClient.B().JsonSet().Key("presence:"+userId).Path("$.state").Value(fmt.Sprintf("\"%s\"", state)).Build()).Error(); err != nil {
		return nil, err
	}

	if err = bumpTTLForUser(userId); err != nil {
		return nil, err
	}

	return p, nil
}

func updateLastSeenForUser(userId string, lastSeen time.Time) (*service_types.UserPresence, error) {
	p, err := getPresenceForUser(userId)
	if err != nil {
		return nil, err
	}

	p.LastSeen = lastSeen.Unix()

	if tx := RedisClient.Do(RedisCtx, RedisClient.B().JsonSet().Key("presence:"+userId).Path("$.lastSeen").Value(fmt.Sprintf("%d", lastSeen.Unix())).Build()); tx.Error() != nil {
		return nil, err
	}

	if err = bumpTTLForUser(userId); err != nil {
		return nil, err
	}

	return p, nil
}

func updateInstanceForUser(userId, instanceId string) (*service_types.UserPresence, error) {
	p, err := getPresenceForUser(userId)
	if err != nil {
		return nil, err
	}

	i, err := models.ParseLocationString(instanceId)
	if err != nil {
		return nil, err
	}

	p.Location = instanceId
	p.WorldId = i.WorldID

	if err != nil {
		return nil, err
	}
	if err = RedisClient.Do(RedisCtx, RedisClient.B().JsonSet().Key("presence:"+userId).Path("$.location").Value(fmt.Sprintf("\"%s\"", p.Location)).Build()).Error(); err != nil {
		return nil, err
	}

	if err = RedisClient.Do(RedisCtx, RedisClient.B().JsonSet().Key("presence:"+userId).Path("$.worldId").Value(fmt.Sprintf("\"%s\"", p.WorldId)).Build()).Error(); err != nil {
		return nil, err
	}

	if err = bumpTTLForUser(userId); err != nil {
		return nil, err
	}

	return p, nil
}

func bumpTTLForUser(userId string) error {
	// Update TTL 5m
	return RedisClient.Do(RedisCtx, RedisClient.B().Expire().Key("presence:"+userId).Seconds(int64(time.Duration(time.Minute*5).Seconds())).Build()).Error()
}

func createDefaultPresenceForUser(id string) (*service_types.UserPresence, error) {
	u, err := models.GetUserById(id)
	if err != nil {
		return nil, err
	}
	p := service_types.UserPresence{
		PresenceCreatedAt: time.Now().Unix(),
		State:             service_types.UserStateOffline,
		Status:            u.Status,
		LastSeen:          u.LastLogin,
		Location:          "",
		WorldId:           "",
	}

	m, err := json.Marshal(p)
	if err != nil {
		return nil, err
	}

	if err = RedisClient.Do(RedisCtx, RedisClient.B().JsonSet().Key("presence:"+id).Path(".").Value(string(m)).Build()).Error(); err != nil {
		return nil, err
	}

	if err = bumpTTLForUser(id); err != nil {
		return nil, err
	}

	return &p, nil
}
