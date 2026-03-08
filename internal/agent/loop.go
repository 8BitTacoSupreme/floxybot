package agent

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/8BitTacoSupreme/floxybot/internal/claude"
	"github.com/8BitTacoSupreme/floxybot/internal/floxctx"
	"github.com/8BitTacoSupreme/floxybot/internal/mcp"
	"github.com/anthropics/anthropic-sdk-go"
)

const maxRounds = 10

// StepCallback is called after each agent step with a human-readable status.
type StepCallback func(step string)

// Loop runs the Claude tool-use agent loop.
type Loop struct {
	claudeClient *claude.Client
	executor     *Executor
	mcpTools     []mcp.ToolInfo
	floxCtx      *floxctx.Context
	autoApprove  bool
	onStep       StepCallback
}

func NewLoop(cc *claude.Client, exec *Executor, tools []mcp.ToolInfo, fctx *floxctx.Context, autoApprove bool, onStep StepCallback) *Loop {
	return &Loop{
		claudeClient: cc,
		executor:     exec,
		mcpTools:     tools,
		floxCtx:      fctx,
		autoApprove:  autoApprove,
		onStep:       onStep,
	}
}

// Run executes the agent loop for a given task, returning the final text response.
func (l *Loop) Run(ctx context.Context, task string) (string, error) {
	tools := buildToolParams(l.mcpTools)

	systemPrompt := claude.BuildSystemPrompt(nil, l.floxCtx)
	systemPrompt += "\n\nYou are in agent/co-pilot mode. Use the available tools to complete the user's task. " +
		"Execute one step at a time, observe the result, then decide the next step. " +
		"When the task is complete, provide a summary of what was done."

	messages := []anthropic.MessageParam{
		anthropic.NewUserMessage(anthropic.NewTextBlock(task)),
	}

	for round := 0; round < maxRounds; round++ {
		resp, err := l.claudeClient.Inner().Messages.New(ctx, anthropic.MessageNewParams{
			Model:     anthropic.Model(l.claudeClient.Model()),
			MaxTokens: 4096,
			System: []anthropic.TextBlockParam{
				{Text: systemPrompt},
			},
			Messages: messages,
			Tools:    tools,
		})
		if err != nil {
			return "", fmt.Errorf("round %d: %w", round, err)
		}

		// Collect text and tool_use blocks from ContentBlockUnion.
		var textParts string
		var toolUses []anthropic.ContentBlockUnion

		for _, block := range resp.Content {
			switch block.Type {
			case "text":
				textParts += block.Text
			case "tool_use":
				toolUses = append(toolUses, block)
			}
		}

		if textParts != "" && l.onStep != nil {
			l.onStep(textParts)
		}

		// If no tool use, we're done.
		if len(toolUses) == 0 || resp.StopReason == "end_turn" {
			return textParts, nil
		}

		// Use ToParam() to convert the response into an assistant message param.
		messages = append(messages, resp.ToParam())

		// Execute each tool call.
		var toolResults []anthropic.ContentBlockParamUnion
		for _, tu := range toolUses {
			if l.onStep != nil {
				l.onStep(fmt.Sprintf("Calling tool: %s", tu.Name))
			}

			var args map[string]any
			json.Unmarshal(tu.Input, &args)

			if !l.autoApprove && IsDestructive(tu.Name) {
				if l.onStep != nil {
					l.onStep(fmt.Sprintf("Skipped destructive tool %s (use --yes to auto-approve)", tu.Name))
				}
				toolResults = append(toolResults, anthropic.NewToolResultBlock(
					tu.ID,
					"Tool call skipped: destructive operation requires --yes flag",
					true,
				))
				continue
			}

			result, err := l.executor.Execute(tu.Name, args)
			if err != nil {
				if l.onStep != nil {
					l.onStep(fmt.Sprintf("Tool %s error: %v", tu.Name, err))
				}
				toolResults = append(toolResults, anthropic.NewToolResultBlock(
					tu.ID,
					fmt.Sprintf("Error: %v", err),
					true,
				))
			} else {
				if l.onStep != nil {
					l.onStep(fmt.Sprintf("Tool %s completed", tu.Name))
				}
				toolResults = append(toolResults, anthropic.NewToolResultBlock(tu.ID, result, false))
			}
		}

		messages = append(messages, anthropic.NewUserMessage(toolResults...))
	}

	return "", fmt.Errorf("agent loop exceeded maximum rounds (%d)", maxRounds)
}

func buildToolParams(mcpTools []mcp.ToolInfo) []anthropic.ToolUnionParam {
	var tools []anthropic.ToolUnionParam
	for _, t := range mcpTools {
		var schema map[string]any
		json.Unmarshal(t.InputSchema, &schema)

		inputSchema := anthropic.ToolInputSchemaParam{}
		if props, ok := schema["properties"]; ok {
			inputSchema.Properties = props
		}
		if req, ok := schema["required"]; ok {
			if reqArr, ok := req.([]any); ok {
				strs := make([]string, len(reqArr))
				for i, v := range reqArr {
					strs[i] = fmt.Sprint(v)
				}
				inputSchema.Required = strs
			}
		}

		tools = append(tools, anthropic.ToolUnionParam{
			OfTool: &anthropic.ToolParam{
				Name:        t.Name,
				Description: anthropic.String(t.Description),
				InputSchema: inputSchema,
			},
		})
	}
	return tools
}
