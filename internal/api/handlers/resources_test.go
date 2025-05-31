package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/deepanshumishra/devcost-api/internal/config"
	"github.com/deepanshumishra/devcost-api/internal/models"
	"github.com/gin-gonic/gin"
)

func TestGetResourcesByTags(t *testing.T) {
	// Initialize config
	cfg, _ := config.NewConfig()

	// Set up Gin router
	r := gin.Default()
	r.GET("/getresourcesbytag", GetResourcesByTags(cfg))

	// Test valid request
	req, _ := http.NewRequest("GET", "/getresourcesbytag?tag_key=project&tag_value=dev-cluster", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// Check status code
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	// Parse response
	var resp struct {
		Resources []models.Resource `json:"resources"`
		Warning   string            `json:"warning,omitempty"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Errorf("Failed to parse response: %v", err)
	}

	// Test invalid request (missing tag_key)
	req, _ = http.NewRequest("GET", "/getresourcesbytag?tag_value=dev-cluster", nil)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}