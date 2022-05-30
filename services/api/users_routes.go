package main

import (
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/tj/go-naturaldate"
	"gitlab.com/george/shoya-go/config"
	"gitlab.com/george/shoya-go/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"strconv"
	"strings"
	"time"
)

func usersRoutes(router *fiber.App) {
	// VRChat is inconsistent with how the do routing. Some are under /user, others /users.
	user := router.Group("/user", ApiKeyMiddleware, AuthMiddleware)
	user.Get("/:id/friendStatus", getUserFriendStatus)
	user.Get("/:id/moderations", AdminMiddleware, getUserModerations)
	user.Post("/:id/moderations", postUserModerations)

	users := router.Group("/users", ApiKeyMiddleware, AuthMiddleware)
	users.Get("/", getUsers)
	users.Get("/:id", getUser)
	users.Post("/:id/addTags", postUserAddTags)
	users.Post("/:id/removeTags", postUserRemoveTags)

	users.Get("/:username/name", getUserByUsername)

	users.Get("/:id/feedback", getUserFeedback)

	users.Put("/:id", putUser)
	users.Delete("/:id", deleteUser)
}

// getUsers | GET /users
// This endpoint allows you to search through the users of the platform.
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
	return c.Status(400).JSON(models.MakeErrorResponse("Bad request", 400))
}

// getUser | GET /users/:id
// Returns a single user based on id.
func getUser(c *fiber.Ctx) error {
	var cu = c.Locals("user").(*models.User)
	var ru *models.User
	var err error

	uid := c.Params("id")
	if cu.ID == uid {
		return c.Status(fiber.StatusOK).JSON(cu.GetAPICurrentUser())
	}

	if ru, err = models.GetUserById(uid); err != nil {
		if err == models.ErrUserNotFound {
			return c.Status(404).JSON(models.MakeErrorResponse(fmt.Sprintf("User %s not found", uid), 404))
		}

		return c.Status(500).JSON(models.MakeErrorResponse(err.Error(), 500))
	}

	return c.Status(fiber.StatusOK).JSON(ru.GetAPIUser(false, false)) // TODO: Implement friendship system. Check friendship.
}

