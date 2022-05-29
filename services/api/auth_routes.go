package main

import (
	"fmt"
	"github.com/gofiber/fiber/v2"
	"gitlab.com/george/shoya-go/config"
	"gitlab.com/george/shoya-go/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"strings"
	"time"
)

func authRoutes(r *fiber.App) {
	auth := r.Group("/auth", ApiKeyMiddleware)
	auth.Get("/", AuthMiddleware, getAuth)
	auth.Get("/exists", getExists)
	auth.Post("/register", postRegister)

	user := auth.Group("/user")
	user.Get("/", LoginMiddleware, AuthMiddleware, getSelf)
	user.Get("/friends", AuthMiddleware, getFriends)
	user.Get("/moderations", AuthMiddleware, getModerations)
	user.Get("/notifications", AuthMiddleware, getNotifications)

	user.Get("/playermoderations", AuthMiddleware, getPlayerModerations)
	user.Post("/playermoderations", AuthMiddleware, postPlayerModerations)
	user.Delete("/playermoderations", AuthMiddleware, deletePlayerModerations)

	user.Get("/playermoderations/:id", AuthMiddleware, getPlayerModeration)
	user.Delete("/playermoderations/:id", AuthMiddleware, deletePlayerModeration)

	user.Put("/unplayermoderate", AuthMiddleware, putUnPlayerModerate)

	// Stubs
	auth.Get("/permissions", AuthMiddleware, getPermissions)
	user.Get("/subscription", AuthMiddleware, getSubscription)
	user.Get("/playermoderated", AuthMiddleware, getPlayerModerated)
}

// getAuth | GET /auth
// Returns the current user's auth token (and refreshes it if necessary).
func getAuth(c *fiber.Ctx) error {
	return c.Status(200).JSON(fiber.Map{
		"ok":    true,
		"token": c.Locals("authCookie").(string),
	})
}

// getExists | GET /auth/exists
// Used to check whether a user with a given username, display name, or email exists.
func getExists(c *fiber.Ctx) error {
	var u *models.User
	var exists = true

	tx := config.DB.Where("username = ?", c.Query("username")).
		Or("display_name = ?", c.Query("displayName")).
		Or("email = ?", c.Query("email")).
		Not("id = ?", c.Query("excludeUserId")). // Exclude the user with the given id if provided.
		Select("id").First(u)

	if tx.Error != nil {
		if tx.Error == gorm.ErrRecordNotFound {
			exists = false
		}
	}
	return c.Status(200).JSON(fiber.Map{
		"userExists": exists,
	})
}

// postRegister | POST /auth/register
// Registers a new user on the platform.
func postRegister(c *fiber.Ctx) error {
	var r RegisterRequest
	var u *models.User

	if config.ApiConfiguration.DisableRegistration.Get() {
		return c.Status(400).JSON(fiber.Map{
			"ok": false,
			"error": fiber.Map{
				"message": "Registrations are currently disabled.",
			},
		})
	}

	if err := c.BodyParser(&r); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"ok": false,
			"error": fiber.Map{
				"message": "Invalid request body",
			},
		})
	}

	if r.Username == "" || r.Email == "" || r.Password == "" {
		return c.Status(400).JSON(fiber.Map{
			"ok": false,
			"error": fiber.Map{
				"message": "Missing required fields",
			},
		})
	}

	if len(r.Username) < 3 || len(r.Username) > 32 {
		return c.Status(400).JSON(fiber.Map{
			"ok": false,
			"error": fiber.Map{
				"message": "Username must be between 3 and 32 characters",
			},
		})
	}

	if len(r.Password) < 8 {
		return c.Status(400).JSON(fiber.Map{
			"ok": false,
			"error": fiber.Map{
				"message": "Password must be at least 8 characters",
			},
		})
	}

	tx := config.DB.Where("username = ?", strings.ToLower(r.Username)).
		Or("display_name = ?", r.Username).
		Or("email = ?", strings.ToLower(r.Email)).First(u)

	if tx.Error != gorm.ErrRecordNotFound {
		return c.Status(400).JSON(fiber.Map{
			"ok": false,
			"error": fiber.Map{
				"message": "Username, display name, or email already exists",
			},
		})
	}

	u = models.NewUser(r.Username, r.Username, r.Email, r.Password)
	tx = config.DB.Create(u)
	if tx.Error != nil {
		return c.Status(500).JSON(fiber.Map{
			"ok": false,
			"error": fiber.Map{
				"message": "Internal server error",
			},
		})
	}

	return c.Status(200).JSON(u.GetAPICurrentUser())
}

