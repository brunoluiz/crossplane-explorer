package xpnavigator

import (
	"time"

	"github.com/brunoluiz/crossplane-explorer/internal/bubbles/component/navigator"
	"github.com/brunoluiz/crossplane-explorer/internal/xplane"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case *xplane.Resource:
		cmd = m.onCrossplaneUpdate(msg)
	case tea.WindowSizeMsg:
		return m, m.onResize(msg)
	case tea.KeyMsg:
		cmd = m.onKey(msg)
	case navigator.EventItemFocused:
		m.statusbar.SetPath(m.pathByData[msg.ID])
	}

	var navigatorCmd tea.Cmd
	m.navigator, navigatorCmd = m.navigator.Update(msg)

	var statusBarCmd tea.Cmd
	m.statusbar, statusBarCmd = m.statusbar.Update(msg)

	return m, tea.Batch(cmd, navigatorCmd, statusBarCmd)
}

func (m *Model) onCrossplaneUpdate(data *xplane.Resource) tea.Cmd {
	if data == nil {
		return nil
	}

	m.setColumns(data.Unstructured.GroupVersionKind().GroupKind())
	m.setData(data)

	if m.watch {
		return tea.Tick(m.watchInterval, func(_ time.Time) tea.Msg {
			return m.getTrace()()
		})
	}
	return nil
}

func (m *Model) onResize(msg tea.WindowSizeMsg) tea.Cmd {
	var navigatorCmd, statusbarCmd tea.Cmd
	m.width = msg.Width
	m.height = msg.Height

	top, _, _, _ := lipgloss.NewStyle().Padding(1).GetPadding()
	m.navigator, navigatorCmd = m.navigator.Update(tea.WindowSizeMsg{Width: m.width, Height: m.height - top})

	m.statusbar, statusbarCmd = m.statusbar.Update(msg)

	return tea.Batch(navigatorCmd, statusbarCmd)
}

func (m *Model) onKey(_ tea.KeyMsg) tea.Cmd {
	return nil
}
