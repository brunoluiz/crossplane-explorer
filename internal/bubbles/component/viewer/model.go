package viewer

import (
	"fmt"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Model struct {
	KeyMap KeyMap

	title     string
	sideTitle string
	content   string

	cmdQuit tea.Cmd
	Styles  Styles

	ready    bool
	height   int
	viewport viewport.Model

	searchInput     textinput.Model
	searchMode      searchMode
	searchResult    string
	searchCursor    int
	searchResultPos []int
}

type searchMode int

const (
	searchModeOff searchMode = iota
	searchModeInit
	searchModeInput
	searchModeFilter
)

type WithOpt func(*Model)

func WithQuitCmd(c tea.Cmd) func(m *Model) {
	return func(m *Model) {
		m.cmdQuit = c
	}
}

func New(opts ...WithOpt) Model {
	ti := textinput.New()
	ti.Prompt = "🔍 "
	ti.Placeholder = "Search..."

	m := Model{
		KeyMap:          DefaultKeyMap(),
		Styles:          DefaultStyles(),
		cmdQuit:         func() tea.Msg { return EventQuit{} },
		searchInput:     ti,
		searchMode:      searchModeOff,
		searchResult:    "",
		searchCursor:    0,
		searchResultPos: []int{},
	}

	for _, opt := range opts {
		opt(&m)
	}

	return m
}

func (m Model) Init() tea.Cmd { return nil }

func (m Model) GetWidth() int {
	w := m.viewport.Width
	borderLeftW := m.Styles.Viewport.GetBorderLeftSize()
	borderRightW := m.Styles.Viewport.GetBorderRightSize()
	return w - borderLeftW - borderRightW
}

func (m Model) View() string {
	if !m.ready {
		return "\n  Initializing..."
	}

	var components []string

	switch m.searchMode {
	case searchModeInit:
		fallthrough
	case searchModeInput:
		searchBar := lipgloss.NewStyle().Render(m.searchInput.View())
		components = append(components, searchBar)
	case searchModeFilter:
		filterBar := lipgloss.NewStyle().Render(fmt.Sprintf("🔍 Showing results for: %s", m.searchInput.Value()))
		components = append(components, filterBar)
	}

	header := m.headerView()
	footer := m.footerView()

	components = append([]string{header, m.viewport.View(), footer}, components...)

	return lipgloss.JoinVertical(lipgloss.Left, components...)
}

type ContentInput struct {
	Title     string
	SideTitle string
	Content   string
}

func (m *Model) setContent(val string) {
	m.viewport.SetContent(
		lipgloss.NewStyle().Width(m.GetWidth()).Render(val),
	)
	m.viewport.GotoTop()
}

func (m *Model) SetContent(msg ContentInput) {
	m.title = msg.Title
	m.sideTitle = msg.SideTitle
	m.content = msg.Content

	m.setContent(m.content)
}

func (m Model) headerView() string {
	return lipgloss.JoinHorizontal(
		lipgloss.Left,
		m.Styles.Title.Render(m.title),
		m.Styles.SideTitle.Render(m.sideTitle),
	)
}

func (m Model) footerView() string {
	return lipgloss.JoinHorizontal(
		lipgloss.Right,
		m.Styles.Footer.Render(fmt.Sprintf("%3.f%%", m.viewport.ScrollPercent()*100)),
	)
}
