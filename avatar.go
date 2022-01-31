package main

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Avatar struct {
	BaseModel
	AuthorID    string
	Name        string
	Description string
	ImageID     string
	Image       File
}

func (a *Avatar) BeforeCreate(tx *gorm.DB) (err error) {
	a.ID = "avtr_" + uuid.New().String()
	return
}

// GetAuthor returns the author of the avatar
func (a *Avatar) GetAuthor() User {
	return User{
		BaseModel: BaseModel{
			ID: a.AuthorID,
		},
	}
}

func NewAvatar() *Avatar {
	return &Avatar{}
}

type APIAvatar struct{}
type APIAvatarWithPackages struct{}
