package main

import (
	"gorm.io/gorm"
)

type BaseModel struct {
	ID        string `gorm:"primarykey"`
	CreatedAt int64
	UpdatedAt int64
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

type RegisterRequest struct {
	AcceptedTOSVersion int    `json:"acceptedTOSVersion"`
	Username           string `json:"username"`
	Password           string `json:"password"`
	Email              string `json:"email"`
	Day                string `json:"day"`
	Month              string `json:"month"`
	Year               string `json:"year"`
	RecaptchaCode      string `json:"recaptchaCode"`
}
