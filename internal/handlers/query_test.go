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

	"github.com/thomaspoignant/go-feature-flag/ffcontext"
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

type mockFFClient struct {
	jsonVariationFunc func(flagKey string, context ffcontext.EvaluationContext, defaultValue map[string]interface{}) (map[string]interface{}, error)
}

func (m *mockFFClient) JSONVariation(flagKey string, context ffcontext.EvaluationContext, defaultValue map[string]interface{}) (map[string]interface{}, error) {
	return m.jsonVariationFunc(flagKey, context, defaultValue)
}

func TestHandleInstruction(t *testing.T) {
	handler := &QueryHandler{
		FFClient: &mockFFClient{
			jsonVariationFunc: func(flagKey string, context ffcontext.EvaluationContext, defaultValue map[string]interface{}) (map[string]interface{}, error) {
				return map[string]interface{}{
					"prompt": "test prompt",
					"schema": "test schema",
				}, nil
			},
		},
		GetLLMClient: func() (provider.LLMClient, error) {
			return &mockLLMClient{
				queryFunc: func(ctx context.Context, query, schema string) (string, error) {
					return `{"response": "test response"}`, nil
				},
			}, nil
		},
	}

	reqBody := `{
		"context": {
			"instruction": "test_instruction"
		},
		"query": {
			"oquery": "test query"
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

	expected := "test response"
	if response["response"] != expected {
		t.Errorf("handler returned unexpected body: got %v want %v",
			response["response"], expected)
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

func TestHandleOpenQuery_Error(t *testing.T) {
	handler := &QueryHandler{
		GetLLMClient: func() (provider.LLMClient, error) {
			return nil, &http.MaxBytesError{}
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

	if status := rr.Code; status != http.StatusInternalServerError {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusInternalServerError)
	}
}

func TestHandleInstruction_FFClient_Error(t *testing.T) {
	handler := &QueryHandler{
		FFClient: &mockFFClient{
			jsonVariationFunc: func(flagKey string, context ffcontext.EvaluationContext, defaultValue map[string]interface{}) (map[string]interface{}, error) {
				return nil, &http.MaxBytesError{}
			},
		},
	}

	reqBody := `{
		"context": {
			"instruction": "test_instruction"
		},
		"query": {
			"oquery": "test query"
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

func TestHandleInstruction_LLM_Error(t *testing.T) {
	handler := &QueryHandler{
		FFClient: &mockFFClient{
			jsonVariationFunc: func(flagKey string, context ffcontext.EvaluationContext, defaultValue map[string]interface{}) (map[string]interface{}, error) {
				return map[string]interface{}{
					"prompt": "test prompt",
					"schema": "test schema",
				}, nil
			},
		},
		GetLLMClient: func() (provider.LLMClient, error) {
			return nil, &http.MaxBytesError{}
		},
	}

	reqBody := `{
		"context": {
			"instruction": "test_instruction"
		},
		"query": {
			"oquery": "test query"
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

func TestNoQueryProvided(t *testing.T) {
	handler := &QueryHandler{}

	reqBody := `{}`
	req := httptest.NewRequest("POST", "/query", bytes.NewBufferString(reqBody))
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusBadRequest)
	}
}
