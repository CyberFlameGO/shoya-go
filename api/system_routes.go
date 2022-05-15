package main

import (
	"github.com/gofiber/fiber/v2"
	"gitlab.com/george/shoya-go/config"
	"strings"
	"time"
)

func systemRoutes(router *fiber.App) {
	router.Get("/ping", getPing)
	router.Get("/health", getHealth)
	router.Get("/time", getTime)
	router.Get("/config", getConfig)

	router.Get("/infoPush", ApiKeyMiddleware, AuthMiddleware, getInfoPush)
	router.Put("/logout", ApiKeyMiddleware, AuthMiddleware, putLogout)

	router.Get("/visits", getVisits)
	router.Put("/visits", ApiKeyMiddleware, AuthMiddleware, putVisits)

	router.Put("/joins", ApiKeyMiddleware, AuthMiddleware, putJoins)

	router.Get("/m_autoConfig", getAutoConfig)
}

func getHealth(c *fiber.Ctx) error {
	return c.Status(200).JSON(fiber.Map{
		"status":     "ok",
		"serverName": config.ApiConfiguration.ServerName.Get(),
	})
}

func getConfig(c *fiber.Ctx) error {
	// Add cookie to response
	c.Cookie(&fiber.Cookie{
		Name:     "apiKey",
		Value:    config.ApiConfiguration.ApiKey.Get(),
		SameSite: "disabled",
	})
	return c.JSON(config.NewApiConfigResponse(&config.ApiConfiguration))
}

func getPing(c *fiber.Ctx) error {
	return c.JSON("pong")
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
	return c.JSON(time.Now().UTC().Format(time.RFC3339))
}

func getInfoPush(c *fiber.Ctx) error {
	//goland:noinspection GoPreferNilSlice
	toPush := []config.ApiInfoPush{}
	requiredTags := strings.Split(c.Query("require"), ",")
	includedTags := strings.Split(c.Query("include"), ",")

	for _, push := range config.ApiConfiguration.InfoPushes.Get() {
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

// TODO: Implement active user count
func getVisits(c *fiber.Ctx) error {
	return c.JSON(0)
}

// TODO: This is a blocker due to requiring the presence service
func putVisits(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"success": fiber.Map{
			"message":     "User pinged room",
			"status_code": 200,
		},
	})
}

// TODO: This is a blocker due to requiring the presence service
func putJoins(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"success": fiber.Map{
			"message":     "User joined room",
			"status_code": 200,
		},
	})
}

func getAutoConfig(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"apiUrl":         config.ApiConfiguration.AutoConfigApiUrl.Get(),
		"websocketUrl":   config.ApiConfiguration.AutoConfigWebsocketUrl.Get(),
		"nameServerHost": config.ApiConfiguration.AutoConfigNameServerHost.Get(),
	})
}
