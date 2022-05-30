package main

import (
	"fmt"
	"github.com/gofiber/fiber/v2"
	"gitlab.com/george/shoya-go/models"
	"strconv"
	"strings"
)

func fileRoutes(router *fiber.App) {
	file := router.Group("/file")
	file.Post("/", ApiKeyMiddleware, AuthMiddleware, postFile)
	file.Get("/:id", getFile)
	file.Get("/:id/:version", getFileVersion)
	file.Get("/:id/:version/:descriptor", getFileVersionDescriptor)

	file.Get("/:id/:version/:descriptor/status", ApiKeyMiddleware, AuthMiddleware, getFileVersionDescriptorStatus)
	file.Put("/:id/:version/:descriptor/start", ApiKeyMiddleware, AuthMiddleware, putFileVersionDescriptorStart)
	file.Put("/:id/:version/:descriptor/finish", ApiKeyMiddleware, AuthMiddleware, putFileVersionDescriptorFinish)
}

// postFile | POST /file
// Creates a file record with version 0.
func postFile(c *fiber.Ctx) error {
	return c.JSON(struct{}{})
}

// getFile | GET /file/:id
// Returns a file.
func getFile(c *fiber.Ctx) error {
	var id = c.Params("id")
	var f *models.File
	var err error
	if f, err = models.GetFile(id); err != nil {
		if err == models.ErrFileNotFound {
			return c.Status(404).JSON(models.MakeErrorResponse(fmt.Sprintf("file %s not found", id), 500))
		}
		return c.JSON(models.MakeErrorResponse(err.Error(), 500))
	}

	return c.JSON(f.GetAPIFile())
}

// getFileVersion | GET /file/:id/:version
// Returns a redirect to the file for that version.
func getFileVersion(c *fiber.Ctx) error {
	var id = c.Params("id")
	var ver, err = strconv.Atoi(c.Params("version"))
	if err != nil {
		return c.Status(400).JSON(models.MakeErrorResponse("invalid file version", 400))
	}

	var f *models.File
	if f, err = models.GetFile(id); err != nil {
		if err == models.ErrFileNotFound {
			return c.Status(404).JSON(models.MakeErrorResponse(fmt.Sprintf("file %s not found", id), 500))
		}
		return c.JSON(models.MakeErrorResponse(err.Error(), 500))
	}

	return c.Redirect(f.GetVersion(ver).FileDescriptor.Url)
}

// getFileVersionDescriptor | GET /file/:id/:version/:descriptor
// Returns a redirect to the specific file descriptor for that version.
// Valid file descriptors: file, delta, signature
func getFileVersionDescriptor(c *fiber.Ctx) error {

	var id = c.Params("id")
	var descriptor = c.Params("descriptor")
	var ver, err = strconv.Atoi(c.Params("version"))
	if err != nil {
		return c.Status(400).JSON(models.MakeErrorResponse("invalid file version", 400))
	}

	var f *models.File
	if f, err = models.GetFile(id); err != nil {
		return c.JSON(models.MakeErrorResponse(err.Error(), 500))
	}

	v := f.GetVersion(ver)
	switch models.FileDescriptorType(descriptor) {
	case models.FileDescriptorTypeFile:
		if v.FileDescriptor.Url != "" {
			return c.Redirect(v.FileDescriptor.Url)
		}
	case models.FileDescriptorTypeDelta:
		if v.DeltaDescriptor.Url != "" {
			return c.Redirect(v.DeltaDescriptor.Url)
		}
	case models.FileDescriptorTypeSignature:
		if v.SignatureDescriptor.Url != "" {
			return c.Redirect(v.SignatureDescriptor.Url)
		}
	}

	fn := strings.Split(v.FileDescriptor.FileName, ".")
	return c.JSON(fiber.Map{ // If a file descriptor url doesn't actually exist, this is a generic "404".
		"fileName":  v.FileDescriptor.FileName,
		"mimeType":  "",
		"extension": fmt.Sprintf(".%s", fn[len(fn)-1]),
		"ownerId":   f.OwnerID,
	})
}

func getFileVersionDescriptorStatus(c *fiber.Ctx) error {
	return c.Next()
}

func putFileVersionDescriptorStart(c *fiber.Ctx) error {
	return c.Next()
}

func putFileVersionDescriptorFinish(c *fiber.Ctx) error {
	return c.Next()
}
