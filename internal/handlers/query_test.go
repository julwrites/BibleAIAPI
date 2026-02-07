package handlers

import (
	"bible-api-service/internal/bible"
	"bible-api-service/internal/chat"
	"bible-api-service/internal/secrets"
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// MockProvider is a simple mock implementation of bible.Provider
type MockProvider struct {
	getVerseFunc    func(book, chapter, verse, version string) (string, error)
	searchWordsFunc func(query, version string) ([]bible.SearchResult, error)
	getVersionsFunc func() ([]bible.ProviderVersion, error)
}

func (m *MockProvider) GetVerse(book, chapter, verse, version string) (string, error) {
	if m.getVerseFunc != nil {
		return m.getVerseFunc(book, chapter, verse, version)
	}
	return "", nil
}

func (m *MockProvider) SearchWords(query, version string) ([]bible.SearchResult, error) {
	if m.searchWordsFunc != nil {
		return m.searchWordsFunc(query, version)
	}
	return nil, nil
}

func (m *MockProvider) GetVersions() ([]bible.ProviderVersion, error) {
	if m.getVersionsFunc != nil {
		return m.getVersionsFunc()
	}
	return nil, nil
}

func createTestVersionManager(t *testing.T) *bible.VersionManager {
	if err := os.MkdirAll("tmp", 0755); err != nil {
		t.Fatal(err)
	}
	tmpDir, err := os.MkdirTemp("tmp", "createTestVersionManager*")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		os.RemoveAll(tmpDir)
	})

	configPath := filepath.Join(tmpDir, "versions.yaml")
	content := `
- code: ESV
  name: English Standard Version
  language: English
  providers:
    biblegateway: ESV
`
	err = os.WriteFile(configPath, []byte(content), 0644)
	require.NoError(t, err)

	vm, err := bible.NewVersionManager(configPath)
	require.NoError(t, err)
	return vm
}

func TestHandleVerseQuery(t *testing.T) {
	vm := createTestVersionManager(t)

	mockP := &MockProvider{
		getVerseFunc: func(book, chapter, verse, version string) (string, error) {
			return "For God so loved the world...", nil
		},
	}

	pm := bible.NewProviderManager(mockP)
	pm.RegisterProvider(bible.DefaultProviderName, mockP)
	pm.RegisterProvider(bible.DefaultProviderName, mockP)

	handler := &QueryHandler{
		ProviderManager: pm,
		VersionManager:  vm,
	}

	reqBody := `{
		"query": {
			"verses": ["John 3:16"]
		}
	}`
	req := httptest.NewRequest("POST", "/query", bytes.NewBufferString(reqBody))
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	var response map[string]string
	if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	expected := "For God so loved the world..."
	if response["verse"] != expected {
		t.Errorf("handler returned unexpected body: got %v want %v",
			response["verse"], expected)
	}
}

func TestHandleWordSearchQuery(t *testing.T) {
	vm := createTestVersionManager(t)

	mockP := &MockProvider{
		searchWordsFunc: func(query, version string) ([]bible.SearchResult, error) {
			return []bible.SearchResult{
				{Verse: "Romans 3:24", URL: "http://example.com/romans3:24"},
			}, nil
		},
	}

	pm := bible.NewProviderManager(mockP)
	pm.RegisterProvider("biblegateway", mockP)

	handler := &QueryHandler{
		ProviderManager: pm,
		VersionManager:  vm,
	}

	reqBody := `{
		"query": {
			"words": ["grace"]
		}
	}`
	req := httptest.NewRequest("POST", "/query", bytes.NewBufferString(reqBody))
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	var response []bible.SearchResult
	if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if len(response) != 1 {
		t.Fatalf("expected 1 result, got %d", len(response))
	}

	expected := "Romans 3:24"
	if response[0].Verse != expected {
		t.Errorf("handler returned unexpected body: got %v want %v",
			response[0].Verse, expected)
	}
}

type mockChatService struct {
	processFunc func(ctx context.Context, req chat.Request) (*chat.Result, error)
}

func (m *mockChatService) Process(ctx context.Context, req chat.Request) (*chat.Result, error) {
	return m.processFunc(ctx, req)
}

