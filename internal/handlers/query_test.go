package handlers

import (
	"bible-api-service/internal/biblegateway"
	"bible-api-service/internal/chat"
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

type mockBibleGatewayClient struct {
	getVerseFunc    func(book, chapter, verse, version string) (string, error)
	searchWordsFunc func(query, version string) ([]biblegateway.SearchResult, error)
}

func (m *mockBibleGatewayClient) GetVerse(book, chapter, verse, version string) (string, error) {
	return m.getVerseFunc(book, chapter, verse, version)
}

func (m *mockBibleGatewayClient) SearchWords(query, version string) ([]biblegateway.SearchResult, error) {
	return m.searchWordsFunc(query, version)
}

func TestHandleVerseQuery(t *testing.T) {
	handler := &QueryHandler{
		BibleGatewayClient: &mockBibleGatewayClient{
			getVerseFunc: func(book, chapter, verse, version string) (string, error) {
				return "For God so loved the world...", nil
			},
		},
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
	handler := &QueryHandler{
		BibleGatewayClient: &mockBibleGatewayClient{
			searchWordsFunc: func(query, version string) ([]biblegateway.SearchResult, error) {
				return []biblegateway.SearchResult{
					{Verse: "Romans 3:24", URL: "http://example.com/romans3:24"},
				}, nil
			},
		},
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

	var response []biblegateway.SearchResult
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
	processFunc func(ctx context.Context, req chat.Request) (chat.Response, error)
}

func (m *mockChatService) Process(ctx context.Context, req chat.Request) (chat.Response, error) {
	return m.processFunc(ctx, req)
}

func TestHandlePromptQuery(t *testing.T) {
	handler := &QueryHandler{
		ChatService: &mockChatService{
			processFunc: func(ctx context.Context, req chat.Request) (chat.Response, error) {
				return chat.Response{"text": "Jesus fed 5,000 men."}, nil
			},
		},
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

	expected := "Jesus fed 5,000 men."
	if response["text"] != expected {
		t.Errorf("handler returned unexpected body: got %v want %v",
			response["text"], expected)
	}
}

func TestHandlePromptQuery_WithContext(t *testing.T) {
	handler := &QueryHandler{
		ChatService: &mockChatService{
			processFunc: func(ctx context.Context, req chat.Request) (chat.Response, error) {
				if len(req.VerseRefs) != 1 || req.VerseRefs[0] != "John 3:16" {
					return nil, &http.MaxBytesError{} // Use an error to signal mismatch
				}
				return chat.Response{"response": "ok"}, nil
			},
		},
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
	handler := &QueryHandler{
		BibleGatewayClient: &mockBibleGatewayClient{
			getVerseFunc: func(book, chapter, verse, version string) (string, error) {
				// Check if the book name with spaces was parsed correctly
				if book != "1 John" || chapter != "1" || verse != "9" {
					return "", nil // Or error
				}
				return "If we confess our sins...", nil
			},
		},
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
	handler := NewQueryHandler()
	if handler.BibleGatewayClient == nil {
		t.Error("expected BibleGatewayClient to be initialized")
	}
	if handler.GetLLMClient == nil {
		t.Error("expected GetLLMClient to be initialized")
	}
	if handler.FFClient == nil {
		t.Error("expected FFClient to be initialized")
	}
}

func TestHandleVerseQuery_Error(t *testing.T) {
	handler := &QueryHandler{
		BibleGatewayClient: &mockBibleGatewayClient{
			getVerseFunc: func(book, chapter, verse, version string) (string, error) {
				return "", &http.MaxBytesError{}
			},
		},
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
	handler := &QueryHandler{
		BibleGatewayClient: &mockBibleGatewayClient{
			searchWordsFunc: func(query, version string) ([]biblegateway.SearchResult, error) {
				return nil, &http.MaxBytesError{}
			},
		},
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
	handler := &QueryHandler{
		ChatService: &mockChatService{
			processFunc: func(ctx context.Context, req chat.Request) (chat.Response, error) {
				return nil, &http.MaxBytesError{}
			},
		},
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
