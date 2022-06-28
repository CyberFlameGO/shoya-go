package models

import (
	"github.com/google/uuid"
	"gitlab.com/george/shoya-go/config"
	"gorm.io/gorm"
	"strings"
	"time"
)

type ModerationType string

const (
	ModerationWarn ModerationType = "warn"
	ModerationKick ModerationType = "kick"
	ModerationBan  ModerationType = "ban"
)

type Moderation struct {
	BaseModel
	SourceID   string `gorm:"primaryKey"`
	TargetID   string `gorm:"primaryKey"`
	WorldID    string
	InstanceID string
	Type       ModerationType
	Reason     string
	ExpiresAt  int64
}

func (m *Moderation) GetAPIModeration(withDisplayname bool) *APIModeration {
	am := &APIModeration{
		ID:           m.ID,
		Type:         string(m.Type),
		TargetUserId: m.TargetID,
		Reason:       m.Reason,
		Details:      map[string]string{}, // ??
		Created:      time.Unix(m.CreatedAt, 0).UTC().Format(time.RFC3339),
		Expires:      time.Unix(m.ExpiresAt, 0).UTC().Format(time.RFC3339),
		WorldId:      m.WorldID,
		InstanceId:   m.InstanceID,
		Active:       m.ExpiresAt == 0 || m.ExpiresAt > time.Now().UTC().Unix(),
		IsPermanent:  m.ExpiresAt == 0,
		Acknowledged: false, // ??
	}

	if withDisplayname {
		u, err := m.GetTarget()
		if err != nil {
			am.TargetDisplayName = "error: could not fetch user"
		}

		am.TargetDisplayName = u.DisplayName
	}

	return am
}

func (m *Moderation) GetSource() (*User, error) {
	var u User

	tx := config.DB.Where("id = ?", m.SourceID).Find(&u)
	if tx.Error != nil {
		return nil, tx.Error
	}

	return &u, nil
}

func (m *Moderation) GetTarget() (*User, error) {
	var u User

	tx := config.DB.Where("id = ?", m.TargetID).Find(&u)
	if tx.Error != nil {
		return nil, tx.Error
	}

	return &u, nil
}

// BeforeCreate is a hook called before the database entry is created.
// It generates a UUID for the PlayerModeration.
func (m *Moderation) BeforeCreate(*gorm.DB) (err error) {
	m.ID = "mod_" + uuid.New().String() // TODO: Possibly do a database lookup to see whether the UUID already exists.
	return
}

type APIModeration struct {
	ID                string            `json:"id"`
	Type              string            `json:"type"`
	TargetUserId      string            `json:"targetUserId"`
	TargetDisplayName string            `json:"targetDisplayName"`
	Reason            string            `json:"reason"`
	Details           map[string]string `json:"details"`
	Created           string            `json:"created"`
	Expires           string            `json:"expires"`
	WorldId           string            `json:"worldId"`
	InstanceId        string            `json:"instanceId"`
	Active            bool              `json:"active"`
	IsPermanent       bool              `json:"isPermanent"`
	Acknowledged      bool              `json:"acknowledged"`
}

type PlayerModerationType string

const (
	PlayerModerationAll         PlayerModerationType = "all"
	PlayerModerationShowAvatar  PlayerModerationType = "showAvatar"
	PlayerModerationHideAvatar  PlayerModerationType = "hideAvatar"
	PlayerModerationMute        PlayerModerationType = "mute"
	PlayerModerationUnmute      PlayerModerationType = "unmute"
	PlayerModerationBlock       PlayerModerationType = "block"
	PlayerModerationUnblock     PlayerModerationType = "unblock"
	PlayerModerationInteractOn  PlayerModerationType = "interactOn"
	PlayerModerationInteractOff PlayerModerationType = "interactOff"
)

func GetPlayerModerationType(s string) PlayerModerationType {
	switch strings.ToLower(s) {
	case "showavatar":
		return PlayerModerationShowAvatar
	case "hideavatar":
		return PlayerModerationHideAvatar
	case "mute":
		return PlayerModerationMute
	case "unmute":
		return PlayerModerationUnmute
	case "block":
		return PlayerModerationBlock
	case "unblock":
		return PlayerModerationUnblock
	case "interacton":
		return PlayerModerationInteractOn
	case "interactoff":
		return PlayerModerationInteractOff
	default:
		return PlayerModerationAll
	}
}

type PlayerModeration struct {
	BaseModel
	SourceID string
	Source   User `gorm:"foreignkey:ID;references:SourceID"`
	TargetID string
	Target   User `gorm:"foreignkey:ID;references:TargetID"`
	Action   PlayerModerationType
}

// BeforeCreate is a hook called before the database entry is created.
// It generates a UUID for the PlayerModeration.
func (p *PlayerModeration) BeforeCreate(*gorm.DB) (err error) {
	p.ID = "pmod_" + uuid.New().String() // TODO: Possibly do a database lookup to see whether the UUID already exists.
	return
}

func (p *PlayerModeration) GetAPIPlayerModeration() *APIPlayerModeration {
	return &APIPlayerModeration{
		ID:                p.ID,
		CreatedAt:         time.Unix(p.CreatedAt, 0).Format(time.RFC3339),
		SourceUserID:      p.SourceID,
		SourceDisplayName: p.Source.DisplayName,
		TargetUserID:      p.TargetID,
		TargetDisplayName: p.Target.DisplayName,
		Type:              p.Action,
	}
}

type APIPlayerModeration struct {
	ID                string               `json:"id"`
	CreatedAt         string               `json:"created"`
	SourceDisplayName string               `json:"sourceDisplayName"`
	SourceUserID      string               `json:"sourceUserId"`
	TargetDisplayName string               `json:"targetDisplayName"`
	TargetUserID      string               `json:"targetUserId"`
	Type              PlayerModerationType `json:"type"`
}
