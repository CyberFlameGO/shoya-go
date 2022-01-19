package main

import "github.com/gofiber/fiber/v2"

func authRoutes(router *fiber.App) {
	router.Get("/auth", AuthMiddleware, getAuth)
	router.Get("/auth/user", DoLoginMiddleware, AuthMiddleware, getSelf)
}

func getAuth(c *fiber.Ctx) error {
	return c.Status(200).JSON(fiber.Map{
		"ok":    true,
		"token": c.Locals("auth_cookie").(string),
	})
}

func getSelf(c *fiber.Ctx) error {
	return c.Status(200).JSON(fiber.Map{
		"ok": true,
	})
}
