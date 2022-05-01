package main

import (
	"github.com/gofiber/fiber/v2"
	"gitlab.com/george/shoya-go/config"
	"gitlab.com/george/shoya-go/models"
	"gorm.io/gorm"
)

func instanceRoutes(router *fiber.App) {
	instances := router.Group("/instances")
	instances.Get("/:instanceId", ApiKeyMiddleware, AuthMiddleware, getInstance)
	instances.Get("/:instanceId/join", ApiKeyMiddleware, AuthMiddleware, joinInstance)
}

func getInstance(c *fiber.Ctx) error {
	// TODO: Fetch instance data from Redis
	id := c.Params("instanceId")
	i, err := models.ParseLocationString(id)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": fiber.Map{
				"message":     err.Error(),
				"status_code": 500,
			},
		})
	}

	var w models.World
	tx := config.DB.Find(&w).Where("id = ?", i.WorldID)
	if tx.Error != nil {
		if tx.Error == gorm.ErrRecordNotFound {
			return c.Status(404).JSON(models.ErrWorldNotFoundResponse)
		}
		return c.Status(500).JSON(fiber.Map{"error": fiber.Map{"message": tx.Error.Error(), "status_code": 500}})
	}

	instanceResp := fiber.Map{
		"id":         id,
		"location":   id,
		"instanceId": i.LocationString,
		"name":       i.InstanceID,
		"worldId":    i.WorldID,
		"type":       i.InstanceType,
		"ownerId":    i.OwnerID,
		"tags":       []string{},
		"active":     true,  // whether the instance currently has players in it
		"full":       false, // requires redis
		"n_users":    0,     // requires redis
		"capacity":   w.Capacity,
		"platforms": fiber.Map{ // requires redis
			"standalonewindows": 0,
			"android":           0,
		},
		"secureName":       "",       // unknown
		"shortName":        "",       // unknown
		"photonRegion":     i.Region, // todo: api -> photon region conversion -- redis?
		"region":           i.Region,
		"canRequestInvite": i.CanRequestInvite, // todo: presence/friends required
		"permanent":        true,               // unknown -- whether access link is permanent??
		"strict":           i.IsStrict,
	}

	if i.InstanceType != "public" {
		instanceResp[i.InstanceType] = i.OwnerID
	}

	return c.JSON(instanceResp)
}
func joinInstance(c *fiber.Ctx) error {
	var w models.World

	instance, err := models.ParseLocationString(c.Params("instanceId"))
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": fiber.Map{
				"message":     err.Error(),
				"status_code": 500,
			},
		})
	}

	tx := config.DB.Find(&w).Where("id = ?", instance.WorldID)
	if tx.Error != nil {
		if tx.Error == gorm.ErrRecordNotFound {
			return c.Status(404).JSON(models.ErrWorldNotFoundResponse)
		}
		return c.Status(500).JSON(fiber.Map{"error": fiber.Map{"message": tx.Error.Error(), "status_code": 500}})
	}

	if config.ApiConfiguration.DiscoveryServiceEnabled.Get() {
		i := DiscoveryService.GetInstance(instance.ID)
		if i == nil {
			i = DiscoveryService.RegisterInstance(instance.ID, w.Capacity)
		}
	}

	t, err := models.CreateJoinToken(c.Locals("user").(*models.User), &w, c.IP(), instance)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": fiber.Map{"message": err.Error(), "status_code": 500}})
	}

	return c.JSON(fiber.Map{
		"token":   t,
		"version": 1,
	})
}
