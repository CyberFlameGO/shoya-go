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
