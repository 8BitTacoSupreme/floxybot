package voyage

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
)

const (
	rerankEndpoint = "https://api.voyageai.com/v1/rerank"
	RerankModel    = "rerank-2"
)

// RerankClient calls the Voyage AI rerank API.
type RerankClient struct {
	apiKey     string
	httpClient *http.Client
}

func NewRerankClient(apiKey string) *RerankClient {
	return &RerankClient{apiKey: apiKey, httpClient: &http.Client{}}
}

type rerankRequest struct {
	Query     string   `json:"query"`
	Documents []string `json:"documents"`
	Model     string   `json:"model"`
	TopK      int      `json:"top_k,omitempty"`
}

type rerankResponse struct {
	Data []RerankResult `json:"data"`
}

// RerankResult is a single reranked document with its relevance score.
type RerankResult struct {
	Index          int     `json:"index"`
	RelevanceScore float64 `json:"relevance_score"`
}

// Rerank takes a query and candidate documents, returns the top-k most relevant indices sorted by score descending.
func (c *RerankClient) Rerank(ctx context.Context, query string, documents []string, topK int) ([]RerankResult, error) {
	body, err := json.Marshal(rerankRequest{
		Query:     query,
		Documents: documents,
		Model:     RerankModel,
		TopK:      topK,
	})
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, rerankEndpoint, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("voyage rerank API %d: %s", resp.StatusCode, b)
	}

	var result rerankResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	sort.Slice(result.Data, func(i, j int) bool {
		return result.Data[i].RelevanceScore > result.Data[j].RelevanceScore
	})

	return result.Data, nil
}
