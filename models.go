package main

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
	"time"
)

// BaseModel is the model used by all Gorm models.
type BaseModel struct {
	ID        string         `gorm:"primarykey" json:"id"`
	CreatedAt int64          `json:"-"`
	UpdatedAt int64          `json:"-"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// Platform is the supported platform for an asset.
// PlatformWindows ("standalonewindows"): The asset has a build for Windows versions of the game.
// PlatformAndroid ("android"): The asset has a build for Android (Quest) versions of the game.
type Platform string

var (
	PlatformWindows Platform = "standalonewindows"
	PlatformAndroid Platform = "android"
)

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

type WorldUnityPackage struct {
	BaseModel
	BelongsToAssetID string
	FileID           string
	File             File     `json:"-"`
	Version          int      `json:"assetVersion"`
	Platform         Platform `json:"platform"`
	UnityVersion     string   `json:"unityVersion"`
	UnitySortNumber  int      `json:"unitySortNumber"`
}

func (u *WorldUnityPackage) BeforeCreate(tx *gorm.DB) (err error) {
	u.ID = "unp_" + uuid.New().String()
	return
}

func (u *WorldUnityPackage) GetAPIUnityPackage(withAssetUrl bool) *APIUnityPackage {
	var assetUrl = ""
	if withAssetUrl {
		assetUrl = u.File.Url
	}
	return &APIUnityPackage{
		ID:              u.ID,
		CreatedAt:       time.Unix(u.CreatedAt, 0).Format("02-01-2006"),
		AssetUrl:        assetUrl,
		Platform:        u.Platform,
		UnityVersion:    u.UnityVersion,
		UnitySortNumber: u.UnitySortNumber,
	}
}

type AvatarUnityPackage struct {
	BaseModel
	BelongsToAssetID string
	FileID           string   `json:"-"`
	File             File     `json:"-"`
	Version          int      `json:"assetVersion"`
	Platform         Platform `json:"platform"`
	UnityVersion     string   `json:"unityVersion"`
	UnitySortNumber  int      `json:"unitySortNumber"`
}

func (u *AvatarUnityPackage) BeforeCreate(tx *gorm.DB) (err error) {
	u.ID = "unp_" + uuid.New().String()
	return
}

func (u *AvatarUnityPackage) GetAPIUnityPackage() *APIUnityPackage {
	return &APIUnityPackage{
		ID:              u.ID,
		CreatedAt:       time.Unix(u.CreatedAt, 0).Format("02-01-2006"),
		AssetUrl:        u.File.Url,
		Platform:        u.Platform,
		UnityVersion:    u.UnityVersion,
		UnitySortNumber: u.UnitySortNumber,
	}
}

type APIUnityPackage struct {
	ID              string      `json:"id"`
	AssetUrl        string      `json:"assetUrl"`
	AssetUrlObject  interface{} `json:"assetUrlObject"`
	CreatedAt       string      `json:"created_at"`
	Platform        Platform    `json:"platform"`
	PluginUrl       string      `json:"pluginUrl"`
	PluginUrlObject interface{} `json:"pluginUrlObject"`
	UnityVersion    string      `json:"unityVersion"`
	UnitySortNumber int         `json:"unitySortNumber"`
}
