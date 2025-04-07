package server

import (
	db "file-transfer/internal/database"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"log"
	"net/http"
	"path/filepath"
)

type Server struct {
	db     *gorm.DB
	router *gin.Engine
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

	unique := false
	for !unique {
		if !db.CheckUUIDInDB(s.db, uniqueID) {
			uniqueID, _ = uuid.NewUUID()
			unique = true
		}
	}

	dbFile := &db.File{
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
