package server

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"net/http/httptest" // Для имитации HTTP-запросов
	"os"                // Для работы с файлами
	"path/filepath"
	"strings"
	"testing" // Стандартная библиотека тестирования
	"time"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert" // Ассерты
	"gorm.io/gorm"                       // ORM
)

type User struct {
	gorm.Model
	UUID     uuid.UUID `json:"uuid" gorm:"uniqueIndex"`
	Login    string    `json:"login" gorm:"uniqueIndex"`
	Password string    `json:"password"`
	Files    []File    `gorm:"foreignKey:UserUUID"`
}

type File struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at,omitempty"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
	UserUUID  uuid.UUID      `json:"user_uuid"`
	FilePath  string         `json:"file_path"`
	FileName  string         `json:"file_name"`
	UUID      uuid.UUID      `json:"uuid" gorm:"uniqueIndex"`
}

func setupTestDB() *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}
	db.AutoMigrate(User{}, File{})
	return db
}

func setupTestServer(t *testing.T) (*Server, *gorm.DB) {
	// Отключаем логи Gin в тестах

	gin.SetMode(gin.TestMode)

	// Создаем временную БД в памяти
	db := setupTestDB()

	// Создаем сервер с тестовой БД
	server := &Server{
		db:     db,
		router: gin.Default(),
	}
	server.setupRoutes()

	return server, db
}
func TestSignUp_Success(t *testing.T) {
	// Arrange
	server, _ := setupTestServer(t)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	// Создаем тело запроса
	requestBody := `{"login":"testuser","password":"password"}`
	c.Request = &http.Request{
		Header: make(http.Header),
		Body:   io.NopCloser(strings.NewReader(requestBody)),
	}

	// Act
	server.SignUp(c)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "Пользователь создан")
	assert.Contains(t, w.Body.String(), "uuid")
}
func TestSignUp_DuplicateLogin(t *testing.T) {
	// Arrange
	server, db := setupTestServer(t)

	// Создаем существующего пользователя
	existingUser := User{
		UUID:     uuid.New(),
		Login:    "existing",
		Password: "pass",
	}
	db.Create(&existingUser)

	// Настраиваем тестовый запрос
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	requestBody := `{"login":"existing","password":"password"}`
	c.Request = &http.Request{
		Header: make(http.Header),
		Body:   io.NopCloser(strings.NewReader(requestBody)),
	}

	// Act
	server.SignUp(c)

	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Логин уже занят")
}

func TestSignUp_EmptyFields(t *testing.T) {
	tests := []struct {
		name       string
		body       string
		wantStatus int
	}{
		{
			name:       "Empty Login",
			body:       `{"login":"","password":"password"}`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "Empty Password",
			body:       `{"login":"testuser","password":""}`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "Both Empty",
			body:       `{"login":"","password":""}`,
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			server, _ := setupTestServer(t)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			c.Request = &http.Request{
				Header: make(http.Header),
				Body:   io.NopCloser(strings.NewReader(tt.body)),
			}

			// Act
			server.SignUp(c)

			// Assert
			assert.Equal(t, tt.wantStatus, w.Code)
			assert.Contains(t, w.Body.String(), "Логин или пароль не могут быть пустыми")
		})
	}
}

func TestSignIn_Success(t *testing.T) {
	// Arrange
	server, db := setupTestServer(t)
	user := User{
		UUID:     uuid.New(),
		Login:    "testuser",
		Password: "password",
	}
	db.Create(&user)

	// Настраиваем тестовый запрос
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	requestBody := `{"login":"testuser","password":"password"}`
	c.Request = &http.Request{
		Header: make(http.Header),
		Body:   io.NopCloser(strings.NewReader(requestBody)),
	}

	// Act
	server.SignIn(c)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "Успешный вход")
	assert.Contains(t, w.Body.String(), user.UUID.String())
}

func TestSignIn_InvalidPassword(t *testing.T) {
	// Arrange
	server, db := setupTestServer(t)
	user := User{
		UUID:     uuid.New(),
		Login:    "testuser",
		Password: "correct_password",
	}
	db.Create(&user)

	// Настраиваем тестовый запрос
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	requestBody := `{"login":"testuser","password":"wrong_password"}`
	c.Request = &http.Request{
		Header: make(http.Header),
		Body:   io.NopCloser(strings.NewReader(requestBody)),
	}

	// Act
	server.SignIn(c)

	// Assert
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "Неверный логин или пароль")
}

