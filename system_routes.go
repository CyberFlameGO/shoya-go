package main

import "github.com/gofiber/fiber/v2"

func systemRoutes(router *fiber.App) {
	router.Get("/config", GetConfig)
	router.Post("/config", func(c *fiber.Ctx) error {
		// Convert request body to map[string]interface
		var body map[string]interface{}
		if err := c.BodyParser(&body); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(err)
		}

		// Update configuration
		if err := ApiConfiguration.Update(body); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(err)
		}

		return c.Next()
	}, GetConfig)
}

func GetConfig(c *fiber.Ctx) error {
	// Add cookie to response
	c.Cookie(&fiber.Cookie{
		Name:  "apiKey",
		Value: ApiConfiguration.ApiKey.Get(),
	})
	return c.JSON(NewApiConfigResponse(&ApiConfiguration))
}
