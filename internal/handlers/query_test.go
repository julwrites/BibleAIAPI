package handlers

import (
	"bible-api-service/internal/biblegateway"
	"bible-api-service/internal/llm/provider"
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

type mockLLMClient struct {
	queryFunc func(ctx context.Context, query, schema string) (string, error)
}

func (m *mockLLMClient) Query(ctx context.Context, query, schema string) (string, error) {
	return m.queryFunc(ctx, query, schema)
}

func TestHandleOpenQuery(t *testing.T) {
	handler := &QueryHandler{
		GetLLMClient: func() (provider.LLMClient, error) {
			return &mockLLMClient{
				queryFunc: func(ctx context.Context, query, schema string) (string, error) {
					return `{"text": "Jesus fed 5,000 men."}`, nil
				},
			}, nil
		},
	}

	reqBody := `{
		"query": {
			"oquery": "How many people did Jesus feed?"
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

	expected := "Jesus fed 5,000 men."
	if response["text"] != expected {
		t.Errorf("handler returned unexpected body: got %v want %v",
			response["text"], expected)
	}
}
