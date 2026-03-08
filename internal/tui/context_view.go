package tui

import (
	"fmt"
	"strings"

	"github.com/8BitTacoSupreme/floxybot/internal/floxctx"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
)

// contextViewModel is the Context tab — read-only Flox env display.
type contextViewModel struct {
	viewport viewport.Model
	ctx      *floxctx.Context
}

func newContextViewModel(fctx *floxctx.Context) contextViewModel {
	vp := viewport.New(80, 20)
	vp.SetContent(formatContextView(fctx))
	return contextViewModel{viewport: vp, ctx: fctx}
}

func (m contextViewModel) Update(msg tea.Msg) (contextViewModel, tea.Cmd) {
	var cmd tea.Cmd
	m.viewport, cmd = m.viewport.Update(msg)
	return m, cmd
}

func (m contextViewModel) View() string {
	return m.viewport.View()
}

func (m *contextViewModel) setSize(w, h int) {
	m.viewport.Width = w
	m.viewport.Height = h
}

func formatContextView(ctx *floxctx.Context) string {
	if ctx == nil {
		return "No Flox context detected."
	}

	var b strings.Builder

	if !ctx.InFloxEnv {
		b.WriteString(sectionStyle.Render("Flox Environment") + "\n")
		b.WriteString("  Not detected. Activate a Flox environment to see details.\n\n")
		b.WriteString(sectionStyle.Render("System") + "\n")
		b.WriteString(fmt.Sprintf("  %s (%s)\n", ctx.System.OS, ctx.System.Arch))
		return b.String()
	}

	b.WriteString(sectionStyle.Render("Flox Environment") + "\n")
	if ctx.FloxEnv != "" {
		b.WriteString(fmt.Sprintf("  FLOX_ENV: %s\n", ctx.FloxEnv))
	}
	if ctx.FloxEnvProject != "" {
		b.WriteString(fmt.Sprintf("  FLOX_ENV_PROJECT: %s\n", ctx.FloxEnvProject))
	}
	b.WriteString("\n")

	if len(ctx.Packages) > 0 {
		b.WriteString(sectionStyle.Render(fmt.Sprintf("Installed Packages (%d)", len(ctx.Packages))) + "\n")
		for _, p := range ctx.Packages {
			b.WriteString(fmt.Sprintf("  - %s\n", p))
		}
		b.WriteString("\n")
	}

	if len(ctx.FloxVars) > 0 {
		b.WriteString(sectionStyle.Render("Flox Variables") + "\n")
		for _, v := range ctx.FloxVars {
			b.WriteString(fmt.Sprintf("  %s = %s\n", v.Key, v.Value))
		}
		b.WriteString("\n")
	}

	if ctx.Manifest != nil && len(ctx.Manifest.Install) > 0 {
		b.WriteString(sectionStyle.Render("Manifest Packages") + "\n")
		for name, spec := range ctx.Manifest.Install {
			b.WriteString(fmt.Sprintf("  %s: %s\n", name, spec.PkgPath))
		}
		b.WriteString("\n")
	}

	b.WriteString(sectionStyle.Render("System") + "\n")
	b.WriteString(fmt.Sprintf("  %s (%s)\n", ctx.System.OS, ctx.System.Arch))

	return b.String()
}
