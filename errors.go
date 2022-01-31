package main

import "github.com/gofiber/fiber/v2"

var MissingCredentialsResponse = fiber.Map{
	"error": fiber.Map{
		"message":     "Missing Credentials!",
		"status_code": 401,
	},
}

var InvalidCredentialsResponse = fiber.Map{
	"error": fiber.Map{
		"message":     "Invalid Credentials!",
		"status_code": 401,
	},
}

var TwoFactorAuthenticationRequiredResponse = fiber.Map{
	"error": fiber.Map{
		"message":     "Two Factor Authentication Required!",
		"status_code": 401,
	},
}
