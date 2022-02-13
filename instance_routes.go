package main

import (
	"github.com/gofiber/fiber/v2"
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
	var w World
	// TODO: parseLocationString
	tx := DB.Find(&w).Where("id = ?", strings.Split(c.Params("instanceId"), ":")[0])
	if tx.Error != nil {
		return c.Status(500).JSON(fiber.Map{"error": fiber.Map{"message": "shit broke", "status_code": 500}})
	}

	t, err := CreateJoinToken(c.Locals("user").(*User), &w, c.IP(), c.Params("instanceId"))
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": fiber.Map{"message": "shit broke", "status_code": 500}})
	}

	return c.JSON(fiber.Map{
		"token":   t,
		"version": 1,
	})

}
