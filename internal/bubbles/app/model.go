package explorer

import (
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/brunoluiz/crossplane-explorer/internal/bubbles/layout/viewer"
	"github.com/brunoluiz/crossplane-explorer/internal/bubbles/shared/navigator"
	"github.com/brunoluiz/crossplane-explorer/internal/bubbles/shared/table"
	"github.com/brunoluiz/crossplane-explorer/internal/xplane"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

const (
	HeaderKeyObject = "OBJECT"

	HeaderKeyVersion       = "VERSION"
	HeaderKeyInstalled     = "INSTALLED"
	HeaderKeyInstalledLast = "INSTALLED LAST"
	HeaderKeyHealthy       = "HEALTHY"
	HeaderKeyHealthyLast   = "HEALTHY LAST"
	HeaderKeyState         = "STATE"

	HeaderKeyGroup      = "GROUP"
	HeaderKeySynced     = "SYNCED"
	HeaderKeySyncedLast = "SYNCED LAST"
	HeaderKeyReady      = "READY"
	HeaderKeyReadyLast  = "READY LAST"

	HeaderKeyStatus = "STATUS"
)

type Pane string

const (
	PaneIrrecoverableError Pane = "error"
	PaneTree               Pane = "tree"
	PaneSummary            Pane = "summary"
)

type Tracer interface {
	GetTrace() (*xplane.Resource, error)
}

type Model struct {
	keyMap        KeyMap
	navigator     navigator.Model
	viewer        viewer.Model
	tracer        Tracer
	width         int
	height        int
	short         bool
	watch         bool
	watchInterval time.Duration
	logger        *slog.Logger

	pane Pane
	err  error

	kind schema.GroupKind
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
	treeModel navigator.Model,
	viewerModel viewer.Model,
	tracer Tracer,
	opts ...WithOpt,
) *Model {
	m := &Model{
		keyMap:        DefaultKeyMap(),
		logger:        logger,
		navigator:     treeModel,
		viewer:        viewerModel,
		tracer:        tracer,
		width:         0,
		height:        0,
		watchInterval: 10 * time.Second,
		short:         true,

		pane: PaneTree,
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
	switch m.pane {
	case PaneIrrecoverableError:
		return fmt.Sprintf("There was a fatal error: %s\nPress q to exit", m.err.Error())
	case PaneSummary:
		return m.viewer.View()
	case PaneTree:
		return lipgloss.JoinVertical(
			lipgloss.Left,
			m.navigator.View(),
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
	case ShortPkgColumnLayout:
		return []table.Column{
			{Title: HeaderKeyObject, Width: 60},
			{Title: HeaderKeyVersion, Width: 8},
			{Title: HeaderKeyInstalled, Width: 8},
			{Title: HeaderKeyHealthy, Width: 7},
			{Title: HeaderKeyState, Width: 7},
			{Title: HeaderKeyStatus, Width: 68},
		}
	case WidePkgColumnLayout:
		return []table.Column{
			{Title: HeaderKeyObject, Width: 60},
			{Title: HeaderKeyVersion, Width: 8},
			{Title: HeaderKeyInstalled, Width: 10},
			{Title: HeaderKeyInstalledLast, Width: 19},
			{Title: HeaderKeyHealthy, Width: 7},
			{Title: HeaderKeyHealthyLast, Width: 19},
			{Title: HeaderKeyState, Width: 7},
			{Title: HeaderKeyStatus, Width: 68},
		}
	default:
		return []table.Column{}
	}
}

func (m Model) getLayout(gk schema.GroupKind) ColumnLayout {
	isPkg := xplane.IsPkg(gk)
	isRes, isWide, isShort := !isPkg, !m.short, m.short

	switch {
	case isPkg && isShort:
		return ShortPkgColumnLayout
	case isPkg && isWide:
		return WidePkgColumnLayout
	case isRes && isShort:
		return ShortObjectColumnLayout
	case isRes && isWide:
		return WideObjectColumnLayout
	default:
		return UnknownColumnLayout
	}
}

func (m *Model) setColumns(gk schema.GroupKind) {
	m.navigator.SetColumns(m.getColumns(m.getLayout(gk)))
}

func (m *Model) setNodes(data *xplane.Resource) {
	rows := []navigator.DataRow{}
	m.kind = data.Unstructured.GroupVersionKind().GroupKind()
	m.traceToRows(data, &rows, 0)
	m.navigator.SetData(rows)
}

func (m *Model) setIrrecoverableError(err error) {
	m.err = err
	m.pane = PaneIrrecoverableError
}

func (m Model) traceToRows(v *xplane.Resource, rows *[]navigator.DataRow, depth int) {
	const treeNodePrefix string = " └─"

	label := fmt.Sprintf("%s/%s", v.Unstructured.GetKind(), v.Unstructured.GetName())
	group := v.Unstructured.GetObjectKind().GroupVersionKind().Group
	row := navigator.DataRow{
		ID:      fmt.Sprintf("%s.%s/%s", v.Unstructured.GetKind(), group, v.Unstructured.GetName()),
		Data:    v,
		Columns: []string{},
	}

	if depth > 0 {
		shape := strings.Repeat(" ", (depth-1)) + treeNodePrefix + " "
		label = shape + label
	}

	if v.Unstructured.GetAnnotations()["crossplane.io/paused"] == "true" {
		label += " (paused)"
		row.Color = lipgloss.ANSIColor(ansi.Yellow)
	}

	data := map[string]string{}
	if xplane.IsPkg(m.kind) {
		resStatus := xplane.GetPkgResourceStatus(v, label)
		data = map[string]string{
			HeaderKeyObject:        label,
			HeaderKeyGroup:         group,
			HeaderKeyVersion:       resStatus.Version,
			HeaderKeyInstalled:     resStatus.Installed,
			HeaderKeyInstalledLast: getTimeStr(resStatus.InstalledLastTransition),
			HeaderKeyHealthy:       resStatus.Healthy,
			HeaderKeyHealthyLast:   getTimeStr(resStatus.HealthyLastTransition),
			HeaderKeyState:         resStatus.State,
			HeaderKeyStatus:        resStatus.Status,
		}
		if !resStatus.Ok {
			row.Color = lipgloss.ANSIColor(ansi.Red)
		}
	} else {
		resStatus := xplane.GetResourceStatus(v, label)
		data = map[string]string{
			HeaderKeyObject:     label,
			HeaderKeyGroup:      group,
			HeaderKeySynced:     resStatus.Synced,
			HeaderKeySyncedLast: getTimeStr(resStatus.SyncedLastTransition),
			HeaderKeyReady:      resStatus.Ready,
			HeaderKeyReadyLast:  getTimeStr(resStatus.ReadyLastTransition),
			HeaderKeyStatus:     resStatus.Status,
		}
		if !resStatus.Ok {
			row.Color = lipgloss.ANSIColor(ansi.Red)
		}
	}

	for _, col := range m.getColumns(m.getLayout(m.kind)) {
		row.Columns = append(row.Columns, data[col.Title])
	}
	*rows = append(*rows, row)

	for _, cv := range v.Children {
		m.traceToRows(cv, rows, depth+1)
	}
}

func getTimeStr(t time.Time) string {
	if t.IsZero() {
		return "-"
	}
	return t.Format(time.RFC822)
}
