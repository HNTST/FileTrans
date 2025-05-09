package server

import (
	"bytes"
	"encoding/base64"
	db "file-transfer/internal/database"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/unidoc/unipdf/v3/model"
	"github.com/unidoc/unipdf/v3/render"
	"gorm.io/gorm"
	"image/png"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
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
func (s *Server) UploadFile(c *gin.Context) {
	const op = "UPLOAD_FILE"

	// Создание директории uploads
	if _, err := os.Stat("uploads"); os.IsNotExist(err) {
		log.Printf("[%s] Создаем папку uploads", op)
		if err := os.Mkdir("uploads", os.ModePerm); err != nil {
			handleError(c, http.StatusInternalServerError, "Не удалось создать папку uploads", err)
			return
		}
	}

	// Получаем файл из формы
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		handleError(c, http.StatusBadRequest, "Ошибка получения файла", err)
		return
	}
	defer func(file multipart.File) {
		err := file.Close()
		if err != nil {

		}
	}(file)

	// Читаем содержимое файла
	data, err := io.ReadAll(file)
	if err != nil {
		handleError(c, http.StatusInternalServerError, "Не удалось прочитать файл", err)
		return
	}

	size := uint64(len(data))

	// Получаем дополнительные параметры
	userUuid := c.PostForm("userUuid")
	if userUuid == "" {
		handleError(c, http.StatusBadRequest, "Отсутствует userUuid", nil)
		return
	}

	// Сохраняем файл на диск
	path := filepath.Join("uploads", header.Filename)
	if err := os.WriteFile(path, data, 0644); err != nil {
		handleError(c, http.StatusInternalServerError, "Не удалось записать файл", err)
		return
	}

	// Генерируем уникальный UUID для файла
	uid, err := s.generateUniqueUUID(func(id uuid.UUID) (bool, error) {
		return db.CheckUUIDFileInDB(s.db, id)
	})
	if err != nil {
		handleError(c, http.StatusInternalServerError, "Не удалось сгенерировать UUID файла", err)
		return
	}

	// Сохраняем информацию о файле в БД
	dbFile := &db.File{
		UUID:     uid,
		UserUUID: uuid.MustParse(userUuid),
		FilePath: path,
		FileName: header.Filename,
		Size:     size,
	}

	if err := db.CreateFile(s.db, dbFile); err != nil {
		handleError(c, http.StatusInternalServerError, "Ошибка сохранения в БД", err)
		return
	}

	log.Printf("[%s] Файл успешно загружен: %s", op, dbFile.FileName)
	c.JSON(http.StatusOK, gin.H{"message": "Файл загружен", "fileUUID": dbFile.UUID})
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

func (s *Server) DeleteFileHandler(c *gin.Context) {
	id := c.Param("uuid")
	err, file := db.DeleteFile(s.db, id)
	if err != nil {
		if strings.Contains(err.Error(), "файл не найден") {
			handleError(c, http.StatusNotFound, "Файл не найден", err)
		} else {
			handleError(c, http.StatusInternalServerError, "Ошибка удаления файла", err)
		}
		return
	}

	log.Printf("[DELETE_FILE] Файл c UUID успешно удален: %s", file.UUID)
	c.JSON(http.StatusOK, gin.H{"message": "Файл удален"})
}

func (s *Server) UpdateFileName(c *gin.Context) {
	uuIdParam := c.Param("uuid")

	var newFile db.File
	if err := parseJSONBody(c, &newFile); err != nil {
		handleError(c, http.StatusBadRequest, "Невалидный ввод", err)
		return
	}

	if newFile.FileName == "" {
		handleError(c, http.StatusBadRequest, "Имя файла не может быть пустым", nil)
		return
	}

	if err := db.UpdateFileName(s.db, uuIdParam, newFile.FileName); err != nil {
		handleError(c, http.StatusInternalServerError, "Ошибка изменения имени файла", err)
		return
	}

	log.Printf("[UPDATE_NAME_FILE] Имя файла обновлено на %s", newFile.FileName)
	c.JSON(http.StatusOK, gin.H{"message": "Имя файла обновлено"})
}

func (s *Server) GetPDFPage(c *gin.Context) {
	const op = "GET_PDF_PAGE"

	uuid := c.Param("uuid")
	pageNum, err := strconv.Atoi(c.Param("page"))
	if err != nil {
		handleError(c, http.StatusBadRequest, "Неверный номер страницы", err)
		return
	}

	// 1. Получаем файл из БД
	file, err := db.GetFileByID(s.db, uuid)
	if err != nil {
		handleError(c, http.StatusNotFound, "Файл не найден", err)
		return
	}

	// 2. Проверяем, что это PDF
	if !strings.HasSuffix(file.FileName, ".pdf") {
		handleError(c, http.StatusBadRequest, "Файл не является PDF", nil)
		return
	}

	// 3. Открываем файл
	f, err := os.Open(file.FilePath)
	if err != nil {
		handleError(c, http.StatusInternalServerError, "Не удалось открыть файл", err)
		return
	}
	defer f.Close()

	// 4. Загружаем PDF
	pdfReader, err := model.NewPdfReader(f)
	if err != nil {
		handleError(c, http.StatusInternalServerError, "Ошибка чтения PDF", err)
		return
	}

	// 5. Проверяем номер страницы
	numPages, err := pdfReader.GetNumPages()
	if err != nil || pageNum < 1 || pageNum > numPages {
		handleError(c, http.StatusBadRequest, "Неверный номер страницы", nil)
		return
	}

	// 6. Получаем страницу
	page, err := pdfReader.GetPage(pageNum)
	if err != nil {
		handleError(c, http.StatusInternalServerError, "Ошибка получения страницы", err)
		return
	}

	// 7. Рендерим страницу в base64
	device := render.NewImageDevice()
	img, err := device.Render(page)
	if err != nil {
		handleError(c, http.StatusInternalServerError, "Ошибка рендера страницы", err)
		return
	}

	// 8. Кодируем изображение в base64
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		handleError(c, http.StatusInternalServerError, "Ошибка кодирования PNG", err)
		return
	}

	// 9. Отправляем ответ
	c.JSON(http.StatusOK, gin.H{
		"image":      "data:image/png;base64," + base64.StdEncoding.EncodeToString(buf.Bytes()),
		"page":       pageNum,
		"totalPages": numPages,
	})
}

// Вспомогательные функции

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
