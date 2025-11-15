package handlers

import (
	"bible-api-service/internal/biblegateway"
	"bible-api-service/internal/llm"
	"bytes"
	"context"
	"encoding/json"
	"errors"
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

type mockLLMClient struct {
	queryFunc func(ctx context.Context, query, schema string) (string, error)
}

func (m *mockLLMClient) Query(ctx context.Context, query, schema string) (string, error) {
	return m.queryFunc(ctx, query, schema)
}

type mockFFClient struct {
	jsonVariationFunc func(flagKey string, context ffcontext.EvaluationContext, defaultValue map[string]interface{}) (map[string]interface{}, error)
}

func (m *mockFFClient) JSONVariation(flagKey string, context ffcontext.EvaluationContext, defaultValue map[string]interface{}) (map[string]interface{}, error) {
	return m.jsonVariationFunc(flagKey, context, defaultValue)
}

func TestServeHTTP(t *testing.T) {
	t.Run("Invalid request payload", func(t *testing.T) {
		handler := NewQueryHandler()
		req := httptest.NewRequest("POST", "/query", bytes.NewBufferString(`{"invalid`))
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)
		if status := rr.Code; status != http.StatusBadRequest {
			t.Errorf("handler returned wrong status code: got %v want %v",
				status, http.StatusBadRequest)
		}
	})

	t.Run("No query provided", func(t *testing.T) {
		handler := NewQueryHandler()
		req := httptest.NewRequest("POST", "/query", bytes.NewBufferString(`{}`))
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)
		if status := rr.Code; status != http.StatusBadRequest {
			t.Errorf("handler returned wrong status code: got %v want %v",
				status, http.StatusBadRequest)
		}
	})
}

