package main

import (
	"encoding/base64"
	"github.com/gofiber/fiber/v2"
	"gitlab.com/george/shoya-go/config"
	"gitlab.com/george/shoya-go/models"
	"gorm.io/gorm/clause"
	"net/url"
	"strings"
	"time"
)

// ApiKeyMiddleware ensures that the request has a valid API key attached.
// The check order is: query > cookie
// If the API key is valid, the request is allowed to continue.
// If the API key is invalid, the request is denied with a 401 ErrMissingCredentialsResponse.
func ApiKeyMiddleware(c *fiber.Ctx) error {
	apiKey := c.Query("apiKey")
	if apiKey == "" {
		apiKey = c.Cookies("apiKey")
		if apiKey == "" {
			return c.Status(401).JSON(models.ErrMissingCredentialsResponse)
		}
	}

	if apiKey != config.ApiConfiguration.ApiKey.Get() {
		// TODO: Check if the API key is valid against the database if it is not the public key.
		return c.Status(401).JSON(models.ErrInvalidCredentialsResponse)
	}

	c.Locals("apiKey", apiKey)
	c.Cookie(&fiber.Cookie{
		Name:     "apiKey",
		Value:    apiKey,
		SameSite: "disabled",
	})
	return c.Next()
}

// DoLoginMiddleware logs the user in if the request contains a valid HTTP Basic Auth header.
// If the credentials are valid, the request is allowed to continue.
// If the credentials are invalid, the request is denied with a 401 ErrInvalidCredentialsResponse.
// If there is no HTTP Basic Auth header, the request is allowed to continue.
func DoLoginMiddleware(c *fiber.Ctx) error {
	authorizationHeader := c.Get("Authorization")

	if authorizationHeader != "" {
		username, password, err := parseVrchatBasicAuth(authorizationHeader)
		if err != nil {
			return c.Status(401).JSON(models.ErrInvalidCredentialsResponse)
		}

		u := models.User{}
		err = config.DB.Preload(clause.Associations).Where("username = ?", username).First(&u).Error
		if err != nil {
			return c.Status(401).JSON(models.ErrInvalidCredentialsResponse)
		}

		m, err := u.CheckPassword(password)
		if !m || err != nil {
			return c.Status(401).JSON(models.ErrInvalidCredentialsResponse)
		}

		isGameReq := c.Locals("isGameRequest").(bool)
		t, err := models.CreateAuthCookie(&u, c.IP(), isGameReq)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "Failed to create auth cookie"})
		}

		c.Locals("user", &u)
		c.Locals("authCookie", t)
		c.Cookie(&fiber.Cookie{
			Name:     "auth",
			Value:    t,
			Expires:  time.Now().Add(time.Hour * 24),
			SameSite: "disabled",
		})
	}
	return c.Next()
}

// AuthMiddleware ensures that a user is logged in.
// If the user is logged in, the request is allowed to continue.
// If the user is not logged in, the request is denied with a 401 ErrMissingCredentialsResponse.
func AuthMiddleware(c *fiber.Ctx) error {
	authCookie := c.Cookies("auth")
	if authCookie == "" {
		authCookie_, ok := c.Locals("authCookie").(string)
		if !ok || authCookie_ == "" {
			return c.Status(401).JSON(models.ErrMissingCredentialsResponse)
		}
		authCookie = authCookie_ // TODO: Look into less hacky solution -- Currently the variable is locally assigned in the if.
	}

	isGameReq := c.Locals("isGameRequest").(bool)
	uid, err := models.ValidateAuthCookie(authCookie, c.IP(), isGameReq, false)
	if err != nil {
		return c.Status(401).JSON(models.ErrInvalidCredentialsResponse)
	}

	u := models.User{}
	err = config.DB.Preload(clause.Associations).
		Preload("CurrentAvatar.Image").
		Preload("FallbackAvatar").
		Where("id = ?", uid).First(&u).Error
	if err != nil {
		return c.Status(401).JSON(models.ErrInvalidCredentialsResponse)
	}

	c.Locals("authCookie", authCookie)
	c.Locals("user", &u)
	return c.Next()
}

// MfaMiddleware ensures that a user has completed MFA before proceeding.
// If the user has completed MFA (or the user does not have MFA enabled), the request is allowed to continue.
// If the user has not completed MFA, the request is denied with a 401 ErrTwoFactorAuthenticationRequiredResponse.
func MfaMiddleware(c *fiber.Ctx) error {
	if c.Locals("user") == nil {
		// TODO: Throw error; user is not logged in, we should not be here.
		return c.Status(401).JSON(models.ErrMissingCredentialsResponse)
	}

	user := c.Locals("user").(*models.User)
	if !user.MfaEnabled {
		return c.Next()
	}
	if c.Cookies("twoFactorAuth") == "" {
		return c.Status(401).JSON(models.ErrTwoFactorAuthenticationRequiredResponse)
	}

	// TODO: Check if the cookie is valid. If it is, the request is allowed to continue.
	//       If the cookie is invalid, return a 401.

	return c.Next()
}

// IsGameRequestMiddleware uses the `X-Requested-With`, `X-MacAddress`, `X-Client-Version`, `X-Platform`, and `User-Agent`
// headers to identify whether a request is coming from the game client or not.
//
// More specifically; for the request to be marked as a game request:
//  >X-Requested-With	must be present
//  >X-MacAddress		must be present
//  >X-Client-Version	must be present
//  >X-Platform		must be present and one of ["standalonewindows", "android"]
//  >User-Agent		must be present and one of ["VRC.Core.BestHTTP", "Transmtn-Pipeline"]
func IsGameRequestMiddleware(c *fiber.Ctx) error {
	headers := c.GetReqHeaders()
	if shouldDoInDepthClientChecks(c.Path()) {
		// When the client uses the Transmtn-Pipeline client, the below headers are not guaranteed to exist,
		if _, ok := headers["X-Requested-With"]; !ok {
			goto failedChecks
		}

		if _, ok := headers["X-Macaddress"]; !ok {
			goto failedChecks
		}

		if _, ok := headers["X-Client-Version"]; !ok {
			goto failedChecks
		}

		if _, ok := headers["X-Platform"]; !ok || (headers["X-Platform"] != "standalonewindows" && headers["X-Platform"] != "android") {
			goto failedChecks
		}
	}

	if _, ok := headers["User-Agent"]; !ok || (headers["User-Agent"] != "VRC.Core.BestHTTP" && headers["User-Agent"] != "Transmtn-Pipeline") {
		goto failedChecks
	}

	c.Locals("isGameRequest", true)
	return c.Next()

failedChecks:
	c.Locals("isGameRequest", false)
	return c.Next()
}

func parseVrchatBasicAuth(authHeader string) (string, string, error) {
	if authHeader == "" {
		return "", "", nil
	}

	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || parts[0] != "Basic" {
		return "", "", nil
	}

	payload, err := base64.StdEncoding.DecodeString(parts[1])
	if err != nil {
		return "", "", err
	}

	pair := strings.SplitN(string(payload), ":", 2)
	if len(pair) != 2 {
		return "", "", nil
	}

	username, err := url.QueryUnescape(pair[0])
	if err != nil {
		return "", "", err
	}
	password, err := url.QueryUnescape(pair[1])
	if err != nil {
		return "", "", err
	}

	return strings.ToLower(username), password, nil
}

func shouldDoInDepthClientChecks(path string) bool {
	if path == "/auth" ||
		path == "/auth/user" ||
		path == "/config" ||
		path == "/time" ||
		path == "/auth/user/notifications" {
		return false
	}
	return true
}
