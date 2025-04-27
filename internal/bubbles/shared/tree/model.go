package tree

import (
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/brunoluiz/crossplane-explorer/internal/bubbles/shared/table"
	"github.com/brunoluiz/crossplane-explorer/internal/bubbles/shared/tree/statusbar"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type searchMode int

type TemporaryGlue struct {
	ID   string
	Data any

	Columns []string
	Color   lipgloss.TerminalColor
}

const (
	searchModeOff searchMode = iota
	searchModeInit
	searchModeInput
	searchModeFilter
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

	Color lipgloss.TerminalColor

	Children []Node
}

type Model struct {
	KeyMap      KeyMap
	Styles      Styles
	Help        help.Model
	table       table.Model
	searchInput textinput.Model
	statusbar   statusbar.Model
	logger      *slog.Logger

	width         int
	height        int
	nodes         []Node
	nodesByCursor map[int]*Node
	pathByNode    map[*Node][]string
	cursor        int

	showHelp     bool
	searchMode   searchMode
	searchResult string

	data []TemporaryGlue
}

func New(
	logger *slog.Logger,
	tableModel table.Model,
	searchInputModel textinput.Model,
	statusBarModel statusbar.Model,
) Model {
	searchInputModel.Prompt = "🔍 "
	searchInputModel.Placeholder = "Search..."
	return Model{
		logger:      logger,
		table:       tableModel,
		searchInput: searchInputModel,
		statusbar:   statusBarModel,
		KeyMap:      DefaultKeyMap(),
		Styles:      DefaultStyles(),

		width:         0,
		height:        0,
		nodesByCursor: map[int]*Node{},
		pathByNode:    map[*Node][]string{},

		showHelp: false,
		Help:     help.New(),
	}
}

func (m *Model) Init() tea.Cmd {
	return nil
}

func (m Model) View() string {
	components := []string{}
	availableHeight := m.height

	if m.showHelp {
		help := m.helpView()
		availableHeight -= lipgloss.Height(help)
		components = append(components, help)
	}

	switch m.searchMode {
	case searchModeInit:
		fallthrough
	case searchModeInput:
		searchBar := lipgloss.NewStyle().Render(m.searchInput.View())
		availableHeight -= lipgloss.Height(searchBar)
		components = append(components, searchBar)
	case searchModeFilter:
		filterBar := lipgloss.NewStyle().Render(fmt.Sprintf("🔍 Showing results for: %s", m.searchInput.Value()))
		availableHeight -= lipgloss.Height(filterBar)
		components = append(components, filterBar)
	}

	components = append(components, m.statusbar.View())

	m.loadTable(m.nodes)
	m.table.SetHeight(availableHeight)
	tree := m.table.View()

	return lipgloss.JoinVertical(lipgloss.Left, append([]string{tree}, components...)...)
}

func (m *Model) SetData(data []TemporaryGlue) {
	m.data = data
}

func (m *Model) SetNodes(nodes []Node) {
	m.nodes = nodes

	m.loadTable(m.nodes)
	m.table.Focus()

	// Set the path to the first item, since it will only render further values on cursor change
	if m.cursor == 0 {
		m.statusbar.SetPath([]string{nodes[0].Label})
	}
}

func (m *Model) loadTable(nodes []Node) []table.Row {
	count := 0 // This is used to keep track of the index of the node we are on (important because we are using a recursive function)
	rows := []table.Row{}
	m.renderTree(&rows, nodes, []string{}, 0, &count)
	m.table.SetRows(rows)
	return rows
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
		m.KeyMap.Copy,
		m.KeyMap.Show,
		m.KeyMap.Search,
		m.KeyMap.Help,
	}

	return append(kb,
		m.KeyMap.Quit,
	)
}

func (m Model) FullHelp() [][]key.Binding {
	kb := [][]key.Binding{{
		m.KeyMap.Up,
		m.KeyMap.Down,
		m.KeyMap.Copy,
		m.KeyMap.Show,
	}}

	return append(kb,
		[]key.Binding{
			m.KeyMap.Quit,
			m.KeyMap.CloseFullHelp,
		})
}

func (m Model) Current() Node              { return *m.nodesByCursor[m.cursor] }
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
	searchTerm := strings.ToLower(m.searchInput.Value())

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

		if m.searchMode == searchModeFilter && strings.Contains(strings.ToLower(node.Key), searchTerm) {
			s = s.Bold(true).Foreground(lipgloss.Color("15")).Background(lipgloss.Color("1")) // White on red
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
	m.Help.ShowAll = false
	return m.Styles.Help.Render(m.Help.View(m))
}
