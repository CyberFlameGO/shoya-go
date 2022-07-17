package api

import "github.com/gofiber/fiber/v2"

func favoriteRoutes(router *fiber.App) {
	favorite := router.Group("/favorite")
	favorite.Get("/", getGroups)
	favorite.Get("/groups", getGroups)
	favorites := router.Group("/favorites")
	favorites.Get("/", getGroups)
}

func getGroups(c *fiber.Ctx) error {
	return c.JSON([]struct{}{})
}