func TestHandleVerseQuery(t *testing.T) {
	t.Run("Successful query", func(t *testing.T) {
		handler := &QueryHandler{
			BibleGatewayClient: &mockBibleGatewayClient{
				getVerseFunc: func(book, chapter, verse, version string) (string, error) {
					return "For God so loved the world...", nil
				},
			},
		}

		reqBody := `{
			"context": { "user": { "version": "ESV" }},
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
	})

	t.Run("Bible Gateway error", func(t *testing.T) {
		handler := &QueryHandler{
			BibleGatewayClient: &mockBibleGatewayClient{
				getVerseFunc: func(book, chapter, verse, version string) (string, error) {
					return "", errors.New("bible gateway error")
				},
			},
		}

		reqBody := `{
			"context": { "user": { "version": "ESV" }},
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
	})
}

func TestHandleWordSearchQuery(t *testing.T) {
	t.Run("Successful query", func(t *testing.T) {
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
			"context": { "user": { "version": "ESV" }},
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
	})

	t.Run("Bible Gateway error", func(t *testing.T) {
		handler := &QueryHandler{
			BibleGatewayClient: &mockBibleGatewayClient{
				searchWordsFunc: func(query, version string) ([]biblegateway.SearchResult, error) {
					return nil, errors.New("bible gateway error")
				},
			},
		}

		reqBody := `{
			"context": { "user": { "version": "ESV" }},
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
	})
}

func TestHandleOpenQuery(t *testing.T) {
	t.Run("Successful query", func(t *testing.T) {
		handler := &QueryHandler{
			GetLLMClient: func() (llm.LLMClient, error) {
				return &mockLLMClient{
					queryFunc: func(ctx context.Context, query, schema string) (string, error) {
						return `{"text": "Jesus fed 5,000 men."}`, nil
					},
				}, nil
			},
		}

		reqBody := `{
			"context": { "user": { "version": "ESV" }},
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
	})

	t.Run("LLM error", func(t *testing.T) {
		handler := &QueryHandler{
			GetLLMClient: func() (llm.LLMClient, error) {
				return &mockLLMClient{
					queryFunc: func(ctx context.Context, query, schema string) (string, error) {
						return "", errors.New("llm error")
					},
				}, nil
			},
		}

		reqBody := `{
			"context": { "user": { "version": "ESV" }},
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
	})

	t.Run("Invalid LLM response", func(t *testing.T) {
		handler := &QueryHandler{
			GetLLMClient: func() (llm.LLMClient, error) {
				return &mockLLMClient{
					queryFunc: func(ctx context.Context, query, schema string) (string, error) {
						return `{"invalid`, nil
					},
				}, nil
			},
		}

		reqBody := `{
			"context": { "user": { "version": "ESV" }},
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
	})
}

func TestHandleInstruction(t *testing.T) {
	t.Run("Successful instruction", func(t *testing.T) {
		handler := &QueryHandler{
			GetLLMClient: func() (llm.LLMClient, error) {
				return &mockLLMClient{
					queryFunc: func(ctx context.Context, query, schema string) (string, error) {
						return `{"summary": "This is a summary."}`, nil
					},
				}, nil
			},
			FFClient: &mockFFClient{
				jsonVariationFunc: func(flagKey string, context ffcontext.EvaluationContext, defaultValue map[string]interface{}) (map[string]interface{}, error) {
					return map[string]interface{}{
						"prompt": "Summarize the following verses: {{.Verses}}",
						"schema": "{}",
					}, nil
				},
			},
		}

		reqBody := `{
			"context": {
				"instruction": "summarize",
				"verses": ["John 1:1", "Genesis 1:1"]
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

		expected := "This is a summary."
		if response["summary"] != expected {
			t.Errorf("handler returned unexpected body: got %v want %v",
				response["summary"], expected)
		}
	})

	t.Run("Feature flag error", func(t *testing.T) {
		handler := &QueryHandler{
			FFClient: &mockFFClient{
				jsonVariationFunc: func(flagKey string, context ffcontext.EvaluationContext, defaultValue map[string]interface{}) (map[string]interface{}, error) {
					return nil, errors.New("feature flag error")
				},
			},
		}

		reqBody := `{
			"context": {
				"instruction": "summarize"
			}
		}`
		req := httptest.NewRequest("POST", "/query", bytes.NewBufferString(reqBody))
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		if status := rr.Code; status != http.StatusInternalServerError {
			t.Errorf("handler returned wrong status code: got %v want %v",
				status, http.StatusInternalServerError)
		}
	})

	t.Run("Invalid prompt in feature flag", func(t *testing.T) {
		handler := &QueryHandler{
			FFClient: &mockFFClient{
				jsonVariationFunc: func(flagKey string, context ffcontext.EvaluationContext, defaultValue map[string]interface{}) (map[string]interface{}, error) {
					return map[string]interface{}{
						"prompt": 123,
						"schema": "{}",
					}, nil
				},
			},
		}

		reqBody := `{
			"context": {
				"instruction": "summarize"
			}
		}`
		req := httptest.NewRequest("POST", "/query", bytes.NewBufferString(reqBody))
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		if status := rr.Code; status != http.StatusInternalServerError {
			t.Errorf("handler returned wrong status code: got %v want %v",
				status, http.StatusInternalServerError)
		}
	})

	t.Run("Invalid schema in feature flag", func(t *testing.T) {
		handler := &QueryHandler{
			FFClient: &mockFFClient{
				jsonVariationFunc: func(flagKey string, context ffcontext.EvaluationContext, defaultValue map[string]interface{}) (map[string]interface{}, error) {
					return map[string]interface{}{
						"prompt": "test",
						"schema": 123,
					}, nil
				},
			},
		}

		reqBody := `{
			"context": {
				"instruction": "summarize"
			}
		}`
		req := httptest.NewRequest("POST", "/query", bytes.NewBufferString(reqBody))
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		if status := rr.Code; status != http.StatusInternalServerError {
			t.Errorf("handler returned wrong status code: got %v want %v",
				status, http.StatusInternalServerError)
		}
	})

	t.Run("Invalid template", func(t *testing.T) {
		handler := &QueryHandler{
			FFClient: &mockFFClient{
				jsonVariationFunc: func(flagKey string, context ffcontext.EvaluationContext, defaultValue map[string]interface{}) (map[string]interface{}, error) {
					return map[string]interface{}{
						"prompt": "{{.Invalid}}",
						"schema": "{}",
					}, nil
				},
			},
		}

		reqBody := `{
			"context": {
				"instruction": "summarize"
			}
		}`
		req := httptest.NewRequest("POST", "/query", bytes.NewBufferString(reqBody))
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		if status := rr.Code; status != http.StatusInternalServerError {
			t.Errorf("handler returned wrong status code: got %v want %v",
				status, http.StatusInternalServerError)
		}
	})
}
