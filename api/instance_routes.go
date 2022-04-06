package main

import (
	"github.com/gofiber/fiber/v2"
	"gitlab.com/george/shoya-go/config"
	"gitlab.com/george/shoya-go/models"
	"strings"
)

func instanceRoutes(router *fiber.App) {
	instances := router.Group("/instances")
	instances.Get("/:instanceId", ApiKeyMiddleware, AuthMiddleware, getInstance)
	instances.Get("/:instanceId/join", ApiKeyMiddleware, AuthMiddleware, joinInstance)
}

func getInstance(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{})
}
func joinInstance(c *fiber.Ctx) error {
	var w models.World
	// TODO: parseLocationString
	tx := config.DB.Find(&w).Where("id = ?", strings.Split(c.Params("instanceId"), ":")[0])
	if tx.Error != nil {
		return c.Status(500).JSON(fiber.Map{"error": fiber.Map{"message": "shit broke", "status_code": 500}})
	}

	t, err := models.CreateJoinToken(c.Locals("user").(*models.User), &w, c.IP(), c.Params("instanceId"))
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": fiber.Map{"message": "shit broke", "status_code": 500}})
	}

	return c.JSON(fiber.Map{
		"token":   t,
		"version": 1,
	})

}
