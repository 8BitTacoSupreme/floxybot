package mcp

import "encoding/json"

// ToolsAsClaudeSchema converts MCP tool definitions to the format Claude expects.
func ToolsAsClaudeSchema(tools []ToolInfo) []map[string]any {
	var schemas []map[string]any
	for _, t := range tools {
		var inputSchema map[string]any
		json.Unmarshal(t.InputSchema, &inputSchema)

		schemas = append(schemas, map[string]any{
			"name":         t.Name,
			"description":  t.Description,
			"input_schema": inputSchema,
		})
	}
	return schemas
}
