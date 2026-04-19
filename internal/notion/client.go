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

// QueryDatabase queries all pages in a Notion database, handling pagination.
// Notion returns max 100 results per request; this loops until all pages are fetched.
func (c *Client) QueryDatabase(databaseID string) (*DatabaseQueryResponse, error) {
	reqURL := fmt.Sprintf("%s/databases/%s/query", baseURL, databaseID)
	var allResults []Page

	body := map[string]string{}
	for {
		bodyBytes, _ := json.Marshal(body)
		req, err := http.NewRequest("POST", reqURL, bytes.NewReader(bodyBytes))
		if err != nil {
			return nil, fmt.Errorf("create request: %w", err)
		}
		c.setHeaders(req)

		resp, err := c.httpClient.Do(req)
		if err != nil {
			return nil, fmt.Errorf("query database: %w", err)
		}

		if resp.StatusCode != http.StatusOK {
			respBody, _ := io.ReadAll(resp.Body)
			_ = resp.Body.Close()
			return nil, fmt.Errorf("notion API error (status %d): %s", resp.StatusCode, string(respBody))
		}

		var page DatabaseQueryResponse
		err = json.NewDecoder(resp.Body).Decode(&page)
		_ = resp.Body.Close()
		if err != nil {
			return nil, fmt.Errorf("decode response: %w", err)
		}

		allResults = append(allResults, page.Results...)

		if !page.HasMore || page.NextCursor == "" {
			break
		}
		body["start_cursor"] = page.NextCursor
	}

	return &DatabaseQueryResponse{Results: allResults}, nil
}

func (c *Client) setHeaders(req *http.Request) {
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Notion-Version", notionVersion)
	req.Header.Set("Content-Type", "application/json")
}