func TestDownloadFile_Success(t *testing.T) {
	server, db := setupTestServer(t)

	user := User{
		UUID:     uuid.New(),
		Login:    "downloader",
		Password: "password",
	}
	db.Create(&user)

	fileUUID := uuid.MustParse("e81aba4c-2921-48f6-8ae8-55d8eeca0a18")

	// Получаем текущую директорию
	dir, err := os.Getwd()
	require.NoError(t, err)
	filePath := filepath.Join(dir, "test_download.txt")

	// Создаём файл до добавления в БД
	err = os.WriteFile(filePath, []byte("download test content"), 0644)
	require.NoError(t, err)
	defer os.Remove(filePath)

	// Добавляем запись о файле
	file := File{
		UserUUID: user.UUID,
		FilePath: filePath,
		FileName: "test_download.txt",
		UUID:     fileUUID,
	}
	db.Create(&file)

	// Настраиваем контекст
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "id", Value: file.UUID.String()}}

	// Вызываем хендлер
	server.DownloadFile(c)

	// Проверяем результат
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Header().Get("Content-Disposition"), "attachment")
	assert.Contains(t, w.Body.String(), "download test content")
}

func TestGetAllFiles_Success(t *testing.T) {
	server, db := setupTestServer(t)

	// Подготавливаем данные
	files := []File{
		{
			UserUUID: uuid.New(),
			FilePath: "/tmp/file1.txt",
			FileName: "file1.txt",
			UUID:     uuid.New(),
		},
		{
			UserUUID: uuid.New(),
			FilePath: "/tmp/file2.txt",
			FileName: "file2.txt",
			UUID:     uuid.New(),
		},
	}
	db.Create(&files)

	// Настраиваем запрос
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = &http.Request{}

	// Вызываем хендлер
	server.GetAllFiles(c)

	// Проверяем
	assert.Equal(t, http.StatusOK, w.Code)

	var result []map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &result)
	assert.NoError(t, err)
	assert.Len(t, result, 2)

	assert.Equal(t, files[0].FileName, result[0]["file_name"])
	assert.Equal(t, files[1].FileName, result[1]["file_name"])
	assert.Equal(t, files[0].UserUUID.String(), result[0]["user_uuid"])
	assert.Equal(t, files[1].UserUUID.String(), result[1]["user_uuid"])
}

func TestGetListFilesByUser_Success(t *testing.T) {
	// Arrange
	server, db := setupTestServer(t)

	// Создаем тестовых пользователей и файлы
	user := User{
		UUID:     uuid.New(),
		Login:    "owner",
		Password: "password",
	}
	db.Create(&user)

	files := []File{
		{
			UserUUID: user.UUID,
			FilePath: "/tmp/user1.txt",
			FileName: "user1.txt",
			UUID:     uuid.New(),
		},
		{
			UserUUID: user.UUID,
			FilePath: "/tmp/user2.txt",
			FileName: "user2.txt",
			UUID:     uuid.New(),
		},
	}
	db.Create(&files)

	// Настраиваем тестовый запрос
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	requestBody := fmt.Sprintf(`{"login":"%s"}`, user.Login)
	c.Request = &http.Request{
		Header: make(http.Header),
		Body:   io.NopCloser(strings.NewReader(requestBody)),
	}

	// Act
	server.GetListFilesByUser(c)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "user1.txt")
	assert.Contains(t, w.Body.String(), "user2.txt")
}

func TestUploadFile_Success(t *testing.T) {
	// Arrange
	server, db := setupTestServer(t)

	// Создаем пользователя
	userUUID := uuid.New()
	user := User{
		UUID:     userUUID,
		Login:    "uploader",
		Password: "password",
	}
	db.Create(&user)

	// Подготавливаем base64-данные файла
	fileName := "testfile.txt"
	fileContent := []byte("This is a test file content.")
	base64Content := base64.StdEncoding.EncodeToString(fileContent)

	// Формируем JSON-запрос
	requestBody := fmt.Sprintf(`{
        "fileName": "%s",
        "userUuid": "%s",
        "fileData": "%s"
    }`, fileName, userUUID.String(), base64Content)

	// Настраиваем тестовый контекст
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = &http.Request{
		Method: "POST",
		Header: map[string][]string{
			"Content-Type": {"application/json"},
		},
		Body: io.NopCloser(strings.NewReader(requestBody)),
	}

	// Act
	server.UploadFile(c)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	// Проверяем наличие файла на диске
	filePath := filepath.Join("uploads", fileName)
	defer os.Remove(filePath) // Чистим после себя

	content, err := os.ReadFile(filePath)
	assert.NoError(t, err)
	assert.Equal(t, fileContent, content)

	// Проверяем, что запись создана в БД
	var dbFile File
	db.Where("uuid = ?", response["uuid"].(string)).First(&dbFile)
	assert.Equal(t, fileName, dbFile.FileName)
	assert.Equal(t, userUUID, dbFile.UserUUID)
	assert.Equal(t, filePath, dbFile.FilePath)
}
