package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"semi-mes/server/internal/config"
	"semi-mes/server/internal/middleware"
	"semi-mes/server/pkg/utils"

	"github.com/gin-gonic/gin"
)

func TestCreateFactory(t *testing.T) {
	config.GlobalConfig = &config.Config{
		JWT: config.JWTConfig{
			Secret:      "test-secret",
			ExpireHours: 24,
		},
	}

	db := setupTestDB(t)
	middleware.InitPermissionMiddleware(db)
	handler := NewBaseDataHandler(db)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/factories", handler.CreateFactory)

	factoryReq := FactoryRequest{
		FactoryCode: "F001",
		FactoryName: "Test Factory",
		Address:     "Test Address",
		Status:      1,
	}
	body, _ := json.Marshal(factoryReq)

	req, _ := http.NewRequest("POST", "/factories", bytes.NewBuffer(body))
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

func TestListFactories(t *testing.T) {
	db := setupTestDB(t)
	handler := NewBaseDataHandler(db)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/factories", handler.ListFactories)

	req, _ := http.NewRequest("GET", "/factories?page=1&page_size=20", nil)
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
