package models

import (
	"errors"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"strings"
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

func (u *WorldUnityPackage) BeforeCreate(*gorm.DB) (err error) {
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
		CreatedAt:       time.Unix(u.CreatedAt, 0).UTC().Format(time.RFC3339Nano),
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

func (u *AvatarUnityPackage) BeforeCreate(*gorm.DB) (err error) {
	u.ID = "unp_" + uuid.New().String()
	return
}

func (u *AvatarUnityPackage) GetAPIUnityPackage() *APIUnityPackage {
	return &APIUnityPackage{
		ID:              u.ID,
		CreatedAt:       time.Unix(u.CreatedAt, 0).UTC().Format(time.RFC3339Nano),
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

type Location struct {
	WorldID          string `json:"worldId"`          // WorldID is the id of the world the instance is for.
	InstanceID       string `json:"instanceId"`       // InstanceID is the instance's identifier (usually 5 numbers).
	LocationString   string `json:"locationString"`   // LocationString is the string minus the world id prefix.
	InstanceType     string `json:"instanceType"`     // InstanceType is the instance's privacy setting. | Valid settings: public, hidden, friends, private.
	OwnerID          string `json:"ownerId"`          // OwnerID is the id of the instance's creator.
	Nonce            string `json:"nonce"`            // Nonce is a "shared key" for use in non-public instances.
	Region           string `json:"region"`           // Region is the Photon region (us, or blank == usw photon) | Valid regions: us, use, eu, jp.
	CanRequestInvite bool   `json:"canRequestInvite"` // CanRequestInvite turns an instance of InstanceType: private (invite-only) to an invite+.
	IsStrict         bool   `json:"strict"`           // IsStrict ensures that the instance is only joinable if the user is friends with the creator.
}

// parseLocationString parses the location string provided in a request.
func parseLocationString(s string) (*Location, error) {
	var location = Location{}
	s1 := strings.Split(s, ":")
	if len(s1) < 2 {
		return nil, errors.New("invalid instance id")
	}

	location.WorldID = s1[0]        // wrld_{uuid}
	location.LocationString = s1[1] // 00000~xxx

	/**
	TODO: Implement regexes for further matching (non-globals!)
		Instance ID: `(?P<instanceId>.*?)~`
		Instance privacy & ownership: `(?P<privacy>hidden|friends|private)\((?P<ownerId>.*?)\)` (No match == public)
		Region: `region\((?P<region>.*?)\)`
		Nonce: `nonce\((?P<nonce>.*?)\)`
	*/

	s2 := strings.Split(s1[1], "~")
	location.InstanceID = s2[0] // 00000

	switch location.InstanceType {
	case "private":
		if strings.Contains(s1[1], "~canRequestInvite") {
			location.CanRequestInvite = true // Invite+
		}
		if strings.Contains(s1[1], "~strict") {
			location.IsStrict = true
		}
		break
	case "friends":
		if strings.Contains(s1[1], "~strict") {
			location.IsStrict = true
		}
	}

	return &location, nil
}
