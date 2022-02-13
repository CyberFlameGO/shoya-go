package main

import "time"

type PhotonValidateJoinJWTResponse struct {
	Valid              bool                 `json:"valid"`
	User               PhotonPropUser       `json:"user"`
	IP                 string               `json:"ip"`
	AvatarDict         PhotonPropAvatarDict `json:"avatarDict"`
	FallbackAvatarDict PhotonPropAvatarDict `json:"favatarDict"`
}

func (p *PhotonValidateJoinJWTResponse) FillFromUser(u *User) {
	avatarImageUrl := u.CurrentAvatar.GetImageUrl()
	avatarImageThumbnailUrl := u.CurrentAvatar.GetThumbnailImageUrl()
	profilePicOverride := u.ProfilePicOverride

	if profilePicOverride != "" {
		avatarImageUrl = profilePicOverride
		avatarImageThumbnailUrl = profilePicOverride
	}
	p.User = PhotonPropUser{
		ID:                             u.ID,
		DisplayName:                    u.DisplayName,
		DeveloperType:                  u.DeveloperType,
		CurrentAvatarImageUrl:          avatarImageUrl,
		CurrentAvatarThumbnailImageUrl: avatarImageThumbnailUrl,
		UserIcon:                       u.UserIcon,
		LastPlatform:                   u.LastPlatform,
		Status:                         string(u.Status),
		StatusDescription:              u.StatusDescription,
		Bio:                            u.Bio,
		Tags:                           u.Tags,
		AllowAvatarCopying:             u.AllowAvatarCopying,
	}
	currAvAuthor, err := u.CurrentAvatar.GetAuthor()
	if err != nil {
		panic("avatar author was nil") // TODO: handle this better
	}
	fbAvAuthor, err := u.FallbackAvatar.GetAuthor()
	if err != nil {
		panic("avatar author was nil") // TODO: handle this better
	}
	p.AvatarDict = PhotonPropAvatarDict{
		ID:                u.CurrentAvatar.ID,
		AssetUrl:          u.CurrentAvatar.GetAssetUrl(),
		AuthorId:          u.CurrentAvatar.AuthorID,
		AuthorName:        currAvAuthor.DisplayName,
		UpdatedAt:         time.Unix(u.CurrentAvatar.UpdatedAt, 0).Format("02-01-2006"),
		Description:       u.CurrentAvatar.Description,
		ImageUrl:          u.CurrentAvatar.GetImageUrl(),
		ThumbnailImageUrl: u.CurrentAvatar.GetThumbnailImageUrl(),
		Name:              u.CurrentAvatar.Name,
		ReleaseStatus:     string(u.CurrentAvatar.ReleaseStatus),
		Version:           u.CurrentAvatar.Version,
		Tags:              u.CurrentAvatar.Tags,
		UnityPackages:     u.CurrentAvatar.GetUnityPackages(),
	}
	p.FallbackAvatarDict = PhotonPropAvatarDict{
		ID:                u.FallbackAvatar.ID,
		AssetUrl:          u.FallbackAvatar.GetAssetUrl(),
		AuthorId:          u.FallbackAvatar.AuthorID,
		AuthorName:        fbAvAuthor.DisplayName,
		UpdatedAt:         time.Unix(u.FallbackAvatar.UpdatedAt, 0).Format("02-01-2006"),
		Description:       u.FallbackAvatar.Description,
		ImageUrl:          u.FallbackAvatar.GetImageUrl(),
		ThumbnailImageUrl: u.FallbackAvatar.GetThumbnailImageUrl(),
		Name:              u.FallbackAvatar.Name,
		ReleaseStatus:     string(u.FallbackAvatar.ReleaseStatus),
		Version:           u.FallbackAvatar.Version,
		Tags:              u.FallbackAvatar.Tags,
		UnityPackages:     u.FallbackAvatar.GetUnityPackages(),
	}
}

type PhotonPropUser struct {
	ID                             string            `json:"id"`
	DisplayName                    string            `json:"displayName"`
	DeveloperType                  string            `json:"developerType"`
	CurrentAvatarImageUrl          string            `json:"currentAvatarImageUrl"`
	CurrentAvatarThumbnailImageUrl string            `json:"currentAvatarThumbnailImageUrl"`
	UserIcon                       string            `json:"userIcon"`
	LastPlatform                   string            `json:"lastPlatform"`
	Status                         string            `json:"status"`
	StatusDescription              string            `json:"statusDescription"`
	Bio                            string            `json:"bio"`
	Tags                           []string          `json:"tags"`
	UnityPackages                  []APIUnityPackage `json:"unityPackages"`
	AllowAvatarCopying             bool              `json:"allowAvatarCopying"`
}

type PhotonPropAvatarDict struct {
	ID                string            `json:"id"`
	AssetUrl          string            `json:"assetUrl"`
	AuthorId          string            `json:"authorId"`
	AuthorName        string            `json:"authorName"`
	UpdatedAt         string            `json:"updated_at"`
	Description       string            `json:"description"`
	Featured          bool              `json:"featured"`
	ImageUrl          string            `json:"imageUrl"`
	ThumbnailImageUrl string            `json:"thumbnailImageUrl"`
	Name              string            `json:"name"`
	ReleaseStatus     string            `json:"releaseStatus"`
	Version           int               `json:"version"`
	Tags              []string          `json:"tags"`
	UnityPackages     []APIUnityPackage `json:"unityPackages"`
}
