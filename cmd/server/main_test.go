package main

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"
)

func TestMain(t *testing.T) {
	os.Setenv("PORT", "8081")
	defer os.Unsetenv("PORT")
	os.Setenv("API_KEY", "test-api-key")
	defer os.Unsetenv("API_KEY")

	go main()

	// Give the server a moment to start
	time.Sleep(500 * time.Millisecond)

	reqBody := `{
		"query": {
			"verses": ["John 3:16"]
		}
	}`

	res, err := http.Post("http://localhost:8081/query", "application/json", bytes.NewBufferString(reqBody))
	if err != nil {
		t.Fatalf("could not send POST request: %v", err)
	}
	defer res.Body.Close()

	// The default client doesn't have an API key, so we expect a 401 Unauthorized
	if res.StatusCode != http.StatusUnauthorized {
		t.Errorf("expected status code %d, got %d", http.StatusUnauthorized, res.StatusCode)
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		t.Fatalf("could not read response body: %v", err)
	}
	if !strings.Contains(string(body), "Missing API Key") {
		t.Errorf("expected response body to contain 'Missing API Key', but it didn't")
	}
}
