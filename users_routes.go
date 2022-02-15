package main

import (
	"fmt"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
	"strconv"
)

func usersRoutes(router *fiber.App) {
	// VRChat is inconsistent with how the do routing. Some are under /user, others /users.
	user := router.Group("/user")
	user.Get("/:id/friendStatus", ApiKeyMiddleware, AuthMiddleware, getUserFriendStatus)

	users := router.Group("/users")
	users.Get("/", getUsers)
	users.Get("/:id", ApiKeyMiddleware, AuthMiddleware, getUser)
	users.Get("/:id/feedback", ApiKeyMiddleware, AuthMiddleware, getUserFeedback)
	users.Post("/", postUser)
	users.Put("/:id", putUser)
	users.Delete("/:id", deleteUser)
}

func getUsers(c *fiber.Ctx) error {
	var users []User
	var rUsers = make([]*APILimitedUser, 0)
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

func getUser(c *fiber.Ctx) error {
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

func postUser(c *fiber.Ctx) error {
	return c.Status(fiber.StatusOK).JSON(User{})
}

func putUser(c *fiber.Ctx) error {
	cu := c.Locals("user").(*User)

	if c.Params("id") != cu.ID && !cu.IsStaff() {
		return c.Status(403).JSON(fiber.Map{
			"error": fiber.Map{
				"message":     "You're not allowed to update another user's profile",
				"status_code": 403,
			},
		})
	}
	return c.Status(fiber.StatusOK).JSON(cu.GetAPICurrentUser())
}

func deleteUser(c *fiber.Ctx) error {
	return c.Status(fiber.StatusOK).JSON(User{})
}

func getUserFeedback(c *fiber.Ctx) error {
	return c.JSON([]interface{}{})
}

func getUserFriendStatus(c *fiber.Ctx) error {
	// TODO: Implement friendships.
	return c.JSON(fiber.Map{
		"incomingRequest": false,
		"isFriend":        false,
		"outgoingRequest": false,
	})
}
