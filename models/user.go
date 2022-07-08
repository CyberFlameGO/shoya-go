package models

import (
	"github.com/alexedwards/argon2id"
	"github.com/google/uuid"
	"github.com/lib/pq"
	"gitlab.com/george/shoya-go/config"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"strings"
	"time"
)

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
	DeveloperType                 string          `json:"developerType" gorm:"default: 'none'"`
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
	LastPlatform                  string          `json:"last_platform"`
	MfaEnabled                    bool            `json:"mfaEnabled"`
	MfaSecret                     string          `json:"-"`
	MfaRecoveryCodes              pq.StringArray  `json:"-" gorm:"type:text[] NOT NULL;default: '{}'::text[]"`
	Permissions                   []Permission    `json:"-"`
	Moderations                   []Moderation    `json:"-" gorm:"references:ID;foreignKey:TargetID"`
	FriendKey                     string          `json:"-"`
	ProfilePicOverride            string          `json:"profilePicOverride"`
	Unsubscribe                   bool            `json:"unsubscribe"`
	UserIcon                      string          `json:"userIcon"`
}

func GetUserById(id string) (*User, error) {
	var u *User
	var err error

	if err = config.DB.Preload(clause.Associations).
		Preload("CurrentAvatar.Image").
		Preload("CurrentAvatar.Image.Versions").
		Preload("CurrentAvatar.Image.Versions.FileDescriptor").
		Preload("CurrentAvatar.Image.Versions.DeltaDescriptor").
		Preload("CurrentAvatar.Image.Versions.SignatureDescriptor").
		Preload("CurrentAvatar.UnityPackages.File").
		Preload("CurrentAvatar.UnityPackages.File.Versions").
		Preload("CurrentAvatar.UnityPackages.File.Versions.FileDescriptor").
		Preload("CurrentAvatar.UnityPackages.File.Versions.DeltaDescriptor").
		Preload("CurrentAvatar.UnityPackages.File.Versions.SignatureDescriptor").
		Preload("FallbackAvatar.Image").
		Preload("FallbackAvatar.Image.Versions").
		Preload("FallbackAvatar.Image.Versions.FileDescriptor").
		Preload("FallbackAvatar.Image.Versions.DeltaDescriptor").
		Preload("FallbackAvatar.Image.Versions.SignatureDescriptor").
		Preload("FallbackAvatar.UnityPackages.File").
		Preload("FallbackAvatar.UnityPackages.File.Versions").
		Preload("FallbackAvatar.UnityPackages.File.Versions.FileDescriptor").
		Preload("FallbackAvatar.UnityPackages.File.Versions.DeltaDescriptor").
		Preload("FallbackAvatar.UnityPackages.File.Versions.SignatureDescriptor").
		Where("id = ?", id).First(&u).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	return u, nil
}

func GetUserByUsername(username string) (*User, error) {
	var u *User
	var err error

	if err = config.DB.Preload(clause.Associations).
		Preload("CurrentAvatar.Image").
		Preload("CurrentAvatar.Image.Versions").
		Preload("CurrentAvatar.Image.Versions.FileDescriptor").
		Preload("CurrentAvatar.Image.Versions.DeltaDescriptor").
		Preload("CurrentAvatar.Image.Versions.SignatureDescriptor").
		Preload("CurrentAvatar.UnityPackages.File").
		Preload("CurrentAvatar.UnityPackages.File.Versions").
		Preload("CurrentAvatar.UnityPackages.File.Versions.FileDescriptor").
		Preload("CurrentAvatar.UnityPackages.File.Versions.DeltaDescriptor").
		Preload("CurrentAvatar.UnityPackages.File.Versions.SignatureDescriptor").
		Preload("FallbackAvatar.Image").
		Preload("FallbackAvatar.Image.Versions").
		Preload("FallbackAvatar.Image.Versions.FileDescriptor").
		Preload("FallbackAvatar.Image.Versions.DeltaDescriptor").
		Preload("FallbackAvatar.Image.Versions.SignatureDescriptor").
		Preload("FallbackAvatar.UnityPackages.File").
		Preload("FallbackAvatar.UnityPackages.File.Versions").
		Preload("FallbackAvatar.UnityPackages.File.Versions.FileDescriptor").
		Preload("FallbackAvatar.UnityPackages.File.Versions.DeltaDescriptor").
		Preload("FallbackAvatar.UnityPackages.File.Versions.SignatureDescriptor").
		Where("username = ?", username).First(&u).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	return u, nil
}

