package main

import (
	"github.com/gofiber/fiber/v2"
	"github.com/lib/pq"
	"gitlab.com/george/shoya-go/config"
	"gitlab.com/george/shoya-go/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"strconv"
	"strings"
)

func worldsRoutes(app *fiber.App) {
	worlds := app.Group("/worlds")
	worlds.Get("/", ApiKeyMiddleware, AuthMiddleware, getWorlds)
	worlds.Get("/favorites", ApiKeyMiddleware, AuthMiddleware, getWorldFavorites)
	worlds.Get("/active", ApiKeyMiddleware, AuthMiddleware, getWorldsActive)
	worlds.Get("/recent", ApiKeyMiddleware, AuthMiddleware, getWorldsRecent)
	worlds.Get("/:id", ApiKeyMiddleware, AuthMiddleware, getWorld)
	worlds.Get("/:id/metadata", ApiKeyMiddleware, AuthMiddleware, getWorldMeta)
	worlds.Get("/:id/publish", ApiKeyMiddleware, AuthMiddleware, getWorldPublish)
}

// getWorlds | GET /worlds
//
// This route retrieves a list of worlds based on various parameters (e.g.: search, offset, number).
// FIXME: This route is extremely unoptimized. Several tons of refactoring and fixing are required.
// TODO: Implement &tag, as well as &notag searching. No clue how to do this in SQL.
func getWorlds(c *fiber.Ctx) error {
	var isGameRequest = c.Locals("isGameRequest").(bool)
	var worlds []models.World
	var apiWorlds = make([]*models.APIWorld, 0)
	var apiWorldsPackages = make([]*models.APIWorldWithPackages, 0)
	var u = c.Locals("user").(*models.User)
	var numberOfWorldsToSearch = 60
	var worldsOffset = 0
	var searchSort = ""
	var searchTerm = ""
	var searchTagsInclude = make([]string, 0)
	var searchTagsExclude = make([]string, 0)
	var searchSelf = false
	var searchUser = ""
	var searchReleaseStatus = models.ReleaseStatusPublic

	// Query preparation
	var tx = config.DB.Model(&models.World{}).
		Preload("Image").
		Preload("UnityPackages.File")

	// Query parameter setup
	if c.Query("n") != "" {
		atoi, err := strconv.Atoi(c.Query("n"))
		if err != nil {
			goto badRequest
		}

		if atoi < 1 || atoi > 100 {
			goto badRequest
		}

		numberOfWorldsToSearch = atoi
	}

	if c.Query("offset") != "" {
		atoi, err := strconv.Atoi(c.Query("offset"))
		if err != nil {
			goto badRequest
		}

		if atoi < 0 {
			goto badRequest
		}

		worldsOffset = atoi
	}

	if c.Query("search") != "" {
		searchTerm = c.Query("search")
	}

	if c.Query("user") == "me" {
		searchSelf = true
	}

	if c.Query("userId") != "" {
		searchUser = c.Query("userId")
	}

	if c.Query("tag") != "" {
		tags := strings.Split(c.Query("tag"), ",")
		for _, tag := range tags {
			searchTagsInclude = append(searchTagsInclude, tag)
		}
	}

	if c.Query("notag") != "" {
		tags := strings.Split(c.Query("notag"), ",")
		for _, tag := range tags {
			searchTagsExclude = append(searchTagsExclude, tag)
		}
	}

	if c.Query("releaseStatus") != "" {
		switch c.Query("releaseStatus") {
		case string(models.ReleaseStatusPublic):
			searchReleaseStatus = models.ReleaseStatusPublic
			break
		case string(models.ReleaseStatusPrivate):
			searchReleaseStatus = models.ReleaseStatusPrivate
			if searchSelf == false {
				searchSelf = true
			}
			if searchUser == "" {
				searchUser = u.ID
			}
			break
		case string(models.ReleaseStatusHidden):
			searchReleaseStatus = models.ReleaseStatusHidden
			break
		}
	}

	if c.Query("sort") != "" {
		searchSort = c.Query("sort")
	}

	// Additional query prep based on parameters
	if searchTerm != "" {
		// TODO: full-text search on world name instead of this jank.
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
			(searchUser != u.ID || searchSelf == false) && u.DeveloperType != "internal" {
			goto badRequest
		}
	}
	tx.Where("release_status = ?", searchReleaseStatus)
	tx.Limit(numberOfWorldsToSearch).Offset(worldsOffset)

	tx.Find(&worlds)

	if isGameRequest {
		for _, world := range worlds {
			wp, err := world.GetAPIWorldWithPackages()
			if err != nil {
				return err
			}
			apiWorldsPackages = append(apiWorldsPackages, wp)
		}
		return c.JSON(apiWorldsPackages)
	} else {
		for _, world := range worlds {
			w, err := world.GetAPIWorld()
			if err != nil {
				return err
			}
			apiWorlds = append(apiWorlds, w)
		}

		return c.JSON(apiWorlds)
	}

badRequest:
	return c.Status(400).JSON(fiber.Map{
		"error": fiber.Map{
			"message":     "Bad request",
			"status_code": 400,
		},
	})
}

func getWorldFavorites(c *fiber.Ctx) error {
	return c.Status(501).JSON([]fiber.Map{})
}

func getWorldsActive(c *fiber.Ctx) error {
	return c.Status(501).JSON([]fiber.Map{})
}

func getWorldsRecent(c *fiber.Ctx) error {
	return c.Status(501).JSON([]fiber.Map{})
}

// getWorld | GET /worlds/:id
//
// This route retrieves information regarding a specific world id.
// The returned JSON is an array of either APIWorld, or APIWorldWithPackages
// It varies based on the request source (see: IsGameRequestMiddleware)
func getWorld(c *fiber.Ctx) error {
	var isGameRequest = c.Locals("isGameRequest").(bool)
	var w models.World
	var aw *models.APIWorld
	var awp *models.APIWorldWithPackages
	var err error

	tx := config.DB.Preload(clause.Associations).Preload("UnityPackages.File").Model(&models.World{}).Where("id = ?", c.Params("id")).First(&w)
	if tx.Error != nil {
		if tx.Error == gorm.ErrRecordNotFound {
			return c.Status(404).JSON(models.ErrWorldNotFoundResponse)
		}
	}

	if isGameRequest {
		awp, err = w.GetAPIWorldWithPackages()
	} else {
		aw, err = w.GetAPIWorld()
	}
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": fiber.Map{
				"message":     "internal server error while trying to get apiworld",
				"status_code": 500,
			},
		})
	}

	if isGameRequest {
		return c.JSON(awp)
	} else {
		return c.JSON(aw)
	}
}

// getWorldMeta | GET /worlds/:id/metadata
//
// This route returns metadata about a specific world id. At this time, there is only a "boilerplate" implementation,
// with no functional metadata sourcing.
func getWorldMeta(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"id":       c.Params("id"),
		"metadata": fiber.Map{},
	})
}

func getWorldPublish(c *fiber.Ctx) error {
	var u *models.User
	var w models.World

	u = c.Locals("user").(*models.User)
	tx := config.DB.Preload(clause.Associations).Preload("UnityPackages.File").Model(&models.World{}).Where("id = ?", c.Params("id")).First(&w)
	if tx.Error != nil {
		if tx.Error == gorm.ErrRecordNotFound {
			return c.Status(404).JSON(models.ErrWorldNotFoundResponse)
		}
	}

	if u.ID == w.AuthorID {
		return c.JSON(fiber.Map{
			"canPublish": true, // always true, not planning on doing labs.
		})
	}

	return c.JSON(fiber.Map{
		"canPublish": false,
	})
}
