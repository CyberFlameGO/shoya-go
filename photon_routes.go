package main

import (
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm/clause"
)

func PhotonRoutes(router *fiber.App) {
	photon := router.Group("/photon")
	photon.Get("/ns", photonSecret, doNsAuth)
	photon.Get("/validateJoin", photonSecret, doJoinTokenValidation)
	photon.Get("/user", photonSecret, doPropertyUpdate)
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

func doJoinTokenValidation(c *fiber.Ctx) error {
	t := c.Query("jwt")
	l := c.Query("roomId")
	if t == "" || l == "" {
		return c.JSON(PhotonValidateJoinJWTResponse{Valid: false})
	}

	claims, err := ValidateJoinToken(t)
	if err != nil {
		return c.JSON(PhotonValidateJoinJWTResponse{Valid: false})
	}

	if claims.Location != l {
		return c.JSON(PhotonValidateJoinJWTResponse{Valid: false})
	}

	var u User
	tx := DB.Model(&User{}).Preload(clause.Associations).Preload("CurrentAvatar.UnityPackages.File").Preload("FallbackAvatar.UnityPackages.File").
		Where("id = ?", claims.UserId).First(&u)
	if tx.Error != nil {
		return c.JSON(PhotonValidateJoinJWTResponse{Valid: false})
	}

	r := PhotonValidateJoinJWTResponse{
		Valid: true,
		IP:    claims.IP,
	}
	r.FillFromUser(&u)
	return c.JSON(r)
}

func doPropertyUpdate(c *fiber.Ctx) error {
	var uid = c.Query("userId")
	var u User
	tx := DB.Model(&User{}).Preload(clause.Associations).Preload("CurrentAvatar.UnityPackages.File").Preload("FallbackAvatar.UnityPackages.File").
		Where("id = ?", uid).First(&u)
	if tx.Error != nil {
		return c.JSON(PhotonValidateJoinJWTResponse{Valid: false})
	}

	r := PhotonValidateJoinJWTResponse{
		Valid: true,
		IP:    "notset",
	}
	r.FillFromUser(&u)
	return c.JSON(r)
}
