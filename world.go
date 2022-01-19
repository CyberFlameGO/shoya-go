package main

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type World struct {
	BaseModel
	AuthorID    string `json:"authorId"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

func (w *World) BeforeCreate(tx *gorm.DB) (err error) {
	w.ID = "wrld_" + uuid.New().String()
	return
}

// GetAuthor returns the author of the world
func (w *World) GetAuthor() User {
	return User{
		BaseModel: BaseModel{
			ID: w.AuthorID,
		},
	}
}

type APIWorld struct{}
type APIWorldWithPackages struct{}
