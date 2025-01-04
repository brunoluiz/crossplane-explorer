package viewer

import (
	"fmt"

	"github.com/brunoluiz/crossplane-explorer/internal/bubbles/viewer"
	"github.com/brunoluiz/crossplane-explorer/internal/xplane"
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

func (m *Model) SetContent(msg ContentInput) {
	val, err := yaml.Marshal(msg.Trace.Unstructured.Object)
	if err != nil {
		panic(err)
	}

	m.viewer.SetContent(viewer.ContentInput{
		Title:     fmt.Sprintf("%s/%s", msg.Trace.Unstructured.GetKind(), msg.Trace.Unstructured.GetName()),
		SideTitle: msg.Trace.Unstructured.GetAPIVersion(),
		Content: m.styles.Main.Render(lipgloss.JoinVertical(
			lipgloss.Top,
			string(val),
		)),
	})
}
