package models

import (
	"github.com/google/uuid"
	"github.com/lib/pq"
	"gitlab.com/george/shoya-go/config"
	"gorm.io/gorm"
	"time"
)

type Avatar struct {
	BaseModel
	AuthorID      string
	Name          string
	Description   string
	ImageID       string
	Image         File
	ReleaseStatus ReleaseStatus
	Tags          pq.StringArray       `json:"tags" gorm:"type:text[] NOT NULL;default: '{}'::text[]"`
	Version       int                  `json:"version" gorm:"type:bigint NOT NULL;default:0"`
	UnityPackages []AvatarUnityPackage `gorm:"foreignKey:BelongsToAssetID"`
}

func (a *Avatar) BeforeCreate(tx *gorm.DB) (err error) {
	a.ID = "avtr_" + uuid.New().String()
	return
}

// GetAuthor returns the author of the avatar
func (a *Avatar) GetAuthor() (*User, error) {
	var u User

	tx := config.DB.Where("id = ?", a.AuthorID).Find(&u)
	if tx.Error != nil {
		return nil, tx.Error
	}

	return &u, nil
}

func (a *Avatar) GetAssetUrl() string {
	var assetUrl string
	maxVersion := 0
	for _, pkg := range a.UnityPackages {
		if pkg.Version >= maxVersion {
			assetUrl = pkg.File.Url
		}
	}

	return assetUrl
}

func (a *Avatar) GetUnityPackages() []APIUnityPackage {
	var pkgs []APIUnityPackage
	for _, pkg := range a.UnityPackages {
		pkgs = append(pkgs, *pkg.GetAPIUnityPackage())
	}

	return pkgs
}

func (a *Avatar) GetImageUrl() string {
	return a.Image.Url
}

func (a *Avatar) GetThumbnailImageUrl() string {
	return a.Image.Url // TODO: Thumbnail service?
}

func (a *Avatar) GetAPIAvatar() (*APIAvatar, error) {
	au, err := a.GetAuthor()
	if err != nil {
		return nil, err
	}

	return &APIAvatar{
		ID:                a.ID,
		AuthorID:          a.AuthorID,
		AuthorName:        au.DisplayName,
		CreatedAt:         time.Unix(a.CreatedAt, 0).UTC().Format(time.RFC3339Nano),
		Description:       a.Description,
		Featured:          false,
		ImageUrl:          a.GetImageUrl(),
		Name:              a.Name,
		ReleaseStatus:     a.ReleaseStatus,
		Tags:              a.Tags,
		ThumbnailImageUrl: a.GetThumbnailImageUrl(),
		Version:           a.Version,
	}, nil
}
func (a *Avatar) GetAPIAvatarWithPackages() (*APIAvatarWithPackages, error) {
	aa, err := a.GetAPIAvatar()
	if err != nil {
		return nil, err
	}
	return &APIAvatarWithPackages{
		APIAvatar:     *aa,
		AssetUrl:      a.GetAssetUrl(),
		UnityPackages: a.GetUnityPackages(),
	}, nil
}

type APIAvatar struct {
	ID                string        `json:"id"`
	AuthorID          string        `json:"authorId"`
	AuthorName        string        `json:"authorName"`
	CreatedAt         string        `json:"created_at"`
	Description       string        `json:"description"`
	Featured          bool          `json:"featured"`
	ImageUrl          string        `json:"imageUrl"`
	Name              string        `json:"name"`
	ReleaseStatus     ReleaseStatus `json:"releaseStatus"`
	Tags              []string      `json:"tags"`
	ThumbnailImageUrl string        `json:"thumbnailImageUrl"`
	Version           int           `json:"version"`
}
type APIAvatarWithPackages struct {
	APIAvatar
	AssetUrl              string            `json:"assetUrl"`
	AssetUrlObject        interface{}       `json:"assetUrlObject"` // Always an empty object.
	UnityPackages         []APIUnityPackage `json:"unityPackages"`
	UnityPackageUrlObject interface{}       `json:"unityPackageUrlObject"` // Always an empty object.
}
