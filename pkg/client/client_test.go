package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
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
