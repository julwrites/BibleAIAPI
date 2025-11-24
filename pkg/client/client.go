package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Client is the client for the Bible API Service.
type Client struct {
	BaseURL    string
	APIKey     string
	HTTPClient *http.Client
}

// NewClient creates a new Bible API Service client.
func NewClient(baseURL, apiKey string) *Client {
	return &Client{
		BaseURL: baseURL,
		APIKey:  apiKey,
		HTTPClient: &http.Client{
			Timeout: 60 * time.Second, // LLM queries can be slow
		},
	}
}

// Query sends a raw QueryRequest and unmarshals the response into result.
func (c *Client) Query(ctx context.Context, req QueryRequest, result interface{}) error {
	url := fmt.Sprintf("%s/query", c.BaseURL)

	body, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("X-API-KEY", c.APIKey)

	resp, err := c.HTTPClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		var errResp ErrorResponse
		if err := json.Unmarshal(respBody, &errResp); err == nil && errResp.Error.Message != "" {
			return fmt.Errorf("api error: %s (code: %d)", errResp.Error.Message, errResp.Error.Code)
		}
		return fmt.Errorf("api returned status %d: %s", resp.StatusCode, string(respBody))
	}

	if result != nil {
		if err := json.Unmarshal(respBody, result); err != nil {
			return fmt.Errorf("failed to unmarshal response: %w", err)
		}
	}

	return nil
}

// GetVerses retrieves verses.
func (c *Client) GetVerses(ctx context.Context, verses []string, version string) (*VerseResponse, error) {
	req := QueryRequest{
		Query:   Query{Verses: verses},
		Context: Context{User: User{Version: version}},
	}
	var resp VerseResponse
	err := c.Query(ctx, req, &resp)
	return &resp, err
}

// SearchWords searches for words.
func (c *Client) SearchWords(ctx context.Context, words []string, version string) (WordSearchResponse, error) {
	req := QueryRequest{
		Query:   Query{Words: words},
		Context: Context{User: User{Version: version}},
	}
	var resp WordSearchResponse
	err := c.Query(ctx, req, &resp)
	return resp, err
}

// OpenQuery performs an open-ended query.
func (c *Client) OpenQuery(ctx context.Context, question string, version string) (*OQueryResponse, error) {
	req := QueryRequest{
		Query:   Query{Prompt: question},
		Context: Context{User: User{Version: version}},
	}
	var resp OQueryResponse
	err := c.Query(ctx, req, &resp)
	return &resp, err
}

// Chat performs a chat query with dynamic response.
func (c *Client) Chat(ctx context.Context, prompt, schema string, verses []string, version string) (map[string]interface{}, error) {
	req := QueryRequest{
		Query: Query{
			Prompt: prompt,
		},
		Context: Context{
			Schema: schema,
			Verses: verses,
			User:   User{Version: version},
		},
	}
	var resp map[string]interface{}
	err := c.Query(ctx, req, &resp)
	return resp, err
}
