package app

import (
	"github.com/atotto/clipboard"
	"github.com/brunoluiz/crossplane-explorer/internal/bubbles/component/navigator"
	"github.com/brunoluiz/crossplane-explorer/internal/bubbles/component/viewer"
	xviewer "github.com/brunoluiz/crossplane-explorer/internal/bubbles/layout/xpsummary"
	"github.com/brunoluiz/crossplane-explorer/internal/xplane"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	m.dumper("new message", msg)

	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		return m, m.onResize(msg)
	case error:
		m.setIrrecoverableError(msg)
		return m, nil
	case *xplane.Resource:
		m.navigator, cmd = m.navigator.Update(msg)
		return m, cmd
	case tea.KeyMsg:
		cmd = m.onKey(msg)
	case viewer.EventQuit:
		m.pane = PaneNavigator
		return m, nil
	case navigator.EventQuitted:
		return m, tea.Interrupt
	case navigator.EventItemCopied:
		//nolint // ignore errors
		clipboard.WriteAll(msg.ID)
	case navigator.EventItemSelected:
		trace, ok := msg.Data.(*xplane.Resource)
		if !ok {
			return m, nil
		}

		if err := m.viewer.SetContent(xviewer.ContentInput{Trace: trace}); err != nil {
			m.setIrrecoverableError(err)
			return m, nil
		}
		m.pane = PaneViewer
		return m, nil
	}

	switch m.pane {
	case PaneViewer:
		var viewerCmd tea.Cmd
		m.viewer, viewerCmd = m.viewer.Update(msg)
		return m, tea.Batch(cmd, viewerCmd)
	case PaneNavigator:
		var navigatorCmd, statusCmd tea.Cmd
		m.navigator, navigatorCmd = m.navigator.Update(msg)

		return m, tea.Batch(cmd, statusCmd, navigatorCmd)
	case PaneIrrecoverableError:
		return m, cmd
	}

	return m, cmd
}

func (m *Model) onResize(msg tea.WindowSizeMsg) tea.Cmd {
	var navigatorCmd, viewerCmd tea.Cmd
	m.navigator, navigatorCmd = m.navigator.Update(msg)
	m.viewer, viewerCmd = m.viewer.Update(msg)

	return tea.Batch(navigatorCmd, viewerCmd)
}

func (m *Model) onKey(msg tea.KeyMsg) tea.Cmd {
	//nolint
	switch {
	case key.Matches(msg, m.keyMap.Quit):
		return tea.Interrupt
	}

	return nil
}
