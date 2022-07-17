package api

import "github.com/gofiber/fiber/v2"

func messageRoutes(app *fiber.App) {
	message := app.Group("/message")
	message.Get("/:userId/:msgType", getMessages)
}

func getMessages(c *fiber.Ctx) error {
	return c.JSON([]fiber.Map{})
}
