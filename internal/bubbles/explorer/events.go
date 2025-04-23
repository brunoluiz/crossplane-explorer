package explorer

import (
	"fmt"
	"time"

	"github.com/brunoluiz/crossplane-explorer/internal/bubbles/explorer/viewer"
	"github.com/brunoluiz/crossplane-explorer/internal/bubbles/tree"
	"github.com/brunoluiz/crossplane-explorer/internal/xplane"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
	"k8s.io/apimachinery/pkg/runtime/schema"
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
	switch msg.String() {
	case "ctrl+c", "ctrl+d":
		return tea.Interrupt
	case "enter", "y":
		if m.pane != PaneTree || m.tree.IsSearchMode() {
			return nil
		}

		curr := m.tree.Current().Value
		trace, ok := curr.(*xplane.Resource)
		if !ok {
			return nil
		}

		err := m.viewer.SetContent(viewer.ContentInput{Trace: trace})
		if err != nil {
			m.setIrrecoverableError(err)
			return nil
		}

		m.pane = PaneSummary
	case "esc":
		if m.pane != PaneTree {
			m.pane = PaneTree
		}
	}

	return nil
}

func addNodes(kind schema.GroupKind, v *xplane.Resource, n *tree.Node) {
	name := fmt.Sprintf("%s/%s", v.Unstructured.GetKind(), v.Unstructured.GetName())
	group := v.Unstructured.GetObjectKind().GroupVersionKind().Group

	n.Label = name
	n.Key = fmt.Sprintf("%s.%s/%s", v.Unstructured.GetKind(), group, v.Unstructured.GetName())
	n.Children = make([]*tree.Node, len(v.Children))

	if v.Unstructured.GetAnnotations()["crossplane.io/paused"] == "true" {
		n.Label += " (paused)"
		n.Color = lipgloss.ANSIColor(ansi.Yellow)
	}

	if xplane.IsPkg(kind) {
		resStatus := xplane.GetPkgResourceStatus(v, name)
		n.Details = map[string]string{
			HeaderKeyVersion:       resStatus.Version,
			HeaderKeyInstalled:     resStatus.Installed,
			HeaderKeyInstalledLast: getTimeStr(resStatus.InstalledLastTransition),
			HeaderKeyHealthy:       resStatus.Healthy,
			HeaderKeyHealthyLast:   getTimeStr(resStatus.HealthyLastTransition),
			HeaderKeyState:         resStatus.State,
			HeaderKeyStatus:        resStatus.Status,
		}
		if !resStatus.Ok {
			n.Color = lipgloss.ANSIColor(ansi.Red)
		}
	} else {
		resStatus := xplane.GetResourceStatus(v, name)
		n.Details = map[string]string{
			HeaderKeyGroup:      group,
			HeaderKeySynced:     resStatus.Synced,
			HeaderKeySyncedLast: getTimeStr(resStatus.SyncedLastTransition),
			HeaderKeyReady:      resStatus.Ready,
			HeaderKeyReadyLast:  getTimeStr(resStatus.ReadyLastTransition),
			HeaderKeyStatus:     resStatus.Status,
		}
		if !resStatus.Ok {
			n.Color = lipgloss.ANSIColor(ansi.Red)
		}
	}
	n.Value = v

	for k, cv := range v.Children {
		n.Children[k] = &tree.Node{}
		addNodes(kind, cv, n.Children[k])
	}
}

func getTimeStr(t time.Time) string {
	if t.IsZero() {
		return "-"
	}
	return t.Format(time.RFC822)
}
