package main

import "github.com/gofiber/fiber/v2"

func systemRoutes(router *fiber.App) {
	router.Get("/config", GetConfig)
	router.Get("/health", GetHealth)
}

func GetHealth(c *fiber.Ctx) error {
	return c.Status(200).JSON(fiber.Map{
		"status":     "ok",
		"serverName": ApiConfiguration.ServerName.Get(),
	})
}

func GetConfig(c *fiber.Ctx) error {
	// Add cookie to response
	c.Cookie(&fiber.Cookie{
		Name:  "apiKey",
		Value: ApiConfiguration.ApiKey.Get(),
	})
	return c.JSON(NewApiConfigResponse(&ApiConfiguration))
}