func GetUserByUsernameOrEmail(usernameOrEmail string) (*User, error) {
	var u *User
	var err error

	if err = config.DB.Preload(clause.Associations).
		Preload("CurrentAvatar.Image").
		Preload("FallbackAvatar").
		Where("username = ?", usernameOrEmail).Or("email = ?", usernameOrEmail).First(&u).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	return u, nil
}

func NewUser(username, displayName, email, password string) *User {
	pw, err := argon2id.CreateHash(password, argon2id.DefaultParams)
	if err != nil {
		panic(err) // panic when crypto fails; sounds good to me tbh.
	}

	return &User{
		AcceptedTermsOfServiceVersion: int(config.ApiConfiguration.CurrentTOSVersion.Get()),
		Username:                      strings.ToLower(username),
		DisplayName:                   displayName,
		Email:                         strings.ToLower(email),
		EmailVerified:                 true,
		Password:                      pw,
		CurrentAvatarID:               config.ApiConfiguration.DefaultAvatar.Get(),
		FallbackAvatarID:              config.ApiConfiguration.DefaultAvatar.Get(),
		HomeWorldID:                   config.ApiConfiguration.HomeWorldId.Get(),
		Status:                        UserStatusActive,
	}
}

// BeforeCreate is a hook called before the database entry is created.
// It generates a UUID for the user.
func (u *User) BeforeCreate(*gorm.DB) (err error) {
	u.ID = "usr_" + uuid.New().String() // TODO: Possibly do a database lookup to see whether the UUID already exists.
	return
}

func (u *User) CheckPassword(password string) (bool, error) {
	m, err := argon2id.ComparePasswordAndHash(password, u.Password)
	if err != nil {
		return false, err
	}

	return m, nil
}

func (u *User) ChangePassword(password string) error {
	pw, err := argon2id.CreateHash(password, argon2id.DefaultParams)
	if err != nil {
		panic(err) // panic when crypto fails; sounds good to me tbh.
	}

	u.Password = pw
	return nil
}

func (u *User) IsStaff() bool {
	for _, tag := range u.Tags {
		if tag == "admin_moderator" {
			return true
		}
	}

	return u.DeveloperType == "internal"
}

func (u *User) IsBanned() (bool, *Moderation) {
	checkTime := time.Now().UTC().Unix()
	for _, mod := range u.Moderations {
		if (mod.Type == ModerationBan) && (mod.ExpiresAt == 0 || mod.ExpiresAt > checkTime) {
			return true, &mod
		}
	}

	return false, nil
}

func (u *User) CanUploadAvatars() bool {
	for _, tag := range u.Tags {
		if tag == "system_avatar_access" || tag == "admin_avatar_access" {
			return true
		}
	}

	return false
}

func (u *User) CanUploadWorlds() bool {
	for _, tag := range u.Tags {
		if tag == "system_avatar_access" || tag == "admin_avatar_access" {
			return true
		}
	}

	return false
}

// GetState returns the state of the user from the presence service.
func (u *User) GetState() UserState { // WIP -- skipcq
	return UserStateActive // TODO: Implement presence service.
}

func (u *User) GetPastDisplayNames() []DisplayNameChangeRecord { // WIP -- skipcq
	return []DisplayNameChangeRecord{} // TODO: Implement display name history.
}

func (u *User) GetPresence() *UserPresence { // WIP -- skipcq
	return &UserPresence{
		ShouldDisclose: false,
	}
}

func (u *User) GetFriends() ([]string, error) {
	var f []FriendRequest

	if tx := config.DB.Where("from_id = ? OR to_id = ?", u.ID, u.ID).Where("state = ?", FriendRequestStateAccepted).Find(&f); tx.Error != nil {
		return nil, tx.Error
	}

	fs := make([]string, len(f))
	for idx, frq := range f {
		if frq.FromID == u.ID {
			fs[idx] = frq.ToID
			continue
		}

		fs[idx] = frq.FromID
	}

	return fs, nil
}

func (u *User) GetNotifications(notificationType NotificationType, showHidden bool, limit, offset int, after time.Time) ([]Notification, error) {
	var n []Notification

	if notificationType == NotificationTypeAll || notificationType == NotificationTypeFriendRequest {
		var frq []FriendRequest
		tx := config.DB.Preload(clause.Associations).Where("to_id = ?", u.ID).Where("state = ?", FriendRequestStateSent)

		if showHidden {
			tx = tx.Or("state = ?", FriendRequestStateIgnored)
		}

		tx.Find(&frq)
		var frDetail = "{}" // Mimic official behavior.
		for _, fr := range frq {
			n = append(n, Notification{
				Type:           NotificationTypeFriendRequest,
				Details:        frDetail,
				CreatedAt:      time.Unix(fr.CreatedAt, 0).Format(time.RFC3339),
				ID:             fr.ID,
				SenderId:       &fr.From.ID,
				SenderUsername: &fr.From.Username,
			})
		}
	}

	// TODO: The following notification types will be implemented at a later date; They will be ephemeral.
	//       They will be stored in Redis for 15 minutes.
	if notificationType == NotificationTypeAll || notificationType == NotificationTypeInvite {

	}

	if notificationType == NotificationTypeAll || notificationType == NotificationTypeInviteResponse {

	}

	if notificationType == NotificationTypeAll || notificationType == NotificationTypeRequestInvite {

	}

	if notificationType == NotificationTypeAll || notificationType == NotificationTypeRequestInviteResponse {

	}

	return n, nil
}

