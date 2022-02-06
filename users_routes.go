package main

import "github.com/gofiber/fiber/v2"

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

	return c.Status(fiber.StatusOK).JSON(User{})
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
