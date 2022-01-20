package main

import "github.com/gofiber/fiber/v2"

func authRoutes(router *fiber.App) {
	router.Get("/auth", ApiKeyMiddleware, AuthMiddleware, getAuth)
	router.Get("/auth/user", ApiKeyMiddleware, DoLoginMiddleware, AuthMiddleware, getSelf)

}

func getAuth(c *fiber.Ctx) error {
	return c.Status(200).JSON(fiber.Map{
		"ok":    true,
		"token": c.Locals("authCookie").(string),
	})
}

func getSelf(c *fiber.Ctx) error {
	u := c.Locals("user").(*User)

	return c.Status(200).JSON(u)
}
