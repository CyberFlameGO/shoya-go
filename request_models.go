package main

import (
	"gorm.io/gorm"
	"strings"
)

// RegisterRequest is the model for requests sent to /auth/register.
type RegisterRequest struct {
	AcceptedTOSVersion int    `json:"acceptedTOSVersion"`
	Username           string `json:"username"`
	Password           string `json:"password"`
	Email              string `json:"email"`
	Day                string `json:"day"`
	Month              string `json:"month"`
	Year               string `json:"year"`
	RecaptchaCode      string `json:"recaptchaCode"`
}

type UpdateUserRequest struct {
	AcceptedTOSVersion     int      `json:"acceptedTOSVersion"`
	Bio                    string   `json:"bio"`
	BioLinks               []string `json:"bioLinks"`
	Birthday               string   `json:"birthday"`
	CurrentPassword        string   `json:"currentPassword"`
	DisplayName            string   `json:"displayName"`
	Email                  string   `json:"email"`
	Password               string   `json:"password"`
	ProfilePictureOverride string   `json:"profilePicOverride"`
	Status                 string   `json:"status"`
	StatusDescription      string   `json:"statusDescription"`
	Tags                   []string `json:"tags"`
	Unsubscribe            bool     `json:"unsubscribe"`
	UserIcon               string   `json:"userIcon"`
	HomeLocation           string   `json:"homeLocation"`
}

var ValidLanguageTags = []string{"eng", "kor", "rus", "spa", "por", "zho", "deu", "jpn", "fra", "swe", "nld", "pol", "dan", "nor", "ita", "tha", "fin", "hun", "ces", "tur", "ara", "ron", "vie", "ukr", "ase", "bfi", "dse", "fsl", "kvk"}

func (r *UpdateUserRequest) EmailChecks(u *User) (bool, error) {
	if r.Email == "" {
		return false, nil
	}

	pwdMatch, err := u.CheckPassword(r.CurrentPassword)
	if !pwdMatch || err != nil {
		return false, invalidCredentialsErrorInUserUpdate
	}

	if DB.Model(&User{}).Where("email = ?", r.Email).Or("pending_email = ?", r.Email).Error != gorm.ErrRecordNotFound {
		return false, userWithEmailAlreadyExistsErrorInUserUpdate
	}

	u.PendingEmail = r.Email
	// TODO: Queue up verification email send
	return true, nil
}

func (r *UpdateUserRequest) StatusChecks(u *User) (bool, error) {
	var status UserStatus
	if r.Status == "" {
		return false, nil
	}

	switch strings.ToLower(r.Status) {
	case "join me":
		status = UserStatus(strings.ToLower(r.Status))
	case "active":
		status = UserStatus(strings.ToLower(r.Status))
	case "ask me":
		status = UserStatus(strings.ToLower(r.Status))
	case "busy":
		status = UserStatus(strings.ToLower(r.Status))
	case "offline":
		if !u.IsStaff() {
			return false, invalidStatusDescriptionErrorInUserUpdate
		}
		status = UserStatus(strings.ToLower(r.Status))
	default:
		return false, invalidUserStatusErrorInUserUpdate
	}

	u.Status = status
	return true, nil
}

func (r *UpdateUserRequest) StatusDescriptionChecks(u *User) (bool, error) {
	if r.StatusDescription == "" {
		return false, nil
	}

	if len(r.StatusDescription) > 32 {
		return false, invalidStatusDescriptionErrorInUserUpdate
	}

	u.StatusDescription = r.StatusDescription
	return true, nil
}

func (r *UpdateUserRequest) BioChecks(u *User) (bool, error) {
	if r.Bio == "" {
		return false, nil
	}

	if len(r.Bio) > 512 {
		return false, invalidBioErrorInUserUpdate
	}

	u.Bio = r.Bio
	return true, nil
}

func (r *UpdateUserRequest) UserIconChecks(u *User) (bool, error) {
	if r.UserIcon == "" {
		return false, nil
	}

	if !u.IsStaff() {
		return false, triedToSetUserIconWithoutBeingStaffErrorInUserUpdate
	}

	u.UserIcon = r.UserIcon
	return true, nil
}

func (r *UpdateUserRequest) ProfilePicOverrideChecks(u *User) (bool, error) {
	if r.ProfilePictureOverride == "" {
		return false, nil
	}

	if !u.IsStaff() {
		return false, triedToSetProfilePicOverrideWithoutBeingStaffErrorInUserUpdate
	}

	u.ProfilePicOverride = r.ProfilePictureOverride
	return true, nil
}

func (r *UpdateUserRequest) TagsChecks(u *User) (bool, error) {
	if len(r.Tags) == 0 {
		return false, nil
	}

	i := 0
	var tagsThatWillApply []string
	for _, tag := range r.Tags {
		if !strings.HasPrefix(tag, "language_") && !u.IsStaff() {
			continue
		} else if strings.HasPrefix("language_", tag) {
			if !isValidLanguageTag(tag) {
				return false, invalidLanguageTagInUserUpdate
			}
			// Ensure that we do not add more that a total of 3 language tags to the user.
			if i++; i > 3 {
				return false, tooManyLanguageTagsInUserUpdate
			}
		}

		tagsThatWillApply = append(tagsThatWillApply, tag)
	}

	for _, tag := range u.Tags {
		if strings.HasPrefix(tag, "system_") || strings.HasPrefix(tag, "admin_") {
			tagsThatWillApply = append(tagsThatWillApply, tag)
		}
	}

	u.Tags = tagsThatWillApply
	return true, nil
}

func (r *UpdateUserRequest) HomeLocationChecks(u *User) (bool, error) {
	if r.HomeLocation == "" {
		return false, nil
	}

	var w World
	tx := DB.Model(&World{}).Where("id = ?", r.HomeLocation).Find(&w)
	if tx.Error != nil {
		if tx.Error == gorm.ErrRecordNotFound {
			return false, worldNotFoundErrorInUserUpdate
		}

		return false, nil
	}

	if w.ReleaseStatus == ReleaseStatusPrivate && (w.AuthorID != u.ID && !u.IsStaff()) {
		return false, worldIsPrivateAndNotOwnedByUser
	}

	u.HomeWorldID = w.ID
	return true, nil
}

func isValidLanguageTag(tag string) bool {
	if !strings.HasPrefix("language_", tag) {
		return false
	}

	split := strings.Split(tag, "_")
	if len(split) != 2 {
		return false
	}

	tagPart := split[1]
	for _, validTag := range ValidLanguageTags {
		if tagPart == validTag {
			return true
		}
	}

	return false
}
