package server

import (
	db "file-transfer/internal/database"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

type Server struct {
	db     *gorm.DB
	router *gin.Engine
}

func (s *Server) SignUp(c *gin.Context) {
	const op = "SING_UP"
	var user db.User
	if err := c.ShouldBindBodyWithJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Недопустимый ввод"})
		return
	}

	if user.Login == "" || user.Password == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Проверте поля логина и пароля, возможно они пусты"})
		return
	}
	if db.CheckLoginInDB(s.db, user.Login) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Такой логин занят"})
		return
	}

	uniqueID, _ := uuid.NewUUID()

	_, unique := db.CheckUUIDUserInDB(s.db, uniqueID)
	if unique {
		uniqueID, _ = uuid.NewUUID()
	}
	user.UUID = uniqueID

	if err := db.CreateUser(s.db, &user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Произошла ошибка создания нового пользователя, попробуйте снова"})
		return
	}
}

func (s *Server) SignIn(c *gin.Context) {
	const op = "SING_IN"
	var user db.User
	if err := c.ShouldBindBodyWithJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Недопустимый ввод"})
		return
	}
	if user.Login == "" || user.Password == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Проверте поля логина и пароля, возможно они пусты"})
		return
	}

	dbUser, err := db.GetUserByLOGIN(s.db, user.Login)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Что пошло не так"})
		log.Printf("Ошибка в получении логина из датабазы: %v", err)
		return
	}

	if dbUser.Password == user.Password {
		c.JSON(http.StatusOK, gin.H{"message": "Успешный вход"})
		return
	}
}

func (s *Server) GetFilesByUser(c *gin.Context) {
	const op = "GET_USER'S_ FILES"
	var user db.User

	if err := c.ShouldBindBodyWithJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Недопустимый ввод"})
		return
	}

	files, err := db.GetListFilesByUser(s.db, user.Login)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Произошла ошибка получения списка файлов"})
		return
	}
	if len(files) == 0 {
		c.JSON(http.StatusOK, gin.H{"message": "У пользователя нет сохранненых файлов"})
		return
	} else {
		c.IndentedJSON(http.StatusOK, files)
	}
}

func (s *Server) GetAllFiles(c *gin.Context) {
	const op = "GET_ALL_FILES_SERVER_FUNC"
	var files []db.File
	files, err := db.GetAllFiles(s.db)
	if err != nil {
		log.Printf("Ошибка получения файлов: %v", err)
		JSONError(c, http.StatusInternalServerError, "Ошибка сервера при получении файлов")
		return
	}
	if len(files) == 0 {
		JSONError(c, http.StatusOK, "Файлы не загруженны")
		return
	} else {
		c.IndentedJSON(http.StatusOK, files)
	}
}

func (s *Server) UploadFile(c *gin.Context) {
	if _, err := os.Stat("uploads"); os.IsNotExist(err) {
		os.Mkdir("uploads", os.ModePerm)
	}
	file, err := c.FormFile("file")
	if err != nil {
		JSONError(c, http.StatusBadRequest, "Не удалось получить файл")
		return
	}

	fileName := filepath.Base(file.Filename)
	path := filepath.Join("uploads", fileName)

	if err := c.SaveUploadedFile(file, path); err != nil {
		JSONError(c, http.StatusInternalServerError, "Не удалось сохранить файл")
		return
	}
	uniqueID, _ := uuid.NewUUID()

	err, unique := db.CheckUUIDFileInDB(s.db, uniqueID)
	if unique {
		uniqueID, _ = uuid.NewUUID()
	}

	var user db.User

	if err := c.ShouldBindBodyWithJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Недопустимый ввод"})
		return
	}

	dbFile := &db.File{
		UserUUID: user.UUID,
		FilePath: path,
		FileName: fileName,
		UUID:     uniqueID,
	}

	if err := db.CreateFile(s.db, dbFile); err != nil {
		JSONError(c, http.StatusInternalServerError, "Ошибка сохранения данных файла")
		return
	}

	c.JSON(http.StatusOK, dbFile)
}

func (s *Server) DownloadFile(c *gin.Context) {
	id := c.Param("id")
	file, err := db.GetFileByID(s.db, id)
	if err != nil {
		JSONError(c, http.StatusNotFound, "Файл не найден")
		return
	}

	c.File(file.FilePath)
}

func JSONMessage(c *gin.Context, text string) {
	c.JSON(http.StatusOK, gin.H{
		"message": text,
	})
}
func JSONError(c *gin.Context, status int, textErr string) {
	c.JSON(status, gin.H{
		"error": textErr,
	})
}
