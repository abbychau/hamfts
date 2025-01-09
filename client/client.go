package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type Client struct {
	baseURL    string
	httpClient *http.Client
}

type SearchRequest struct {
	Query string `json:"query"`
}

type DocumentRequest struct {
	ID      string                 `json:"id"`
	Content string                 `json:"content"`
	Meta    map[string]interface{} `json:"metadata,omitempty"`
}

func NewClient(baseURL string) *Client {
	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (c *Client) Search(query string) ([]interface{}, error) {
	req := SearchRequest{Query: query}
	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Post(c.baseURL+"/search", "application/json", bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("search failed with status: %d", resp.StatusCode)
	}

	var results []interface{}
	if err := json.NewDecoder(resp.Body).Decode(&results); err != nil {
		return nil, err
	}

	return results, nil
}

func (c *Client) AddDocument(id, content string, metadata map[string]interface{}) error {
	req := DocumentRequest{
		ID:      id,
		Content: content,
		Meta:    metadata,
	}
	body, err := json.Marshal(req)
	if err != nil {
		return err
	}

	resp, err := c.httpClient.Post(c.baseURL+"/documents", "application/json", bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("add document failed with status: %d", resp.StatusCode)
	}

	return nil
}

func (c *Client) ListDocuments() ([]string, error) {
	resp, err := c.httpClient.Get(c.baseURL + "/documents")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("list documents failed with status: %d", resp.StatusCode)
	}

	var docs []string
	if err := json.NewDecoder(resp.Body).Decode(&docs); err != nil {
		return nil, err
	}

	return docs, nil
}

func (c *Client) DeleteDocument(id string) error {
	req, err := http.NewRequest(http.MethodDelete, fmt.Sprintf("%s/documents/%s", c.baseURL, id), nil)
	if err != nil {
		return err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("delete document failed with status: %d", resp.StatusCode)
	}

	return nil
}

func (c *Client) GetStats() (map[string]interface{}, error) {
	resp, err := c.httpClient.Get(c.baseURL + "/stats")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("get stats failed with status: %d", resp.StatusCode)
	}

	var stats map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&stats); err != nil {
		return nil, err
	}

	return stats, nil
}
