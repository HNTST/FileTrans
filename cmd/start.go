package cmd

import (
	"file-transfer/internal/database"
	"file-transfer/internal/server"
	"log"
)

func Start() {
	initDatabase, err := database.InitDatabase()
	if err != nil {
		log.Fatal(err)
	}
	srv := server.NewServer(initDatabase)
	if err := srv.Run(":8080"); err != nil {
		log.Fatal(err)
	}
}
