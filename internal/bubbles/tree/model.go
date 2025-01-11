package tree

import (
	"log/slog"
	"strings"
	"time"

	"github.com/brunoluiz/crossplane-explorer/internal/bubbles/table"
	"github.com/brunoluiz/crossplane-explorer/internal/bubbles/tree/statusbar"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type State struct {
	LastTransitionTime time.Time
	Status             string
}

type ColorConfig struct {
	Foreground lipgloss.ANSIColor
	Background lipgloss.ANSIColor
}

type Node struct {
	Key   string
	Value any

	Label   string
	Details map[string]string

	Selected ColorConfig
	Color    lipgloss.TerminalColor

	Children []Node
}

type Model struct {
	KeyMap    KeyMap
	Styles    Styles
	Help      help.Model
	table     table.Model
	statusbar statusbar.Model
	logger    *slog.Logger

	width         int
	height        int
	nodes         []Node
	nodesByCursor map[int]*Node
	pathByNode    map[*Node][]string
	cursor        int

	showHelp bool
}

func New(
	l *slog.Logger,
	t table.Model,
	s statusbar.Model,
) Model {
	return Model{
		logger:    l,
		table:     t,
		statusbar: s,
		KeyMap:    DefaultKeyMap(),
		Styles:    DefaultStyles(),

		width:         0,
		height:        0,
		nodesByCursor: map[int]*Node{},
		pathByNode:    map[*Node][]string{},

		showHelp: true,
		Help:     help.New(),
	}
}

func (m *Model) Init() tea.Cmd {
	return nil
}

func (m Model) View() string {
	availableHeight := m.height

	var help string
	if m.showHelp {
		help = m.helpView()
		availableHeight -= lipgloss.Height(help)
	}

	availableHeight -= m.statusbar.GetHeight()
	m.table.SetHeight(availableHeight)

	return lipgloss.JoinVertical(lipgloss.Left,
		lipgloss.NewStyle().Height(m.height-m.statusbar.GetHeight()).Render(m.table.View()),
		help,
		m.statusbar.View(),
	)
}

func (m *Model) SetNodes(nodes []Node) {
	m.nodes = nodes

	count := 0 // This is used to keep track of the index of the node we are on (important because we are using a recursive function)
	rows := []table.Row{}
	m.renderTree(&rows, m.nodes, []string{}, 0, &count)
	m.table.SetRows(rows)
	m.table.Focus()
}

func (m *Model) SetColumns(cc []table.Column) {
	m.table.SetColumns(cc)
	cols := m.table.Columns()

	// Adding `2` due to borders and all
	if len(cols) > 2 {
		w := 0
		for _, col := range cols[:len(cols)-1] {
			w += col.Width + 3
		}
		cols[len(cols)-1].Width = (m.width - w + 2)
	}
}

func (m Model) ShortHelp() []key.Binding {
	kb := []key.Binding{
		m.KeyMap.Up,
		m.KeyMap.Down,
		m.KeyMap.Yank,
		m.KeyMap.Describe,
	}

	return append(kb,
		m.KeyMap.Quit,
	)
}

func (m Model) FullHelp() [][]key.Binding {
	kb := [][]key.Binding{{
		m.KeyMap.Up,
		m.KeyMap.Down,
		m.KeyMap.Yank,
		m.KeyMap.Describe,
	}}

	return append(kb,
		[]key.Binding{
			m.KeyMap.Quit,
			m.KeyMap.CloseFullHelp,
		})
}

func (m Model) Current() Node              { return *m.nodesByCursor[m.cursor] }
func (m *Model) SetShowHelp() bool         { return m.showHelp }
func (m *Model) setSize(width, height int) { m.width = width; m.height = height }

func (m *Model) numberOfNodes() int {
	count := 0

	var countNodes func([]Node)
	countNodes = func(nodes []Node) {
		for _, node := range nodes {
			count++
			if node.Children != nil {
				countNodes(node.Children)
			}
		}
	}

	countNodes(m.nodes)

	return count
}

func (m *Model) renderTree(rows *[]table.Row, remainingNodes []Node, currentPath []string, indent int, count *int) {
	const treeNodePrefix string = " └─"

	for _, node := range remainingNodes {
		// If we aren't at the root, we add the arrow shape to the string
		shape := ""
		if indent > 0 {
			shape = strings.Repeat(" ", (indent-1)) + treeNodePrefix + " "
		}

		// Generate the correct index for the node
		idx := *count
		*count++

		s := lipgloss.NewStyle()
		if m.cursor != idx {
			s = s.Foreground(node.Color)
		}

		cols := []table.Cell{{Value: shape + node.Label, Style: s}}
		for _, v := range m.table.Columns()[1:] {
			cols = append(cols, table.Cell{Value: node.Details[v.Title], Style: s})
		}

		*rows = append(*rows, cols)
		m.nodesByCursor[idx] = &node

		// Used to be able to trace back the path on the tree
		path := make([]string, len(currentPath))
		copy(path, currentPath)
		path = append(path, node.Label)
		m.pathByNode[&node] = path

		if node.Children != nil {
			m.renderTree(rows, node.Children, path, indent+1, count)
		}
	}
}

func (m Model) helpView() string {
	return m.Styles.Help.Render(m.Help.View(m))
}
