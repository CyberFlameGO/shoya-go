package main

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type FavoriteGroup struct {
	*BaseModel
	UserId string         `gorm:"type:uuid;not null"`
	Name   string         `json:"name"`
	Items  []FavoriteItem `json:"-"`
}

func (u *FavoriteGroup) BeforeCreate(tx *gorm.DB) (err error) {
	u.ID = "fvgrp_" + uuid.New().String() // TODO: Possibly do a database lookup to see whether the UUID already exists.
	return
}

type FavoriteItem struct {
	*BaseModel
	FavoriteGroupId string `json:"groupId"`
	OwnerId         string `json:"ownerId"`
	ItemId          string `json:"itemId"`
}

func (u *FavoriteItem) BeforeCreate(tx *gorm.DB) (err error) {
	u.ID = "fvrt_" + uuid.New().String() // TODO: Possibly do a database lookup to see whether the UUID already exists.
	return
}
