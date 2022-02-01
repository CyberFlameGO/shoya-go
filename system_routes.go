package main

import (
	"github.com/gofiber/fiber/v2"
	"time"
)

func systemRoutes(router *fiber.App) {
	router.Get("/config", GetConfig)
	router.Get("/health", GetHealth)
	router.Put("/logout", Logout)
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

func Logout(c *fiber.Ctx) error {
	c.Cookie(&fiber.Cookie{
		Name:    "auth",
		Value:   "",
		Expires: time.Now().Add(time.Hour * -1),
	})
	c.Cookie(&fiber.Cookie{
		Name:    "twoFactorAuth",
		Value:   "",
		Expires: time.Now().Add(time.Hour * -1),
	})
	return c.Status(200).JSON(fiber.Map{
		"status": "ok",
	})
}
