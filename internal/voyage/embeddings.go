package voyage

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	embedEndpoint = "https://api.voyageai.com/v1/embeddings"
	EmbedModel    = "voyage-3-lite"
	EmbedDim      = 512
)

// EmbeddingClient calls the Voyage AI embeddings API.
type EmbeddingClient struct {
	apiKey     string
	httpClient *http.Client
}

func NewEmbeddingClient(apiKey string) *EmbeddingClient {
	return &EmbeddingClient{apiKey: apiKey, httpClient: &http.Client{}}
}

type embedRequest struct {
	Input []string `json:"input"`
	Model string   `json:"model"`
}

type embedResponse struct {
	Data []struct {
		Embedding []float32 `json:"embedding"`
	} `json:"data"`
	Usage struct {
		TotalTokens int `json:"total_tokens"`
	} `json:"usage"`
}

// Embed returns embeddings for one or more texts. Retries on 429 with backoff.
func (c *EmbeddingClient) Embed(ctx context.Context, texts []string) ([][]float32, error) {
	body, err := json.Marshal(embedRequest{Input: texts, Model: EmbedModel})
	if err != nil {
		return nil, err
	}

	var lastErr error
	for attempt := 0; attempt < 5; attempt++ {
		if attempt > 0 {
			wait := time.Duration(attempt*20) * time.Second
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(wait):
			}
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, embedEndpoint, bytes.NewReader(body))
		if err != nil {
			return nil, err
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+c.apiKey)

		resp, err := c.httpClient.Do(req)
		if err != nil {
			return nil, err
		}

		if resp.StatusCode == 429 {
			resp.Body.Close()
			lastErr = fmt.Errorf("voyage embed API 429 (rate limited)")
			continue
		}

		if resp.StatusCode != http.StatusOK {
			b, _ := io.ReadAll(resp.Body)
			resp.Body.Close()
			return nil, fmt.Errorf("voyage embed API %d: %s", resp.StatusCode, b)
		}

		var result embedResponse
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			resp.Body.Close()
			return nil, err
		}
		resp.Body.Close()

		embeddings := make([][]float32, len(result.Data))
		for i, d := range result.Data {
			embeddings[i] = d.Embedding
		}
		return embeddings, nil
	}
	return nil, lastErr
}

// EmbedOne is a convenience for embedding a single text.
func (c *EmbeddingClient) EmbedOne(ctx context.Context, text string) ([]float32, error) {
	vecs, err := c.Embed(ctx, []string{text})
	if err != nil {
		return nil, err
	}
	if len(vecs) == 0 {
		return nil, fmt.Errorf("voyage returned empty embeddings")
	}
	return vecs[0], nil
}