func (u *User) GetAPIUser(isFriend bool, shouldGetLocation bool) *APIUser {
	var friendKey = ""
	var worldId = "offline"
	var location = "offline"
	var instanceId = "offline"

	if isFriend {
		friendKey = u.FriendKey
	}

	if shouldGetLocation {
		userPresence := u.GetPresence()
		if userPresence.ShouldDisclose {
			worldId = userPresence.WorldId
			location = userPresence.Location
			instanceId = userPresence.Location
		} else {
			worldId = "private"
			location = "private"
			instanceId = "private"
		}
	}

	avatarImageUrl := u.CurrentAvatar.GetImageUrl()
	avatarImageThumbnailUrl := u.CurrentAvatar.GetThumbnailImageUrl()
	profilePicOverride := u.ProfilePicOverride

	if profilePicOverride != "" {
		avatarImageUrl = ""
		avatarImageThumbnailUrl = ""
	}

	return &APIUser{
		BaseModel: BaseModel{
			ID:        u.ID,
			CreatedAt: u.CreatedAt,
			UpdatedAt: u.UpdatedAt,
			DeletedAt: u.DeletedAt,
		},
		AllowAvatarCopying:             u.AllowAvatarCopying,
		Bio:                            u.Bio,
		BioLinks:                       u.BioLinks,
		CurrentAvatarImageUrl:          avatarImageUrl,
		CurrentAvatarThumbnailImageUrl: avatarImageThumbnailUrl,
		DateJoined:                     time.Unix(u.CreatedAt, 0).Format("02-01-2006"),
		DeveloperType:                  u.DeveloperType,
		DisplayName:                    u.DisplayName,
		FriendKey:                      friendKey,
		InstanceId:                     instanceId,
		IsFriend:                       isFriend,
		LastLogin:                      time.Unix(u.LastLogin, 0).Format(time.RFC3339),
		LastPlatform:                   Platform(u.LastPlatform),
		Location:                       location,
		ProfilePictureOverride:         profilePicOverride,
		State:                          u.GetState(),
		Status:                         u.Status,
		StatusDescription:              u.StatusDescription,
		Tags:                           u.Tags,
		UserIcon:                       u.UserIcon,
		Username:                       u.Username,
		WorldId:                        worldId,
	}
}

func (u *User) GetAPILimitedUser(isFriend bool, shouldGetLocation bool) *APILimitedUser {
	var friendKey = ""
	var location = "offline"

	if isFriend {
		friendKey = u.FriendKey
	}

	if shouldGetLocation {
		userPresence := u.GetPresence()
		if userPresence.ShouldDisclose {
			location = userPresence.Location
		} else {
			location = "private"
		}
	}

	avatarImageUrl := u.CurrentAvatar.GetImageUrl()
	avatarImageThumbnailUrl := u.CurrentAvatar.GetThumbnailImageUrl()
	profilePicOverride := u.ProfilePicOverride

	if profilePicOverride != "" {
		avatarImageUrl = ""
		avatarImageThumbnailUrl = ""
	}

	return &APILimitedUser{
		BaseModel: BaseModel{
			ID:        u.ID,
			CreatedAt: u.CreatedAt,
			UpdatedAt: u.UpdatedAt,
			DeletedAt: u.DeletedAt,
		},
		Bio:                            u.Bio,
		CurrentAvatarImageUrl:          avatarImageUrl,
		CurrentAvatarThumbnailImageUrl: avatarImageThumbnailUrl,
		DeveloperType:                  u.DeveloperType,
		DisplayName:                    u.DisplayName,
		FallbackAvatarId:               u.FallbackAvatarID,
		IsFriend:                       isFriend,
		LastPlatform:                   Platform(u.LastPlatform),
		ProfilePictureOverride:         profilePicOverride,
		Status:                         u.Status,
		StatusDescription:              u.StatusDescription,
		Tags:                           u.Tags,
		Location:                       &location,
		FriendKey:                      &friendKey,
	}
}

