package main

import (
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

func authRoutes(router *fiber.App) {
	router.Get("/auth", ApiKeyMiddleware, AuthMiddleware, getAuth)
	router.Get("/auth/exists", ApiKeyMiddleware, getExists)
	router.Get("/auth/user", ApiKeyMiddleware, DoLoginMiddleware, AuthMiddleware, getSelf)
	router.Get("/auth/user/notifications", ApiKeyMiddleware, AuthMiddleware, getNotifications)
}

func getAuth(c *fiber.Ctx) error {
	return c.Status(200).JSON(fiber.Map{
		"ok":    true,
		"token": c.Locals("authCookie").(string),
	})
}

func getExists(c *fiber.Ctx) error {
	var u User
	var exists = true

	tx := DB.Where("username = ?", c.Query("username")).
		Or("display_name = ?", c.Query("displayName")).
		Or("email = ?", c.Query("email")).Select("id").First(&u)

	if tx.Error != nil {
		if tx.Error == gorm.ErrRecordNotFound {
			exists = false
		}
	}
	return c.Status(200).JSON(fiber.Map{
		"userExists": exists,
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
