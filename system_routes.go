package main

import (
	"github.com/gofiber/fiber/v2"
	"strings"
	"time"
)

func systemRoutes(router *fiber.App) {
	router.Get("/config", getConfig)
	router.Get("/health", getHealth)
	router.Get("/time", getTime)
	router.Put("/logout", putLogout)
	router.Get("/infoPush", getInfoPush)
	router.Get("/visits", getVisits)
}

func getHealth(c *fiber.Ctx) error {
	return c.Status(200).JSON(fiber.Map{
		"status":     "ok",
		"serverName": ApiConfiguration.ServerName.Get(),
	})
}

func getConfig(c *fiber.Ctx) error {
	// Add cookie to response
	c.Cookie(&fiber.Cookie{
		Name:     "apiKey",
		Value:    ApiConfiguration.ApiKey.Get(),
		SameSite: "disabled",
	})
	return c.JSON(NewApiConfigResponse(&ApiConfiguration))
}

func putLogout(c *fiber.Ctx) error {
	c.Cookie(&fiber.Cookie{
		Name:     "auth",
		Value:    "",
		Expires:  time.Now().Add(time.Hour * -1),
		SameSite: "disabled",
	})
	c.Cookie(&fiber.Cookie{
		Name:     "twoFactorAuth",
		Value:    "",
		Expires:  time.Now().Add(time.Hour * -1),
		SameSite: "disabled",
	})
	return c.Status(200).JSON(fiber.Map{
		"status": "ok",
	})
}

func getTime(c *fiber.Ctx) error {
	return c.JSON(time.Now().UTC().Format("2006-01-02T15:04:05+00:00"))
}

func getInfoPush(c *fiber.Ctx) error {
	var toPush []ApiInfoPush
	requiredTags := strings.Split(c.Query("require"), ",")
	includedTags := strings.Split(c.Query("include"), ",")

	for _, push := range ApiConfiguration.InfoPushes.Get() {
		for _, pushed := range toPush {
			if push.Id == pushed.Id {
				break
			}
		}

	nextPush:
		for _, tag := range requiredTags {
			for _, pushTag := range push.Tags {
				if tag == pushTag {
					toPush = append(toPush, push)
					break nextPush
				}
			}
		}

	nextPush2:
		for _, tag := range includedTags {
			for _, pushTag := range push.Tags {
				if tag == pushTag {
					toPush = append(toPush, push)
					break nextPush2
				}
			}
		}
	}

	return c.JSON(toPush)
}

func getVisits(c *fiber.Ctx) error {
	return c.JSON(0)
}
