package agent

import (
	"fmt"
	"strings"
)

// destructiveTools lists MCP tool names that require user confirmation.
var destructiveTools = map[string]bool{
	"flox_install":   true,
	"flox_uninstall": true,
	"flox_edit":      true,
	"flox_delete":    true,
	"flox_push":      true,
	"flox_activate":  true,
}

// IsDestructive returns true if the tool name requires confirmation.
func IsDestructive(toolName string) bool {
	return destructiveTools[toolName]
}

// FormatConfirmation returns a human-readable prompt for confirming a destructive action.
func FormatConfirmation(toolName string, args map[string]any) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("The agent wants to run: %s\n", toolName))
	for k, v := range args {
		sb.WriteString(fmt.Sprintf("  %s: %v\n", k, v))
	}
	sb.WriteString("Allow? [y/N] ")
	return sb.String()
}
