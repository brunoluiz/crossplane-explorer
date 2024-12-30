package viewer

import (
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type EventSetup struct {
	Title     string
	SideTitle string
	Content   string
}

func (m *Model) onKey(msg tea.KeyMsg) tea.Cmd {
	if k := msg.String(); k == "ctrl+c" || k == "q" || k == "esc" {
		return m.cmdQuit
	}
	return nil
}

func (m *Model) onSetup(msg EventSetup) tea.Cmd {
	m.title = msg.Title
	m.sideTitle = msg.SideTitle
	m.content = msg.Content
	m.viewport.SetContent(m.content)
	return nil
}

func (m *Model) onResize(msg tea.WindowSizeMsg) tea.Cmd {
	headerHeight := lipgloss.Height(m.headerView())
	footerHeight := lipgloss.Height(m.footerView())
	verticalMarginHeight := headerHeight + footerHeight

	if !m.ready {
		// Since this program is using the full size of the viewport we
		// need to wait until we've received the window dimensions before
		// we can initialize the viewport. The initial dimensions come in
		// quickly, though asynchronously, which is why we wait for them
		// here.
		m.viewport = viewport.New(msg.Width, msg.Height-verticalMarginHeight)
		m.viewport.Style = m.viewportStyle
		m.viewport.YPosition = headerHeight
		m.viewport.HighPerformanceRendering = m.useHighPerformanceRenderer
		m.viewport.SetContent(m.content)
		m.ready = true

		// This is only necessary for high performance rendering, which in
		// most cases you won't need.
		//
		// Render the viewport one line below the header.
		m.viewport.YPosition = headerHeight + 1
	} else {
		m.viewport.Width = msg.Width
		m.viewport.Height = msg.Height - verticalMarginHeight
	}

	if m.useHighPerformanceRenderer {
		// Render (or re-render) the whole viewport. Necessary both to
		// initialize the viewport and when the window is resized.
		//
		// This is needed for high-performance rendering only.
		return viewport.Sync(m.viewport)
	}

	return nil
}