package main

import (
	"fmt"
	"github.com/gofiber/fiber/v2"
	"gitlab.com/george/shoya-go/config"
	"gitlab.com/george/shoya-go/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"strconv"
	"strings"
)

func usersRoutes(router *fiber.App) {
	// VRChat is inconsistent with how the do routing. Some are under /user, others /users.
	user := router.Group("/user")
	user.Get("/:id/friendStatus", ApiKeyMiddleware, AuthMiddleware, getUserFriendStatus)

	users := router.Group("/users")
	users.Get("/", ApiKeyMiddleware, AuthMiddleware, getUsers)
	users.Get("/:id", ApiKeyMiddleware, AuthMiddleware, getUser)
	users.Get("/:username/name", ApiKeyMiddleware, AuthMiddleware, getUserByUsername)
	users.Get("/:id/feedback", ApiKeyMiddleware, AuthMiddleware, getUserFeedback)
	users.Post("/", ApiKeyMiddleware, AuthMiddleware, postUser)
	users.Put("/:id", ApiKeyMiddleware, AuthMiddleware, putUser)
	users.Delete("/:id", ApiKeyMiddleware, AuthMiddleware, deleteUser)
}

func getUsers(c *fiber.Ctx) error {
	var users []models.User
	var rUsers = make([]*models.APILimitedUser, 0)
	var searchTerm string
	var searchDeveloperType string
	var searchOffset = 0
	var numberOfUsersToSearch = 60

	tx := config.DB.Model(models.User{}).
		Preload("CurrentAvatar.Image")

	// Query parameter setup
	if _n := c.Query("n"); _n != "" {
		atoi, err := strconv.Atoi(_n)
		if err != nil {
			goto badRequest
		}

		if atoi < 1 || atoi > 100 {
			goto badRequest
		}

		numberOfUsersToSearch = atoi
	}

	if _o := c.Query("offset"); _o != "" {
		atoi, err := strconv.Atoi(_o)
		if err != nil {
			goto badRequest
		}

		if atoi < 0 {
			goto badRequest
		}

		searchOffset = atoi
	}

	if _d := c.Query("developerType"); _d != "" {
		searchDeveloperType = _d
	}

	if _s := c.Query("search"); _s != "" {
		searchTerm = "%" + _s + "%" // TODO: FTS
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
	cu, ok := c.Locals("user").(*models.User)
	if !ok {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Couldn't retrieve current user.",
		})
	}

	uid := c.Params("id")
	if cu.ID == uid {
		return c.Status(fiber.StatusOK).JSON(cu.GetAPICurrentUser())
	}

	ru := &models.User{}
	tx := config.DB.Where("id = ?", uid).
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

func getUserByUsername(c *fiber.Ctx) error {
	cu, ok := c.Locals("user").(*models.User)
	if !ok {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Couldn't retrieve current user.",
		})
	}

	username := strings.ToLower(c.Params("username"))
	if cu.Username == username {
		return c.Status(fiber.StatusOK).JSON(cu.GetAPICurrentUser())
	}

	ru := &models.User{}
	tx := config.DB.Where("username = ?", username).
		Preload("CurrentAvatar.Image").
		Preload("FallbackAvatar").
		Find(&ru)

	if tx.Error != nil {
		if tx.Error == gorm.ErrRecordNotFound {
			return c.Status(404).JSON(fiber.Map{
				"error": fiber.Map{
					"message":     fmt.Sprintf("User %s not found", username),
					"status_code": 404,
				},
			})
		}
	}

	return c.Status(fiber.StatusOK).JSON(ru.GetAPIUser(false, false)) // TODO: Implement friendship system. Check friendship.
}

func postUser(c *fiber.Ctx) error {
	return c.Status(fiber.StatusOK).JSON(models.User{})
}

// putUser | PUT /users/:id
// This endpoint is used to update the information of a user.
//
// The following fields can be updated via this endpoint:
//  - acceptedTOSVersion
//  - bio
//  - bioLinks [Not Implemented]
//  - status
//  - statusDescription
//  - email
//  - displayName [Not Implemented]
//  - userIcon [Staff-only]
//  - profilePicOverride [Staff-only]
//
// FIXME: This endpoint is horribly unoptimized.
func putUser(c *fiber.Ctx) error {
	// dear client team, why are you sending separate PUT requests for status, statusDescription?
	var r UpdateUserRequest
	var u models.User
	var cu = c.Locals("user").(*models.User)
	var changes = map[string]interface{}{}
	var bioChanged bool
	var emailChanged bool
	var statusChanged bool
	var statusDescriptionChanged bool
	var userIconChanged bool
	var profilePicOverrideChanged bool
	var tagsChanged bool
	var homeWorldChanged bool

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
		config.DB.Where("id = ?", c.Params("id")).Find(&u)
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
		if err == models.InvalidBioErrorInUserUpdate {
			goto badRequest
		}
	}
	emailChanged, err = r.EmailChecks(&u)
	if err != nil {
		if err == models.InvalidCredentialsErrorInUserUpdate {
			goto wrongPassword
		}

		if err == models.UserWithEmailAlreadyExistsErrorInUserUpdate {
			goto badRequest
		}
	}

	statusChanged, err = r.StatusChecks(&u)
	if err != nil {
		if err == models.InvalidUserStatusErrorInUserUpdate {
			goto badRequest
		}
	}

	statusDescriptionChanged, err = r.StatusDescriptionChecks(&u)
	if err != nil {
		if err == models.InvalidStatusDescriptionErrorInUserUpdate {
			goto badRequest
		}
	}

	userIconChanged, err = r.UserIconChecks(&u)
	if err != nil {
		if err == models.TriedToSetUserIconWithoutBeingStaffErrorInUserUpdate {
			goto badRequest
		}
	}

	profilePicOverrideChanged, err = r.ProfilePicOverrideChecks(&u)
	if err != nil {
		if err == models.TriedToSetProfilePicOverrideWithoutBeingStaffErrorInUserUpdate {
			goto badRequest
		}
	}

	tagsChanged, err = r.TagsChecks(&u)
	if err != nil {
		goto badRequest
	}

	homeWorldChanged, err = r.HomeLocationChecks(&u)
	if err != nil {
		goto badRequest
	}

	if bioChanged {
		changes["bio"] = u.Bio
	}

	if emailChanged {
		changes["pending_email"] = u.Email
		// TODO: Queue up email verification sending
	}

	if statusChanged {
		changes["status"] = u.Status
	}

	if statusDescriptionChanged {
		changes["status_description"] = u.StatusDescription
	}

	if userIconChanged {
		changes["user_icon"] = u.UserIcon
	}

	if profilePicOverrideChanged {
		changes["profile_pic_override"] = u.ProfilePicOverride
	}

	if tagsChanged {
		changes["tags"] = u.Tags
	}

	if homeWorldChanged {
		changes["home_world_id"] = u.HomeWorldID
	}

	config.DB.Omit(clause.Associations).Model(&u).Updates(changes)

	return c.Status(fiber.StatusOK).JSON(u.GetAPICurrentUser())

wrongPassword:
	return c.Status(400).JSON(models.ErrInvalidCredentialsResponse)

badRequest:
	return c.Status(400).JSON(fiber.Map{
		"error": fiber.Map{
			"message":     "Bad request",
			"status_code": 400,
		},
	})
}

func deleteUser(c *fiber.Ctx) error {
	return c.Status(fiber.StatusOK).JSON(models.User{})
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
