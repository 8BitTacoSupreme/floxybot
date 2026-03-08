package feedback

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// Vote represents a user vote on a response.
type Vote struct {
	Query        string   `json:"query"`
	ResponseHash string   `json:"response_hash"`
	Vote         string   `json:"vote"` // "up" or "down"
	Comment      string   `json:"comment,omitempty"`
	Sources      []string `json:"sources,omitempty"`
}

// Client sends anonymous feedback to the Linode backend.
type Client struct {
	backendURL string
	httpClient *http.Client
}

func NewClient(backendURL string) *Client {
	return &Client{
		backendURL: backendURL,
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}
}

// SendVote submits a vote to the feedback API. Non-blocking best-effort.
func (c *Client) SendVote(ctx context.Context, v Vote) error {
	body, err := json.Marshal(v)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.backendURL+"/feedback", bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("sending feedback: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("feedback API returned %d", resp.StatusCode)
	}
	return nil
}
