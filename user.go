package main

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// UserState represents the activity state of a user.
type UserState string

var (
	UserStateOffline UserState = "offline"
	UserStateActive  UserState = "active"
	UserStateOnline  UserState = "online"
)

// UserStatus is the status of a user. It can be offline, active, join me, ask me, or busy.
type UserStatus string

var (
	UserStatusOffline UserStatus = "offline"
	UserStatusActive  UserStatus = "active"
	UserStatusJoinMe  UserStatus = "join me"
	UserStatusAskMe   UserStatus = "ask me"
	UserStatusBusy    UserStatus = "busy"
)

func NewUserStatus(s string) UserStatus {
	switch s {
	case "offline":
		return UserStatusOffline
	case "active":
		return UserStatusActive
	case "join me":
		return UserStatusJoinMe
	case "ask me":
		return UserStatusAskMe
	case "busy":
		return UserStatusBusy
	default:
		return UserStatusActive
	}
}

func (s UserStatus) String() string {
	return string(s)
}

// User is a user of the application.
type User struct {
	BaseModel
	AcceptedTermsOfServiceVersion int             `json:"acceptedTOSVersion"`
	Username                      string          `json:"username"`
	DisplayName                   string          `json:"displayName"`
	Email                         string          `json:"-"`
	Password                      string          `json:"-"`
	CurrentAvatarId               string          `json:"currentAvatarId"`
	CurrentAvatar                 Avatar          `json:"-"`
	FallbackAvatarId              string          `json:"fallbackAvatarId"`
	FallbackAvatar                Avatar          `json:"-"`
	HomeWorldId                   string          `json:"homeLocation"`
	HomeWorld                     World           `json:"-"`
	Status                        UserStatus      `json:"status"`
	StatusDescription             string          `json:"statusDescription"`
	Tags                          []string        `json:"tags" gorm:"type:text[]"`
	UserFavorites                 []FavoriteGroup `json:"-"`
	WorldFavorites                []FavoriteGroup `json:"-"`
	AvatarFavorites               []FavoriteGroup `json:"-"`
	LastLogin                     int64           `json:"lastLogin"`
	LastPlatform                  string          `json:"lastPlatform"`
	MfaEnabled                    bool            `json:"mfaEnabled"`
	MfaSecret                     string          `json:"-"`
	MfaRecoveryCodes              []string        `json:"-" gorm:"type:text[]"`
	Permissions                   []Permission    `json:"-"`
	Moderations                   []Moderation    `json:"-"`
}

// BeforeCreate is a hook called before the database entry is created.
// It generates a UUID for the user.
func (u *User) BeforeCreate(tx *gorm.DB) (err error) {
	u.ID = "usr_" + uuid.New().String() // TODO: Possibly do a database lookup to see whether the UUID already exists.
	return
}

func (u *User) CheckPassword(password string) bool {
	return u.Password == password // TODO: Implement password hashing. This is a placeholder.
}

// APIUser is a data structure used for API responses as well as fetching relevant data from the database.
type APIUser struct {
	ID          string     `json:"id"`
	Username    string     `json:"username"`
	DisplayName string     `json:"displayName"`
	State       UserState  `gorm:"-" json:"state"`
	Status      UserStatus `json:"status"`
}
type APILimitedUser struct{}
type APICurrentUser struct{}
