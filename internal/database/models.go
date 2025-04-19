package database

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	UUID     uuid.UUID `json:"uuid" gorm:"uniqueIndex"`
	Login    string    `json:"login" gorm:"uniqueIndex"`
	Password string    `json:"password"`
	Files    []File    `gorm:"foreignKey:UUID""`
}
type File struct {
	gorm.Model
	FilePath string    `json:"file_path"`
	FileName string    `json:"file_name"`
	UUID     uuid.UUID `json:"uuid" gorm:"uniqueIndex"`
}
