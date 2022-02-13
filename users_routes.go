package main

import (
	"fmt"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

func UsersRoutes(router *fiber.App) {
	users := router.Group("/users")
	users.Get("/", GetUsers)
	users.Get("/:id", ApiKeyMiddleware, AuthMiddleware, GetUser)
	users.Get("/:id/feedback", ApiKeyMiddleware, AuthMiddleware, GetUserFeedback)
	users.Post("/", PostUser)
	users.Put("/:id", PutUser)
	users.Delete("/:id", DeleteUser)
}

func GetUsers(c *fiber.Ctx) error {
	return c.Status(fiber.StatusOK).JSON([]User{})
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
