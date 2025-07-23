package app

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/atotto/clipboard"
	"github.com/brunoluiz/crossplane-explorer/internal/bubbles/component/navigator"
	"github.com/brunoluiz/crossplane-explorer/internal/bubbles/component/viewer"
	xviewer "github.com/brunoluiz/crossplane-explorer/internal/bubbles/layout/xpsummary"
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

func (m Model) exec(c string, args ...string) tea.Cmd {
	cmd := exec.Command(c, args...)
	// Inherit environment so $EDITOR is respected
	cmd.Env = os.Environ()
	// Attach to the user's terminal
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	// cmd.Stderr = os.Stderr

	return tea.ExecProcess(cmd, func(err error) tea.Msg {
		return nil
	})
}

func (m Model) pager(c string, args ...string) tea.Cmd {
	cmd := c + " " + strings.Join(args, " ")
	pager := os.Getenv("PAGER")
	if pager == "" {
		pager = "less"
	}
	viewCmd := fmt.Sprintf("%s | %s", cmd, pager)

	return m.exec(os.Getenv("SHELL"), "-c", viewCmd)
}

func (m Model) kubectlEdit(ns, resource string) tea.Cmd {
	args := []string{"edit", resource}
	if ns != "" {
		args = append(args, "-n", ns)
	}
	return m.exec("kubectl", args...)
}

func (m Model) kubectlDescribe(ns, resource string) tea.Cmd {
	args := []string{"describe", resource}
	if ns != "" {
		args = append(args, "-n", ns)
	}
	return m.pager("kubectl", args...)
}

func (m Model) kubectlDelete(ns, resource string) tea.Cmd {
	args := []string{"delete", resource}
	if ns != "" {
		args = append(args, "-n", ns)
	}
	return m.exec("kubectl", args...)
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
	case viewer.EventQuit:
		m.pane = PaneNavigator
		return m, nil
	case navigator.EventQuitted:
		return m, tea.Interrupt
	case navigator.EventItemDescribe:
		trace, ok := msg.Data.(*xplane.Resource)
		if !ok {
			return m, nil
		}
		ns, _ := GetPath[string](trace.Unstructured.Object, "metadata", "namespace")
		return m, tea.Batch(tea.HideCursor, m.kubectlDescribe(ns, msg.ID))
	case navigator.EventItemEdit:
		trace, ok := msg.Data.(*xplane.Resource)
		if !ok {
			return m, nil
		}
		ns, _ := GetPath[string](trace.Unstructured.Object, "metadata", "namespace")
		return m, tea.Batch(tea.HideCursor, m.kubectlEdit(ns, msg.ID))
	case navigator.EventItemDelete:
		trace, ok := msg.Data.(*xplane.Resource)
		if !ok {
			return m, nil
		}
		ns, _ := GetPath[string](trace.Unstructured.Object, "metadata", "namespace")
		return m, tea.Batch(tea.HideCursor, m.kubectlDelete(ns, msg.ID))
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
