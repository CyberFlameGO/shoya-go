package main

import (
	"fmt"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"strconv"
)

func usersRoutes(router *fiber.App) {
	// VRChat is inconsistent with how the do routing. Some are under /user, others /users.
	user := router.Group("/user")
	user.Get("/:id/friendStatus", ApiKeyMiddleware, AuthMiddleware, getUserFriendStatus)

	users := router.Group("/users")
	users.Get("/", ApiKeyMiddleware, AuthMiddleware, getUsers)
	users.Get("/:id", ApiKeyMiddleware, AuthMiddleware, getUser)
	users.Get("/:id/feedback", ApiKeyMiddleware, AuthMiddleware, getUserFeedback)
	users.Post("/", ApiKeyMiddleware, AuthMiddleware, postUser)
	users.Put("/:id", ApiKeyMiddleware, AuthMiddleware, putUser)
	users.Delete("/:id", ApiKeyMiddleware, AuthMiddleware, deleteUser)
}

func getUsers(c *fiber.Ctx) error {
	var users []User
	var rUsers = make([]*APILimitedUser, 0)
	var searchTerm string
	var searchDeveloperType string
	var searchOffset = 0
	var numberOfUsersToSearch = 60

	tx := DB.Model(User{}).
		Preload("CurrentAvatar.Image")

	// Query parameter setup
	if c.Query("n") != "" {
		atoi, err := strconv.Atoi(c.Query("n"))
		if err != nil {
			goto badRequest
		}

		if atoi < 1 || atoi > 100 {
			goto badRequest
		}

		numberOfUsersToSearch = atoi
	}

	if c.Query("offset") != "" {
		atoi, err := strconv.Atoi(c.Query("offset"))
		if err != nil {
			goto badRequest
		}

		if atoi < 0 {
			goto badRequest
		}

		searchOffset = atoi
	}

	if c.Query("developerType") != "" {
		searchDeveloperType = c.Query("developerType")
	}

	if c.Query("search") != "" {
		searchTerm = "%" + c.Query("search") + "%" // TODO: FTS
	}

	if searchDeveloperType != "" {
		tx.Where("developer_type = ?", searchDeveloperType)
	}

	if searchTerm != "" {
		tx.Where("display_name ILIKE ?", searchTerm)
	}

	tx.Limit(numberOfUsersToSearch).Offset(searchOffset)
	tx.Find(&users)

	for _, user := range users {
		lu := user.GetAPILimitedUser(false, false) // TODO: Friendship check.
		rUsers = append(rUsers, lu)
	}
	return c.Status(fiber.StatusOK).JSON(rUsers)

badRequest:
	return c.Status(400).JSON(fiber.Map{
		"error": fiber.Map{
			"message":     "Bad request",
			"status_code": 400,
		},
	})
}

func getUser(c *fiber.Ctx) error {
	cu, ok := c.Locals("user").(*User)
	if !ok {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Couldn't retrieve current user.",
		})
	}

	uid := c.Params("id")
	if cu.ID == uid {
		return c.Status(fiber.StatusOK).JSON(cu.GetAPICurrentUser())
	}

	ru := &User{}
	tx := DB.Where("id = ?", uid).
		Preload("CurrentAvatar.Image").
		Preload("FallbackAvatar").
		Find(&ru)

	if tx.Error != nil {
		if tx.Error == gorm.ErrRecordNotFound {
			return c.Status(404).JSON(fiber.Map{
				"error": fiber.Map{
					"message":     fmt.Sprintf("User %s not found", uid),
					"status_code": 404,
				},
			})
		}
	}

	return c.Status(fiber.StatusOK).JSON(ru.GetAPIUser(false, false)) // TODO: Implement friendship system. Check friendship.
}

func postUser(c *fiber.Ctx) error {
	return c.Status(fiber.StatusOK).JSON(User{})
}

func putUser(c *fiber.Ctx) error {
	// dear client team, why are you sending separate PUT requests for status, statusDescription?
	var r UpdateUserRequest
	var u User
	var cu = c.Locals("user").(*User)
	var changes = map[string]interface{}{}
	var bioChanged bool
	var emailChanged bool
	var statusChanged bool
	var statusDescriptionChanged bool

	if c.Params("id") != cu.ID && !cu.IsStaff() {
		return c.Status(403).JSON(fiber.Map{
			"error": fiber.Map{
				"message":     "You're not allowed to update another user's profile",
				"status_code": 403,
			},
		})
	}

	if c.Params("id") == cu.ID {
		u = *cu
	} else {
		DB.Where("id = ?", c.Params("id")).Find(&u)
	}

	err := c.BodyParser(&r)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": fiber.Map{
				"message":     err,
				"status_code": 500,
			},
		})
	}

	bioChanged, err = r.BioChecks(&u)
	if err != nil {
		if err == invalidBioErrorInUserUpdate {
			goto badRequest
		}
	}
	emailChanged, err = r.EmailChecks(&u)
	if err != nil {
		if err == invalidCredentialsErrorInUserUpdate {
			goto wrongPassword
		}

		if err == userWithEmailAlreadyExistsErrorInUserUpdate {
			goto badRequest
		}
	}

	statusChanged, err = r.StatusChecks(&u)
	if err != nil {
		if err == invalidUserStatusErrorInUserUpdate {
			goto badRequest
		}
	}

	statusDescriptionChanged, err = r.StatusDescriptionChecks(&u)
	if err != nil {
		if err == invalidStatusDescriptionErrorInUserUpdate {
			goto badRequest
		}
	}

	if bioChanged {
		changes["bio"] = u.Bio
	}
	if emailChanged {
		changes["email"] = u.Email
	}

	if statusChanged {
		changes["status"] = u.Status
	}

	if statusDescriptionChanged {
		changes["status_description"] = u.StatusDescription
	}
	DB.Omit(clause.Associations).Model(&u).Updates(changes)

	return c.Status(fiber.StatusOK).JSON(u.GetAPICurrentUser())

wrongPassword:
	return c.Status(400).JSON(ErrInvalidCredentialsResponse)

badRequest:
	return c.Status(400).JSON(fiber.Map{
		"error": fiber.Map{
			"message":     "Bad request",
			"status_code": 400,
		},
	})
}

func deleteUser(c *fiber.Ctx) error {
	return c.Status(fiber.StatusOK).JSON(User{})
}

func getUserFeedback(c *fiber.Ctx) error {
	return c.JSON([]interface{}{})
}

func getUserFriendStatus(c *fiber.Ctx) error {
	// TODO: Implement friendships.
	return c.JSON(fiber.Map{
		"incomingRequest": false,
		"isFriend":        false,
		"outgoingRequest": false,
	})
}