// getUserByUsername | GET /users/:username/name
// Returns a single user based on username.
func getUserByUsername(c *fiber.Ctx) error {
	var cu = c.Locals("user").(*models.User)
	var ru *models.User
	var err error

	username := strings.ToLower(c.Params("username"))
	if cu.Username == username {
		return c.Status(fiber.StatusOK).JSON(cu.GetAPICurrentUser())
	}

	if ru, err = models.GetUserByUsername(username); err != nil {
		if err == models.ErrUserNotFound {
			return c.Status(404).JSON(models.MakeErrorResponse(fmt.Sprintf("User %s not found", username), 404))
		}

		return c.Status(500).JSON(models.MakeErrorResponse(err.Error(), 500))
	}

	return c.Status(fiber.StatusOK).JSON(ru.GetAPIUser(false, false)) // TODO: Implement friendship system. Check friendship.
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
func putUser(c *fiber.Ctx) error {
	var cu = c.Locals("user").(*models.User)
	var r UpdateUserRequest
	var u models.User
	var changes = map[string]interface{}{}
	var bioChanged bool
	var emailChanged bool
	var passwordChanged bool
	var statusChanged bool
	var statusDescriptionChanged bool
	var userIconChanged bool
	var profilePicOverrideChanged bool
	var tagsChanged bool
	var homeWorldChanged bool

	if c.Params("id") != cu.ID && !cu.IsStaff() {
		return c.Status(403).JSON(models.MakeErrorResponse("You're not allowed to update another user's profile", 403))
	}

	if c.Params("id") == cu.ID {
		u = *cu
	} else {
		config.DB.Where("id = ?", c.Params("id")).Find(&u)
	}

	err := c.BodyParser(&r)
	if err != nil {
		return c.Status(500).JSON(models.MakeErrorResponse(err.Error(), 500))
	}

	bioChanged, err = r.BioChecks(&u)
	if err != nil {
		if err == models.ErrInvalidBioInUserUpdate {
			goto badRequest
		}
	}
	emailChanged, err = r.EmailChecks(&u)
	if err != nil {
		if err == models.ErrInvalidCredentialsInUserUpdate {
			goto wrongPassword
		}

		if err == models.ErrEmailAlreadyExistsInUserUpdate {
			goto badRequest
		}
	}

	passwordChanged, err = r.PasswordChecks(&u)
	if err != nil {
		goto badRequest
	}

	statusChanged, err = r.StatusChecks(&u)
	if err != nil {
		if err == models.ErrInvalidUserStatusInUserUpdate {
			goto badRequest
		}
	}

	statusDescriptionChanged, err = r.StatusDescriptionChecks(&u)
	if err != nil {
		if err == models.ErrInvalidStatusDescriptionInUserUpdate {
			goto badRequest
		}
	}

	userIconChanged, err = r.UserIconChecks(&u)
	if err != nil {
		if err == models.ErrSetUserIconWhenNotStaffInUserUpdate {
			goto badRequest
		}
	}

	profilePicOverrideChanged, err = r.ProfilePicOverrideChecks(&u)
	if err != nil {
		if err == models.ErrSetProfilePicOverrideWhenNotStaffInUserUpdate {
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

	if passwordChanged {
		changes["password"] = u.Password
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
	return c.Status(400).JSON(models.MakeErrorResponse("Bad request", 400))
}

// deleteUser | DELETE /users/:id
// Marks a user for deletion.
// TODO: User deletion.
func deleteUser(c *fiber.Ctx) error {
	return c.Status(fiber.StatusOK).JSON(models.User{})
}

// getUserFeedback | GET /users/:id/feedback
// Returns the reports created by this user.
// TODO: Implement reporting system
func getUserFeedback(c *fiber.Ctx) error {
	return c.JSON([]struct{}{})
}

// getUserModerations | GET /user/:id/moderations
// Admin-only. Returns the active moderations against this user.
func getUserModerations(c *fiber.Ctx) error {
	uid := c.Params("id")
	ru := models.User{}
	tx := config.DB.Where("id = ?", uid).
		Preload("Moderations").
		Find(&ru)

	if tx.Error != nil {
		if tx.Error == gorm.ErrRecordNotFound {
			return c.Status(404).JSON(models.MakeErrorResponse(fmt.Sprintf("User %s not found", uid), 404))
		}
	}

	//goland:noinspection GoPreferNilSlice
	r := []*models.APIModeration{}
	for _, moderation := range ru.Moderations {
		if moderation.ExpiresAt == 0 || moderation.ExpiresAt > time.Now().UTC().Unix() {
			am := moderation.GetAPIModeration(false)
			am.TargetDisplayName = ru.DisplayName
			r = append(r, am)
		}
	}

	return c.JSON(r)
}

// postUserModerations | POST /user/:id/moderations
// Adds a moderation to a user.
// TODO: Implement rate-limiter so people can't spam moderation actions
func postUserModerations(c *fiber.Ctx) error {
	var u = c.Locals("user").(*models.User)
	var mod *models.Moderation
	var req ModerationRequest
	var exp time.Time

	err := c.BodyParser(&req)
	if err != nil {
		return c.Status(500).JSON(models.MakeErrorResponse(err.Error(), 500))
	}

	if (req.Type == models.ModerationBan) && !u.IsStaff() {
		return c.Status(401).JSON(models.ErrMissingAdminCredentialsResponse)
	}

	if req.Type == models.ModerationKick || req.Type == models.ModerationWarn {
		// TODO: Validate whether the instance is actually active (discovery & presence svc required)
		//       and whether the user is allowed to moderate that instance once
		//       multi-mod is implemented
		i, err := models.ParseLocationString(fmt.Sprintf("%s:%s", req.WorldID, req.InstanceID))
		if err != nil {
			return c.Status(500).JSON(models.MakeErrorResponse(err.Error(), 500))
		}

		if i.OwnerID != u.ID && !u.IsStaff() {
			return c.Status(403).JSON(models.MakeErrorResponse("not authorized to moderate this instance", 403))
		}
	}

	if boolConvert(req.IsPermanent) {
		if !u.IsStaff() {
			return c.Status(403).JSON(models.MakeErrorResponse("not authorized to create permanent moderations", 403))
		}

		exp = time.Unix(0, 0) // If expiry is `0`, we'll assume it's permanent.
	} else {
		exp, err = naturaldate.Parse(strings.ReplaceAll(req.ExpiresAt, "_", " "), time.Now().UTC(), naturaldate.WithDirection(naturaldate.Future))
		if err != nil {
			return c.Status(500).JSON(models.MakeErrorResponse(err.Error(), 500))
		}

		if exp.Before(time.Now().UTC()) {
			return c.Status(400).JSON(models.MakeErrorResponse("cannot create moderation in the past", 400))
		}
	}

	mod = &models.Moderation{
		SourceID:   u.ID,
		TargetID:   req.TargetID,
		WorldID:    req.WorldID,
		InstanceID: req.InstanceID,
		Type:       req.Type,
		Reason:     req.Reason,
		ExpiresAt:  exp.Unix(),
	}

	err = config.DB.Create(mod).Error
	if err != nil {
		return c.Status(500).JSON(models.MakeErrorResponse(err.Error(), 500))
	}

	return c.JSON(fiber.Map{
		"id": mod.ID,
	})
}

// getUserFriendStatus | GET /user/:id/friendStatus
// Gets the status of an incoming or outgoing friend request toward that user (or if they are already friends).
// TODO: This depends on friendships being implemented.
func getUserFriendStatus(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"incomingRequest": false,
		"isFriend":        false,
		"outgoingRequest": false,
	})
}

// postUserAddTags | POST /users/:id/addTags
// Adds tags to a user.
func postUserAddTags(c *fiber.Ctx) error {
	var cu = c.Locals("user").(*models.User)
	var u *models.User
	var r AddTagsRequest

	var tagsChanged bool
	var changes = map[string]interface{}{}
	var err error

	if c.Params("id") != cu.ID && !cu.IsStaff() {
		return c.Status(403).JSON(models.MakeErrorResponse("You're not allowed to update another user's profile", 403))
	}

	if c.Params("id") == cu.ID {
		u = cu
	} else {
		if u, err = models.GetUserById(c.Params("id")); err != nil {
			if err == models.ErrUserNotFound {
				return c.Status(404).JSON(models.MakeErrorResponse(fmt.Sprintf("User %s not found", c.Params("id")), 404))
			}
			return c.Status(500).JSON(models.MakeErrorResponse(err.Error(), 500))
		}
	}

	err = c.BodyParser(&r)
	if err != nil {
		return c.Status(500).JSON(models.MakeErrorResponse(err.Error(), 500))
	}

	tagsChanged, err = r.TagsChecks(u)
	if err != nil {
		goto badRequest
	}

	if tagsChanged {
		changes["tags"] = u.Tags
	}

	config.DB.Omit(clause.Associations).Model(&u).Updates(changes)

	return c.Status(fiber.StatusOK).JSON(u.GetAPICurrentUser())

badRequest:
	return c.Status(400).JSON(models.MakeErrorResponse("Bad request", 400))
}

// postUserRemoveTags | POST /users/:id/removeTags
// Removes tags from a user.
func postUserRemoveTags(c *fiber.Ctx) error {
	var cu = c.Locals("user").(*models.User)
	var u *models.User
	var r RemoveTagsRequest

	var tagsChanged bool
	var changes = map[string]interface{}{}
	var err error

	if c.Params("id") != cu.ID && !cu.IsStaff() {
		return c.Status(403).JSON(models.MakeErrorResponse("You're not allowed to update another user's profile", 403))
	}

	if c.Params("id") == cu.ID {
		u = cu
	} else {
		if u, err = models.GetUserById(c.Params("id")); err != nil {
			if err == models.ErrUserNotFound {
				return c.Status(404).JSON(models.MakeErrorResponse(fmt.Sprintf("User %s not found", c.Params("id")), 404))
			}
			return c.Status(500).JSON(models.MakeErrorResponse(err.Error(), 500))
		}
	}

	err = c.BodyParser(&r)
	if err != nil {
		return c.Status(500).JSON(models.MakeErrorResponse(err.Error(), 500))
	}

	tagsChanged, err = r.TagsChecks(u)
	if err != nil {
		goto badRequest
	}

	if tagsChanged {
		changes["tags"] = u.Tags
	}

	config.DB.Omit(clause.Associations).Model(&u).Updates(changes)

	return c.Status(fiber.StatusOK).JSON(u.GetAPICurrentUser())

badRequest:
	return c.Status(400).JSON(models.MakeErrorResponse("Bad request", 400))
}
