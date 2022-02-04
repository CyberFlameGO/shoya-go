package main

import (
	"github.com/google/uuid"
	"github.com/lib/pq"
	"gorm.io/gorm"
	"strings"
	"time"
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
	AllowAvatarCopying            bool            `json:"allowAvatarCopying"`
	Bio                           string          `json:"bio"`
	BioLinks                      pq.StringArray  `json:"bioLinks" gorm:"type:text[] NOT NULL;default: '{}'::text[]"`
	Username                      string          `json:"username"`
	DisplayName                   string          `json:"displayName"`
	DeveloperType                 string          `json:"developerType"`
	Email                         string          `json:"-"`
	PendingEmail                  string          `json:"pendingEmail"`
	EmailVerified                 bool            `json:"emailVerified"`
	Password                      string          `json:"-"`
	CurrentAvatarID               string          `json:"currentAvatarId"`
	CurrentAvatar                 Avatar          `json:"-"`
	FallbackAvatarID              string          `json:"fallbackAvatarId"`
	FallbackAvatar                Avatar          `json:"-"`
	HomeWorldID                   string          `json:"homeLocation"`
	HomeWorld                     World           `json:"-"`
	Status                        UserStatus      `json:"status"`
	StatusDescription             string          `json:"statusDescription"`
	Tags                          pq.StringArray  `json:"tags" gorm:"type:text[] NOT NULL;default: '{}'::text[]"`
	UserFavorites                 []FavoriteGroup `json:"-"`
	WorldFavorites                []FavoriteGroup `json:"-"`
	AvatarFavorites               []FavoriteGroup `json:"-"`
	LastLogin                     int64           `json:"lastLogin"`
	LastPlatform                  string          `json:"lastPlatform"`
	MfaEnabled                    bool            `json:"mfaEnabled"`
	MfaSecret                     string          `json:"-"`
	MfaRecoveryCodes              pq.StringArray  `json:"-" gorm:"type:text[] NOT NULL;default: '{}'::text[]"`
	Permissions                   []Permission    `json:"-"`
	Moderations                   []Moderation    `json:"-"`
	FriendKey                     string          `json:"-"`
	ProfilePicOverride            string          `json:"profilePicOverride"`
	Unsubscribe                   bool            `json:"unsubscribe"`
	UserIcon                      string          `json:"userIcon"`
}

