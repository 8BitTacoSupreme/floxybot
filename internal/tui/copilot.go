package tui

import (
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
)

// copilotModel is the Co-Pilot tab state. Wired in Phase 4.
type copilotModel struct {
	viewport viewport.Model
	log      string
}

func newCopilotModel() copilotModel {
	vp := viewport.New(80, 20)
	vp.SetContent("Co-Pilot mode — not yet implemented (Phase 4)\nDescribe a task and the agent will execute it via Flox MCP.")
	return copilotModel{viewport: vp}
}

func (m copilotModel) Update(msg tea.Msg) (copilotModel, tea.Cmd) {
	var cmd tea.Cmd
	m.viewport, cmd = m.viewport.Update(msg)
	return m, cmd
}

func (m copilotModel) View() string {
	return m.viewport.View()
}

func (m *copilotModel) setSize(w, h int) {
	m.viewport.Width = w
	m.viewport.Height = h
}
