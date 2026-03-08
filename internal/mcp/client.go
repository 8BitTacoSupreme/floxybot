package mcp

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"sync"
)

// Client manages a subprocess MCP server (flox-mcp) and communicates via JSON-RPC over stdio.
type Client struct {
	cmd       *exec.Cmd
	transport *Transport
	reqID     int
	mu        sync.Mutex
	tools     []ToolInfo
}

// NewClient starts the flox-mcp subprocess and initializes the MCP session.
func NewClient(mcpBinary string) (*Client, error) {
	cmd := exec.Command(mcpBinary)
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, fmt.Errorf("stdin pipe: %w", err)
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("stdout pipe: %w", err)
	}
	// Discard stderr or log it.
	cmd.Stderr = nil

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("starting %s: %w", mcpBinary, err)
	}

	c := &Client{
		cmd:       cmd,
		transport: NewTransport(stdout, stdin),
	}

	// Initialize MCP session.
	if err := c.initialize(); err != nil {
		cmd.Process.Kill()
		return nil, fmt.Errorf("MCP initialize: %w", err)
	}

	// Discover tools.
	if err := c.discoverTools(); err != nil {
		cmd.Process.Kill()
		return nil, fmt.Errorf("MCP tool discovery: %w", err)
	}

	return c, nil
}

func (c *Client) initialize() error {
	params, _ := json.Marshal(map[string]any{
		"protocolVersion": "2024-11-05",
		"capabilities":    map[string]any{},
		"clientInfo": map[string]string{
			"name":    "floxybot",
			"version": "0.1.0",
		},
	})

	_, err := c.call("initialize", params)
	if err != nil {
		return err
	}

	// Send initialized notification.
	return c.transport.Send(Notification{
		JSONRPC: "2.0",
		Method:  "notifications/initialized",
	})
}

func (c *Client) discoverTools() error {
	resp, err := c.call("tools/list", nil)
	if err != nil {
		return err
	}
	var result ToolListResult
	if err := json.Unmarshal(resp, &result); err != nil {
		return fmt.Errorf("parsing tools list: %w", err)
	}
	c.tools = result.Tools
	return nil
}

// Tools returns the list of available MCP tools.
func (c *Client) Tools() []ToolInfo {
	return c.tools
}

// CallTool invokes an MCP tool by name with the given arguments.
func (c *Client) CallTool(name string, args map[string]any) (*CallToolResult, error) {
	params, err := json.Marshal(CallToolParams{Name: name, Arguments: args})
	if err != nil {
		return nil, err
	}
	resp, err := c.call("tools/call", params)
	if err != nil {
		return nil, err
	}
	var result CallToolResult
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, fmt.Errorf("parsing tool result: %w", err)
	}
	return &result, nil
}

// Close shuts down the MCP subprocess.
func (c *Client) Close() error {
	if c.cmd.Process != nil {
		c.cmd.Process.Kill()
	}
	return c.cmd.Wait()
}

func (c *Client) call(method string, params json.RawMessage) (json.RawMessage, error) {
	c.mu.Lock()
	c.reqID++
	id := c.reqID
	c.mu.Unlock()

	req := Request{
		JSONRPC: "2.0",
		ID:      id,
		Method:  method,
		Params:  params,
	}
	if err := c.transport.Send(req); err != nil {
		return nil, fmt.Errorf("sending %s: %w", method, err)
	}

	// Read responses until we get our ID (skip notifications).
	for {
		raw, err := c.transport.Receive()
		if err != nil {
			return nil, fmt.Errorf("receiving %s response: %w", method, err)
		}

		var resp Response
		if err := json.Unmarshal(raw, &resp); err != nil {
			// Might be a notification — skip.
			continue
		}
		if resp.ID == id {
			if resp.Error != nil {
				return nil, resp.Error
			}
			return resp.Result, nil
		}
	}
}
