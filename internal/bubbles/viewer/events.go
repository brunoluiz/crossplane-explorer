package viewer

import (
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/samber/lo"
)

type EventQuit struct{}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		cmds = append(cmds, m.onKey(msg))
	case tea.WindowSizeMsg:
		cmds = append(cmds, m.onResize(msg))
	}

	// Handle keyboard and mouse events in the viewport
	m.viewport, cmd = m.viewport.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m *Model) onKey(msg tea.KeyMsg) tea.Cmd {
	switch {
	case key.Matches(msg, m.KeyMap.Quit):
		return m.cmdQuit
	}
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
		m.viewport.Style = m.Styles.Viewport
		m.viewport.YPosition = headerHeight
		m.viewport.HighPerformanceRendering = m.useHighPerformanceRenderer
		m.viewport.KeyMap.PageUp.SetKeys(lo.Flatten([][]string{m.KeyMap.PageUp.Keys(), m.viewport.KeyMap.PageUp.Keys()})...)
		m.viewport.KeyMap.PageDown.SetKeys(lo.Flatten([][]string{m.KeyMap.PageDown.Keys(), m.viewport.KeyMap.PageDown.Keys()})...)
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
