package agent

import (
	"fmt"

	"github.com/8BitTacoSupreme/floxybot/internal/mcp"
)

// Executor dispatches tool calls to the MCP client.
type Executor struct {
	mcpClient *mcp.Client
}

func NewExecutor(mc *mcp.Client) *Executor {
	return &Executor{mcpClient: mc}
}

// Execute runs a tool call and returns the text result.
func (e *Executor) Execute(toolName string, args map[string]any) (string, error) {
	result, err := e.mcpClient.CallTool(toolName, args)
	if err != nil {
		return "", fmt.Errorf("calling tool %s: %w", toolName, err)
	}

	if result.IsError {
		var errText string
		for _, c := range result.Content {
			if c.Type == "text" {
				errText += c.Text
			}
		}
		return "", fmt.Errorf("tool %s error: %s", toolName, errText)
	}

	var text string
	for _, c := range result.Content {
		if c.Type == "text" {
			text += c.Text
		}
	}
	return text, nil
}