// getSelf | GET /auth/user
// Returns the current user's information.
func getSelf(c *fiber.Ctx) error {
	var u = c.Locals("user").(*models.User)

	return c.Status(200).JSON(u.GetAPICurrentUser())
}

// getFriends | GET /auth/user/friends
// Returns a list of the user's friends.
// TODO: This requires the implementation of friends.
func getFriends(c *fiber.Ctx) error {
	return c.JSON([]struct{}{})
}

// getNotifications | GET /auth/user/notifications
// Returns the current user's notifications.
// TODO: This requires the implementation of presence. (presence has to be built before notifications).
func getNotifications(c *fiber.Ctx) error {
	return c.Status(200).JSON([]struct{}{})
}

// getModerations | GET /auth/user/moderations
// Returns the active moderations against the user.
func getModerations(c *fiber.Ctx) error {
	var u = c.Locals("user").(*models.User)
	//goland:noinspection GoPreferNilSlice
	r := []*models.APIModeration{}
	for _, moderation := range u.Moderations {
		if moderation.ExpiresAt == 0 || moderation.ExpiresAt > time.Now().UTC().Unix() {
			am := moderation.GetAPIModeration(false)
			am.TargetDisplayName = u.DisplayName
			r = append(r, am)
		}
	}

	return c.JSON(r)
}

// getPlayerModerations | GET /auth/user/playermoderations
// Returns the player moderations this user has enacted.
func getPlayerModerations(c *fiber.Ctx) error {
	var u = c.Locals("user").(*models.User)
	var mods []models.PlayerModeration
	//goland:noinspection GoPreferNilSlice
	var resp = []*models.APIPlayerModeration{}

	modType := models.PlayerModerationAll

	if t := c.Query("type"); t != "" {
		modType = models.GetPlayerModerationType(t)
	}

	tx := config.DB.Preload(clause.Associations).Where("source_id = ?", u.ID)
	if modType != models.PlayerModerationAll {
		tx.Where("action = ?", modType)
	}

	if t := c.Query("targetUserId"); t != "" {
		tx.Where("target_id = ?", t)
	}

	tx.Find(&mods)

	for _, mod := range mods {
		resp = append(resp, mod.GetAPIPlayerModeration())
	}

	return c.JSON(resp)
}

// postPlayerModerations | POST /auth/user/playermoderations
// Creates a new player moderation.
func postPlayerModerations(c *fiber.Ctx) error {
	var u = c.Locals("user").(*models.User)
	var mod *models.PlayerModeration
	var req PlayerModerationRequest
	err := c.BodyParser(&req)
	if err != nil {
		return c.Status(500).JSON(models.MakeErrorResponse(err.Error(), 500))
	}

	tx := config.DB.Preload(clause.Associations).Where("id = ?", c.Params("id")).Where("source_id = ?", u.ID).First(&mod)
	if tx.RowsAffected != 0 {
		return c.JSON(mod.GetAPIPlayerModeration())
	}

	if err != nil {
		if err != gorm.ErrRecordNotFound {
			return c.Status(500).JSON(fiber.Map{
				"error": fiber.Map{
					"message":     err.Error(),
					"status_code": 500,
				},
			})
		}
	}

	mod = &models.PlayerModeration{
		SourceID: u.ID,
		TargetID: req.Against,
		Action:   req.Type,
	}

	err = config.DB.Create(mod).Error
	if err != nil {
		return c.Status(500).JSON(models.MakeErrorResponse(err.Error(), 500))
	}

	return c.JSON(mod.GetAPIPlayerModeration())
}