func (u *User) GetAPICurrentUser() *APICurrentUser {
	avatarImageUrl := u.CurrentAvatar.GetImageUrl()
	avatarImageThumbnailUrl := u.CurrentAvatar.GetThumbnailImageUrl()
	profilePicOverride := u.ProfilePicOverride

	if profilePicOverride != "" {
		avatarImageUrl = ""
		avatarImageThumbnailUrl = ""
	}

	//Slightly differs from the real api. For some reason, bio links will show empty strings at the same location they are set on first update
	//so setting ["https://youtube.com", "", "https://twitter.com"] will show ["https://youtube.com", "", "https://twitter.com"]
	//but when you refresh the page, it will show ["https://youtube.com", "https://twitter.com", ""]
	//We will sort them immediately
	tempBioLinks := make([]string, 0)
	for _, link := range u.BioLinks {
		if link != "" {
			tempBioLinks = append(tempBioLinks, link)
		}
	}

	//api seems to always return 3, even when empty is present
	for len(tempBioLinks) != 3 {
		tempBioLinks = append(tempBioLinks, "")
	}

	u.BioLinks = tempBioLinks

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
		CurrentAvatarImageUrl:          avatarImageUrl,
		CurrentAvatarThumbnailImageUrl: avatarImageThumbnailUrl,
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
		ProfilePicOverride:             profilePicOverride,
		State:                          u.GetState(),
		Status:                         u.Status,
		StatusDescription:              u.StatusDescription,
		StatusFirstTime:                false, // Hardcoded to false. This data won't be collected.
		StatusHistory:                  []string{u.StatusDescription},
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
	BaseModel
	AllowAvatarCopying             bool       `json:"allowAvatarCopying"`
	Bio                            string     `json:"bio"`
	BioLinks                       []string   `json:"bioLinks"`
	CurrentAvatarImageUrl          string     `json:"currentAvatarImageUrl"`
	CurrentAvatarThumbnailImageUrl string     `json:"currentAvatarThumbnailImageUrl"`
	DateJoined                     string     `json:"date_joined"`
	DeveloperType                  string     `json:"developerType"`
	DisplayName                    string     `json:"displayName"`
	FriendKey                      string     `json:"friendKey,omitempty"`
	InstanceId                     string     `json:"instanceId"`
	IsFriend                       bool       `json:"isFriend"`
	LastLogin                      string     `json:"last_login"`
	LastActivity                   string     `json:"last_activity"`
	LastPlatform                   Platform   `json:"last_platform"`
	Location                       string     `json:"location"`
	ProfilePictureOverride         string     `json:"profilePicOverride"`
	State                          UserState  `json:"state"`
	Status                         UserStatus `json:"status"`
	StatusDescription              string     `json:"statusDescription"`
	Tags                           []string   `json:"tags"`
	UserIcon                       string     `json:"userIcon"`
	Username                       string     `json:"username"`
	WorldId                        string     `json:"worldId"`
	FriendRequestStatus            string     `json:"friendRequestStatus"` // Requires implementation of friendship

	// The following have not been implemented so far, and they seem to have undocumented behavior on official.
	// They have been seen as an empty string (""), "private", or "offline", but not once as what they describe.
	// Additionally, the implementation of them requires the presence service.
	TravelingToInstance string `json:"travelingToInstance"`
	TravelingToLocation string `json:"travelingToLocation"`
	TravelingToWorld    string `json:"travelingToWorld"`
}
type APILimitedUser struct {
	BaseModel
	Bio                            string     `json:"bio"`
	CurrentAvatarImageUrl          string     `json:"currentAvatarImageUrl"`
	CurrentAvatarThumbnailImageUrl string     `json:"currentAvatarThumbnailImageUrl"`
	DeveloperType                  string     `json:"developerType"`
	DisplayName                    string     `json:"displayName"`
	FallbackAvatarId               string     `json:"fallbackAvatar"`
	IsFriend                       bool       `json:"isFriend"`
	LastPlatform                   Platform   `json:"last_platform"`
	ProfilePictureOverride         string     `json:"profilePicOverride"`
	Status                         UserStatus `json:"status"`
	StatusDescription              string     `json:"statusDescription"`
	Tags                           []string   `json:"tags"`
	UserIcon                       string     `json:"userIcon"`
	Username                       string     `json:"username"`
	Location                       *string    `json:"location,omitempty"`
	FriendKey                      *string    `json:"friendKey,omitempty"`
}
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
	LastPlatform                   string                    `json:"last_platform"`
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

type UserPresence struct {
	ShouldDisclose bool
	WorldId        string
	Location       string
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
