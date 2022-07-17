package api

import (
	"encoding/base64"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"gitlab.com/george/shoya-go/config"
	"gitlab.com/george/shoya-go/models"
	"gorm.io/gorm/clause"
	"math/rand"
	"net/url"
	"strings"
	"time"
)

// AddXPoweredByHeader adds an `X-Powered-By` header to every response
// with a randomly-selected string from the models.XPoweredByHeaders slice.
func AddXPoweredByHeader(c *fiber.Ctx) error {
	c.Set("X-Powered-By", models.XPoweredByHeaders[rand.Intn(len(models.XPoweredByHeaders))]) // #nosec skipcq
	return c.Next()
}

// ApiKeyMiddleware ensures that the request has a valid API key attached.
func ApiKeyMiddleware(c *fiber.Ctx) error {
	var apiKey string

	if apiKey = c.Query("apiKey"); apiKey == "" {
		if apiKey = c.Cookies("apiKey"); apiKey == "" {
			return c.Status(401).JSON(models.ErrMissingCredentialsResponse)
		}
	}

	if apiKey != config.ApiConfiguration.ApiKey.Get() {
		return c.Status(401).JSON(models.ErrInvalidCredentialsResponse)
	}

	c.Locals("apiKey", apiKey)
	c.Cookie(&fiber.Cookie{Name: "apiKey", Value: apiKey, SameSite: "disabled"})
	return c.Next()
}

// LoginMiddleware logs a user in if there's an Authorization header present in the request.
func LoginMiddleware(c *fiber.Ctx) error {
	authorizationHeader := c.Get("Authorization")

	if authorizationHeader != "" {
		var username string
		var password string
		var err error
		var u *models.User
		var m bool // Password matched
		var banned bool
		var moderation *models.Moderation
		var isGameReq bool
		var ok bool
		var t string

		username, password, err = parseVrchatBasicAuth(authorizationHeader)
		if err != nil {
			return c.Status(401).JSON(models.ErrInvalidCredentialsResponse)
		}

		if u, err = models.GetUserByUsernameOrEmail(username); err != nil {
			return c.Status(401).JSON(models.ErrInvalidCredentialsResponse)
		}

		m, err = u.CheckPassword(password)
		if !m || err != nil {
			return c.Status(401).JSON(models.ErrInvalidCredentialsResponse)
		}

		if banned, moderation = u.IsBanned(); banned {
			return produceBanResponse(c, u, moderation)
		}

		if isGameReq, ok = c.Locals("isGameRequest").(bool); !ok {
			isGameReq = false
		}

		if t, err = models.CreateAuthCookie(u, c.IP(), isGameReq); err != nil {
			return c.Status(500).JSON(models.MakeErrorResponse("failed to create auth cookie", 500))
		}

		u.LastLogin = time.Now().Unix()
		config.DB.Omit(clause.Associations).Save(&u)

		c.Locals("user", u)
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

func AuthMiddleware(c *fiber.Ctx) error {
	var authCookie string
	var ok bool
	var isGameReq bool
	var uid string
	var err error
	var u *models.User
	var banned bool
	var moderation *models.Moderation

	if authCookie = c.Cookies("auth"); authCookie == "" {
		if authCookie, ok = c.Locals("authCookie").(string); !ok || authCookie == "" {
			return c.Status(401).JSON(models.ErrMissingCredentialsResponse)
		}
	}

	if isGameReq, ok = c.Locals("isGameRequest").(bool); !ok {
		isGameReq = false
	}

	if uid, err = models.ValidateAuthCookie(authCookie, c.IP(), isGameReq, false); err != nil {
		return c.Status(401).JSON(models.ErrInvalidCredentialsResponse)
	}

	if u, err = models.GetUserById(uid); err != nil {
		return c.Status(401).JSON(models.ErrInvalidCredentialsResponse)
	}

	if banned, moderation = u.IsBanned(); banned {
		return produceBanResponse(c, u, moderation)
	}

	c.Locals("authCookie", authCookie)
	c.Locals("user", u)
	return c.Next()

}
func MfaMiddleware() {} // later

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
	var ok bool

	headers := c.GetReqHeaders()
	if shouldDoInDepthClientChecks(c.Path()) {
		// When the client uses the Transmtn-Pipeline client, the below headers are not guaranteed to exist,
		if _, ok = headers["X-Requested-With"]; !ok {
			goto failedChecks
		}

		if _, ok = headers["X-Macaddress"]; !ok {
			goto failedChecks
		}

		if _, ok = headers["X-Client-Version"]; !ok {
			if _, ok = headers["X-Unity-Version"]; !ok {
				goto failedChecks
			}
		}

		if _, ok = headers["X-Platform"]; !ok || (headers["X-Platform"] != "standalonewindows" && headers["X-Platform"] != "android") {
			goto failedChecks
		}
	}

	if _, ok = headers["User-Agent"]; !ok || (headers["User-Agent"] != "VRC.Core.BestHTTP" && headers["User-Agent"] != "Transmtn-Pipeline") {
		goto failedChecks
	}

	c.Locals("isGameRequest", true)
	return c.Next()

failedChecks:
	c.Locals("isGameRequest", false)
	return c.Next()
}

func AdminMiddleware(c *fiber.Ctx) error {
	var u = c.Locals("user").(*models.User)

	if !u.IsStaff() {
		return c.Status(401).JSON(models.ErrMissingAdminCredentialsResponse)
	}

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
		strings.HasPrefix(path, "/auth/user/notifications") {
		return false
	}
	return true
}

func produceBanResponse(c *fiber.Ctx, u *models.User, moderation *models.Moderation) error {
	var r fiber.Map
	if moderation == nil {
		return c.Status(403).SendString("Ban") // Not even joking, from what I can recall, this is what Official actually responds with when your moderation can't be found/is cached.
	}

	if moderation.ExpiresAt == 0 {
		r = models.MakeErrorResponse(fmt.Sprintf("Account permanently banned: %s", moderation.Reason), 403)
		r["target"] = u.Username
		r["reason"] = moderation.Reason
		r["isPermanent"] = true

		return c.Status(403).JSON(r)
	}

	banExpiresAt := time.Unix(moderation.ExpiresAt, 0)
	r = models.MakeErrorResponse(fmt.Sprintf("Account temporarily suspended until %s (in %d days): %s", banExpiresAt.Format("Jan 02, 2006 15:04 MST"), int(banExpiresAt.Sub(time.Now().UTC()).Hours()/24), moderation.Reason), 403)
	r["target"] = u.Username
	r["reason"] = moderation.Reason
	r["expires"] = banExpiresAt.Format(time.RFC3339)
	r["isPermanent"] = false

	return c.Status(403).JSON(r)
}
