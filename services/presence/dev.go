package presence

import (
	"fmt"
	"github.com/gofiber/fiber/v2"
	"gitlab.com/george/shoya-go/models/service_types"
	"net/url"
	"strconv"
	"time"
)

func initRoutes(app *fiber.App) {
	initDev(app)
}

func initDev(app *fiber.App) {
	dev := app.Group("/dev")
	dev.Post("/batch", postBatch)
	dev.Get("/:id", getPresence)
	dev.Put("/:id/state/:state", postState)
	dev.Put("/:id/status/:status", postStatus)
	dev.Put("/:id/instance/:instance", postInstance)
	dev.Put("/:id/lastSeen", postLastSeen)
	dev.Put("/:id/lastSeen/:time", postLastSeen)
}

func getPresence(c *fiber.Ctx) error {
	p, err := getPresenceForUser(c.Params("id"))
	if err != nil {
		fmt.Println("getPresenceForUser failed:", err)
		return c.Status(500).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	return c.JSON(p)
}

func postBatch(c *fiber.Ctx) error {
	var users []string
	err := c.BodyParser(&users)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	var presences map[string]*service_types.UserPresence
	presences = make(map[string]*service_types.UserPresence)
	for _, user := range users {
		p, err := getPresenceForUser(user)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{
				"error": err.Error(),
			})
		}
		presences[user] = p
	}

	return c.JSON(presences)
}

func postState(c *fiber.Ctx) error {
	var state = service_types.UserState(c.Params("state"))
	p, err := updateStateForUser(c.Params("id"), state)
	if err != nil {
		fmt.Println("updateStateForUser failed:", err)
		return c.Status(500).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(p)
}

func postStatus(c *fiber.Ctx) error {
	var sStatus, err = url.PathUnescape(c.Params("status"))
	var status = service_types.UserStatus(sStatus)
	p, err := updateStatusForUser(c.Params("id"), status)
	if err != nil {
		fmt.Println("updateStatusForUser failed:", err)
		return c.Status(500).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(p)
}

func postInstance(c *fiber.Ctx) error {
	var instance = c.Params("instance")
	p, err := updateInstanceForUser(c.Params("id"), instance)
	if err != nil {
		fmt.Println("updateInstanceForUser failed:", err)
		return c.Status(500).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(p)
}

func postLastSeen(c *fiber.Ctx) error {
	var lsParam = c.Params("time")
	var lsTime time.Time
	if lsParam == "" {
		lsTime = time.Now()
	} else {
		atoi, err := strconv.Atoi(lsParam)
		if err != nil {
			fmt.Println("Atoi failed")
			return c.Status(500).JSON(fiber.Map{
				"error": err.Error(),
			})
		}

		lsTime = time.Unix(int64(atoi), 0)
	}

	p, err := updateLastSeenForUser(c.Params("id"), lsTime)
	if err != nil {
		fmt.Println("updateLastSeenForUser failed:", err)
		return c.Status(500).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(p)
}
