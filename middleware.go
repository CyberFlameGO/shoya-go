package main

import "github.com/gofiber/fiber/v2"

// ApiKeyMiddleware ensures that the request has a valid API key attached.
// The check order is: query > cookie
// If the API key is valid, the request is allowed to continue.
// If the API key is invalid, the request is denied with a 401 MissingCredentialsResponse.
func ApiKeyMiddleware(c *fiber.Ctx) error {
	apiKey := c.Query("apiKey")
	if apiKey == "" {
		apiKey = c.Cookies("apiKey")
	}

	if apiKey != ApiConfiguration.ApiKey.Get() {
		return c.Status(401).JSON(MissingCredentialsResponse)
	}

	return c.Next()
}

// DoLoginMiddleware logs the user in if the request contains a valid HTTP Basic Auth header.
// If the credentials are valid, the request is allowed to continue.
// If the credentials are invalid, the request is denied with a 401 InvalidCredentialsResponse.
func DoLoginMiddleware(c *fiber.Ctx) error {
	authorizationHeader := c.Get("Authorization")

	if authorizationHeader != "" {
		// TODO: URL-decode the username and password (VRChat specifically encodes them)
		//       and then use them to authenticate the user, attaching ana auth cookie to the request & response.
		//       If the user is authenticated, the request is allowed to continue.
		//       If the user is not authenticated, the request is denied with a 401 InvalidCredentialsResponse.
	}
	return c.Next()
}

// AuthMiddleware ensures that a user is logged in.
// If the user is logged in, the request is allowed to continue.
// If the user is not logged in, the request is denied with a 401 MissingCredentialsResponse.
func AuthMiddleware(c *fiber.Ctx) error {
	if c.Cookies("auth") == "" {
		return c.Status(401).JSON(MissingCredentialsResponse)
	}
	return c.Next()
}

// MfaMiddleware ensures that a user has completed MFA before proceeding.
// If the user has completed MFA, the request is allowed to continue.
// If the user has not completed MFA, the request is denied with a 401 TwoFactorAuthenticationRequiredResponse.
func MfaMiddleware(c *fiber.Ctx) error {
	if c.Locals("user") == nil {
		// TODO: Throw error; user is not logged in, we should not be here.
		return c.Status(401).JSON(MissingCredentialsResponse)
	}
	if c.Cookies("twoFactorAuth") == "" {
		return c.Status(401).JSON(TwoFactorAuthenticationRequiredResponse)
	}

	// TODO: Check if the user has enabled 2FA, if so, check if the cookie is valid.
	//       If the cookie is invalid, return a 401.

	return c.Next()
}
