package main

import (
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/lib/pq"
	"gitlab.com/george/shoya-go/config"
	"gitlab.com/george/shoya-go/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"strconv"
	"strings"
)

func avatarsRoutes(router *fiber.App) {
	avatars := router.Group("/avatars", ApiKeyMiddleware, AuthMiddleware)
	avatars.Get("/", getAvatars)
	avatars.Get("/favorites", getAvatarFavorites)
	avatars.Get("/licensed", getLicensedAvatars)
	avatars.Get("/:id", getAvatar)
	avatars.Put("/:id/select", selectAvatar)
}

// getAvatars | GET /avatars
// This endpoint allows you to discover avatars on the platform.
func getAvatars(c *fiber.Ctx) error {
	var isGameRequest = c.Locals("isGameRequest").(bool)
	var avatars []models.Avatar
	var apiAvatars = make([]*models.APIAvatar, 0)
	var apiAvatarsWithPackages = make([]*models.APIAvatarWithPackages, 0)
	var u = c.Locals("user").(*models.User)
	var numberOfAvatarsToSearch = 60
	var avatarsOffset = 0
	var searchSort = ""
	var searchTerm = ""
	var searchTagsInclude = make([]string, 0)
	var searchTagsExclude = make([]string, 0)
	var searchSelf = false
	var searchUser = ""
	var searchReleaseStatus = models.ReleaseStatusPublic
	var limitToReleaseStatus = true

	var tx = config.DB.Model(&models.Avatar{}).
		Preload("Image").
		Preload("UnityPackages.File")

	if c.Query("n") != "" {
		atoi, err := strconv.Atoi(c.Query("n"))
		if err != nil {
			goto badRequest
		}

		if atoi < 1 || atoi > 100 {
			goto badRequest
		}

		numberOfAvatarsToSearch = atoi
	}

	if c.Query("offset") != "" {
		atoi, err := strconv.Atoi(c.Query("offset"))
		if err != nil {
			goto badRequest
		}

		if atoi < 0 {
			goto badRequest
		}

		avatarsOffset = atoi
	}

	if c.Query("search") != "" {
		if !u.IsStaff() {
			goto badRequest
		}
		searchTerm = c.Query("search")
	}

	if c.Query("user") == "me" {
		limitToReleaseStatus = false
		searchSelf = true
	}

	if c.Query("userId") != "" {
		searchUser = c.Query("userId")
	}

	if c.Query("tag") != "" {
		tags := strings.Split(c.Query("tag"), ",")
		searchTagsInclude = append(searchTagsInclude, tags...)
	}

	if c.Query("notag") != "" {
		tags := strings.Split(c.Query("notag"), ",")
		searchTagsExclude = append(searchTagsExclude, tags...)
	}

	if c.Query("releaseStatus") != "" {
		switch c.Query("releaseStatus") {
		case string(models.ReleaseStatusPublic):
			searchReleaseStatus = models.ReleaseStatusPublic

		case string(models.ReleaseStatusPrivate):
			searchReleaseStatus = models.ReleaseStatusPrivate
			if !searchSelf {
				searchSelf = true
			}
			if searchUser == "" {
				searchUser = u.ID
			}

		case string(models.ReleaseStatusHidden):
			searchReleaseStatus = models.ReleaseStatusHidden
		}
	}

	if c.Query("sort") != "" {
		searchSort = c.Query("sort")
	}

	if searchTerm != "" {
		searchTerm = "%" + searchTerm + "%"
		tx = tx.Where("name ILIKE ?", searchTerm)
	}

	if searchSelf {
		tx = tx.Where("author_id = ?", u.ID)
	}

	if searchUser != "" {
		tx = tx.Where("author_id = ?", searchUser)
	}

	if len(searchTagsInclude) > 0 {
		tx.Where("(?::text[] && tags) IS true", pq.StringArray(searchTagsInclude))
	}

	if len(searchTagsExclude) > 0 {
		tx.Where("(?::text[] && tags) IS NOT true", pq.StringArray(searchTagsExclude))
	}

	if searchSort != "" {
		if searchSort == "shuffle" {
			tx.Order("random()")
		}
	}

	if searchReleaseStatus != models.ReleaseStatusPublic {
		if searchReleaseStatus == models.ReleaseStatusHidden && u.DeveloperType != "internal" {
			goto badRequest
		}

		if searchReleaseStatus == models.ReleaseStatusPrivate &&
			(searchUser != u.ID || !searchSelf) && u.DeveloperType != "internal" {
			goto badRequest
		}
	}

	if limitToReleaseStatus {
		tx.Where("release_status = ?", searchReleaseStatus)
	}
	tx.Limit(numberOfAvatarsToSearch).Offset(avatarsOffset)

	tx.Find(&avatars)

	if isGameRequest {
		for _, avatar := range avatars {
			ap, err := avatar.GetAPIAvatarWithPackages()
			if err != nil {
				return err
			}
			apiAvatarsWithPackages = append(apiAvatarsWithPackages, ap)
		}
		return c.JSON(apiAvatarsWithPackages)
	} else {
		for _, avatar := range avatars {
			a, err := avatar.GetAPIAvatar()
			if err != nil {
				return err
			}
			apiAvatars = append(apiAvatars, a)
		}

		return c.JSON(apiAvatars)
	}

badRequest:
	return c.Status(400).JSON(models.MakeErrorResponse("Bad request", 400))
}

