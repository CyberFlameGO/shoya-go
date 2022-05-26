package main

import "github.com/gofiber/fiber/v2"

func favoriteRoutes(router *fiber.App) {
	favorites := router.Group("/favorite")
	favorites.Get("/", getGroups)
	favorites.Get("/groups", getGroups)
}

func getGroups(c *fiber.Ctx) error {
	return c.JSON([]fiber.Map{})
}
