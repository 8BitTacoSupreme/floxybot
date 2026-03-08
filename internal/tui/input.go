package tui

import (
	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
)

// inputModel wraps a textarea for the shared input area.
type inputModel struct {
	textarea textarea.Model
}

func newInputModel() inputModel {
	ta := textarea.New()
	ta.Placeholder = "Ask a question..."
	ta.CharLimit = 4000
	ta.SetHeight(1)
	ta.ShowLineNumbers = false
	ta.FocusedStyle.CursorLine = ta.FocusedStyle.CursorLine.UnsetBackground()
	ta.Focus()
	return inputModel{textarea: ta}
}

func (m inputModel) Update(msg tea.Msg) (inputModel, tea.Cmd) {
	var cmd tea.Cmd
	m.textarea, cmd = m.textarea.Update(msg)
	return m, cmd
}

func (m inputModel) View() string {
	return inputStyle.Render(m.textarea.View())
}

func (m inputModel) Value() string {
	return m.textarea.Value()
}

func (m *inputModel) Reset() {
	m.textarea.Reset()
}

func (m *inputModel) SetWidth(w int) {
	m.textarea.SetWidth(w - 4) // account for padding/border
}
