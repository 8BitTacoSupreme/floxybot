package claude

import (
	"context"

	"github.com/anthropics/anthropic-sdk-go"
)

// TokenCallback is called with each text delta during streaming.
type TokenCallback func(text string)

// StreamChat sends a message with streaming and calls the callback for each token.
func (c *Client) StreamChat(ctx context.Context, systemPrompt string, messages []anthropic.MessageParam, onToken TokenCallback) (string, error) {
	stream := c.inner.Messages.NewStreaming(ctx, anthropic.MessageNewParams{
		Model:     anthropic.Model(c.model),
		MaxTokens: 4096,
		System: []anthropic.TextBlockParam{
			{Text: systemPrompt},
		},
		Messages: messages,
	})
	defer stream.Close()

	var fullText string
	for stream.Next() {
		evt := stream.Current()
		if evt.Type == "content_block_delta" && evt.Delta.Type == "text_delta" {
			onToken(evt.Delta.Text)
			fullText += evt.Delta.Text
		}
	}
	if err := stream.Err(); err != nil {
		return fullText, err
	}
	return fullText, nil
}

// Chat sends a non-streaming message and returns the full response.
func (c *Client) Chat(ctx context.Context, systemPrompt string, messages []anthropic.MessageParam) (string, error) {
	resp, err := c.inner.Messages.New(ctx, anthropic.MessageNewParams{
		Model:     anthropic.Model(c.model),
		MaxTokens: 4096,
		System: []anthropic.TextBlockParam{
			{Text: systemPrompt},
		},
		Messages: messages,
	})
	if err != nil {
		return "", err
	}

	var text string
	for _, block := range resp.Content {
		if block.Type == "text" {
			text += block.Text
		}
	}
	return text, nil
}
