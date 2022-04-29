package models

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
	"time"
)

type Moderation struct {
	BaseModel
	UserID string
}

type PlayerModerationType string

var (
	PlayerModerationShowAvatar  PlayerModerationType = "showAvatar"
	PlayerModerationHideAvatar  PlayerModerationType = "hideAvatar"
	PlayerModerationMute        PlayerModerationType = "mute"
	PlayerModerationUnmute      PlayerModerationType = "unmute"
	PlayerModerationBlock       PlayerModerationType = "block"
	PlayerModerationUnblock     PlayerModerationType = "unblock"
	PlayerModerationInteractOn  PlayerModerationType = "interactOn"
	PlayerModerationInteractOff PlayerModerationType = "interactOff"
)

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
