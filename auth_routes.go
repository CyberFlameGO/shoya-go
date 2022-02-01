package main

import (
	"fmt"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

func authRoutes(router *fiber.App) {
	router.Get("/auth", ApiKeyMiddleware, AuthMiddleware, getAuth)
	router.Get("/auth/exists", ApiKeyMiddleware, getExists)
	router.Post("/auth/register", ApiKeyMiddleware, postRegister)
	router.Get("/auth/user", ApiKeyMiddleware, DoLoginMiddleware, AuthMiddleware, getSelf)
	router.Get("/auth/user/notifications", ApiKeyMiddleware, AuthMiddleware, getNotifications)
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
	var u User
	var exists = true

	tx := DB.Where("username = ?", c.Query("username")).
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

func postRegister(c *fiber.Ctx) error {
	var r RegisterRequest
	var _u User
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

	tx := DB.Where("username = ?", c.Query("username")).
		Or("display_name = ?", c.Query("username")).
		Or("email = ?", c.Query("email")).First(&_u)

	if tx.Error != gorm.ErrRecordNotFound {
		return c.Status(400).JSON(fiber.Map{
			"ok": false,
			"error": fiber.Map{
				"message": "Username, display name, or email already exists",
			},
		})
	}

	u := NewUser(r.Username, r.Username, r.Email, r.Password)
	tx = DB.Create(&u)
	if tx.Error != nil {
		fmt.Println(tx.Error)
		return c.Status(500).JSON(fiber.Map{
			"ok": false,
			"error": fiber.Map{
				"message": "Internal server error",
			},
		})
	}

	return c.Status(200).JSON(fiber.Map{
		"ok": true,
	})
}

func getSelf(c *fiber.Ctx) error {
	u := c.Locals("user").(*User)

	return c.Status(200).JSON(u)
}

func getNotifications(c *fiber.Ctx) error {
	u := c.Locals("user").(*User)

	return c.Status(200).JSON(fiber.Map{
		"user": u,
	}) // TODO: implement
}
