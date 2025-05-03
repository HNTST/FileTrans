package server

import (
	"log"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// Создание нового сервера и настройка маршрутов
// db - соединение с базой данных через GORM
// Возвращает указатель на Server
func NewServer(db *gorm.DB) *Server {
	log.Println("[NEW_SERVER] Инициализация сервера")
	server := &Server{
		db:     db,
		router: gin.Default(),
	}
	server.setupRoutes()
	return server
}

// Настройка маршрутов сервера
// Регистрирует все HTTP-эндпоинты и связывает их с обработчиками
func (s *Server) setupRoutes() {
	log.Println("[SETUP_ROUTES] Настраиваем маршруты")

	// Статические файлы
	s.router.Static("/static", "./public")

	// Главная страница
	s.router.GET("/", func(c *gin.Context) {
		c.File("./public/index.html")
	})

	s.router.DELETE("/delete-file/:uuid", s.DeleteFileHandler)

	// Загрузка файла
	s.router.POST("/upload", s.UploadFile)

	// Получение всех файлов
	s.router.GET("/files", s.GetAllFiles)

	// Скачивание файла по ID
	s.router.GET("/download/:id", s.DownloadFile)

	// Регистрация пользователя
	s.router.POST("/register", s.SignUp)

	// Получение файлов пользователя
	s.router.POST("/usersFiles", s.GetListFilesByUser)

	// Авторизация пользователя
	s.router.POST("/signInPage", s.SignIn)

	s.router.PATCH("/updateFileName/:uuid", s.UpdateFileName)

	s.router.GET("/api/pdf/:uuid/page/:page", s.GetPDFPage)

}

// Запуск HTTP-сервера
// addr - адрес в формате "host:port" (например, ":8080")
// Возвращает ошибку при проблемах с запуском сервера
func (s *Server) Run(addr string) error {
	log.Printf("[RUN_SERVER] Сервер запущен на %s", addr)
	return s.router.Run(addr)
}
