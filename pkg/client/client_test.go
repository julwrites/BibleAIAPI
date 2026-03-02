package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestGetVerses(t *testing.T) {
	mockResponse := VerseResponse{
		Verse: "John 3:16 (ESV) ...",
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/query" {
			t.Errorf("Expected path /query, got %s", r.URL.Path)
		}
		if r.Header.Get("X-API-KEY") != "test-key" {
			t.Errorf("Expected API key test-key, got %s", r.Header.Get("X-API-KEY"))
		}
		json.NewEncoder(w).Encode(mockResponse)
	}))
	defer ts.Close()

	client := NewClient(ts.URL, "test-key")
	resp, err := client.GetVerses(context.Background(), []string{"John 3:16"}, "ESV")
	if err != nil {
		t.Fatalf("GetVerses failed: %v", err)
	}

	if resp.Verse != mockResponse.Verse {
		t.Errorf("Expected verse %s, got %s", mockResponse.Verse, resp.Verse)
	}
}

func TestChat(t *testing.T) {
	mockResponse := map[string]interface{}{
		"summary": "This is a summary",
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(mockResponse)
	}))
	defer ts.Close()

	client := NewClient(ts.URL, "test-key")
	resp, err := client.Chat(context.Background(), "summarize", "{}", nil, "ESV")
	if err != nil {
		t.Fatalf("Chat failed: %v", err)
	}

	if resp["summary"] != "This is a summary" {
		t.Errorf("Expected summary, got %v", resp["summary"])
	}
}

func TestSearchWords(t *testing.T) {
	mockResponse := []interface{}{
		map[string]interface{}{"title": "Verse 1", "text": "Text 1"},
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(mockResponse)
	}))
	defer ts.Close()

	client := NewClient(ts.URL, "test-key")
	resp, err := client.SearchWords(context.Background(), []string{"test"}, "ESV")
	if err != nil {
		t.Fatalf("SearchWords failed: %v", err)
	}

	if len(resp) != 1 {
		t.Errorf("Expected 1 result, got %d", len(resp))
	}
}

func TestOpenQuery(t *testing.T) {
	mockResponse := OQueryResponse{
		Text: "Here is the answer",
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(mockResponse)
	}))
	defer ts.Close()

	client := NewClient(ts.URL, "test-key")
	resp, err := client.OpenQuery(context.Background(), "question", "ESV")
	if err != nil {
		t.Fatalf("OpenQuery failed: %v", err)
	}

	if resp.Text != "Here is the answer" {
		t.Errorf("Expected response, got %v", resp.Text)
	}
}

func TestNewClient(t *testing.T) {
	baseURL := "http://api.example.com"
	apiKey := "test-api-key"
	client := NewClient(baseURL, apiKey)

	if client.BaseURL != baseURL {
		t.Errorf("Expected BaseURL %s, got %s", baseURL, client.BaseURL)
	}

	if client.APIKey != apiKey {
		t.Errorf("Expected APIKey %s, got %s", apiKey, client.APIKey)
	}

	if client.HTTPClient == nil {
		t.Fatal("Expected HTTPClient to be initialized, got nil")
	}

	expectedTimeout := 60 * time.Second
	if client.HTTPClient.Timeout != expectedTimeout {
		t.Errorf("Expected Timeout %v, got %v", expectedTimeout, client.HTTPClient.Timeout)
	}
}
