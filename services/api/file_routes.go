package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"gitlab.com/george/shoya-go/config"
	pb "gitlab.com/george/shoya-go/gen/v1/proto"
	"gitlab.com/george/shoya-go/models"
	"gorm.io/gorm/clause"
	"strconv"
	"strings"
	"time"
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
	var fileVersion = &models.FileVersion{
		FileID:  f.ID,
		Version: f.GetLatestVersion().Version + 1,
		Status:  models.FileUploadStatusWaiting,
	}
	var fileDescriptor = &models.FileDescriptor{
		FileID:      f.ID,
		Type:        models.FileDescriptorTypeFile,
		Status:      models.FileUploadStatusNone,
		Category:    models.FileUploadCategoryQueued,
		SizeInBytes: 0,
		FileName:    fmt.Sprintf("%s.%s.%d%s", strings.ReplaceAll(f.Name, " ", "-")[:32], f.ID, fileVersion.Version, f.Extension),
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
		FileName:    fmt.Sprintf("%s.%s.%d%s.delta", strings.ReplaceAll(f.Name, " ", "-")[:32], f.ID, fileVersion.Version, f.Extension),
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
		FileName:    fmt.Sprintf("%s.%s.%d%s.signature", strings.ReplaceAll(f.Name, " ", "-")[:32], f.ID, fileVersion.Version, f.Extension),
		Url:         "",
		Md5:         "",
		UploadId:    "",
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
	tx := config.DB.Unscoped().Delete(&f)
	if tx.Error != nil {
		return c.Status(500).JSON(models.MakeErrorResponse(tx.Error.Error(), 500))
	}

	return c.JSON(fiber.Map{"ok": true})
}

// getFileVersion | GET /file/:id/:version
// Returns a redirect to the file for that version.
func getFileVersion(c *fiber.Ctx) error {
	var id = c.Params("id")
	var ver, err = strconv.Atoi(c.Params("version"))
	var v *models.FileVersion
	var r *pb.GetFileResponse
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

	v = f.GetVersion(ver)
	if v.FileDescriptor.Status == models.FileUploadStatusComplete {
		if v.FileDescriptor.Status == models.FileUploadStatusComplete {
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()
			r, err = FilesService.GetFile(ctx, &pb.GetFileRequest{Name: &v.FileDescriptor.FileName})
			if err != nil {
				return c.Status(500).JSON(models.MakeErrorResponse("failed to generate file url", 500))
			}

			return c.Redirect(r.GetUrl())
		}
	}

	return c.JSON(fiber.Map{ // If a file descriptor url doesn't actually exist, this is a generic "404".
		"fileName":  v.FileDescriptor.FileName,
		"mimeType":  f.MimeType,
		"extension": f.Extension,
		"ownerId":   f.OwnerID,
	})
}

// getFileVersionDescriptor | GET /file/:id/:version/:descriptor
// Returns a redirect to the specific file descriptor for that version.
// Valid file descriptors: file, delta, signature
func getFileVersionDescriptor(c *fiber.Ctx) error {

	var id = c.Params("id")
	var descriptor = c.Params("descriptor")
	var ver, err = strconv.Atoi(c.Params("version"))
	var r *pb.GetFileResponse
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
		if v.FileDescriptor.Status == models.FileUploadStatusComplete {
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()
			r, err = FilesService.GetFile(ctx, &pb.GetFileRequest{Name: &v.FileDescriptor.FileName})
			if err != nil {
				return c.Status(500).JSON(models.MakeErrorResponse("failed to generate file url", 500))
			}

			return c.Redirect(r.GetUrl())
		}
	case models.FileDescriptorTypeDelta:
		if v.DeltaDescriptor.Status == models.FileUploadStatusComplete {
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()
			r, err = FilesService.GetFile(ctx, &pb.GetFileRequest{Name: &v.DeltaDescriptor.FileName})
			if err != nil {
				return c.Status(500).JSON(models.MakeErrorResponse("failed to generate file url", 500))
			}

			return c.Redirect(r.GetUrl())
		}
	case models.FileDescriptorTypeSignature:
		if v.SignatureDescriptor.Status == models.FileUploadStatusComplete {
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()
			r, err = FilesService.GetFile(ctx, &pb.GetFileRequest{Name: &v.SignatureDescriptor.FileName})
			if err != nil {
				return c.Status(500).JSON(models.MakeErrorResponse("failed to generate file url", 500))
			}

			return c.Redirect(r.GetUrl())
		}
	}

	return c.JSON(fiber.Map{ // If a file descriptor url doesn't actually exist, this is a generic "404".
		"fileName":  v.FileDescriptor.FileName,
		"mimeType":  f.MimeType,
		"extension": f.Extension,
		"ownerId":   f.OwnerID,
	})
}

// getFileVersionDescriptorStatus | GET /file/:id/:version/:descriptor/status
func getFileVersionDescriptorStatus(c *fiber.Ctx) error {
	var f = c.Locals("file").(*models.File)
	var v int
	var err error

	if v, err = strconv.Atoi(c.Params("version")); v < 0 || err != nil {
		return c.Status(400).JSON(models.MakeErrorResponse("could not parse version", 400))
	}

	switch models.FileDescriptorType(c.Params("descriptor")) {
	case models.FileDescriptorTypeFile:
		return c.JSON(f.GetVersion(v).GetAPIFileVersion().File)
	case models.FileDescriptorTypeDelta:
		return c.JSON(f.GetVersion(v).GetAPIFileVersion().Delta)
	case models.FileDescriptorTypeSignature:
		return c.JSON(f.GetVersion(v).GetAPIFileVersion().Signature)
	default:
		return c.Status(400).JSON(models.MakeErrorResponse("invalid file descriptor type", 400))
	}
}

