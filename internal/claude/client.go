package claude

import (
	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
)

// Client wraps the Anthropic SDK client.
type Client struct {
	inner anthropic.Client
	model string
}

// NewClient creates a Claude API client.
func NewClient(apiKey, model string) *Client {
	inner := anthropic.NewClient(option.WithAPIKey(apiKey))
	return &Client{inner: inner, model: model}
}

// Model returns the configured model name.
func (c *Client) Model() string {
	return c.model
}

// Inner returns the underlying SDK client for direct use (e.g., tool-use calls).
func (c *Client) Inner() *anthropic.Client {
	return &c.inner
}
