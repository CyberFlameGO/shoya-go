package service_types

type UserPresence struct {
	PresenceCreatedAt int64      `json:"_createdAt"`
	IsOnline          bool       `json:"isOnline"`
	ShouldDisclose    bool       `json:"shouldDisclose"`
	CanRequestInvite  bool       `json:"canRequestInvite"`
	Status            UserStatus `json:"status"`
	State             UserState  `json:"state"`
	LastSeen          int64      `json:"lastSeen"`
	WorldId           string     `json:"worldId"`
	Location          string     `json:"location"`
}

// UserState represents the activity state of a user.
type UserState string

const (
	UserStateOffline UserState = "offline"
	UserStateActive  UserState = "active"
	UserStateOnline  UserState = "online"
)

// UserStatus is the status of a user. It can be offline, active, join me, ask me, or busy.
type UserStatus string

const (
	UserStatusOffline UserStatus = "offline"
	UserStatusActive  UserStatus = "active"
	UserStatusJoinMe  UserStatus = "join me"
	UserStatusAskMe   UserStatus = "ask me"
	UserStatusBusy    UserStatus = "busy"
)

func (s UserStatus) String() string {
	return string(s)
}