func putFileVersionDescriptorStart(c *fiber.Ctx) error {
	var f = c.Locals("file").(*models.File)
	var v int
	var ver *models.FileVersion
	var fileName string
	var fileMd5 string
	var err error

	if v, err = strconv.Atoi(c.Params("version")); v < 0 || err != nil {
		return c.Status(400).JSON(models.MakeErrorResponse("could not parse version", 400))
	}
	ver = f.GetVersion(v)

	switch models.FileDescriptorType(c.Params("descriptor")) {
	case models.FileDescriptorTypeFile:
		if ver.FileDescriptor.Status == models.FileUploadStatusComplete {
			return c.Status(400).JSON(models.MakeErrorResponse("already completed", 400))
		}
		fileName = ver.FileDescriptor.FileName
		fileMd5 = ver.FileDescriptor.Md5
	case models.FileDescriptorTypeDelta:
		if ver.DeltaDescriptor.Status == models.FileUploadStatusComplete {
			return c.Status(400).JSON(models.MakeErrorResponse("already completed", 400))
		}
		fileName = ver.DeltaDescriptor.FileName
		fileMd5 = ver.DeltaDescriptor.Md5
	case models.FileDescriptorTypeSignature:
		if ver.SignatureDescriptor.Status == models.FileUploadStatusComplete {
			return c.Status(400).JSON(models.MakeErrorResponse("already completed", 400))
		}
		fileName = ver.SignatureDescriptor.FileName
		fileMd5 = ver.SignatureDescriptor.Md5
	default:
		return c.Status(400).JSON(models.MakeErrorResponse("invalid descriptor", 400))
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	r, err := FilesService.CreateFile(ctx, &pb.CreateFileRequest{Name: &fileName, Md5: &fileMd5, ContentType: &f.MimeType})
	if err != nil {
		return c.Status(500).JSON(models.MakeErrorResponse(err.Error(), 500))
	}

	return c.JSON(fiber.Map{
		"url": r.GetUrl(),
	})
}

func putFileVersionDescriptorFinish(c *fiber.Ctx) error {
	var f = c.Locals("file").(*models.File)
	var v int
	var ver *models.FileVersion
	var fd *models.FileDescriptor
	var err error

	if v, err = strconv.Atoi(c.Params("version")); v < 0 || err != nil {
		return c.Status(400).JSON(models.MakeErrorResponse("could not parse version", 400))
	}
	ver = f.GetVersion(v)

	switch models.FileDescriptorType(c.Params("descriptor")) {
	case models.FileDescriptorTypeFile:
		if ver.FileDescriptor.Status == models.FileUploadStatusComplete {
			return c.Status(400).JSON(models.MakeErrorResponse("already completed", 400))
		}
		tx := config.DB.Where("id = ?", ver.FileDescriptorID).First(&fd)
		if tx.Error != nil {
			return c.Status(500).JSON(models.MakeErrorResponse("error getting file descriptor", 500))
		}
		ver.FileDescriptor.Status = models.FileUploadStatusComplete
	case models.FileDescriptorTypeDelta:
		if ver.DeltaDescriptor.Status == models.FileUploadStatusComplete {
			return c.Status(400).JSON(models.MakeErrorResponse("already completed", 400))
		}
		tx := config.DB.Where("id = ?", ver.DeltaDescriptorID).First(&fd)
		if tx.Error != nil {
			return c.Status(500).JSON(models.MakeErrorResponse("error getting file descriptor", 500))
		}
		ver.DeltaDescriptor.Status = models.FileUploadStatusComplete
	case models.FileDescriptorTypeSignature:
		if ver.SignatureDescriptor.Status == models.FileUploadStatusComplete {
			return c.Status(400).JSON(models.MakeErrorResponse("already completed", 400))
		}
		tx := config.DB.Where("id = ?", ver.SignatureDescriptorID).First(&fd)
		if tx.Error != nil {
			return c.Status(500).JSON(models.MakeErrorResponse("error getting file descriptor", 500))
		}
		ver.SignatureDescriptor.Status = models.FileUploadStatusComplete
	default:
		return c.Status(400).JSON(models.MakeErrorResponse("invalid descriptor", 400))
	}

	fd.Status = models.FileUploadStatusComplete
	if config.DB.Omit(clause.Associations).Updates(fd).Error != nil {
		return c.Status(500).JSON(models.MakeErrorResponse("could not update database object", 500))
	}

	if ver.FileDescriptor.Status == models.FileUploadStatusComplete && ver.SignatureDescriptor.Status == models.FileUploadStatusComplete {
		ver.Status = models.FileUploadStatusComplete
		if config.DB.Omit(clause.Associations).Updates(ver).Error != nil {
			return c.Status(500).JSON(models.MakeErrorResponse("could not update database object", 500))
		}
	} else if ver.DeltaDescriptor.Status == models.FileUploadStatusComplete && ver.SignatureDescriptor.Status == models.FileUploadStatusComplete {
		ver.Status = models.FileUploadStatusComplete
		if config.DB.Omit(clause.Associations).Updates(ver).Error != nil {
			return c.Status(500).JSON(models.MakeErrorResponse("could not update database object", 500))
		}
	}

	return c.JSON(ver.GetAPIFileVersion())
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
