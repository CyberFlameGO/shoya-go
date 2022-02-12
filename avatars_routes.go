package main

import (
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func AvatarsRoutes(router *fiber.App) {
	avatars := router.Group("/avatars")
	avatars.Get("/", ApiKeyMiddleware, AuthMiddleware, getAvatars)
	avatars.Get("/favorites", ApiKeyMiddleware, AuthMiddleware, getAvatarFavorites)
	avatars.Get("/licensed", ApiKeyMiddleware, AuthMiddleware, getLicensedAvatars)
	avatars.Get("/:id", ApiKeyMiddleware, AuthMiddleware, getAvatar)
}

func getAvatars(c *fiber.Ctx) error {
	return c.Status(501).JSON([]fiber.Map{})
}

func getAvatarFavorites(c *fiber.Ctx) error {
	return c.Status(501).JSON([]fiber.Map{})
}

func getLicensedAvatars(c *fiber.Ctx) error {
	return c.Status(501).JSON([]fiber.Map{})
}

func getAvatar(c *fiber.Ctx) error {
	var a Avatar
	tx := DB.Preload(clause.Associations).Preload("UnityPackages.File").Model(&Avatar{}).Where("id = ?", c.Params("id")).First(&a)
	if tx.Error != nil {
		if tx.Error == gorm.ErrRecordNotFound {
			return c.Status(404).JSON(ErrWorldNotFoundResponse)
		}
	}

	// aa, err := w.GetAPIAvatar()
	aa, err := a.GetAPIAvatarWithPackages() // TODO: Flip based on request context. currently like this for testing.
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": fiber.Map{
				"message":     "internal server error while trying to get apiavatar",
				"status_code": 500,
			},
		})
	}

	return c.JSON(aa)
}
