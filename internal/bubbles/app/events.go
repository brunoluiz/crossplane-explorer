package app

import (
	"github.com/atotto/clipboard"
	"github.com/brunoluiz/crossplane-explorer/internal/bubbles/component/navigator"
	"github.com/brunoluiz/crossplane-explorer/internal/xplane"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

func GetPath[V any](m map[string]any, path ...string) (V, bool) {
	var zero V
	curr := any(m)
	for i, key := range path {
		mm, ok := curr.(map[string]any)
		if !ok {
			return zero, false
		}
		v, exists := mm[key]
		if !exists {
			return zero, false
		}
		curr = v
		// If this is the last key, try to cast to V
		if i == len(path)-1 {
			val, ok := curr.(V)
			if ok {
				return val, true
			}
			return zero, false
		}
	}
	return zero, false
}

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
	case navigator.EventQuitted:
		return m, tea.Interrupt
	case navigator.EventItemDescribe:
		trace, ok := msg.Data.(*xplane.Resource)
		if !ok {
			return m, nil
		}
		ns, _ := GetPath[string](trace.Unstructured.Object, "metadata", "namespace")
		return m, tea.Batch(tea.HideCursor, m.kubectl.Describe(ns, msg.ID))
	case navigator.EventItemEdit:
		trace, ok := msg.Data.(*xplane.Resource)
		if !ok {
			return m, nil
		}
		ns, _ := GetPath[string](trace.Unstructured.Object, "metadata", "namespace")
		return m, tea.Batch(tea.HideCursor, m.kubectl.Edit(ns, msg.ID))
	case navigator.EventItemDelete:
		trace, ok := msg.Data.(*xplane.Resource)
		if !ok {
			return m, nil
		}
		ns, _ := GetPath[string](trace.Unstructured.Object, "metadata", "namespace")
		return m, tea.Batch(tea.HideCursor, m.kubectl.Delete(ns, msg.ID))
	case navigator.EventItemCopied:
		//nolint // ignore errors
		clipboard.WriteAll(msg.ID)
	case navigator.EventItemSelected:
		trace, ok := msg.Data.(*xplane.Resource)
		if !ok {
			return m, nil
		}
		ns, _ := GetPath[string](trace.Unstructured.Object, "metadata", "namespace")
		return m, tea.Batch(tea.HideCursor, m.kubectl.Get(ns, msg.ID))
	}

	switch m.pane {
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