func NewUser(username, displayName, email, password string) *User {
	return &User{
		AcceptedTermsOfServiceVersion: int(ApiConfiguration.CurrentTOSVersion.Get()),
		Username:                      username,
		DisplayName:                   displayName,
		Email:                         email,
		Password:                      password, // TODO: Implement password hashing. This is a placeholder.
		CurrentAvatarID:               ApiConfiguration.DefaultAvatar.Get(),
		FallbackAvatarID:              ApiConfiguration.DefaultAvatar.Get(),
		HomeWorldID:                   ApiConfiguration.HomeWorldId.Get(),
		Status:                        UserStatusActive,
	}
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

// GetState returns the state of the user from the presence service.
func (u *User) GetState() UserState {
	return UserStateActive // TODO: Implement presence service.
}

func (u *User) GetPastDisplayNames() []DisplayNameChangeRecord {
	return []DisplayNameChangeRecord{} // TODO: Implement display name history.
}

func (u *User) GetAPICurrentUser() *APICurrentUser {
	return &APICurrentUser{
		BaseModel: BaseModel{
			ID:        u.ID,
			CreatedAt: u.CreatedAt,
			UpdatedAt: u.UpdatedAt,
			DeletedAt: u.DeletedAt,
		},
		AcceptedTermsOfServiceVersion:  u.AcceptedTermsOfServiceVersion,
		AccountDeletionDate:            nil,        // TODO: Implement account deletion.
		ActiveFriends:                  []string{}, // TODO: Implement friends.
		AllowAvatarCopying:             u.AllowAvatarCopying,
		Bio:                            u.Bio,
		BioLinks:                       u.BioLinks,
		CurrentAvatarID:                u.CurrentAvatarID,
		CurrentAvatarAssetUrl:          u.CurrentAvatar.GetAssetUrl(),
		CurrentAvatarImageUrl:          u.CurrentAvatar.GetImageUrl(),
		CurrentAvatarThumbnailImageUrl: u.CurrentAvatar.GetThumbnailImageUrl(),
		DateJoined:                     time.Unix(u.CreatedAt, 0).Format("02-01-2006"),
		DeveloperType:                  u.DeveloperType,
		DisplayName:                    u.DisplayName,
		EmailVerified:                  u.EmailVerified,
		FallbackAvatarID:               u.FallbackAvatarID,
		FriendKey:                      u.FriendKey,
		Friends:                        []string{}, // TODO: Implement friends.
		HasBirthday:                    true,       // Hardcoded to true. This data won't be collected.
		HasEmail:                       u.Email != "" && u.EmailVerified,
		HasLoggedInFromClient:          true, // Hardcoded to true. Likely unnecessary.
		HomeLocationID:                 u.HomeWorldID,
		IsFriend:                       false, // TODO: Implement friends.
		LastLogin:                      u.LastLogin,
		LastPlatform:                   u.LastPlatform,
		ObfuscatedEmail:                ObfuscateEmail(u.Email),
		ObfuscatedPendingEmail:         ObfuscateEmail(u.PendingEmail),
		OfflineFriends:                 []string{}, // TODO: Implement friends.
		OnlineFriends:                  []string{}, // TODO: Implement friends.
		PastDisplayNames:               u.GetPastDisplayNames(),
		ProfilePicOverride:             u.ProfilePicOverride,
		State:                          u.GetState(),
		Status:                         u.Status,
		StatusDescription:              u.StatusDescription,
		StatusFirstTime:                false,                         // Hardcoded to false. This data won't be collected.
		StatusHistory:                  []string{u.StatusDescription}, // TODO: Implement status history.
		Tags:                           u.Tags,
		TwoFactorAuthEnabled:           u.MfaEnabled,
		Unsubscribe:                    u.Unsubscribe,
		UserIcon:                       u.UserIcon,
		Username:                       u.Username,
		FriendGroupNames:               []string{},
	}
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
type APICurrentUser struct {
	BaseModel
	AcceptedTermsOfServiceVersion  int                       `json:"acceptedTOSVersion"`
	AccountDeletionDate            *string                   `json:"accountDeletionDate"`
	ActiveFriends                  []string                  `json:"activeFriends"`
	AllowAvatarCopying             bool                      `json:"allowAvatarCopying"`
	Bio                            string                    `json:"bio"`
	BioLinks                       []string                  `json:"bioLinks"`
	CurrentAvatarID                string                    `json:"currentAvatar"`
	CurrentAvatarAssetUrl          string                    `json:"currentAvatarAssetUrl"`
	CurrentAvatarImageUrl          string                    `json:"currentAvatarImageUrl"`
	CurrentAvatarThumbnailImageUrl string                    `json:"currentAvatarThumbnailImageUrl"`
	DateJoined                     string                    `json:"date_joined"`
	DeveloperType                  string                    `json:"developerType"`
	DisplayName                    string                    `json:"displayName"`
	EmailVerified                  bool                      `json:"emailVerified"`
	FallbackAvatarID               string                    `json:"fallbackAvatar"`
	FriendGroupNames               []string                  `json:"friendGroupNames"`
	FriendKey                      string                    `json:"friendKey"`
	Friends                        []string                  `json:"friends"`
	HasBirthday                    bool                      `json:"hasBirthday"`
	HasEmail                       bool                      `json:"hasEmail"`
	HasLoggedInFromClient          bool                      `json:"hasLoggedInFromClient"`
	HomeLocationID                 string                    `json:"homeLocation"`
	IsFriend                       bool                      `json:"isFriend"`
	LastLogin                      int64                     `json:"lastLogin"`
	LastPlatform                   string                    `json:"lastPlatform"`
	ObfuscatedEmail                string                    `json:"obfuscatedEmail"`
	ObfuscatedPendingEmail         string                    `json:"obfuscatedPendingEmail"`
	OculusID                       string                    `json:"oculusId"`
	OfflineFriends                 []string                  `json:"offlineFriends"`
	OnlineFriends                  []string                  `json:"onlineFriends"`
	PastDisplayNames               []DisplayNameChangeRecord `json:"pastDisplayNames"`
	ProfilePicOverride             string                    `json:"profilePicOverride"`
	State                          UserState                 `json:"state"`
	Status                         UserStatus                `json:"status"`
	StatusDescription              string                    `json:"statusDescription"`
	StatusFirstTime                bool                      `json:"statusFirstTime"`
	StatusHistory                  []string                  `json:"statusHistory"`
	SteamDetails                   interface{}               `json:"steamDetails,omitempty"`
	SteamID                        string                    `json:"steamId,omitempty"`
	Tags                           []string                  `json:"tags"`
	TwoFactorAuthEnabled           bool                      `json:"twoFactorAuthEnabled"`
	Unsubscribe                    bool                      `json:"unsubscribe"`
	UserIcon                       string                    `json:"userIcon"`
	Username                       string                    `json:"username"`
}

type DisplayNameChangeRecord struct {
	DisplayName string `json:"displayName"`
	Timestamp   int64  `json:"updated_at"`
	Reverted    bool   `json:"reverted,omitempty"`
}

func ObfuscateEmail(email string) string {
	if len(email) == 0 {
		return ""
	}

	atIndex := strings.Index(email, "@")
	if atIndex == -1 {
		return ""
	}

	return email[0:1] + strings.Repeat("*", atIndex-1) + email[atIndex:]
}
