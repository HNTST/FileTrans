package database

import (
	"fmt"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"log"
)

// checkUUIDExistence — общая функция для проверки существования записи по UUID
func checkUUIDExistence(db *gorm.DB, model interface{}, uuID uuid.UUID) (bool, error) {
	const op = "CHECK_UUID_EXISTENCE"
	var count int64

	if err := db.Model(model).Where("uuid = ?", uuID).Count(&count).Error; err != nil {
		log.Printf("[%s] Ошибка проверки UUID: %v", op, err)
		return false, fmt.Errorf("%s: %w", op, err)
	}

	return count > 0, nil
}

// CheckUUIDFileInDB — проверяет существование файла по UUID
func CheckUUIDFileInDB(db *gorm.DB, uuID uuid.UUID) (bool, error) {
	return checkUUIDExistence(db, &File{}, uuID)
}

// CheckUUIDUserInDB — проверяет существование пользователя по UUID
func CheckUUIDUserInDB(db *gorm.DB, uuID uuid.UUID) (bool, error) {
	return checkUUIDExistence(db, &User{}, uuID)
}

// getRecordByField — общая функция для получения записи по полю
func getRecordByField(db *gorm.DB, out interface{}, field string, value interface{}) error {
	const op = "GET_RECORD_BY_FIELD"

	if err := db.Where(fmt.Sprintf("%s = ?", field), value).First(out).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			log.Printf("[%s] Запись не найдена: %v", op, err)
			return gorm.ErrRecordNotFound
		}
		log.Printf("[%s] Ошибка запроса: %v", op, err)
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

// GetFileByID — получает файл из БД по ID
func GetFileByID(db *gorm.DB, uuid string) (*File, error) {
	var file File
	if err := getRecordByField(db, &file, "uuid", uuid); err != nil {
		return nil, err // уже возвращается "Файл не найден"
	}
	return &file, nil
}

// GetUserByLogin — получает пользователя из БД по логину
func GetUserByLogin(db *gorm.DB, login string) (User, error) {
	var user User
	if err := getRecordByField(db, &user, "login", login); err != nil {
		return User{}, err
	}
	return user, nil
}

// CheckLoginInDB — проверяет существование логина в БД
func CheckLoginInDB(db *gorm.DB, login string) bool {
	var user User
	if err := getRecordByField(db, &user, "login", login); err != nil {
		return false
	}
	return true
}

// CreateUser — создаёт нового пользователя в БД
func CreateUser(db *gorm.DB, user *User) error {
	const op = "CREATE_USER"
	if err := db.Create(user).Error; err != nil {
		log.Printf("[%s] Ошибка создания пользователя: %v", op, err)
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

// GetAllFiles — получает все файлы из БД
func GetAllFiles(db *gorm.DB) ([]File, error) {
	const op = "GET_ALL_FILES"
	var files []File
	if err := db.Find(&files).Error; err != nil {
		log.Printf("[%s] Ошибка получения файлов: %v", op, err)
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return files, nil
}

// CreateFile — создаёт новый файл в БД
func CreateFile(db *gorm.DB, file *File) error {
	const op = "CREATE_FILE"
	if err := db.Create(file).Error; err != nil {
		log.Printf("[%s] Ошибка создания файла: %v", op, err)
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

// GetListFilesByUser — получает список файлов по логину пользователя
func GetListFilesByUser(db *gorm.DB, login string) ([]File, error) {
	var files []File
	if err := db.Where("user_uuid = (SELECT uuid FROM users WHERE login = ?)", login).Find(&files).Error; err != nil {
		return nil, fmt.Errorf("GET_LIST_FILES_BY_USER: %w", err)
	}
	return files, nil
}
