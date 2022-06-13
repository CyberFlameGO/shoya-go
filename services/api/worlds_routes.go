package api

import (
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/lib/pq"
	"gitlab.com/george/shoya-go/config"
	"gitlab.com/george/shoya-go/models"
	"gorm.io/gorm/clause"
	"strconv"
	"strings"
)

func worldsRoutes(app *fiber.App) {
	worlds := app.Group("/worlds", ApiKeyMiddleware, AuthMiddleware)
	worlds.Get("/", getWorlds)
	worlds.Post("/", postWorlds)
	worlds.Get("/favorites", getWorldFavorites)
	worlds.Get("/active", getWorldsActive)
	worlds.Get("/recent", getWorldsRecent)
	worlds.Get("/:id", getWorld)
	worlds.Put("/:id", putWorld)
	worlds.Get("/:id/metadata", getWorldMeta)
	worlds.Get("/:id/publish", getWorldPublish)
	worlds.Put("/:id/publish", putWorldPublish)
	worlds.Delete("/:id/publish", deleteWorldPublish)
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
		Preload("Image.Versions").
		Preload("Image.Versions.FileDescriptor").
		Preload("Image.Versions.DeltaDescriptor").
		Preload("Image.Versions.SignatureDescriptor").
		Preload("UnityPackages.File").
		Preload("UnityPackages.File.Versions").
		Preload("UnityPackages.File.Versions.FileDescriptor").
		Preload("UnityPackages.File.Versions.DeltaDescriptor").
		Preload("UnityPackages.File.Versions.SignatureDescriptor")

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

func putWorld(c *fiber.Ctx) error {
	var r *CreateWorldRequest
	var u = c.Locals("user").(*models.User)
	var w *models.World
	var fileId string
	var imageId string
	var changes = map[string]interface{}{}
	var aw *models.APIWorldWithPackages
	var unp *models.WorldUnityPackage
	var lunp *models.APIUnityPackage
	var fv int
	var err error

	if !u.CanUploadWorlds() {
		return c.Status(403).JSON(models.MakeErrorResponse("cannot upload worlds at this time", 403))
	}

	if err = c.BodyParser(&r); err != nil {
		return c.Status(400).JSON(models.MakeErrorResponse("bad request", 400))
	}

	if w, err = models.GetWorldById(c.Params("id")); w == nil || err != nil {
		if err == models.ErrWorldNotFound {
			return c.Status(404).JSON(models.ErrWorldNotFoundResponse)
		}
		return c.Status(500).JSON(models.MakeErrorResponse(err.Error(), 500))
	}

	if w.AuthorID != u.ID {
		return c.Status(403).JSON(models.MakeErrorResponse("not authorized to update this world", 403))
	}

	if r.AssetUrl != "" || r.ImageUrl != "" {
		if !r.HasValidUrls() {
			return c.Status(400).JSON(models.MakeErrorResponse("bad request", 400))
		}

		if r.AssetUrl != "" {
			if fileId, err = r.GetFileID(); err != nil {
				return c.Status(400).JSON(models.MakeErrorResponse("bad request", 400))
			}
		}

		if r.ImageUrl != "" {
			if imageId, err = r.GetImageID(); err != nil {
				return c.Status(400).JSON(models.MakeErrorResponse("bad request", 400))
			}
		}
	}

	if fileId != "" {
		lv := 0
		lvidx := 0
		for idx, vunp := range w.GetUnityPackages(true) {
			if vunp.AssetVersion > lv {
				lv = vunp.AssetVersion
				lvidx = idx
			}
		}

		lunp = &w.GetUnityPackages(true)[lvidx]
		if r.UnityVersion == "" {
			r.UnityVersion = lunp.UnityVersion
		}

		fv, err = r.GetFileVersion()
		if err != nil {
			return c.Status(500).JSON(models.MakeErrorResponse(err.Error(), 500))
		}
		unp = &models.WorldUnityPackage{
			BelongsToAssetID: w.ID,
			FileID:           fileId,
			FileVersion:      fv,
			Version:          r.AssetVersion,
			Platform:         "standalonewindows",
			UnityVersion:     r.UnityVersion,
		}

		config.DB.Create(&unp)
		w, err = models.GetWorldById(w.ID)
		if err != nil {
			return c.Status(500).JSON(models.MakeErrorResponse(err.Error(), 500))
		}
	}

	if imageId != "" {
		changes["image_id"] = imageId
	}

	if r.Name != "" {
		changes["name"] = r.Name
	}

	if r.Description != "" {
		changes["description"] = r.Description
	}

	if r.ReleaseStatus != "" {
		switch models.ReleaseStatus(r.ReleaseStatus) {
		case models.ReleaseStatusPrivate:
			changes["release_status"] = models.ReleaseStatusPrivate
		case models.ReleaseStatusPublic:
			changes["release_status"] = models.ReleaseStatusPublic
		case models.ReleaseStatusHidden:
			if u.IsStaff() {
				changes["release_status"] = models.ReleaseStatusHidden
			}
		}
	}

	if r.Capacity != 0 {
		if r.Capacity > 128 {
			return c.Status(400).JSON(models.MakeErrorResponse("world cannot have a soft-cap of more than 128", 400))
		}
		changes["capacity"] = r.Capacity
	}

	if err = config.DB.Omit(clause.Associations).Model(&w).Updates(changes).Error; err != nil {
		return c.Status(500).JSON(models.MakeErrorResponse(err.Error(), 500))
	}

	if aw, err = w.GetAPIWorldWithPackages(); err != nil {
		return c.Status(500).JSON(models.MakeErrorResponse(err.Error(), 500))
	}
	return c.JSON(aw)
}

func postWorlds(c *fiber.Ctx) error {
	var r *CreateWorldRequest
	var u = c.Locals("user").(*models.User)
	var w *models.World
	var fileId string
	var imageId string
	var aw *models.APIWorldWithPackages
	var fv int
	var err error

	if !u.CanUploadWorlds() {
		return c.Status(403).JSON(models.MakeErrorResponse("cannot upload worlds at this time", 403))
	}

	if err = c.BodyParser(&r); err != nil {
		return c.Status(400).JSON(models.MakeErrorResponse("bad request", 400))
	}

	if w, err = models.GetWorldById(r.ID); w != nil || err == nil {
		return c.Status(403).JSON(models.MakeErrorResponse("not allowed to overwrite an already-existing world", 403))
	}

	if !r.HasValidUrls() {
		return c.Status(400).JSON(models.MakeErrorResponse("bad request", 400))
	}

	if fileId, err = r.GetFileID(); err != nil {
		return c.Status(400).JSON(models.MakeErrorResponse("bad request", 400))
	}

	if imageId, err = r.GetImageID(); err != nil {
		return c.Status(400).JSON(models.MakeErrorResponse("bad request", 400))
	}

	w = &models.World{
		AuthorID:      u.ID,
		Name:          r.Name,
		Description:   r.Description,
		ImageID:       imageId,
		ReleaseStatus: models.ReleaseStatusPrivate,
		Tags:          r.ParseTags(),
		Version:       0,
		Capacity:      r.Capacity,
	}
	w.ID = r.ID
	r.Tags = append(r.ParseTags(), "system_approved")

	if r.ReleaseStatus != "" {
		switch models.ReleaseStatus(r.ReleaseStatus) {
		case models.ReleaseStatusPrivate:
			w.ReleaseStatus = models.ReleaseStatusPrivate
		case models.ReleaseStatusPublic:
			w.ReleaseStatus = models.ReleaseStatusPublic
		case models.ReleaseStatusHidden:
			if u.IsStaff() {
				w.ReleaseStatus = models.ReleaseStatusHidden
			}
		}
	}

	if tx := config.DB.Omit(clause.Associations).Create(&w); tx.Error != nil {
		return c.Status(500).JSON(models.MakeErrorResponse(tx.Error.Error(), 500))
	}

	if fv, err = r.GetFileVersion(); err != nil {
		return c.Status(500).JSON(models.MakeErrorResponse(err.Error(), 500))
	}

	unp := &models.WorldUnityPackage{
		BelongsToAssetID: w.ID,
		FileID:           fileId,
		FileVersion:      fv,
		Version:          r.AssetVersion,
		Platform:         r.Platform,
		UnityVersion:     r.UnityVersion,
		UnitySortNumber:  0,
	}

	if tx := config.DB.Create(&unp); tx.Error != nil {
		return c.Status(500).JSON(models.MakeErrorResponse(tx.Error.Error(), 500))
	}

	w, err = models.GetWorldById(w.ID)
	if err != nil {
		if err == models.ErrWorldNotFound {
			return c.Status(404).JSON(models.ErrWorldNotFoundResponse)
		}
		return c.Status(500).JSON(models.MakeErrorResponse(err.Error(), 500))
	}

	if aw, err = w.GetAPIWorldWithPackages(); err != nil {
		return c.Status(500).JSON(models.MakeErrorResponse(err.Error(), 500))
	}

	return c.JSON(aw)
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

	var w *models.World
	var aw *models.APIWorld
	var awp *models.APIWorldWithPackages

	var is [][]string
	var i []*models.WorldInstance

	var err error

	if w, err = models.GetWorldById(c.Params("id")); err != nil {
		if err == models.ErrWorldNotFound {
			return c.Status(404).JSON(models.ErrWorldNotFoundResponse)
		}

		return c.Status(500).JSON(models.MakeErrorResponse(err.Error(), 500))
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
	var w *models.World
	var err error
	if w, err = models.GetWorldById(c.Params("id")); err != nil {
		if err == models.ErrWorldNotFound {
			return c.Status(404).JSON(models.ErrWorldNotFoundResponse)
		}

		return c.Status(500).JSON(models.MakeErrorResponse(err.Error(), 500))
	}

	if u.ID != w.AuthorID && !u.IsStaff() {
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
	var w *models.World
	var err error

	if w, err = models.GetWorldById(c.Params("id")); err != nil {
		if err == models.ErrWorldNotFound {
			return c.Status(404).JSON(models.ErrWorldNotFoundResponse)
		}

		return c.Status(500).JSON(models.MakeErrorResponse(err.Error(), 500))
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

func putWorldPublish(c *fiber.Ctx) error {
	var u = c.Locals("user").(*models.User)
	var w *models.World
	var aw *models.APIWorld
	var changes = map[string]interface{}{}
	var err error

	if w, err = models.GetWorldById(c.Params("id")); err != nil {
		if err == models.ErrWorldNotFound {
			return c.Status(404).JSON(models.ErrWorldNotFoundResponse)
		}

		return c.Status(500).JSON(models.MakeErrorResponse(err.Error(), 500))
	}

	if u.ID != w.AuthorID {
		return c.Status(403).JSON(models.MakeErrorResponse("not allowed to set publish status for this world", 403))
	}

	changes["release_status"] = models.ReleaseStatusPublic
	config.DB.Omit(clause.Associations).Model(&w).Updates(changes)

	if aw, err = w.GetAPIWorld(); err != nil {
		return c.Status(500).JSON(models.MakeErrorResponse(err.Error(), 500))
	}
	return c.JSON(aw)
}

func deleteWorldPublish(c *fiber.Ctx) error {
	var u = c.Locals("user").(*models.User)
	var w *models.World
	var aw *models.APIWorld
	var changes = map[string]interface{}{}
	var err error

	if w, err = models.GetWorldById(c.Params("id")); err != nil {
		if err == models.ErrWorldNotFound {
			return c.Status(404).JSON(models.ErrWorldNotFoundResponse)
		}

		return c.Status(500).JSON(models.MakeErrorResponse(err.Error(), 500))
	}

	if u.ID != w.AuthorID {
		return c.Status(403).JSON(models.MakeErrorResponse("not allowed to set publish status for this world", 403))
	}

	changes["release_status"] = models.ReleaseStatusPrivate
	config.DB.Omit(clause.Associations).Model(&w).Updates(changes)

	if aw, err = w.GetAPIWorld(); err != nil {
		return c.Status(500).JSON(models.MakeErrorResponse(err.Error(), 500))
	}
	return c.JSON(aw)
}
