package server

import (
	"encoding/base64"
	db "file-transfer/internal/database"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

// Server Структура сервера, содержащая соединение с БД и роутер Gin
type Server struct {
	db     *gorm.DB    // Соединение с базой данных через GORM
	router *gin.Engine // Роутер Gin для обработки HTTP-запросов
}

// SignUp Регистрация нового пользователя
func (s *Server) SignUp(c *gin.Context) {
	const op = "SIGN_UP"

	var user db.User
	if err := parseJSONBody(c, &user); err != nil {
		handleError(c, http.StatusBadRequest, "Невалидный ввод", err)
		return
	}

	if user.Login == "" || user.Password == "" {
		handleError(c, http.StatusBadRequest, "Логин или пароль не могут быть пустыми", nil)
		return
	}

	if db.CheckLoginInDB(s.db, user.Login) {
		handleError(c, http.StatusBadRequest, "Логин уже занят", nil)
		return
	}

	uid, err := s.generateUniqueUUID(func(id uuid.UUID) (bool, error) {
		return db.CheckUUIDUserInDB(s.db, id)
	})
	if err != nil {
		handleError(c, http.StatusInternalServerError, "Не удалось сгенерировать уникальный ID", err)
		return
	}

	user.UUID = uid

	if err := db.CreateUser(s.db, &user); err != nil {
		handleError(c, http.StatusInternalServerError, "Ошибка создания пользователя", err)
		return
	}

	log.Printf("[%s] Пользователь успешно зарегистрирован: %s", op, user.Login)
	c.JSON(http.StatusOK, gin.H{"message": "Пользователь создан", "uuid": user.UUID})
}

// SignIn Авторизация пользователя
func (s *Server) SignIn(c *gin.Context) {
	const op = "SIGN_IN"

	var credentials struct {
		Login    string `json:"login"`
		Password string `json:"password"`
	}

	if err := parseJSONBody(c, &credentials); err != nil {
		handleError(c, http.StatusBadRequest, "Невалидный ввод", err)
		return
	}

	dbUser, err := db.GetUserByLogin(s.db, credentials.Login)
	if err != nil {
		handleError(c, http.StatusUnauthorized, "Неверный логин или пароль", err)
		return
	}

	if dbUser.Password != credentials.Password {
		handleError(c, http.StatusUnauthorized, "Неверный логин или пароль", nil)
		return
	}

	log.Printf("[%s] Пользователь вошёл: %s", op, dbUser.Login)
	c.JSON(http.StatusOK, gin.H{"message": "Успешный вход", "uuid": dbUser.UUID})
}

// GetListFilesByUser Получение списка файлов конкретного пользователя
func (s *Server) GetListFilesByUser(c *gin.Context) {
	const op = "GET_FILES_BY_USER"

	var user struct {
		Login string `json:"login"` // JSON использует любые имена — это нормально
	}

	if err := parseJSONBody(c, &user); err != nil {
		handleError(c, http.StatusBadRequest, "Невалидный ввод", err)
		return
	}

	files, err := db.GetListFilesByUser(s.db, user.Login)
	if err != nil {
		handleError(c, http.StatusInternalServerError, "Ошибка получения списка файлов", err)
		return
	}

	if len(files) == 0 {
		c.JSON(http.StatusOK, gin.H{"message": "Файлов нет"})
		return
	}

	log.Printf("[%s] Получено файлов для %s: %d", op, user.Login, len(files))
	c.IndentedJSON(http.StatusOK, files)
}

// GetAllFiles Получение всех файлов из БД
func (s *Server) GetAllFiles(c *gin.Context) {
	const op = "GET_ALL_FILES"

	files, err := db.GetAllFiles(s.db)
	if err != nil {
		handleError(c, http.StatusInternalServerError, "Ошибка получения всех файлов", err)
		return
	}

	if len(files) == 0 {
		c.JSON(http.StatusOK, gin.H{"message": "Файлы не найдены"})
		return
	}

	log.Printf("[%s] Общее количество файлов: %d", op, len(files))
	c.IndentedJSON(http.StatusOK, files)
}

// UploadFile Загрузка файла на сервер
// UploadFile Загрузка файла через JSON (например, с base64 содержимым)
func (s *Server) UploadFile(c *gin.Context) {
	const op = "UPLOAD_FILE"

	// Создание директории uploads, если её нет
	if _, err := os.Stat("uploads"); os.IsNotExist(err) {
		log.Printf("[%s] Создаём папку uploads", op)
		if err := os.Mkdir("uploads", os.ModePerm); err != nil {
			handleError(c, http.StatusInternalServerError, "Не удалось создать папку uploads", err)
			return
		}
	}

	// Парсим JSON из тела запроса
	var req struct {
		FileName string    `json:"fileName"` // Имя файла
		UserUUID uuid.UUID `json:"userUuid"` // UUID пользователя
		FileData string    `json:"fileData"` // Base64-кодированный контент файла
	}

	if err := parseJSONBody(c, &req); err != nil {
		handleError(c, http.StatusBadRequest, "Невалидный JSON", err)
		return
	}

	if req.FileName == "" || req.FileData == "" || req.UserUUID == uuid.Nil {
		handleError(c, http.StatusBadRequest, "Все поля обязательны", nil)
		return
	}

	// Декодируем base64-строку в байты
	data, err := base64.StdEncoding.DecodeString(req.FileData)
	if err != nil {
		handleError(c, http.StatusBadRequest, "Ошибка декодирования данных файла", err)
		return
	}

	// Путь для сохранения файла
	path := filepath.Join("uploads", req.FileName)
	if err := os.WriteFile(path, data, 0644); err != nil {
		handleError(c, http.StatusInternalServerError, "Не удалось записать файл", err)
		return
	}

	// Генерация UUID файла
	uid, err := s.generateUniqueUUID(func(id uuid.UUID) (bool, error) {
		return db.CheckUUIDFileInDB(s.db, id)
	})
	if err != nil {
		handleError(c, http.StatusInternalServerError, "Не удалось сгенерировать UUID файла", err)
		return
	}

	// Сохраняем информацию о файле в БД
	dbFile := &db.File{
		UserUUID: req.UserUUID,
		FilePath: path,
		FileName: req.FileName,
		UUID:     uid,
	}

	if err := db.CreateFile(s.db, dbFile); err != nil {
		handleError(c, http.StatusInternalServerError, "Ошибка сохранения файла в БД", err)
		return
	}

	log.Printf("[%s] Файл успешно загружен: %s", op, dbFile.FileName)
	c.JSON(http.StatusOK, dbFile)
}

// DownloadFile Скачивание файла по UUID
func (s *Server) DownloadFile(c *gin.Context) {
	id := c.Param("id")

	file, err := db.GetFileByID(s.db, id)
	if err != nil {
		handleError(c, http.StatusNotFound, "Файл не найден", err)
		return
	}

	// Защита от отсутствия физического файла
	if _, err := os.Stat(file.FilePath); os.IsNotExist(err) {
		handleError(c, http.StatusNotFound, "Физический файл не найден", err)
		return
	}

	log.Printf("[DOWNLOAD_FILE] Отправляем файл: %s", file.FileName)
	c.File(file.FilePath)
}

// Вспомогательные функции

func JSONMessage(c *gin.Context, text string) {
	c.JSON(http.StatusOK, gin.H{"message": text})
}

func JSONError(c *gin.Context, status int, textErr string) {
	log.Printf("[ERROR] %s", textErr)
	c.JSON(status, gin.H{"error": textErr})
}

func handleError(c *gin.Context, status int, message string, err error) {
	log.Printf("[%s] Ошибка: %v", message, err)
	c.JSON(status, gin.H{"error": message})
}

func parseJSONBody(c *gin.Context, obj interface{}) error {
	const op = "PARSE_JSON_BODY"
	if err := c.ShouldBindBodyWithJSON(obj); err != nil {
		log.Printf("[%s] Ошибка парсинга JSON: %v", op, err)
		return fmt.Errorf("недопустимый формат ввода")
	}
	return nil
}

func (s *Server) generateUniqueUUID(checkFn func(uuid.UUID) (bool, error)) (uuid.UUID, error) {
	const op = "GENERATE_UNIQUE_UUID"
	for i := 0; i < 5; i++ { // max retries
		uid := uuid.New()
		exists, err := checkFn(uid)
		if err != nil {
			log.Printf("[%s] Ошибка проверки UUID: %v", op, err)
			continue
		}
		if !exists {
			return uid, nil
		}
	}
	err := fmt.Errorf("не удалось сгенерировать уникальный UUID за 5 попыток")
	log.Printf("[%s] Ошибка: %v", op, err)
	return uuid.Nil, err
}
