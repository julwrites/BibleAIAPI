package util

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestJSONError(t *testing.T) {
	tests := []struct {
		name           string
		code           int
		message        string
		expectedBody   string
		expectedStatus int
	}{
		{
			name:           "Bad Request",
			code:           http.StatusBadRequest,
			message:        "Invalid request body",
			expectedBody:   `{"error":{"code":400,"message":"Invalid request body"}}`,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Internal Server Error",
			code:           http.StatusInternalServerError,
			message:        "Something went wrong",
			expectedBody:   `{"error":{"code":500,"message":"Something went wrong"}}`,
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:           "Empty Message",
			code:           http.StatusBadRequest,
			message:        "",
			expectedBody:   `{"error":{"code":400,"message":""}}`,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Special Characters",
			code:           http.StatusInternalServerError,
			message:        "Error with \"quotes\" and \n newlines",
			expectedBody:   `{"error":{"code":500,"message":"Error with \"quotes\" and \n newlines"}}`,
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			JSONError(rr, tt.code, tt.message)

			if status := rr.Code; status != tt.expectedStatus {
				t.Errorf("handler returned wrong status code: got %v want %v",
					status, tt.expectedStatus)
			}

			// Check Content-Type header
			if contentType := rr.Header().Get("Content-Type"); contentType != "application/json" {
				t.Errorf("handler returned wrong content type: got %v want %v",
					contentType, "application/json")
			}

			// Trim the newline character from the body
			body := strings.TrimSpace(rr.Body.String())
			if body != tt.expectedBody {
				t.Errorf("handler returned unexpected body: got %v want %v",
					body, tt.expectedBody)
			}
		})
	}
}
