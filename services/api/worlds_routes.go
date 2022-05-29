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

func worldsRoutes(app *fiber.App) {
	worlds := app.Group("/worlds", ApiKeyMiddleware, AuthMiddleware)
	worlds.Get("/", getWorlds)
	worlds.Get("/favorites", getWorldFavorites)
	worlds.Get("/active", getWorldsActive)
	worlds.Get("/recent", getWorldsRecent)
	worlds.Get("/:id", getWorld)
	worlds.Get("/:id/metadata", getWorldMeta)
	worlds.Get("/:id/publish", getWorldPublish)
	worlds.Get("/:id/:version/feedback", getWorldFeedback)
}

// getWorlds | GET /worlds
//
// This route retrieves a list of worlds based on various parameters (e.g.: search, offset, number).
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

	var is [][]string
	var i []*models.WorldInstance

	// Query preparation
	var tx = config.DB.Model(&models.World{}).
		Preload("Image").
		Preload("UnityPackages.File")

	// Query parameter setup
	if _n := c.Query("n"); _n != "" {
		atoi, err := strconv.Atoi(_n)
		if err != nil {
			goto badRequest
		}

		if atoi < 1 || atoi > 100 {
			goto badRequest
		}

		numberOfWorldsToSearch = atoi
	}

	if _o := c.Query("offset"); _o != "" {
		atoi, err := strconv.Atoi(_o)
		if err != nil {
			goto badRequest
		}

		if atoi < 0 {
			goto badRequest
		}

		worldsOffset = atoi
	}

	if _s := c.Query("search"); _s != "" {
		searchTerm = _s
	}

	if c.Query("user") == "me" {
		searchSelf = true
	}

	if _uid := c.Query("userId"); _uid != "" {
		searchUser = _uid
	}

	if _tags := c.Query("tag"); _tags != "" {
		tags := strings.Split(_tags, ",")
		searchTagsInclude = append(searchTagsInclude, tags...)
	}

	if _exclTags := c.Query("notag"); _exclTags != "" {
		tags := strings.Split(_exclTags, ",")
		searchTagsExclude = append(searchTagsExclude, tags...)
	}

	if _r := c.Query("releaseStatus"); _r != "" {
		switch _r {
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

	if _s := c.Query("sort"); _s != "" {
		searchSort = _s
	}

	// Additional query prep based on parameters
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
	tx.Where("release_status = ?", searchReleaseStatus)
	tx.Limit(numberOfWorldsToSearch).Offset(worldsOffset)

	tx.Find(&worlds)

	if isGameRequest {
		for _, world := range worlds {
			wp, err := world.GetAPIWorldWithPackages()
			if err != nil {
				return err
			}

			if config.ApiConfiguration.DiscoveryServiceEnabled.Get() {
				i = DiscoveryService.GetInstancesForWorld(wp.ID)
				is = make([][]string, len(i))
				for idx, _i := range i {
					is[idx] = []string{_i.InstanceID, fmt.Sprintf("%d", _i.PlayerCount.Total)}
				}
				wp.Instances = is
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

			if config.ApiConfiguration.DiscoveryServiceEnabled.Get() {
				i = DiscoveryService.GetInstancesForWorld(w.ID)
				is = make([][]string, len(i))
				for idx, _i := range i {
					is[idx] = []string{_i.InstanceID, fmt.Sprintf("%d", _i.PlayerCount.Total)}
				}
				w.Instances = is
			}

			apiWorlds = append(apiWorlds, w)
		}

		return c.JSON(apiWorlds)
	}

badRequest:
	return c.Status(400).JSON(models.MakeErrorResponse("Bad request", 400))
}

// getWorldFavorites | GET /worlds/favorites
// Returns the user's favorite worlds.
// TODO: Implement favorites
func getWorldFavorites(c *fiber.Ctx) error {
	return c.JSON([]struct{}{})
}

// getWorldsActive | GET /worlds/active
// Returns the most active (in terms of ccu) worlds.
// TODO: Implement presence & world metrics
func getWorldsActive(c *fiber.Ctx) error {
	return c.JSON([]struct{}{})
}

// getWorldsRecent | GET /worlds/recent
// Returns the most recent worlds the user has been in.
// TODO: Implement presence.
func getWorldsRecent(c *fiber.Ctx) error {
	return c.JSON([]struct{}{})
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

	var is [][]string
	var i []*models.WorldInstance

	var err error

	tx := config.DB.Preload(clause.Associations).Preload("UnityPackages.File").Model(&models.World{}).Where("id = ?", c.Params("id")).First(&w)
	if tx.Error != nil {
		if tx.Error == gorm.ErrRecordNotFound {
			return c.Status(404).JSON(models.ErrWorldNotFoundResponse)
		}
	}

	if isGameRequest {
		awp, err = w.GetAPIWorldWithPackages()
		if config.ApiConfiguration.DiscoveryServiceEnabled.Get() {
			i = DiscoveryService.GetInstancesForWorld(awp.ID)
		}
	} else {
		aw, err = w.GetAPIWorld()
		if config.ApiConfiguration.DiscoveryServiceEnabled.Get() {
			i = DiscoveryService.GetInstancesForWorld(aw.ID)
		}
	}

	if config.ApiConfiguration.DiscoveryServiceEnabled.Get() {
		is = make([][]string, len(i))
		for idx, _i := range i {
			is[idx] = []string{_i.InstanceID, fmt.Sprintf("%d", _i.PlayerCount.Total)}
		}

		if isGameRequest {
			awp.Instances = is
		} else {
			aw.Instances = is
		}
	}

	if err != nil {
		return c.Status(500).JSON(models.MakeErrorResponse("internal server error while trying to get apiworld", 500))
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
		"metadata": struct{}{},
	})
}

// getWorldFeedback | GET /worlds/:id/0/feedback
// Returns the reports created against this world.
// TODO: Implement reporting system
func getWorldFeedback(c *fiber.Ctx) error {
	var u = c.Locals("user").(*models.User)
	var w models.World
	tx := config.DB.Preload(clause.Associations).Preload("UnityPackages.File").Model(&models.World{}).Where("id = ?", c.Params("id")).First(&w)
	if tx.Error != nil {
		if tx.Error == gorm.ErrRecordNotFound {
			return c.Status(404).JSON(models.ErrWorldNotFoundResponse)
		}
	}

	if u.ID != w.AuthorID {
		return c.Status(403).JSON(models.MakeErrorResponse("not allowed to access feedback for this world", 403))
	}

	return c.JSON(fiber.Map{
		"reportScore":   0,
		"reportCount":   0,
		"reportReasons": []struct{}{},
	})
}

// getWorldPublish | GET /worlds/:id/publish
// Returns whether this world can be published to labs(?).
func getWorldPublish(c *fiber.Ctx) error {
	var u = c.Locals("user").(*models.User)
	var w models.World

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
