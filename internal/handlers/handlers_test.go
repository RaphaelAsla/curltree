package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"curltree/internal/database"
	"curltree/internal/models"
)

func setupTestHandler(t *testing.T) *Handler {
	tmpFile := t.TempDir() + "/test.db"
	db, err := database.NewSQLiteDB(tmpFile)
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	return NewHandler(db)
}

func TestGetProfile(t *testing.T) {
	handler := setupTestHandler(t)

	req := &models.CreateUserRequest{
		SSHPublicKey: "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQ test",
		FullName:     "Test User",
		Username:     "testuser",
		About:        "Test about",
		Links: []models.LinkInput{
			{Name: "Website", URL: "https://example.com"},
		},
	}

	_, err := handler.db.CreateUser(req)
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	t.Run("JSON response", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/testuser", nil)
		req.Header.Set("Accept", "application/json")
		w := httptest.NewRecorder()

		handler.GetProfile(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		var profile models.PublicProfile
		if err := json.Unmarshal(w.Body.Bytes(), &profile); err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		if profile.Username != "testuser" {
			t.Errorf("Expected username 'testuser', got '%s'", profile.Username)
		}
	})

	t.Run("Text response", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/testuser", nil)
		req.Header.Set("User-Agent", "curl/7.68.0")
		w := httptest.NewRecorder()

		handler.GetProfile(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		contentType := w.Header().Get("Content-Type")
		if contentType != "text/plain" {
			t.Errorf("Expected content-type 'text/plain', got '%s'", contentType)
		}

		body := w.Body.String()
		if !contains(body, "Test User (@testuser)") {
			t.Errorf("Expected body to contain user info, got: %s", body)
		}
	})

	t.Run("User not found", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/nonexistent", nil)
		w := httptest.NewRecorder()

		handler.GetProfile(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("Expected status 404, got %d", w.Code)
		}
	})

	t.Run("Empty username", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		w := httptest.NewRecorder()

		handler.GetProfile(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status 400, got %d", w.Code)
		}
	})
}

func TestCreateProfile(t *testing.T) {
	handler := setupTestHandler(t)

	t.Run("Valid profile", func(t *testing.T) {
		createReq := &models.CreateUserRequest{
			SSHPublicKey: "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQ test",
			FullName:     "Test User",
			Username:     "testuser",
			About:        "Test about",
			Links: []models.LinkInput{
				{Name: "Website", URL: "https://example.com"},
			},
		}

		body, _ := json.Marshal(createReq)
		req := httptest.NewRequest("POST", "/api/profiles", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.CreateProfile(w, req)

		if w.Code != http.StatusCreated {
			t.Errorf("Expected status 201, got %d. Body: %s", w.Code, w.Body.String())
		}

		var user models.User
		if err := json.Unmarshal(w.Body.Bytes(), &user); err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		if user.Username != createReq.Username {
			t.Errorf("Expected username '%s', got '%s'", createReq.Username, user.Username)
		}
	})

	t.Run("Duplicate username", func(t *testing.T) {
		createReq := &models.CreateUserRequest{
			SSHPublicKey: "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQ test2",
			FullName:     "Test User 2",
			Username:     "testuser", // Same username as above
			About:        "Test about",
			Links:        []models.LinkInput{},
		}

		body, _ := json.Marshal(createReq)
		req := httptest.NewRequest("POST", "/api/profiles", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.CreateProfile(w, req)

		if w.Code != http.StatusConflict {
			t.Errorf("Expected status 409, got %d", w.Code)
		}
	})

	t.Run("Invalid JSON", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/api/profiles", bytes.NewReader([]byte("invalid json")))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.CreateProfile(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status 400, got %d", w.Code)
		}
	})

	t.Run("Missing required fields", func(t *testing.T) {
		createReq := &models.CreateUserRequest{
			SSHPublicKey: "",
			FullName:     "",
			Username:     "",
			About:        "",
			Links:        []models.LinkInput{},
		}

		body, _ := json.Marshal(createReq)
		req := httptest.NewRequest("POST", "/api/profiles", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.CreateProfile(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status 400, got %d", w.Code)
		}
	})
}

func contains(s, substr string) bool {
	return bytes.Contains([]byte(s), []byte(substr))
}