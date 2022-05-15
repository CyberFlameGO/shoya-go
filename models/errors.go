package models

import (
	"errors"
	"github.com/gofiber/fiber/v2"
)

var XPoweredByHeaders = []string{
	"uwu", // ... i have no words other than "why?". This is actually present on Official btw.
	"Chaos Monkey",
	"Bone Hurting Juice",
	"SAL 9000",
	"; DROP TABLE avatars;",
	"Human-readable Slorgs",
	"Honestly, we have no clue, and we're afraid to find out",
	"Thats it! You people have stood in my way long enough. Im going to clown college!",
}

var (
	ErrMissingCredentialsResponse = fiber.Map{
		"error": fiber.Map{
			"message":     "Missing Credentials!",
			"status_code": 401,
		},
	}

	ErrMissingAdminCredentialsResponse = fiber.Map{
		"error": fiber.Map{
			"message":     "Missing Admin Credentials!",
			"status_code": 401,
		},
	}

	ErrInvalidCredentialsResponse = fiber.Map{
		"error": fiber.Map{
			"message":     "Invalid Credentials!",
			"status_code": 401,
		},
	}

	ErrTwoFactorAuthenticationRequiredResponse = fiber.Map{
		"error": fiber.Map{
			"message":     "Two Factor Authentication Required!",
			"status_code": 401,
		},
	}

	ErrNotImplementedResponse = fiber.Map{
		"error": fiber.Map{
			"message":     "The route you're looking for is not (yet) implemented.",
			"status_code": 501,
		},
	}

	ErrWorldNotFoundResponse = fiber.Map{
		"error": fiber.Map{
			"message":     "World not found",
			"status_code": 404,
		},
	}

	ErrAvatarNotFoundResponse = fiber.Map{
		"error": fiber.Map{
			"message":     "Avatar not found",
			"status_code": 404,
		},
	}

	ErrInstanceNotFoundResponse = fiber.Map{
		"error": fiber.Map{
			"message":     "Instance not found",
			"status_code": 404,
		},
	}

	ErrInvalidCredentialsInUserUpdate                = errors.New("invalid credentials presented during user update")
	ErrEmailAlreadyExistsInUserUpdate                = errors.New("user with email already exists")
	ErrInvalidUserStatusInUserUpdate                 = errors.New("invalid user status")
	ErrInvalidStatusDescriptionInUserUpdate          = errors.New("invalid status description")
	ErrInvalidBioInUserUpdate                        = errors.New("invalid bio")
	ErrTooManyLanguageTagsInUserUpdate               = errors.New("too many language tags")
	ErrInvalidLanguageTagInUserUpdate                = errors.New("invalid language tag")
	ErrSetUserIconWhenNotStaffInUserUpdate           = errors.New("tried to set user icon without being staff")
	ErrSetProfilePicOverrideWhenNotStaffInUserUpdate = errors.New("tried to set profile pic override without being staff")
	ErrWorldNotFoundInUserUpdate                     = errors.New("world not found")
	ErrWorldPrivateNotOwnedByUserInUserUpdate        = errors.New("world is private and not owned by current user")

	ErrInvalidJoinJWT = errors.New("invalid join token")
)
