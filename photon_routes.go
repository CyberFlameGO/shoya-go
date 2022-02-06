package main

import (
	"github.com/gofiber/fiber/v2"
)

func PhotonRoutes(router *fiber.App) {
	photon := router.Group("/photon")
	photon.Get("/ns", photonSecret, doNsAuth)
}

var PhotonInvalidParametersResponse = fiber.Map{"ResultCode": 3}
var PhotonCustomAuthFailedResponse = fiber.Map{"ResultCode": 2}
var PhotonCustomAuthSuccessResponse = fiber.Map{"ResultCode": 1}

func photonSecret(c *fiber.Ctx) error {
	if c.Query("secret") != ApiConfiguration.PhotonSecret.Get() {
		return c.JSON(fiber.Map{"ResultCode": 3})
	}
	return c.Next()
}

func doNsAuth(c *fiber.Ctx) error {
	t := c.Query("token")
	u := c.Query("user")
	if t == "" || u == "" {
		return c.JSON(PhotonInvalidParametersResponse)
	}

	uid, err := ValidateAuthCookie(t, c.IP(), false, true)
	if err != nil || uid != u {
		return c.JSON(PhotonCustomAuthFailedResponse)
	}

	return c.JSON(PhotonCustomAuthSuccessResponse)
}
