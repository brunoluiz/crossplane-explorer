package xpnavigator

import (
	"time"

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
	}

	var navigatorCmd tea.Cmd
	m.navigator, navigatorCmd = m.navigator.Update(msg)

	return m, tea.Batch(cmd, navigatorCmd)
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
	m.width = msg.Width
	m.height = msg.Height

	top, right, _, left := lipgloss.NewStyle().Padding(1).GetPadding()
	m.navigator, _ = m.navigator.Update(tea.WindowSizeMsg{Width: m.width - right - left, Height: m.height - top})

	return nil
}

func (m *Model) onKey(msg tea.KeyMsg) tea.Cmd {
	return nil
}
