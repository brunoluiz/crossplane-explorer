package navigator

import (
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/brunoluiz/crossplane-explorer/internal/bubbles/shared/navigator/statusbar"
	"github.com/brunoluiz/crossplane-explorer/internal/bubbles/shared/table"

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
	tree := m.renderTable(availableHeight)

	return lipgloss.JoinVertical(lipgloss.Left, append([]string{tree}, components...)...)
}

func (m *Model) SetData(data []TemporaryGlue) {
	m.data = data
	m.table.Focus()
}

func (m *Model) renderTable(height int) string {
	rows := []table.Row{}
	searchTerm := strings.ToLower(m.searchInput.Value())
	for k, v := range m.data {
		s := lipgloss.NewStyle()
		if m.cursor != k {
			s = s.Foreground(v.Color)
		}

		if m.searchMode == searchModeFilter && strings.Contains(strings.ToLower(v.ID), searchTerm) {
			s = s.Bold(true).Foreground(lipgloss.Color("15")).Background(lipgloss.Color("1")) // White on red
		}

		cols := []table.Cell{}
		for _, col := range v.Columns {
			cols = append(cols, table.Cell{Value: col, Style: s})
		}

		rows = append(rows, cols)
	}

	m.table.SetHeight(height)
	m.table.SetRows(rows)
	return m.table.View()
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

func (m Model) Current() *TemporaryGlue    { return &m.data[m.cursor] }
func (m *Model) setSize(width, height int) { m.width = width; m.height = height }

func (m Model) helpView() string {
	m.Help.ShowAll = false
	return m.Styles.Help.Render(m.Help.View(m))
}
