package database

import (
	"fmt"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"log"
)

// Структура для хранения соединения с базой данных
type Storage struct {
	db *gorm.DB
}

// Глобальная переменная для доступа к БД
var db Storage

// NewDatabase создает новое соединение с базой данных SQLite
// Инициализирует подключение и выполняет автоматическую миграцию таблиц
// Возвращает:
// - *gorm.DB - указатель на объект базы данных
// - error - ошибка при подключении или миграции
func NewDatabase() (*gorm.DB, error) {
	log.Println("[DB_INIT] Инициализация подключения к базе данных")

	// Подключение к SQLite базе данных
	database, err := gorm.Open(sqlite.Open("files"), &gorm.Config{})
	if err != nil {
		log.Printf("[DB_INIT] Ошибка подключения к SQLite: %v", err)
		return nil, fmt.Errorf("не удалось подключиться к базе данных: %w", err)
	}

	log.Println("[DB_MIGRATE] Начинаем миграцию моделей")

	// Миграция модели файла
	if err := database.AutoMigrate(&File{}); err != nil {
		log.Printf("[DB_MIGRATE] Ошибка миграции модели File: %v", err)
		return nil, fmt.Errorf("ошибка миграции модели File: %w", err)
	}

	// Миграция модели пользователя
	if err := database.AutoMigrate(&User{}); err != nil {
		log.Printf("[DB_MIGRATE] Ошибка миграции модели User: %v", err)
		return nil, fmt.Errorf("ошибка миграции модели User: %w", err)
	}

	log.Println("[DB_INIT] База данных успешно инициализирована")

	// Сохраняем соединение в глобальной переменной
	db.db = database

	return database, nil
}
