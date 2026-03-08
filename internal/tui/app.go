package tui

import (
	"github.com/8BitTacoSupreme/floxybot/internal/claude"
	"github.com/8BitTacoSupreme/floxybot/internal/floxctx"
	"github.com/8BitTacoSupreme/floxybot/internal/rag"
	tea "github.com/charmbracelet/bubbletea"
)

// App is the top-level bubbletea model with tabbed layout.
type App struct {
	activeTab tabID
	width     int
	height    int

	input       inputModel
	chat        chatModel
	copilot     copilotModel
	contextView contextViewModel
}

// NewApp creates the TUI application model.
func NewApp(cc *claude.Client, ret *rag.Retriever, fctx *floxctx.Context) App {
	return App{
		activeTab:   tabChat,
		input:       newInputModel(),
		chat:        newChatModel(cc, ret, fctx),
		copilot:     newCopilotModel(),
		contextView: newContextViewModel(fctx),
	}
}

func (a App) Init() tea.Cmd {
	return nil
}

func (a App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return a, tea.Quit
		case "tab":
			a.activeTab = (a.activeTab + 1) % tabID(len(tabNames))
			return a, nil
		case "shift+tab":
			a.activeTab = (a.activeTab - 1 + tabID(len(tabNames))) % tabID(len(tabNames))
			return a, nil
		case "enter":
			if a.activeTab == tabChat || a.activeTab == tabCoPilot {
				text := a.input.Value()
				if text == "" {
					return a, nil
				}
				a.input.Reset()
				if a.activeTab == tabChat {
					cmd := a.chat.submit(text)
					return a, cmd
				}
				// Co-Pilot submit will be wired in Phase 4.
			}
		}

	case tea.WindowSizeMsg:
		a.width = msg.Width
		a.height = msg.Height
		contentHeight := a.height - 5 // tab bar + input + status
		a.chat.setSize(a.width, contentHeight)
		a.copilot.setSize(a.width, contentHeight)
		a.contextView.setSize(a.width, contentHeight)
		a.input.SetWidth(a.width)
	}

	// Route updates to active tab.
	var cmd tea.Cmd
	switch a.activeTab {
	case tabChat:
		a.chat, cmd = a.chat.Update(msg)
	case tabCoPilot:
		a.copilot, cmd = a.copilot.Update(msg)
	case tabContext:
		a.contextView, cmd = a.contextView.Update(msg)
	}

	// Always update input for keystrokes (except on Context tab).
	if a.activeTab != tabContext {
		var inputCmd tea.Cmd
		a.input, inputCmd = a.input.Update(msg)
		if cmd == nil {
			cmd = inputCmd
		} else {
			cmd = tea.Batch(cmd, inputCmd)
		}
	}

	return a, cmd
}

func (a App) View() string {
	tabBar := renderTabBar(a.activeTab, a.width)

	var content string
	switch a.activeTab {
	case tabChat:
		content = a.chat.View()
	case tabCoPilot:
		content = a.copilot.View()
	case tabContext:
		content = a.contextView.View()
	}

	status := statusStyle.Render("Tab/Shift+Tab: switch tabs | Enter: send | Ctrl+C: quit")

	if a.activeTab == tabContext {
		return tabBar + "\n" + content + "\n" + status
	}
	return tabBar + "\n" + content + "\n" + a.input.View() + "\n" + status
}
