package database

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// User представляет модель пользователя в системе
// Содержит учетные данные и связанные с пользователем файлы
type User struct {
	gorm.Model           // Стандартные поля GORM (ID, CreatedAt, UpdatedAt, DeletedAt)
	UUID       uuid.UUID `json:"uuid" gorm:"uniqueIndex"`  // Уникальный идентификатор пользователя
	Login      string    `json:"login" gorm:"uniqueIndex"` // Логин пользователя (уникален)
	Password   string    `json:"password"`                 // Пароль пользователя (хранится в открытом виде*)
	Files      []File    `gorm:"foreignKey:UserUUID"`      // Связанные с пользователем файлы
}

// *ВНИМАНИЕ*: В реальном приложении пароли должны храниться в захэшированном виде

// File представляет модель файла в системе
// Связывается с пользователем через внешний ключ
type File struct {
	gorm.Model           // Стандартные поля GORM (ID, CreatedAt, UpdatedAt, DeletedAt)
	UserUUID   uuid.UUID `json:"user_uuid"`               // Идентификатор владельца файла
	FilePath   string    `json:"file_path"`               // Физический путь к файлу на сервере
	FileName   string    `json:"file_name"`               // Оригинальное имя файла
	UUID       uuid.UUID `json:"uuid" gorm:"uniqueIndex"` // Уникальный идентификатор файла
	Size       uint64    `json:"size"`                    // Размер файла в байтах
}