func TestHandlePromptQuery(t *testing.T) {
	vm := createTestVersionManager(t)

	// We don't need real provider here as ChatService is mocked
	pm := bible.NewProviderManager(nil)

	handler := &QueryHandler{
		ChatService: &mockChatService{
			processFunc: func(ctx context.Context, req chat.Request) (*chat.Result, error) {
				if len(req.PrioritizedProviders) == 0 || req.PrioritizedProviders[0].Name != bible.DefaultProviderName {
					return nil, &http.MaxBytesError{} // Indicate failure
				}
				return &chat.Result{
					Data:     chat.Response{"text": "Jesus fed 5,000 men."},
					IsStream: false,
					Meta:     map[string]interface{}{"ai_provider": "openai"},
				}, nil
			},
		},
		VersionManager:  vm,
		ProviderManager: pm,
	}

	reqBody := `{
		"query": {
			"prompt": "How many people did Jesus feed?"
		},
		"context": {
			"user": {
				"version": "ESV"
			}
		}
	}`
	req := httptest.NewRequest("POST", "/query", bytes.NewBufferString(reqBody))
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	var response map[string]interface{}
	if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	data, ok := response["data"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected data field in response")
	}

	expected := "Jesus fed 5,000 men."
	if data["text"] != expected {
		t.Errorf("handler returned unexpected body: got %v want %v",
			data["text"], expected)
	}
}

func TestHandlePromptQuery_Stream(t *testing.T) {
	vm := createTestVersionManager(t)
	pm := bible.NewProviderManager(nil)

	handler := &QueryHandler{
		ChatService: &mockChatService{
			processFunc: func(ctx context.Context, req chat.Request) (*chat.Result, error) {
				ch := make(chan string, 2)
				ch <- "Jesus "
				ch <- "wept."
				close(ch)
				return &chat.Result{
					Stream:   ch,
					IsStream: true,
					Meta:     map[string]interface{}{"ai_provider": "openai"},
				}, nil
			},
		},
		VersionManager:  vm,
		ProviderManager: pm,
	}

	reqBody := `{
		"query": {
			"prompt": "Shortest verse?"
		},
		"options": {
			"stream": true
		}
	}`
	req := httptest.NewRequest("POST", "/query", bytes.NewBufferString(reqBody))
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	if contentType := rr.Header().Get("Content-Type"); contentType != "text/event-stream" {
		t.Errorf("handler returned wrong content type: got %v want %v",
			contentType, "text/event-stream")
	}

	body := rr.Body.String()
	if !strings.Contains(body, `event: meta`) {
		t.Error("body does not contain meta event")
	}
	if !strings.Contains(body, `event: chunk`) {
		t.Error("body does not contain chunk event")
	}
	if !strings.Contains(body, `event: done`) {
		t.Error("body does not contain done event")
	}
}

func TestHandlePromptQuery_WithContext(t *testing.T) {
	vm := createTestVersionManager(t)
	pm := bible.NewProviderManager(nil)

	handler := &QueryHandler{
		ChatService: &mockChatService{
			processFunc: func(ctx context.Context, req chat.Request) (*chat.Result, error) {
				if len(req.VerseRefs) != 1 || req.VerseRefs[0] != "John 3:16" {
					return nil, &http.MaxBytesError{} // Use an error to signal mismatch
				}
				return &chat.Result{
					Data: chat.Response{"response": "ok"},
				}, nil
			},
		},
		VersionManager:  vm,
		ProviderManager: pm,
	}

	reqBody := `{
		"query": {
			"prompt": "Explain this verse"
		},
		"context": {
			"verses": ["John 3:16"]
		}
	}`
	req := httptest.NewRequest("POST", "/query", bytes.NewBufferString(reqBody))
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}
}

func TestInvalidRequest_MultipleQueries(t *testing.T) {
	handler := &QueryHandler{}

	reqBody := `{
		"query": {
			"verses": ["John 3:16"],
			"prompt": "Explain this"
		}
	}`
	req := httptest.NewRequest("POST", "/query", bytes.NewBufferString(reqBody))
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusBadRequest)
	}
}

func TestInvalidRequest_ContextWithoutPrompt(t *testing.T) {
	handler := &QueryHandler{}

	reqBody := `{
		"query": {
			"verses": ["John 3:16"]
		},
		"context": {
			"verses": ["John 3:16"]
		}
	}`
	req := httptest.NewRequest("POST", "/query", bytes.NewBufferString(reqBody))
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusBadRequest)
	}
}

