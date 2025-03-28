package database

import "gorm.io/gorm"

func GetAllFiles(db *gorm.DB) ([]File, error) {
	const op = "GET_ALL_FILES_DATABASE"
	var files []File
	if err := db.Find(&files).Error; err != nil {
		return nil, err
	}
	return files, nil
}

func CreateFile(db *gorm.DB, file *File) error {
	return db.Create(file).Error
}

func GetFileByID(db *gorm.DB, id string) (File, error) {
	var file File
	if err := db.First(&file, id).Error; err != nil {
		return File{}, err
	}
	return file, nil
}
