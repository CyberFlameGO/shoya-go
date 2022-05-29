package main

import (
	"github.com/gofiber/fiber/v2"
	"gitlab.com/george/shoya-go/config"
	"gitlab.com/george/shoya-go/models"
	"gorm.io/gorm"
)

func instanceRoutes(router *fiber.App) {
	instances := router.Group("/instances", ApiKeyMiddleware, AuthMiddleware)
	instances.Get("/:instanceId", getInstance)
	instances.Get("/:instanceId/join", joinInstance)
}

// getInstance | GET /instances/:instanceId
// Returns an instance.
func getInstance(c *fiber.Ctx) error {
	var instance *models.WorldInstance
	id := c.Params("instanceId")
	i, err := models.ParseLocationString(id)
	if err != nil {
		return c.Status(500).JSON(models.MakeErrorResponse(err.Error(), 500))
	}

	if config.ApiConfiguration.DiscoveryServiceEnabled.Get() {
		instance = DiscoveryService.GetInstance(id)
		if instance == nil {
			return c.Status(404).JSON(models.ErrInstanceNotFoundResponse)
		}
	}

	var w models.World
	tx := config.DB.Find(&w).Where("id = ?", i.WorldID)
	if tx.Error != nil {
		if tx.Error == gorm.ErrRecordNotFound {
			return c.Status(404).JSON(models.ErrWorldNotFoundResponse)
		}
		return c.Status(500).JSON(models.MakeErrorResponse(tx.Error.Error(), 500))
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
		"active":     true,
		"full":       instance.OverCapacity,
		"n_users":    instance.PlayerCount.Total, // requires redis
		"capacity":   w.Capacity,
		"platforms": fiber.Map{
			"standalonewindows": instance.PlayerCount.PlatformWindows,
			"android":           instance.PlayerCount.PlatformAndroid,
		},
		"secureName":       "", // unknown
		"shortName":        "", // unknown
		"photonRegion":     i.Region,
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

// joinInstance | GET /instances/:instanceId/join
// Generates and returns a room join token.
func joinInstance(c *fiber.Ctx) error {
	var w models.World

	instance, err := models.ParseLocationString(c.Params("instanceId"))
	if err != nil {
		return c.Status(500).JSON(models.MakeErrorResponse(err.Error(), 500))
	}

	tx := config.DB.Find(&w).Where("id = ?", instance.WorldID)
	if tx.Error != nil {
		if tx.Error == gorm.ErrRecordNotFound {
			return c.Status(404).JSON(models.ErrWorldNotFoundResponse)
		}
		return c.Status(500).JSON(models.MakeErrorResponse(tx.Error.Error(), 500))
	}

	if config.ApiConfiguration.DiscoveryServiceEnabled.Get() {
		if DiscoveryService.GetInstance(instance.ID) == nil {
			DiscoveryService.RegisterInstance(instance.ID, w.Capacity)
		}
	}

	t, err := models.CreateJoinToken(c.Locals("user").(*models.User), &w, c.IP(), instance)
	if err != nil {
		return c.Status(500).JSON(models.MakeErrorResponse(err.Error(), 500))
	}

	return c.JSON(fiber.Map{
		"canModerateInstance": false, // So, err… the official API also returns this as false at all times, because it's not implemented on their end.
		"token":               t,
		"version":             1,
	})
}
