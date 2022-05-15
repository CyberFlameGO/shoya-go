package main

import (
	"fmt"
	"github.com/gofiber/fiber/v2"
	"gitlab.com/george/shoya-go/config"
	"gitlab.com/george/shoya-go/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"strings"
)

func authRoutes(router *fiber.App) {
	router.Get("/auth", ApiKeyMiddleware, AuthMiddleware, getAuth)
	router.Get("/auth/exists", ApiKeyMiddleware, getExists)
	router.Post("/auth/register", ApiKeyMiddleware, postRegister)
	router.Get("/auth/user", ApiKeyMiddleware, DoLoginMiddleware, AuthMiddleware, getSelf)
	router.Get("/auth/user/friends", ApiKeyMiddleware, AuthMiddleware, getFriends)
	router.Get("/auth/user/moderations", ApiKeyMiddleware, AuthMiddleware, getModerations)
	router.Get("/auth/user/notifications", ApiKeyMiddleware, AuthMiddleware, getNotifications)

	router.Get("/auth/user/playermoderations", ApiKeyMiddleware, AuthMiddleware, getPlayerModerations)
	router.Post("/auth/user/playermoderations", ApiKeyMiddleware, AuthMiddleware, postPlayerModerations)
	router.Delete("/auth/user/playermoderations", ApiKeyMiddleware, AuthMiddleware, deletePlayerModerations)

	router.Put("/auth/user/unplayermoderate", ApiKeyMiddleware, AuthMiddleware, putUnPlayerModerate)

	router.Get("/auth/user/playermoderations/:id", ApiKeyMiddleware, AuthMiddleware, getPlayerModeration)
	router.Delete("/auth/user/playermoderations/:id", ApiKeyMiddleware, AuthMiddleware, deletePlayerModeration)

	// Stub routes | Will not implement
	router.Get("/auth/user/subscription", ApiKeyMiddleware, AuthMiddleware, getSubscription)
	router.Get("/auth/permissions", ApiKeyMiddleware, AuthMiddleware, getPermissions)
	router.Get("/auth/user/playermoderated", ApiKeyMiddleware, AuthMiddleware, getPlayerModerated)
}

// getAuth | /auth
// Returns the current user's auth token (and refreshes it if necessary)
func getAuth(c *fiber.Ctx) error {
	return c.Status(200).JSON(fiber.Map{
		"ok":    true,
		"token": c.Locals("authCookie").(string),
	})
}

// getExists | /auth/exists
// Used to check whether a user with a given username, display name, or email exists.
func getExists(c *fiber.Ctx) error {
	var u models.User
	var exists = true

	tx := config.DB.Where("username = ?", c.Query("username")).
		Or("display_name = ?", c.Query("displayName")).
		Or("email = ?", c.Query("email")).
		Not("id = ?", c.Query("excludeUserId")). // Exclude the user with the given id if provided.
		Select("id").First(&u)

	if tx.Error != nil {
		if tx.Error == gorm.ErrRecordNotFound {
			exists = false
		}
	}
	return c.Status(200).JSON(fiber.Map{
		"userExists": exists,
	})
}

// postRegister | /auth/register
// Used to register a new user.
func postRegister(c *fiber.Ctx) error {
	var r RegisterRequest
	var _u models.User

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
		Or("email = ?", strings.ToLower(r.Email)).First(&_u)

	if tx.Error != gorm.ErrRecordNotFound {
		return c.Status(400).JSON(fiber.Map{
			"ok": false,
			"error": fiber.Map{
				"message": "Username, display name, or email already exists",
			},
		})
	}

	u := models.NewUser(r.Username, r.Username, r.Email, r.Password)
	tx = config.DB.Create(&u)
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

// getSelf | /auth/user
// Returns the current user's information.
func getSelf(c *fiber.Ctx) error {
	u := c.Locals("user").(*models.User)

	return c.Status(200).JSON(u.GetAPICurrentUser())
}

func getFriends(c *fiber.Ctx) error {
	return c.JSON([]fiber.Map{})
}

// getNotifications | /auth/user/notifications
// Returns the current user's notifications.
func getNotifications(c *fiber.Ctx) error {
	return c.Status(200).JSON([]fiber.Map{}) // TODO: implement
}

func getSubscription(c *fiber.Ctx) error {
	return c.JSON([]interface{}{})
}

func getPermissions(c *fiber.Ctx) error {
	return c.JSON([]interface{}{})
}

func getModerations(c *fiber.Ctx) error {
	return c.JSON([]interface{}{})
}

func getPlayerModerations(c *fiber.Ctx) error {
	var mods []models.PlayerModeration
	//goland:noinspection GoPreferNilSlice,GoPreferNilSlice
	var resp = []*models.APIPlayerModeration{}

	u := c.Locals("user").(*models.User)
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

func postPlayerModerations(c *fiber.Ctx) error {
	var mod *models.PlayerModeration
	var req PlayerModerationRequest
	u := c.Locals("user").(*models.User)
	err := c.BodyParser(&req)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": fiber.Map{
				"message":     err.Error(),
				"status_code": 500,
			},
		})
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
		return c.Status(500).JSON(fiber.Map{
			"error": fiber.Map{
				"message":     err.Error(),
				"status_code": 500,
			},
		})
	}

	return c.JSON(mod.GetAPIPlayerModeration())
}

