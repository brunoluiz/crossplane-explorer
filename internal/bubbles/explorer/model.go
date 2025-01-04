package explorer

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/brunoluiz/crossplane-explorer/internal/bubbles/explorer/statusbar"
	"github.com/brunoluiz/crossplane-explorer/internal/bubbles/explorer/viewer"
	"github.com/brunoluiz/crossplane-explorer/internal/bubbles/table"
	"github.com/brunoluiz/crossplane-explorer/internal/bubbles/tree"
	"github.com/brunoluiz/crossplane-explorer/internal/xplane"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	HeaderKeyObject     = "OBJECT"
	HeaderKeyGroup      = "GROUP"
	HeaderKeySynced     = "SYNCED"
	HeaderKeySyncedLast = "SYNCED LAST"
	HeaderKeyReady      = "READY"
	HeaderKeyReadyLast  = "READY LAST"
	HeaderKeyStatus     = "STATUS"
)

type Pane string

const (
	PaneTree    Pane = "tree"
	PaneSummary Pane = "summary"
)

type Tracer interface {
	GetTrace() (*xplane.Resource, error)
}

type Model struct {
	tree          tree.Model
	statusbar     *statusbar.Model // requires pointer here
	viewer        viewer.Model
	tracer        Tracer
	width         int
	height        int
	short         bool
	watch         bool
	watchInterval time.Duration
	logger        *slog.Logger

	pane      Pane
	err       error
	resByNode map[*tree.Node]*xplane.Resource
}

type WithOpt func(*Model)

func WithWatch(enabled bool) func(*Model) {
	return func(m *Model) {
		m.watch = enabled
	}
}

func WithWatchInterval(t time.Duration) func(*Model) {
	return func(m *Model) {
		m.watchInterval = t
	}
}

func WithShortColumns(enabled bool) func(*Model) {
	return func(m *Model) {
		m.short = enabled
	}
}

func New(
	logger *slog.Logger,
	treeModel tree.Model,
	viewerModel viewer.Model,
	statusbarModel statusbar.Model,
	tracer Tracer,
	opts ...WithOpt,
) *Model {
	m := &Model{
		logger:        logger,
		tree:          treeModel,
		statusbar:     &statusbarModel,
		viewer:        viewerModel,
		tracer:        tracer,
		width:         0,
		height:        0,
		watchInterval: 10 * time.Second,
		short:         true,

		pane:      PaneTree,
		resByNode: map[*tree.Node]*xplane.Resource{},
	}
	m.tree.OnSelectionChange = func(n *tree.Node) {
		m.statusbar.SetPath(n.Path)
	}

	for _, opt := range opts {
		opt(m)
	}

	return m
}

func (m Model) getTrace() tea.Cmd {
	return func() tea.Msg {
		res, err := m.tracer.GetTrace()
		if err != nil {
			return err
		}
		return res
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(m.getTrace())
}

func (m Model) View() string {
	if m.err != nil {
		return fmt.Sprintf("There was a fatal error: %s\nPress q to exit", m.err.Error())
	}

	switch m.pane {
	case PaneSummary:
		return m.viewer.View()
	case PaneTree:
		return lipgloss.JoinVertical(
			lipgloss.Left,
			lipgloss.NewStyle().Height(m.height-m.statusbar.GetHeight()).Render(m.tree.View()),
			m.statusbar.View(),
		)
	default:
		return "No pane selected"
	}
}

type ColumnLayout int

const (
	UnknownColumnLayout ColumnLayout = iota
	ShortObjectColumnLayout
	WideObjectColumnLayout
	ShortPkgColumnLayout
	WidePkgColumnLayout
)

func (m Model) getColumns(layout ColumnLayout) []table.Column {
	switch layout {
	case ShortObjectColumnLayout:
		return []table.Column{
			{Title: HeaderKeyObject, Width: 60},
			{Title: HeaderKeyGroup, Width: 30},
			{Title: HeaderKeySynced, Width: 7},
			{Title: HeaderKeyReady, Width: 7},
			{Title: HeaderKeyStatus, Width: 68},
		}
	case WideObjectColumnLayout:
		return []table.Column{
			{Title: HeaderKeyObject, Width: 60},
			{Title: HeaderKeyGroup, Width: 30},
			{Title: HeaderKeySynced, Width: 7},
			{Title: HeaderKeySyncedLast, Width: 19},
			{Title: HeaderKeyReady, Width: 7},
			{Title: HeaderKeyReadyLast, Width: 19},
			{Title: HeaderKeyStatus, Width: 68},
		}
	default:
		return []table.Column{}
	}
}
