package explorer

import (
	"time"

	xviewer "github.com/brunoluiz/crossplane-explorer/internal/bubbles/layout/viewer"
	"github.com/brunoluiz/crossplane-explorer/internal/bubbles/shared/tree"
	"github.com/brunoluiz/crossplane-explorer/internal/bubbles/shared/viewer"
	"github.com/brunoluiz/crossplane-explorer/internal/xplane"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case error:
		m.setIrrecoverableError(msg)
		return m, nil
	case *xplane.Resource:
		cmd = m.onLoad(msg)
	case tea.WindowSizeMsg:
		return m, m.onResize(msg)
	case tea.KeyMsg:
		cmd = m.onKey(msg)
	case viewer.EventQuit:
		m.pane = PaneTree
		return m, nil
	case tree.EventQuit:
		return m, tea.Interrupt
	case tree.EventShow:
		trace, ok := msg.Node.Value.(*xplane.Resource)
		if !ok {
			return m, nil
		}

		if err := m.viewer.SetContent(xviewer.ContentInput{Trace: trace}); err != nil {
			m.setIrrecoverableError(err)
			return m, nil
		}
		m.pane = PaneSummary
	}

	switch m.pane {
	case PaneSummary:
		var viewerCmd tea.Cmd
		m.viewer, viewerCmd = m.viewer.Update(msg)
		return m, tea.Batch(cmd, viewerCmd)
	case PaneTree:
		var treeCmd, statusCmd tea.Cmd
		m.tree, treeCmd = m.tree.Update(msg)

		return m, tea.Batch(cmd, statusCmd, treeCmd)
	case PaneIrrecoverableError:
		return m, cmd
	}

	return m, cmd
}

func (m *Model) onLoad(data *xplane.Resource) tea.Cmd {
	if data == nil {
		return nil
	}

	m.setColumns(data.Unstructured.GroupVersionKind().GroupKind())
	m.setNodes(data)

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
	m.tree, _ = m.tree.Update(tea.WindowSizeMsg{Width: m.width - right - left, Height: m.height - top})
	m.viewer, _ = m.viewer.Update(msg)

	return nil
}

func (m *Model) onKey(msg tea.KeyMsg) tea.Cmd {
	switch {
	case key.Matches(msg, m.keyMap.Quit):
		return tea.Interrupt
	}

	return nil
}
