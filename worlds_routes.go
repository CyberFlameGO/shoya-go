package main

import (
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"strconv"
)

func worldsRoutes(app *fiber.App) {
	worlds := app.Group("/worlds")
	worlds.Get("/", ApiKeyMiddleware, AuthMiddleware, getWorlds)
	worlds.Get("/favorites", ApiKeyMiddleware, AuthMiddleware, getWorldFavorites)
	worlds.Get("/active", ApiKeyMiddleware, AuthMiddleware, getWorldsActive)
	worlds.Get("/recent", ApiKeyMiddleware, AuthMiddleware, getWorldsRecent)
	worlds.Get("/:id", ApiKeyMiddleware, AuthMiddleware, getWorld)
	worlds.Get("/:id/metadata", ApiKeyMiddleware, AuthMiddleware, getWorldMeta)
}

// getWorlds | /worlds
//
// This route retrieves a list of worlds based on various parameters (e.g.: search, offset, number).
// FIXME: This route is extremely unoptimized. Several tons of refactoring and fixing are required.
// TODO: Implement &tag, as well as &notag searching. No clue how to do this in SQL.
func getWorlds(c *fiber.Ctx) error {
	var isGameRequest = c.Locals("isGameRequest").(bool)
	var worlds []World
	var apiWorlds []*APIWorld
	var apiWorldsPackages []*APIWorldWithPackages
	var u = c.Locals("user").(*User)
	var numberOfWorldsToSearch = 60
	var worldsOffset = 0
	var searchTerm = ""
	var searchSelf = false
	var searchUser = ""
	var searchReleaseStatus = ReleaseStatusPublic

	// Query preparation
	var tx = DB.Model(&World{}).
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

	if c.Query("releaseStatus") != "" {
		switch c.Query("releaseStatus") {
		case string(ReleaseStatusPublic):
			searchReleaseStatus = ReleaseStatusPublic
			break
		case string(ReleaseStatusPrivate):
			searchReleaseStatus = ReleaseStatusPrivate
			if searchSelf == false {
				searchSelf = true
			}
			if searchUser == "" {
				searchUser = u.ID
			}
			break
		case string(ReleaseStatusHidden):
			searchReleaseStatus = ReleaseStatusHidden
			break
		}
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

	if searchReleaseStatus != ReleaseStatusPublic {
		if searchReleaseStatus == ReleaseStatusHidden && u.DeveloperType != "internal" {
			goto badRequest
		}

		if searchReleaseStatus == ReleaseStatusPrivate &&
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

func getWorld(c *fiber.Ctx) error {
	var isGameRequest = c.Locals("isGameRequest").(bool)
	var w World
	var aw *APIWorld
	var awp *APIWorldWithPackages
	var err error

	tx := DB.Preload(clause.Associations).Preload("UnityPackages.File").Model(&World{}).Where("id = ?", c.Params("id")).First(&w)
	if tx.Error != nil {
		if tx.Error == gorm.ErrRecordNotFound {
			return c.Status(404).JSON(ErrWorldNotFoundResponse)
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

func getWorldMeta(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"id":       c.Params("id"),
		"metadata": fiber.Map{},
	})
}
