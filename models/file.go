package models

import (
	"fmt"
	"github.com/google/uuid"
	"gitlab.com/george/shoya-go/config"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"time"
)

var FileAllowedExtensions = []string{".vrca", ".vrcw", ".png", ".jpg", ".jpeg"}

type FileDescriptorType string

const (
	FileDescriptorTypeFile      FileDescriptorType = "file"
	FileDescriptorTypeDelta     FileDescriptorType = "delta"
	FileDescriptorTypeSignature FileDescriptorType = "signature"
)

type FileType string

const (
	FileTypeWorld  FileType = "world"
	FileTypeAvatar FileType = "avatar"
)

type FileUploadStatus string

const (
	FileUploadStatusNone     FileUploadStatus = "none"
	FileUploadStatusWaiting  FileUploadStatus = "waiting"
	FileUploadStatusQueued   FileUploadStatus = "queued"
	FileUploadStatusComplete FileUploadStatus = "complete"
	FileUploadStatusError    FileUploadStatus = "error"
)

type FileUploadCategory string

const (
	FileUploadCategorySimple    FileUploadCategory = "simple"
	FileUploadCategoryMultipart FileUploadCategory = "multipart"
	FileUploadCategoryQueued    FileUploadCategory = "queued"
)

type File struct {
	BaseModel
	OwnerID   string        `json:"ownerId"`
	Name      string        `json:"name"`
	MimeType  string        `json:"mimeType"`
	Extension string        `json:"extension"`
	Versions  []FileVersion `json:"versions" gorm:"foreignKey:FileID"`
}

func (f *File) BeforeCreate(*gorm.DB) (err error) {
	f.ID = "file_" + uuid.New().String()
	return
}

func GetFile(id string) (*File, error) {
	var f *File
	var err error

	if err = config.DB.Preload(clause.Associations).
		Preload("Versions.FileDescriptor").
		Preload("Versions.DeltaDescriptor").
		Preload("Versions.SignatureDescriptor").
		Where("id = ?", id).First(&f).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrFileNotFound
		}
		return nil, err
	}

	return f, nil
}

func (f *File) GetVersion(ver int) *FileVersion {
	for _, fv := range f.Versions {
		if fv.Version == ver {
			return &fv
		}
	}

	return &FileVersion{}
}

func (f *File) GetLatestVersion() *FileVersion {
	if len(f.Versions) == 0 {
		return &FileVersion{}
	}

	var i, v int

	for idx, fv := range f.Versions {
		if fv.Version > v {
			v = fv.Version
			i = idx
		}
	}

	return &f.Versions[i]
}

type FileVersion struct {
	BaseModel
	FileID                string           `json:"-"` // Relation to File
	Version               int              `json:"version"`
	Status                FileUploadStatus `json:"status"`
	FileDescriptorID      string           `json:"-"`
	FileDescriptor        FileDescriptor   `json:"file" gorm:"foreignKey:ID;references:FileDescriptorID"`
	DeltaDescriptorID     string           `json:"-"`
	DeltaDescriptor       FileDescriptor   `json:"delta" gorm:"foreignKey:ID;references:DeltaDescriptorID"`
	SignatureDescriptorID string           `json:"-"`
	SignatureDescriptor   FileDescriptor   `json:"signature" gorm:"foreignKey:ID;references:SignatureDescriptorID"`
}

func (f *FileVersion) GetFileUrl() string {
	return fmt.Sprintf("%s/file/%s/%d/%s", config.ApiConfiguration.ApiUrl.Get(), f.FileID, f.Version, FileDescriptorTypeFile)
}
func (f *FileVersion) GetDeltaUrl() string {
	return fmt.Sprintf("%s/file/%s/%d/%s", config.ApiConfiguration.ApiUrl.Get(), f.FileID, f.Version, FileDescriptorTypeDelta)
}
func (f *FileVersion) GetSignatureUrl() string {
	return fmt.Sprintf("%s/file/%s/%d/%s", config.ApiConfiguration.ApiUrl.Get(), f.FileID, f.Version, FileDescriptorTypeSignature)
}

func (f *FileVersion) BeforeCreate(*gorm.DB) (err error) {
	f.ID = "filever_" + uuid.New().String()
	return
}

type FileDescriptor struct {
	BaseModel
	FileID      string             `json:"-"`
	Type        FileDescriptorType `json:"-"`
	Status      FileUploadStatus   `json:"status"`
	Category    FileUploadCategory `json:"category"`
	SizeInBytes int                `json:"sizeInBytes"`
	FileName    string             `json:"fileName"`
	Url         string             `json:"url"`
	Md5         string             `json:"md5"`
	UploadId    string             `json:"uploadId"` // ?
}

func (f *FileDescriptor) BeforeCreate(*gorm.DB) (err error) {
	f.ID = "filedesc_" + uuid.New().String()
	return
}

type APIFile struct {
	ID        string           `json:"id"`
	Name      string           `json:"name"`
	OwnerID   string           `json:"ownerId"`
	MimeType  string           `json:"mimeType"`
	Extension string           `json:"extension"`
	Versions  []APIFileVersion `json:"versions"`
	Tags      []string         `json:"tags"`
}

type APIFileVersion struct {
	Version   int                `json:"version"`
	Status    FileUploadStatus   `json:"status"`
	CreatedAt string             `json:"created_at"`
	File      *APIFileDescriptor `json:"file,omitempty"`
	Delta     *APIFileDescriptor `json:"delta,omitempty"`
	Signature *APIFileDescriptor `json:"signature,omitempty"`
}

type APIFileDescriptor struct {
	FileName    string             `json:"fileName"`
	Url         string             `json:"url"`
	Md5         string             `json:"md5"`
	SizeInBytes int                `json:"sizeInBytes"`
	Status      FileUploadStatus   `json:"status"`
	Category    FileUploadCategory `json:"category"`
	UploadId    string             `json:"uploadId"`
}

func (f *File) GetAPIFile() *APIFile {
	var fvs []APIFileVersion
	for _, fv := range f.Versions {
		fvs = append(fvs, *fv.GetAPIFileVersion())
	}

	return &APIFile{
		ID:        f.ID,
		Name:      f.Name,
		OwnerID:   f.OwnerID,
		MimeType:  f.MimeType,
		Extension: f.Extension,
		Versions:  fvs,
		Tags:      []string{},
	}
}

func (f *FileVersion) GetAPIFileVersion() *APIFileVersion {
	var fv = &APIFileVersion{
		Version:   f.Version,
		Status:    f.Status,
		CreatedAt: time.Unix(f.CreatedAt, 0).UTC().Format(time.RFC3339),
		File:      f.FileDescriptor.GetAPIFileDescriptor(),
		Delta:     f.DeltaDescriptor.GetAPIFileDescriptor(),
		Signature: f.SignatureDescriptor.GetAPIFileDescriptor(),
	}

	if fv.File != nil {
		fv.File.Url = f.GetFileUrl()
	}

	if fv.Delta != nil {
		fv.Delta.Url = f.GetDeltaUrl()
	}

	if fv.Signature != nil {
		fv.Signature.Url = f.GetSignatureUrl()
	}

	return fv
}

func (f *FileDescriptor) GetAPIFileDescriptor() *APIFileDescriptor {
	if f.Url == "" { // If URL is empty, we'll assume this descriptor does not exist.
		return nil
	}

	return &APIFileDescriptor{
		FileName:    f.FileName,
		Url:         f.Url,
		Md5:         f.Md5,
		SizeInBytes: f.SizeInBytes,
		Status:      f.Status,
		Category:    f.Category,
		UploadId:    f.UploadId,
	}
}
