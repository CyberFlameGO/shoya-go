package models

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type FavoriteGroupType string

var (
	FavoriteGroupTypeUser   FavoriteGroupType = "user"
	FavoriteGroupTypeWorld  FavoriteGroupType = "world"
	FavoriteGroupTypeAvatar FavoriteGroupType = "avatar"
)

type FavoriteGroup struct {
	*BaseModel
	UserID    string
	GroupType FavoriteGroupType
	Name      string         `json:"name"`
	MaxItems  int            `json:"max_items"`
	Items     []FavoriteItem `json:"-"`
}

func (f *FavoriteGroup) BeforeCreate(tx *gorm.DB) (err error) {
	f.ID = "fvgrp_" + uuid.New().String() // TODO: Possibly do a database lookup to see whether the UUID already exists.
	return
}

func (f *FavoriteGroup) AddItem(item FavoriteItem) {
	if len(f.Items) >= f.MaxItems {
		return
	}
	f.Items = append(f.Items, item)
}

func (f *FavoriteGroup) RemoveItem(item FavoriteItem) {
	for i, v := range f.Items {
		if v.ID == item.ID {
			f.Items = append(f.Items[:i], f.Items[i+1:]...)
			return
		}
	}
}

func NewFavoriteGroup(uid string, groupType FavoriteGroupType, name string, maxItems int) *FavoriteGroup {
	return &FavoriteGroup{
		UserID:    uid,
		GroupType: groupType,
		Name:      name,
		MaxItems:  maxItems,
	}
}

type FavoriteItem struct {
	*BaseModel
	FavoriteGroupId string `json:"groupId"`
	OwnerId         string `json:"ownerId"`
	ItemId          string `json:"itemId"`
}

func (f *FavoriteItem) BeforeCreate(tx *gorm.DB) (err error) {
	f.ID = "fvrt_" + uuid.New().String() // TODO: Possibly do a database lookup to see whether the UUID already exists.
	return
}

func NewFavoriteItem(groupId string, ownerId string, itemId string) *FavoriteItem {
	return &FavoriteItem{
		FavoriteGroupId: groupId,
		OwnerId:         ownerId,
		ItemId:          itemId,
	}
}
