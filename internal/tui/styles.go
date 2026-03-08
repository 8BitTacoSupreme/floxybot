package tui

import "github.com/charmbracelet/lipgloss"

var (
	// Tab bar styles.
	activeTabStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("205")).
			Border(lipgloss.NormalBorder(), false, false, true, false).
			BorderForeground(lipgloss.Color("205")).
			Padding(0, 2)

	inactiveTabStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("240")).
				Padding(0, 2)

	tabGapStyle = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder(), false, false, true, false).
			BorderForeground(lipgloss.Color("236"))

	// Content area.
	contentStyle = lipgloss.NewStyle().
			Padding(1, 2)

	// Input area.
	inputStyle = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder(), true, false, false, false).
			BorderForeground(lipgloss.Color("236")).
			Padding(0, 1)

	// Status bar.
	statusStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")).
			Padding(0, 1)

	// Spinner / loading.
	spinnerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("205"))

	// Assistant message.
	assistantStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("252"))

	// User message.
	userStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("117")).
			Bold(true)

	// Error.
	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("196"))

	// Section header in context view.
	sectionStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("205")).
			Bold(true)
)
