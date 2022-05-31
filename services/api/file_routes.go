package main

import (
	"encoding/base64"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"gitlab.com/george/shoya-go/config"
	"gitlab.com/george/shoya-go/models"
	"gorm.io/gorm/clause"
	"strconv"
	"strings"
)

func fileRoutes(router *fiber.App) {
	file := router.Group("/file")
	file.Post("/", ApiKeyMiddleware, AuthMiddleware, createFile)
	file.Get("/:id", getFile)
	file.Post("/:id", ApiKeyMiddleware, AuthMiddleware, IsFileOwnerMiddleware, postFile)
	file.Delete("/:id", ApiKeyMiddleware, AuthMiddleware, IsFileOwnerMiddleware, deleteFile)
	file.Get("/:id/:version", getFileVersion)
	file.Get("/:id/:version/:descriptor", getFileVersionDescriptor)

	file.Get("/:id/:version/:descriptor/status", ApiKeyMiddleware, AuthMiddleware, IsFileOwnerMiddleware, getFileVersionDescriptorStatus)
	file.Put("/:id/:version/:descriptor/start", ApiKeyMiddleware, AuthMiddleware, IsFileOwnerMiddleware, putFileVersionDescriptorStart)
	file.Put("/:id/:version/:descriptor/finish", ApiKeyMiddleware, AuthMiddleware, IsFileOwnerMiddleware, putFileVersionDescriptorFinish)
}