// deletePlayerModerations | DELETE /auth/user/playermoderations
// Deletes all active player moderations from the user.
func deletePlayerModerations(c *fiber.Ctx) error {
	var u = c.Locals("user").(*models.User)
	err := config.DB.Unscoped().Where("source_id = ?", u.ID).Delete(&models.PlayerModeration{}).Error
	if err != nil {
		return c.Status(500).JSON(models.MakeErrorResponse(err.Error(), 500))
	}

	return c.JSON(fiber.Map{
		"success": fiber.Map{
			"message":     "OK",
			"status_code": 200,
		}})
}

// putUnPlayerModerate | PUT /auth/user/unplayermoderate
// Removes a player moderation against another player (or all of a player moderation type).
func putUnPlayerModerate(c *fiber.Ctx) error {
	var u = c.Locals("user").(*models.User)
	var req PlayerModerationRequest
	err := c.BodyParser(&req)
	if err != nil {
		return c.Status(500).JSON(models.MakeErrorResponse(err.Error(), 500))
	}

	query := config.DB.Unscoped().Where("source_id = ?", u.ID)
	if req.Against != "" {
		query.Where("target_id = ?", req.Against)
	}

	if req.Type != "" {
		query.Where("action = ?", req.Type)
	}

	err = query.Delete(&models.PlayerModeration{}).Error
	if err != nil {
		return c.Status(500).JSON(models.MakeErrorResponse(err.Error(), 500))
	}

	return c.JSON(fiber.Map{
		"success": fiber.Map{
			"message":     fmt.Sprintf("User %s unmoderated", req.Against),
			"status_code": 200,
		},
	})

}

// getPlayerModeration | GET /auth/user/playermoderations/:id
// Returns a single player moderation.
func getPlayerModeration(c *fiber.Ctx) error {
	var u = c.Locals("user").(*models.User)
	var mod *models.PlayerModeration

	err := config.DB.Preload(clause.Associations).Where("id = ?", c.Params("id")).Where("source_id = ?", u.ID).First(mod).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(404).JSON(models.MakeErrorResponse("can't find playerModeration!", 404))
		}

		return c.Status(500).JSON(models.MakeErrorResponse(err.Error(), 500))
	}

	return c.JSON(mod.GetAPIPlayerModeration())
}

// deletePlayerModeration | GET /auth/user/playermoderations/:id
// Deletes a single player moderation.
func deletePlayerModeration(c *fiber.Ctx) error {
	var u = c.Locals("user").(*models.User)
	var mod *models.PlayerModeration

	err := config.DB.Preload(clause.Associations).Where("id = ?", c.Params("id")).First(&mod).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(404).JSON(models.MakeErrorResponse("can't find playerModeration!", 404))
		}

		return c.Status(500).JSON(models.MakeErrorResponse(err.Error(), 500))
	}

	if mod.SourceID != u.ID {
		return c.Status(403).JSON(models.MakeErrorResponse("You definitely can't delete a playerModeration you didn't create", 403))
	}

	err = config.DB.Unscoped().Where("id = ?", c.Params("id")).Delete(&models.PlayerModeration{}).Error
	if err != nil {
		return c.Status(500).JSON(models.MakeErrorResponse(err.Error(), 500))
	}

	return c.JSON(fiber.Map{
		"success": fiber.Map{
			"message":     fmt.Sprintf("PlayerModeration %s removed", c.Params("id")),
			"status_code": 200,
		},
	})
}

// getPlayerModerated | GET /auth/user/playermoderated
// Stub route which will not receive an implementation; Circa build 333.
func getPlayerModerated(c *fiber.Ctx) error {
	return c.JSON([]struct{}{})
}

// getSubscription | GET /auth/user/subscription
// Stub route which will not receive an implementation.
func getSubscription(c *fiber.Ctx) error {
	return c.JSON([]struct{}{})
}

// getPermissions | GET /auth/permissions
// Stub route which will not receive an implementation.
func getPermissions(c *fiber.Ctx) error {
	if c.Query("condensed") == "true" { // MUST be "true", not True, or TRUE. GG's.
		return c.JSON(fiber.Map{}) // In the case of condensed=true, an object is expected.
	}
	return c.JSON([]struct{}{})
}
