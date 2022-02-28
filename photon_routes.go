package main

import (
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm/clause"
	"strconv"
	"time"
)

func photonRoutes(router *fiber.App) {
	photon := router.Group("/photon")
	photon.Get("/ns", photonSecret, doNsAuth)
	photon.Get("/validateJoin", photonSecret, doJoinTokenValidation)
	photon.Get("/user", photonSecret, doPropertyUpdate)
	photon.Get("/getConfig", photonSecret, getPhotonConfig)
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
	tx := DB.Model(&User{}).Preload(clause.Associations).
		Preload("CurrentAvatar.Image").
		Preload("FallbackAvatar.Image").
		Preload("CurrentAvatar.UnityPackages.File").
		Preload("FallbackAvatar.UnityPackages.File").
		Where("id = ?", claims.UserId).First(&u)
	if tx.Error != nil {
		return c.JSON(PhotonValidateJoinJWTResponse{Valid: false})
	}

	r := PhotonValidateJoinJWTResponse{
		Time:  strconv.Itoa(int(time.Now().Unix())),
		Valid: true,
		IP:    claims.IP,
	}
	r.FillFromUser(&u)
	return c.JSON(r)
}

func doPropertyUpdate(c *fiber.Ctx) error {
	var uid = c.Query("userId")
	var u User
	tx := DB.Model(&User{}).Preload(clause.Associations).
		Preload("CurrentAvatar.Image").
		Preload("FallbackAvatar.Image").
		Preload("CurrentAvatar.UnityPackages.File").
		Preload("FallbackAvatar.UnityPackages.File").
		Where("id = ?", uid).First(&u)
	if tx.Error != nil {
		return c.JSON(PhotonValidateJoinJWTResponse{Valid: false})
	}

	r := PhotonValidateJoinJWTResponse{
		Time:  strconv.Itoa(int(time.Now().Unix())),
		Valid: true,
		IP:    "notset",
	}
	r.FillFromUser(&u)
	return c.JSON(r)
}

func getPhotonConfig(c *fiber.Ctx) error {
	// TODO: Make this dynamic.
	return c.JSON(&PhotonConfig{
		MaxAccountsPerIPAddress: 5,
		RateLimitList: map[int]int{
			1:   60,
			3:   5,
			4:   200,
			5:   50,
			6:   400,
			7:   500,
			8:   1,
			9:   75,
			33:  2,
			40:  1,
			42:  1, // ?
			202: 1,
			209: 20,
			210: 90,
		},
		RateLimitUnknownBool: true,
	})
}
