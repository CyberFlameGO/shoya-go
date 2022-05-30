package main

import (
	"github.com/gofiber/fiber/v2"
	"gitlab.com/george/shoya-go/models"
	"strconv"
)

func fileRoutes(router *fiber.App) {
	file := router.Group("/file")
	file.Get("/:id", getFile)
	file.Get("/:id/:version", getFileVersion)
	file.Get("/:id/:version/:descriptor", getFileVersionDescriptor)
	file.Get("/:id/:version/:descriptor/start", getFileVersionDescriptor)
	file.Get("/:id/:version/:descriptor/status", getFileVersionDescriptor)
	file.Get("/:id/:version/:descriptor/finish", getFileVersionDescriptor)
	file.Get("/:id/:version/:descriptor/file", getFileVersionDescriptor)
}

func getFile(c *fiber.Ctx) error {
	var id = c.Params("id")
	var f *models.File
	var err error
	if f, err = models.GetFile(id); err != nil {
		return c.JSON(models.MakeErrorResponse(err.Error(), 500))
	}

	return c.JSON(f.GetAPIFile())
}

func getFileVersion(c *fiber.Ctx) error {
	var id = c.Params("id")
	var ver, err = strconv.Atoi(c.Params("version"))
	if err != nil {
		return c.Status(400).JSON(models.MakeErrorResponse("invalid file version", 400))
	}

	var f *models.File
	if f, err = models.GetFile(id); err != nil {
		return c.JSON(models.MakeErrorResponse(err.Error(), 500))
	}

	return c.Redirect(f.GetVersion(ver).FileDescriptor.Url)
}

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

	switch models.FileDescriptorType(descriptor) {
	case models.FileDescriptorTypeFile:
		return c.Redirect(f.GetVersion(ver).FileDescriptor.Url)
	case models.FileDescriptorTypeDelta:
	case models.FileDescriptorTypeSignature:
		break
	}

	return c.JSON(fiber.Map{"TODO": true})
}
