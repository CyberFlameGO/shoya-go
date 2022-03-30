package main

import (
	"errors"
	"github.com/gofiber/fiber/v2"
)

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

var ErrAvatarNotFoundResponse = fiber.Map{
	"error": fiber.Map{
		"message":     "Avatar not found",
		"status_code": 404,
	},
}

var (
	invalidCredentialsErrorInUserUpdate                            = errors.New("invalid credentials presented during user update")
	userWithEmailAlreadyExistsErrorInUserUpdate                    = errors.New("user with email already exists")
	invalidUserStatusErrorInUserUpdate                             = errors.New("invalid user status")
	invalidStatusDescriptionErrorInUserUpdate                      = errors.New("invalid status description")
	invalidBioErrorInUserUpdate                                    = errors.New("invalid bio")
	tooManyLanguageTagsInUserUpdate                                = errors.New("too many language tags")
	invalidLanguageTagInUserUpdate                                 = errors.New("invalid language tag")
	triedToSetUserIconWithoutBeingStaffErrorInUserUpdate           = errors.New("tried to set user icon without being staff")
	triedToSetProfilePicOverrideWithoutBeingStaffErrorInUserUpdate = errors.New("tried to set profile pic override without being staff")
	worldNotFoundErrorInUserUpdate                                 = errors.New("world not found")
	worldIsPrivateAndNotOwnedByUser                                = errors.New("world is private and not owned by current user")
)
