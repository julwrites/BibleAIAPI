package handlers

import (
	"bible-api-service/internal/storage"
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockSecretsClient struct {
	mock.Mock
}

func (m *MockSecretsClient) GetSecret(ctx context.Context, name string) (string, error) {
	args := m.Called(ctx, name)
	return args.String(0), args.Error(1)
}

func (m *MockSecretsClient) Close() error {
	return nil
}

type ErrorStorageClient struct {
	*storage.MockClient
}

func (m *ErrorStorageClient) CreateAPIKey(ctx context.Context, clientName string, dailyLimit int) (*storage.APIKey, error) {
	return nil, assert.AnError
}

func TestCreateKey(t *testing.T) {
	mockStorage := storage.NewMockClient()
	mockSecrets := new(MockSecretsClient)

	// Setup handler
	handler := NewAdminHandler(mockStorage, mockSecrets)

	// Test Case 1: Success
	t.Run("Success", func(t *testing.T) {
		mockSecrets.On("GetSecret", mock.Anything, "ADMIN_PASSWORD").Return("secret123", nil).Once()

		reqBody := CreateKeyRequest{
			Password:   "secret123",
			ClientName: "TestClient",
			DailyLimit: 500,
		}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest("POST", "/api/admin/keys", bytes.NewBuffer(body))
		w := httptest.NewRecorder()

		handler.CreateKey(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp storage.APIKey
		json.NewDecoder(w.Body).Decode(&resp)
		assert.Equal(t, "TestClient", resp.ClientName)
		assert.Equal(t, 500, resp.DailyLimit)
		assert.NotEmpty(t, resp.Key)
	})

	// Test Case 2: Wrong Password
	t.Run("WrongPassword", func(t *testing.T) {
		mockSecrets.On("GetSecret", mock.Anything, "ADMIN_PASSWORD").Return("secret123", nil).Once()

		reqBody := CreateKeyRequest{
			Password:   "wrong",
			ClientName: "TestClient",
		}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest("POST", "/api/admin/keys", bytes.NewBuffer(body))
		w := httptest.NewRecorder()

		handler.CreateKey(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	// Test Case 3: Missing Client Name
	t.Run("MissingClientName", func(t *testing.T) {
		reqBody := CreateKeyRequest{
			Password: "secret123",
		}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest("POST", "/api/admin/keys", bytes.NewBuffer(body))
		w := httptest.NewRecorder()

		handler.CreateKey(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	// Test Case 4: Secrets Error
	t.Run("SecretsError", func(t *testing.T) {
		mockSecrets.On("GetSecret", mock.Anything, "ADMIN_PASSWORD").Return("", assert.AnError).Once()

		reqBody := CreateKeyRequest{
			Password:   "any",
			ClientName: "TestClient",
		}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest("POST", "/api/admin/keys", bytes.NewBuffer(body))
		w := httptest.NewRecorder()

		handler.CreateKey(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	// Test Case 5: Storage Error
	t.Run("StorageError", func(t *testing.T) {
		mockSecrets.On("GetSecret", mock.Anything, "ADMIN_PASSWORD").Return("secret123", nil).Once()

		errStorage := &ErrorStorageClient{MockClient: storage.NewMockClient()}
		errHandler := NewAdminHandler(errStorage, mockSecrets)

		reqBody := CreateKeyRequest{
			Password:   "secret123",
			ClientName: "TestClient",
		}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest("POST", "/api/admin/keys", bytes.NewBuffer(body))
		w := httptest.NewRecorder()

		errHandler.CreateKey(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}
