package main

import (
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

func worldsRoutes(app *fiber.App) {
	worlds := app.Group("/worlds")
	worlds.Get("/:id", ApiKeyMiddleware, AuthMiddleware, getWorld)
}

func getWorld(c *fiber.Ctx) error {
	w := World{BaseModel: BaseModel{ID: c.Params("id")}}
	tx := DB.Find(&w)
	if tx.Error != nil {
		if tx.Error == gorm.ErrRecordNotFound {
			return c.Status(404).JSON(ErrWorldNotFoundResponse)
		}
	}

	// aw, err := w.GetAPIWorld()
	aw, err := w.GetAPIWorldWithPackages() // TODO: Flip based on request context. currently like this for testing.
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": fiber.Map{
				"message":     "internal server error while trying to get apiworld",
				"status_code": 500,
			},
		})
	}

	return c.JSON(aw)
}