// createFile | POST /file
// Creates a file record with version 0.
func createFile(c *fiber.Ctx) error {
	var u = c.Locals("user").(*models.User)
	var r CreateFileRequest
	var f models.File
	var fv models.FileVersion
	var err error

	if err = c.BodyParser(&r); err != nil {
		return c.Status(500).JSON(models.MakeErrorResponse("failed to parse request body", 500))
	}

	f = models.File{
		OwnerID:   u.ID,
		Name:      r.Name,
		MimeType:  r.MimeType,
		Extension: r.Extension,
	}
	config.DB.Create(&f)

	fv = models.FileVersion{
		FileID:  f.ID,
		Version: 0,
		Status:  models.FileUploadStatusComplete,
	}
	config.DB.Create(&fv)

	f.Versions = []models.FileVersion{fv}
	return c.JSON(f.GetAPIFile())
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

// postFile | POST /file/:id
// Creates a new file version.
func postFile(c *fiber.Ctx) error {
	var f = c.Locals("file").(*models.File)
	var r CreateFileVersionRequest
	var fileMd5Hash []byte
	var signatureMd5Hash []byte
	var fileDescriptor = &models.FileDescriptor{
		FileID:      f.ID,
		Type:        models.FileDescriptorTypeFile,
		Status:      models.FileUploadStatusNone,
		Category:    models.FileUploadCategoryQueued,
		SizeInBytes: 0,
		FileName:    fmt.Sprintf("%s.%s%s", strings.ReplaceAll(f.Name, " ", "-")[:32], f.ID, f.Extension),
		Url:         "",
		Md5:         "",
		UploadId:    "",
	}
	var deltaDescriptor = &models.FileDescriptor{
		FileID:      f.ID,
		Type:        models.FileDescriptorTypeSignature,
		Status:      models.FileUploadStatusNone,
		Category:    models.FileUploadCategoryQueued,
		SizeInBytes: 0,
		FileName:    fmt.Sprintf("%s.%s%s.signature", strings.ReplaceAll(f.Name, " ", "-")[:32], f.ID, f.Extension),
		Url:         "",
		Md5:         "",
		UploadId:    "",
	}
	var signatureDescriptor = &models.FileDescriptor{
		FileID:      f.ID,
		Type:        models.FileDescriptorTypeDelta,
		Status:      models.FileUploadStatusNone,
		Category:    models.FileUploadCategorySimple,
		SizeInBytes: 0,
		FileName:    fmt.Sprintf("%s.%s%s.delta", strings.ReplaceAll(f.Name, " ", "-")[:32], f.ID, f.Extension),
		Url:         "",
		Md5:         "",
		UploadId:    "",
	}
	var fileVersion = &models.FileVersion{
		FileID:  f.ID,
		Version: f.GetLatestVersion().Version + 1,
		Status:  models.FileUploadStatusWaiting,
	}
	var err error

	if err = c.BodyParser(&r); err != nil {
		return c.Status(500).JSON(models.MakeErrorResponse("failed to parse request body", 500))
	}

	if r.FileMd5 != "" {
		fileMd5Hash, err = base64.StdEncoding.DecodeString(r.FileMd5)
		if err != nil {
			return c.Status(500).JSON(models.MakeErrorResponse("file md5 invalid", 500))
		}
	}

	signatureMd5Hash, err = base64.StdEncoding.DecodeString(r.SignatureMd5)
	if err != nil {
		return c.Status(500).JSON(models.MakeErrorResponse("signature md5 invalid", 500))
	}

	if len(fileMd5Hash) != 16 || len(signatureMd5Hash) != 16 {
		return c.Status(500).JSON(models.MakeErrorResponse("file or signature md5 invalid", 500))
	}

	if r.FileMd5 != "" && r.FileSizeInBytes != 0 {
		fileDescriptor.Status = models.FileUploadStatusWaiting
		fileDescriptor.Category = models.FileUploadCategorySimple
		fileDescriptor.Md5 = r.FileMd5
		fileDescriptor.SizeInBytes = r.FileSizeInBytes
	}

	if r.DeltaMd5 != "" && r.DeltaSizeInBytes != 0 {
		deltaDescriptor.Status = models.FileUploadStatusWaiting
		deltaDescriptor.Category = models.FileUploadCategorySimple
		deltaDescriptor.Md5 = r.DeltaMd5
		deltaDescriptor.SizeInBytes = r.DeltaSizeInBytes
	}

	if r.SignatureMd5 != "" && r.SignatureSizeInBytes != 0 {
		signatureDescriptor.Status = models.FileUploadStatusWaiting
		signatureDescriptor.Category = models.FileUploadCategorySimple
		signatureDescriptor.Md5 = r.SignatureMd5
		signatureDescriptor.SizeInBytes = r.SignatureSizeInBytes
	}

	err = config.DB.Create(&fileDescriptor).Error
	if err != nil {
		fmt.Println(err)
	}
	err = config.DB.Create(&deltaDescriptor).Error
	if err != nil {
		fmt.Println(err)
	}
	err = config.DB.Create(&signatureDescriptor).Error
	if err != nil {
		fmt.Println(err)
	}

	fileVersion.FileDescriptorID = fileDescriptor.ID
	fileVersion.DeltaDescriptorID = deltaDescriptor.ID
	fileVersion.SignatureDescriptorID = signatureDescriptor.ID

	config.DB.Omit(clause.Associations).Create(&fileVersion)

	fileVersion.FileDescriptor = *fileDescriptor
	fileVersion.DeltaDescriptor = *deltaDescriptor
	fileVersion.SignatureDescriptor = *signatureDescriptor

	f.Versions = append(f.Versions, *fileVersion)

	return c.JSON(f.GetAPIFile())
}

func deleteFile(c *fiber.Ctx) error {
	var f = c.Locals("file").(*models.File)
	config.DB.Unscoped().Delete(&f)

	return c.JSON(fiber.Map{"ok": true})
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

	return c.JSON(fiber.Map{ // If a file descriptor url doesn't actually exist, this is a generic "404".
		"fileName":  v.FileDescriptor.FileName,
		"mimeType":  f.MimeType,
		"extension": f.Extension,
		"ownerId":   f.OwnerID,
	})
}

func getFileVersionDescriptorStatus(c *fiber.Ctx) error {
	var u = c.Locals("user").(*models.User)
	var f = c.Locals("file").(*models.File)
	var err error

	if u.ID == "" {
		return c.Next()
	}

	if f.ID == "" {
		return c.Next()
	}

	if err != nil {
		return c.Next()
	}

	return c.Next()
}

func putFileVersionDescriptorStart(c *fiber.Ctx) error {
	var u = c.Locals("user").(*models.User)
	var f = c.Locals("file").(*models.File)
	var err error

	if u.ID == "" {
		return c.Next()
	}

	if f.ID == "" {
		return c.Next()
	}

	if err != nil {
		return c.Next()
	}

	return c.Next()
}

func putFileVersionDescriptorFinish(c *fiber.Ctx) error {
	var u = c.Locals("user").(*models.User)
	var f = c.Locals("file").(*models.File)
	var err error

	if u.ID == "" {
		return c.Next()
	}

	if f.ID == "" {
		return c.Next()
	}

	if err != nil {
		return c.Next()
	}
	return c.Next()
}

func IsFileOwnerMiddleware(c *fiber.Ctx) error {
	var u = c.Locals("user").(*models.User)
	var fid = c.Params("id")
	var f *models.File
	var err error

	if f, err = models.GetFile(fid); err != nil {
		if err == models.ErrFileNotFound {
			return c.Status(404).JSON(models.MakeErrorResponse(fmt.Sprintf("file %s not found", fid), 500))
		}
		return c.JSON(models.MakeErrorResponse(err.Error(), 500))
	}

	if f.OwnerID != u.ID && !u.IsStaff() {
		return c.Status(403).JSON(models.MakeErrorResponse("not allowed to update another user's file", 403))
	}

	c.Locals("file", f)

	return c.Next()
}
