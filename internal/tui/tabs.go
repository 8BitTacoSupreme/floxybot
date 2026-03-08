package tui

import "strings"

type tabID int

const (
	tabChat tabID = iota
	tabCoPilot
	tabContext
)

var tabNames = []string{"Chat", "Co-Pilot", "Context"}

func renderTabBar(active tabID, width int) string {
	var tabs []string
	for i, name := range tabNames {
		if tabID(i) == active {
			tabs = append(tabs, activeTabStyle.Render(name))
		} else {
			tabs = append(tabs, inactiveTabStyle.Render(name))
		}
	}
	row := strings.Join(tabs, "")
	gap := tabGapStyle.Render(strings.Repeat(" ", max(0, width-lipglossWidth(row))))
	return row + gap
}

func lipglossWidth(s string) int {
	// Count visible characters (rough — lipgloss handles ANSI).
	// For tab bar sizing this is good enough.
	count := 0
	inEscape := false
	for _, r := range s {
		if r == '\x1b' {
			inEscape = true
			continue
		}
		if inEscape {
			if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') {
				inEscape = false
			}
			continue
		}
		count++
	}
	return count
}
