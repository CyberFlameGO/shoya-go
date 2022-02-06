package main

import "github.com/gofiber/fiber/v2"

var ErrMissingCredentialsResponse = fiber.Map{
	"error": fiber.Map{
		"message":     "Missing Credentials!",
		"status_code": 401,
	},
}

var ErrInvalidCredentialsResponse = fiber.Map{
	"error": fiber.Map{
		"message":     "Invalid Credentials!",
		"status_code": 401,
	},
}

var ErrTwoFactorAuthenticationRequiredResponse = fiber.Map{
	"error": fiber.Map{
		"message":     "Two Factor Authentication Required!",
		"status_code": 401,
	},
}

var ErrNotImplementedResponse = fiber.Map{
	"error": fiber.Map{
		"message":     "The route you're looking for is not (yet) implemented.",
		"status_code": 501,
	},
}

var ErrWorldNotFoundResponse = fiber.Map{
	"error": fiber.Map{
		"message":     "World not found",
		"status_code": 404,
	},
}
