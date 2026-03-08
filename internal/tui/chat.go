package tui

import (
	"context"
	"strings"

	"github.com/8BitTacoSupreme/floxybot/internal/claude"
	"github.com/8BitTacoSupreme/floxybot/internal/floxctx"
	"github.com/8BitTacoSupreme/floxybot/internal/rag"
	"github.com/anthropics/anthropic-sdk-go"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
)

// chatMessage stores a conversation turn.
type chatMessage struct {
	role    string // "user" or "assistant"
	content string
}

// chatModel is the Chat tab state.
type chatModel struct {
	viewport  viewport.Model
	messages  []chatMessage
	history   []anthropic.MessageParam
	streaming bool

	claudeClient *claude.Client
	retriever    *rag.Retriever
	floxCtx      *floxctx.Context
}

func newChatModel(cc *claude.Client, ret *rag.Retriever, fctx *floxctx.Context) chatModel {
	vp := viewport.New(80, 20)
	vp.SetContent("Welcome to Floxybot! Ask anything about Flox.\n")
	return chatModel{
		viewport:     vp,
		claudeClient: cc,
		retriever:    ret,
		floxCtx:      fctx,
	}
}

// streamDoneMsg signals that streaming is complete.
type streamDoneMsg struct {
	fullText string
	err      error
}

// streamTokenMsg carries a single streamed token.
type streamTokenMsg struct {
	text string
}

func (m chatModel) Update(msg tea.Msg) (chatModel, tea.Cmd) {
	switch msg := msg.(type) {
	case streamTokenMsg:
		if len(m.messages) > 0 && m.messages[len(m.messages)-1].role == "assistant" {
			m.messages[len(m.messages)-1].content += msg.text
			m.viewport.SetContent(m.renderMessages())
			m.viewport.GotoBottom()
		}
	case streamDoneMsg:
		m.streaming = false
		if msg.err != nil {
			m.messages = append(m.messages, chatMessage{
				role:    "assistant",
				content: errorStyle.Render("Error: " + msg.err.Error()),
			})
		}
		if msg.fullText != "" {
			m.history = append(m.history, anthropic.NewAssistantMessage(
				anthropic.NewTextBlock(msg.fullText),
			))
		}
		m.viewport.SetContent(m.renderMessages())
		m.viewport.GotoBottom()
	}

	var cmd tea.Cmd
	m.viewport, cmd = m.viewport.Update(msg)
	return m, cmd
}

func (m *chatModel) submit(userText string) tea.Cmd {
	m.messages = append(m.messages, chatMessage{role: "user", content: userText})
	m.messages = append(m.messages, chatMessage{role: "assistant", content: ""})
	m.streaming = true

	m.history = append(m.history, anthropic.NewUserMessage(
		anthropic.NewTextBlock(userText),
	))

	m.viewport.SetContent(m.renderMessages())
	m.viewport.GotoBottom()

	historyCopy := make([]anthropic.MessageParam, len(m.history))
	copy(historyCopy, m.history)

	return func() tea.Msg {
		ctx := context.Background()

		// Retrieve RAG context.
		var ragChunks []rag.RetrievedChunk
		if m.retriever != nil {
			chunks, err := m.retriever.Retrieve(ctx, userText, 5)
			if err == nil {
				ragChunks = chunks
			}
		}

		systemPrompt := claude.BuildSystemPrompt(ragChunks, m.floxCtx)

		if m.claudeClient == nil {
			// No API key — show RAG results only.
			if len(ragChunks) > 0 {
				return streamDoneMsg{fullText: rag.FormatForPrompt(ragChunks)}
			}
			return streamDoneMsg{err: nil, fullText: "No API key configured. Set ANTHROPIC_API_KEY to enable chat."}
		}

		fullText, err := m.claudeClient.StreamChat(ctx, systemPrompt, historyCopy, func(text string) {
			// Note: we can't send tea.Msg from here directly in a blocking stream.
			// The streaming tokens will be accumulated in fullText.
		})
		return streamDoneMsg{fullText: fullText, err: err}
	}
}

func (m chatModel) renderMessages() string {
	var sb strings.Builder
	for _, msg := range m.messages {
		switch msg.role {
		case "user":
			sb.WriteString(userStyle.Render("> "+msg.content) + "\n\n")
		case "assistant":
			rendered := renderMarkdown(msg.content)
			sb.WriteString(assistantStyle.Render(rendered) + "\n")
		}
	}
	return sb.String()
}

func (m *chatModel) setSize(w, h int) {
	m.viewport.Width = w
	m.viewport.Height = h
}

func (m chatModel) View() string {
	return m.viewport.View()
}
