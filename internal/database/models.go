package database

import "gorm.io/gorm"

type File struct {
	gorm.Model
	FilePath string `json:"file_path"`
	FileName string `json:"file_name"`
	UUID     string `json:"uuid"@gmail.com@`
}
