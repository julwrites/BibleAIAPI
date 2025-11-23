package handlers

import (
	"bible-api-service/internal/storage"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestServeAdminUI(t *testing.T) {
	mockStorage := storage.NewMockClient()
	mockSecrets := new(MockSecretsClient) // Defined in admin_test.go
	handler := NewAdminHandler(mockStorage, mockSecrets)

	t.Run("GET returns HTML", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/admin", nil)
		w := httptest.NewRecorder()

		handler.ServeAdminUI(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}

		if contentType := w.Header().Get("Content-Type"); contentType != "text/html" {
			t.Errorf("expected Content-Type text/html, got %s", contentType)
		}

		body := w.Body.String()
		if !strings.Contains(body, "<!DOCTYPE html>") {
			t.Error("body does not contain HTML doctype")
		}
		if !strings.Contains(body, "Bible API Admin") {
			t.Error("body does not contain title")
		}
	})

	t.Run("POST returns 405", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/admin", nil)
		w := httptest.NewRecorder()

		handler.ServeAdminUI(w, req)

		if w.Code != http.StatusMethodNotAllowed {
			t.Errorf("expected status 405, got %d", w.Code)
		}
	})
}