func TestInvalidRequest_NoQuery(t *testing.T) {
	handler := &QueryHandler{}

	reqBody := `{
		"query": {}
	}`
	req := httptest.NewRequest("POST", "/query", bytes.NewBufferString(reqBody))
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusBadRequest)
	}
}

func TestHandleVerseQuery_BookWithSpaces(t *testing.T) {
	vm := createTestVersionManager(t)
	mockP := &MockProvider{
		getVerseFunc: func(book, chapter, verse, version string) (string, error) {
			// Check if the book name with spaces was parsed correctly
			if book != "1 John" || chapter != "1" || verse != "9" {
				return "", nil // Or error
			}
			return "If we confess our sins...", nil
		},
	}
	pm := bible.NewProviderManager(mockP)
	pm.RegisterProvider(bible.DefaultProviderName, mockP)

	handler := &QueryHandler{
		ProviderManager: pm,
		VersionManager:  vm,
	}

	reqBody := `{
		"query": {
			"verses": ["1 John 1:9"]
		}
	}`
	req := httptest.NewRequest("POST", "/query", bytes.NewBufferString(reqBody))
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	var response map[string]string
	if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	expected := "If we confess our sins..."
	if response["verse"] != expected {
		t.Errorf("handler returned unexpected body: got %v want %v",
			response["verse"], expected)
	}
}

func TestNewQueryHandler(t *testing.T) {
	vm := createTestVersionManager(t)
	handler := NewQueryHandler(&secrets.EnvClient{}, vm)
	if handler.ProviderManager == nil {
		t.Error("expected ProviderManager to be initialized")
	}
	if handler.GetLLMClient == nil {
		t.Error("expected GetLLMClient to be initialized")
	}
	if handler.FFClient == nil {
		t.Error("expected FFClient to be initialized")
	}
	if handler.VersionManager == nil {
		t.Error("expected VersionManager to be initialized")
	}
}

func TestHandleVerseQuery_Error(t *testing.T) {
	vm := createTestVersionManager(t)
	mockP := &MockProvider{
		getVerseFunc: func(book, chapter, verse, version string) (string, error) {
			return "", &http.MaxBytesError{}
		},
	}
	pm := bible.NewProviderManager(mockP)
	pm.RegisterProvider(bible.DefaultProviderName, mockP)

	handler := &QueryHandler{
		ProviderManager: pm,
		VersionManager:  vm,
	}

	reqBody := `{
		"query": {
			"verses": ["John 3:16"]
		}
	}`
	req := httptest.NewRequest("POST", "/query", bytes.NewBufferString(reqBody))
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusInternalServerError {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusInternalServerError)
	}
}

func TestHandleWordSearchQuery_Error(t *testing.T) {
	vm := createTestVersionManager(t)
	mockP := &MockProvider{
		searchWordsFunc: func(query, version string) ([]bible.SearchResult, error) {
			return nil, &http.MaxBytesError{}
		},
	}
	pm := bible.NewProviderManager(mockP)
	pm.RegisterProvider(bible.DefaultProviderName, mockP)

	handler := &QueryHandler{
		ProviderManager: pm,
		VersionManager:  vm,
	}

	reqBody := `{
		"query": {
			"words": ["grace"]
		}
	}`
	req := httptest.NewRequest("POST", "/query", bytes.NewBufferString(reqBody))
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusInternalServerError {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusInternalServerError)
	}
}

func TestHandlePromptQuery_Error(t *testing.T) {
	vm := createTestVersionManager(t)
	pm := bible.NewProviderManager(nil)

	handler := &QueryHandler{
		ChatService: &mockChatService{
			processFunc: func(ctx context.Context, req chat.Request) (*chat.Result, error) {
				return nil, &http.MaxBytesError{}
			},
		},
		VersionManager:  vm,
		ProviderManager: pm,
	}

	reqBody := `{
		"query": {
			"prompt": "Explain this"
		},
		"context": {
			"user": {
				"version": "ESV"
			}
		}
	}`
	req := httptest.NewRequest("POST", "/query", bytes.NewBufferString(reqBody))
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusInternalServerError {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusInternalServerError)
	}
}

func TestInvalidRequestPayload(t *testing.T) {
	handler := &QueryHandler{}

	reqBody := `{invalid json}`
	req := httptest.NewRequest("POST", "/query", bytes.NewBufferString(reqBody))
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusBadRequest)
	}
}
