package main

import (
	"gorm.io/gorm"
)

// BaseModel is the model used by all Gorm models.
type BaseModel struct {
	ID        string         `gorm:"primarykey" json:"id"`
	CreatedAt int64          `json:"-"`
	UpdatedAt int64          `json:"-"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// ReleaseStatus is the state an asset is released in.
// ReleaseStatusPublic ("public"):
//    For worlds, this means that anyone can search for it, create instances, and favorite it.
//    For avatars, this means that anyone can use it.
// ReleaseStatusPrivate ("private"):
//    For worlds, this means that only people who have the ID can create instances. It cannot be favorited.
//    For avatars, this means that only the avatar author can use it.
// ReleaseStatusHidden ("hidden"): This is the status that deleted content goes into.
type ReleaseStatus string

var (
	ReleaseStatusPublic  ReleaseStatus = "public"
	ReleaseStatusPrivate ReleaseStatus = "private"
	ReleaseStatusHidden  ReleaseStatus = "hidden"
)

type UnityPackage struct{}