func putUnPlayerModerate(c *fiber.Ctx) error {
	var req PlayerModerationRequest
	u := c.Locals("user").(*models.User)
	err := c.BodyParser(&req)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": fiber.Map{
				"message":     err.Error(),
				"status_code": 500,
			},
		})
	}

	err = config.DB.Unscoped().Where("source_id = ?", u.ID).Where("target_id = ?", req.Against).Where("action = ?", req.Type).Delete(&models.PlayerModeration{}).Error
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": fiber.Map{
				"message":     err.Error(),
				"status_code": 500,
			},
		})
	}

	return c.JSON(fiber.Map{
		"success": fiber.Map{
			"message":     fmt.Sprintf("User %s unmoderated", req.Against),
			"status_code": 200,
		},
	})

}

func deletePlayerModerations(c *fiber.Ctx) error {
	u := c.Locals("user").(*models.User)
	err := config.DB.Unscoped().Where("source_id = ?", u.ID).Delete(&models.PlayerModeration{}).Error
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": fiber.Map{
				"message":     err.Error(),
				"status_code": 500,
			},
		})
	}

	return c.JSON(fiber.Map{
		"success": fiber.Map{
			"message":     "OK",
			"status_code": 200,
		}})
}

func getPlayerModeration(c *fiber.Ctx) error {
	var mod *models.PlayerModeration
	u := c.Locals("user").(*models.User)

	err := config.DB.Preload(clause.Associations).Where("id = ?", c.Params("id")).Where("source_id = ?", u.ID).First(&mod).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(404).JSON(fiber.Map{
				"error": fiber.Map{
					"message":     "Can't find playerModerationǃ",
					"status_code": 404,
				},
			})
		}

		return c.Status(500).JSON(fiber.Map{
			"error": fiber.Map{
				"message":     err.Error(),
				"status_code": 500,
			},
		})
	}

	return c.JSON(mod.GetAPIPlayerModeration())
}

func deletePlayerModeration(c *fiber.Ctx) error {
	var mod *models.PlayerModeration
	u := c.Locals("user").(*models.User)

	err := config.DB.Preload(clause.Associations).Where("id = ?", c.Params("id")).First(&mod).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(404).JSON(fiber.Map{
				"error": fiber.Map{
					"message":     "Can't find playerModerationǃ",
					"status_code": 404,
				},
			})
		}

		return c.Status(500).JSON(fiber.Map{
			"error": fiber.Map{
				"message":     err.Error(),
				"status_code": 500,
			},
		})
	}

	if mod.SourceID != u.ID {
		return c.Status(403).JSON(fiber.Map{
			"error": fiber.Map{
				"message":     "You definitely can't delete a playerModeration you didn't create",
				"status_code": 403,
			}},
		)
	}

	err = config.DB.Unscoped().Where("id = ?", c.Params("id")).Delete(&models.PlayerModeration{}).Error
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": fiber.Map{
				"message":     err.Error(),
				"status_code": 500,
			},
		})
	}

	return c.JSON(fiber.Map{
		"success": fiber.Map{
			"message":     fmt.Sprintf("PlayerModeration %s removed", c.Params("id")),
			"status_code": 200,
		},
	})
}

func getPlayerModerated(c *fiber.Ctx) error {
	// Stub route. Will likely not be implemented due to it no-longer existing in recent builds of the game.
	return c.JSON([]interface{}{})
}
