// Package notion implements a client for the Notion API.
package notion

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

const (
	baseURL       = "https://api.notion.com/v1"
	notionVersion = "2022-06-28"
)

// Client is a Notion API client.
type Client struct {
	token      string
	httpClient *http.Client
}

// NewClient creates a new Notion API client with the given integration token.
func NewClient(token string) *Client {
	return &Client{
		token:      token,
		httpClient: &http.Client{},
	}
}

// NewClientWithHTTP creates a client with a custom HTTP client (for testing).
func NewClientWithHTTP(token string, httpClient *http.Client) *Client {
	return &Client{
		token:      token,
		httpClient: httpClient,
	}
}

// QueryDatabase queries all pages in a Notion database.
func (c *Client) QueryDatabase(databaseID string) (*DatabaseQueryResponse, error) {
	url := fmt.Sprintf("%s/databases/%s/query", baseURL, databaseID)

	req, err := http.NewRequest("POST", url, bytes.NewReader([]byte("{}")))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	c.setHeaders(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("query database: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("notion API error (status %d): %s", resp.StatusCode, string(body))
	}

	var result DatabaseQueryResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return &result, nil
}

func (c *Client) setHeaders(req *http.Request) {
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Notion-Version", notionVersion)
	req.Header.Set("Content-Type", "application/json")
}
