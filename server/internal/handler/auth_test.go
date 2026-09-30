package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"semi-mes/server/internal/config"
	"semi-mes/server/migrations"
	"semi-mes/server/pkg/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	if err := migrations.AutoMigrate(db); err != nil {
		t.Fatalf("Failed to migrate test database: %v", err)
	}

	if err := migrations.SeedData(db); err != nil {
		t.Fatalf("Failed to seed test data: %v", err)
	}

	return db
}

func TestLogin(t *testing.T) {
	config.GlobalConfig = &config.Config{
		JWT: config.JWTConfig{
			Secret:      "test-secret",
			ExpireHours: 24,
		},
	}

	db := setupTestDB(t)
	handler := NewAuthHandler(db)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/login", handler.Login)

	loginReq := LoginRequest{
		Username: "admin",
		Password: "admin123",
	}
	body, _ := json.Marshal(loginReq)

	req, _ := http.NewRequest("POST", "/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
	}

	var resp utils.Response
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if resp.Code != 0 {
		t.Errorf("Expected response code 0, got %d", resp.Code)
	}
}

func TestLoginInvalidCredentials(t *testing.T) {
	config.GlobalConfig = &config.Config{
		JWT: config.JWTConfig{
			Secret:      "test-secret",
			ExpireHours: 24,
		},
	}

	db := setupTestDB(t)
	handler := NewAuthHandler(db)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/login", handler.Login)

	loginReq := LoginRequest{
		Username: "admin",
		Password: "wrongpassword",
	}
	body, _ := json.Marshal(loginReq)

	req, _ := http.NewRequest("POST", "/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	var resp utils.Response
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if resp.Code != 40001 {
		t.Errorf("Expected response code 40001, got %d", resp.Code)
	}
}
