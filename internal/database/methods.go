package database

import (
	"fmt"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"log"
)

func CheckLoginInDB(db *gorm.DB, login string) bool {
	const op = "CHECK_LOGIN_IN_DATABASE"
	var user User

	// Ищем пользователя с указанным логином
	if err := db.Where("login = ?", login).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			// Логин не найден в базе данных
			return false
		}
		// Обработка других ошибок (например, проблемы с подключением к БД)
		log.Printf("[%s] Ошибка проверки логина: %v", op, err)
		return false
	}

	// Логин найден в базе данных
	return true
}

func GetUserByLOGIN(db *gorm.DB, login string) (User, error) {
	const op = "CHECK_PASSWORD_BY_LOGIN"
	var user User

	if err := db.Where("Login = ?", login).First(&user).Error; err != nil {
		return User{}, err
	}
	return user, nil
}

func CreateUser(db *gorm.DB, user *User) error {
	return db.Create(user).Error
}

func GetAllFiles(db *gorm.DB) ([]File, error) {
	const op = "GET_ALL_FILES_DATABASE"
	var files []File
	if err := db.Find(&files).Error; err != nil {
		return nil, err
	}
	return files, nil
}

func CreateFile(db *gorm.DB, file *File) error {
	return db.Create(file).Error
}

func GetFileByID(db *gorm.DB, uuid string) (File, error) {
	var file File
	if err := db.Where("uuid = ?", uuid).First(&file).Error; err != nil {
		return File{}, err
	}
	return file, nil
}

func CheckUUIDFileInDB(db *gorm.DB, uuID uuid.UUID) (error, bool) {
	var files File
	var count int64
	if err := db.Model(&File{}).Where("uuid = ?", uuID).Count(&count).Error; err != nil {
		return err, false
	}
	return nil, count > 0

	if uuID == files.UUID {
		return nil, true
	}
	return nil, false
}

func CheckUUIDUserInDB(db *gorm.DB, uuID uuid.UUID) (error, bool) {
	var user User
	var count int64
	if err := db.Model(&User{}).Where("uuid = ?", uuID).Count(&count).Error; err != nil {
		return err, false
	}
	return nil, count > 0
	if uuID == user.UUID {
		return nil, true
	}
	return nil, false

}

func GetListFilesByUser(db *gorm.DB, login string) ([]File, error) {
	var files []File

	if err := db.Where("Login = ?", login).Find(&files).Error; err != nil {
		log.Printf("не удалось получить файлы пользователя: %v", err)
		return nil, fmt.Errorf("не удалось получить файлы пользователя")
	}
	return files, nil
}
