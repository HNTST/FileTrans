package database

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type File struct {
	gorm.Model
	FilePath string    `json:"file_path"`
	FileName string    `json:"file_name"`
	UUID     uuid.UUID `json:"uuid"@gmail.com@`
}
