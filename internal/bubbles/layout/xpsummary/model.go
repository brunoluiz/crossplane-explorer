package xpsummary

import (
	"fmt"

	"github.com/brunoluiz/xpdig/internal/bubbles/component/viewer"
	"github.com/brunoluiz/xpdig/internal/ds"
	"github.com/brunoluiz/xpdig/internal/xplane"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/goccy/go-yaml"
)

type Model struct {
	viewer viewer.Model

	styles Styles
}

func New() Model {
	return Model{
		viewer: viewer.New(),
		styles: DefaultStyles(),
	}
}

func (m Model) Init() tea.Cmd { return nil }
func (m Model) View() string  { return m.viewer.View() }

type ContentInput struct {
	Trace *xplane.Resource
}

func (m *Model) SetContent(msg ContentInput) error {
	obj := msg.Trace.Unstructured.Object
	ds.WalkMap(obj, func(key string, value any) (any, bool) {
		// These fields are usually injected server side and make checking objects quite hard
		if key == "managedFields" || keyHasSuffix(key, ".managedFields") {
			return nil, false
		}
		return value, true
	})

	val, err := yaml.Marshal(obj)
	if err != nil {
		return err
	}

	m.viewer.SetContent(viewer.ContentInput{
		Title:     fmt.Sprintf("%s/%s", msg.Trace.Unstructured.GetKind(), msg.Trace.Unstructured.GetName()),
		SideTitle: msg.Trace.Unstructured.GetAPIVersion(),
		Content: m.styles.Main.Render(lipgloss.JoinVertical(
			lipgloss.Top, string(val),
		)),
	})
	return nil
}

// Helper function to check if a key has a specific suffix
func keyHasSuffix(key, suffix string) bool {
	if len(key) < len(suffix) {
		return false
	}
	return key[len(key)-len(suffix):] == suffix
}
