package main

import "github.com/gofiber/fiber/v2"

func UsersRoutes(router *fiber.App) {
	users := router.Group("/users")
	users.Get("/", GetUsers)
	users.Get("/:id", GetUser)
	users.Post("/", PostUser)
	users.Put("/:id", PutUser)
	users.Delete("/:id", DeleteUser)
}

func GetUsers(c *fiber.Ctx) error {
	return c.Status(fiber.StatusOK).JSON([]User{})
}

func GetUser(c *fiber.Ctx) error {
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
