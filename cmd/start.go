package cmd

import (
	"file-transfer/internal/database"
	"file-transfer/internal/server"
	"log"
)

// Start - точка входа приложения
// Инициализирует базу данных и запускает HTTP-сервер
// Вызывает os.Exit(1) при фатальных ошибках
func Start() {
	log.Println("[START] Инициализация приложения")

	// Инициализация базы данных
	log.Println("[DB_INIT] Подключение к базе данных")
	initDatabase, err := database.NewDatabase()
	if err != nil {
		log.Fatalf("[DB_INIT] Ошибка подключения к БД: %v", err)
	}

	// Создание таблиц (миграция моделей)
	log.Println("[DB_MIGRATE] Миграция моделей")

	// Создание экземпляра сервера
	log.Println("[SERVER_INIT] Инициализация сервера")
	srv := server.NewServer(initDatabase)

	// Запуск сервера
	log.Println("[SERVER_START] Запуск сервера на :8080")
	if err := srv.Run(":8080"); err != nil {
		log.Fatalf("[SERVER_START] Ошибка запуска сервера: %v", err)
	}
}
