package database

import (
	"fmt"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

type Storage struct {
	db *gorm.DB
}

var db Storage

func NewDatabase() (*gorm.DB, error) {

	fmt.Println("Подлючение к базе данных")
	var err error
	db.db, err = gorm.Open(sqlite.Open("files"), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	err = db.db.AutoMigrate(&File{})
	if err != nil {
		return nil, err
	}
	err = db.db.AutoMigrate(&User{})
	if err != nil {
		return nil, err
	}
	return db.db, nil
}
