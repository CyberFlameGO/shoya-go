package main

import "github.com/gofiber/fiber/v2"

var MissingCredentialsResponse = fiber.Map{
	"error": "Missing Credentials!",
}

var InvalidCredentialsResponse = fiber.Map{
	"error": "Invalid Credentials!",
}

var TwoFactorAuthenticationRequiredResponse = fiber.Map{
	"error": "Two Factor Authentication Required!",
}
