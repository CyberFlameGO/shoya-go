package main

import (
	"fmt"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
	"strconv"
)

func usersRoutes(router *fiber.App) {
	users := router.Group("/users")
	users.Get("/", GetUsers)
	users.Get("/:id", ApiKeyMiddleware, AuthMiddleware, GetUser)
	users.Get("/:id/feedback", ApiKeyMiddleware, AuthMiddleware, GetUserFeedback)
	users.Post("/", PostUser)
	users.Put("/:id", PutUser)
	users.Delete("/:id", DeleteUser)
}

func GetUsers(c *fiber.Ctx) error {
	var users []User
	var rUsers []*APILimitedUser
	var searchTerm string
	var searchDeveloperType string
	var searchOffset = 0
	var numberOfUsersToSearch = 60

	tx := DB.Model(User{}).
		Preload("CurrentAvatar.Image")

	// Query parameter setup
	if c.Query("n") != "" {
		atoi, err := strconv.Atoi(c.Query("n"))
		if err != nil {
			goto badRequest
		}

		if atoi < 1 || atoi > 100 {
			goto badRequest
		}

		numberOfUsersToSearch = atoi
	}

	if c.Query("offset") != "" {
		atoi, err := strconv.Atoi(c.Query("offset"))
		if err != nil {
			goto badRequest
		}

		if atoi < 0 {
			goto badRequest
		}

		searchOffset = atoi
	}

	if c.Query("developerType") != "" {
		searchDeveloperType = c.Query("developerType")
	}

	if c.Query("search") != "" {
		searchTerm = "%" + c.Query("search") + "%" // TODO: FTS
	}

	if searchDeveloperType != "" {
		tx.Where("developer_type = ?", searchDeveloperType)
	}

	if searchTerm != "" {
		tx.Where("display_name ILIKE ?", searchTerm)
	}

	tx.Limit(numberOfUsersToSearch).Offset(searchOffset)
	tx.Find(&users)

	for _, user := range users {
		lu := user.GetAPILimitedUser(false, false) // TODO: Friendship check.
		rUsers = append(rUsers, lu)
	}
	return c.Status(fiber.StatusOK).JSON(rUsers)

badRequest:
	return c.Status(400).JSON(fiber.Map{
		"error": fiber.Map{
			"message":     "Bad request",
			"status_code": 400,
		},
	})
}

func GetUser(c *fiber.Ctx) error {
	cu, ok := c.Locals("user").(*User)
	if !ok {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Couldn't retrieve current user.",
		})
	}

	uid := c.Params("id")
	if cu.ID == uid {
		return c.Status(fiber.StatusOK).JSON(cu.GetAPICurrentUser())
	}

	ru := &User{}
	tx := DB.Where("id = ?", uid).
		Preload("CurrentAvatar.Image").
		Preload("FallbackAvatar").
		Find(&ru)

	if tx.Error != nil {
		if tx.Error == gorm.ErrRecordNotFound {
			return c.Status(404).JSON(fiber.Map{
				"error": fiber.Map{
					"message":     fmt.Sprintf("User %s not found", uid),
					"status_code": 404,
				},
			})
		}
	}

	return c.Status(fiber.StatusOK).JSON(ru.GetAPIUser(false, false)) // TODO: Implement friendship system. Check friendship.
}

func PostUser(c *fiber.Ctx) error {
	return c.Status(fiber.StatusOK).JSON(User{})
}

func PutUser(c *fiber.Ctx) error {
	return c.Status(fiber.StatusOK).JSON(User{})
}

func DeleteUser(c *fiber.Ctx) error {
	return c.Status(fiber.StatusOK).JSON(User{})
}

func GetUserFeedback(c *fiber.Ctx) error {
	return c.JSON([]interface{}{})
}
