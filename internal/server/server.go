package server

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func NewServer(db *gorm.DB) *Server {
	server := &Server{
		db:     db,
		router: gin.Default(),
	}
	server.setupRoutes()
	return server
}

func (s *Server) setupRoutes() {

	s.router.Static("/static", "./public")
	s.router.GET("/", func(c *gin.Context) {
		c.File("./public/index.html")
	})

	s.router.POST("/upload", s.UploadFile)
	s.router.GET("/files", s.GetAllFiles)
	s.router.GET("/download/:id", s.DownloadFile)
	s.router.POST("/register", s.SignUp)
	s.router.GET("/usersFiles", s.GetFilesByUser)
	s.router.POST("/signInPage", s.SignIn)
}

func (s *Server) Run(addr string) error {
	return s.router.Run(addr)
}