// getAvatarFavorites | GET /avatars/favorites
// Returns a list of the user's favorited avatars.
// TODO: Implement favorites.
func getAvatarFavorites(c *fiber.Ctx) error {
	return c.JSON([]struct{}{})
}

// getLicensedAvatars | GET /avatars/licensed
// Returns a list of the user's favorited avatars.
// Won't do: relates to commerce features.
func getLicensedAvatars(c *fiber.Ctx) error {
	return c.JSON([]struct{}{})
}

// getAvatar | GET /avatars/:id
// Returns an avatar.
func getAvatar(c *fiber.Ctx) error {
	var isGameRequest = c.Locals("isGameRequest").(bool)
	var a models.Avatar
	tx := config.DB.Preload(clause.Associations).Preload("UnityPackages.File").Model(&models.Avatar{}).Where("id = ?", c.Params("id")).First(&a)
	if tx.Error != nil {
		if tx.Error == gorm.ErrRecordNotFound {
			return c.Status(404).JSON(models.ErrAvatarNotFoundResponse)
		}
	}

	var aa *models.APIAvatar
	var aap *models.APIAvatarWithPackages
	var err error

	if isGameRequest {
		aap, err = a.GetAPIAvatarWithPackages()
	} else {
		aa, err = a.GetAPIAvatar()
	}
	if err != nil {
		return c.Status(500).JSON(models.MakeErrorResponse("internal server error while trying to get apiavatar", 500))
	}

	if isGameRequest {
		return c.JSON(aap)
	} else {
		return c.JSON(aa)
	}
}

// selectAvatar | PUT /avatars/:id/select
// Sets the avatar the user is currently in.
func selectAvatar(c *fiber.Ctx) error {
	var u = c.Locals("user").(*models.User)
	var a models.Avatar
	var changes = map[string]interface{}{}

	tx := config.DB.Preload(clause.Associations).Preload("UnityPackages.File").Model(&models.Avatar{}).Where("id = ?", c.Params("id")).First(&a)
	if tx.Error != nil {
		if tx.Error == gorm.ErrRecordNotFound {
			return c.Status(404).JSON(models.ErrAvatarNotFoundResponse)
		}
	}

	if !u.IsStaff() && a.ReleaseStatus != models.ReleaseStatusPublic && u.ID != a.AuthorID {
		return c.Status(403).JSON(models.MakeErrorResponse("trying to switch into private avatar not uploaded by self", 403))
	}

	changes["current_avatar_id"] = a.ID
	changes["fallback_avatar_id"] = a.ID

	tx = config.DB.Omit(clause.Associations).Model(&u).Updates(changes)
	fmt.Println(tx.Error)

	u.CurrentAvatarID = a.ID
	u.CurrentAvatar = a

	u.FallbackAvatarID = a.ID
	u.FallbackAvatar = a
	return c.JSON(u.GetAPICurrentUser())
}
