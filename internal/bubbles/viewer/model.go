package viewer

import (
	"fmt"

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

	// You generally won't need this unless you're processing stuff with
	// complicated ANSI escape sequences. Turn it on if you notice flickering.
	//
	// Also keep in mind that high performance rendering only works for programs
	// that use the full size of the terminal. We're enabling that below with
	// tea.EnterAltScreen().
	useHighPerformanceRenderer bool

	ready    bool
	viewport viewport.Model
}

type WithOpt func(*Model)

func WithQuitCmd(c tea.Cmd) func(m *Model) {
	return func(m *Model) {
		m.cmdQuit = c
	}
}

func WithHighPerformanceRenderer(enabled bool) func(m *Model) {
	return func(m *Model) {
		m.useHighPerformanceRenderer = enabled
	}
}

func New(opts ...WithOpt) Model {
	m := Model{
		KeyMap:                     DefaultKeyMap(),
		Styles:                     DefaultStyles(),
		cmdQuit:                    func() tea.Msg { return EventQuit{} },
		useHighPerformanceRenderer: false,
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
	return fmt.Sprintf("%s\n%s\n%s", m.headerView(), m.viewport.View(), m.footerView())
}

type ContentInput struct {
	Title     string
	SideTitle string
	Content   string
}

func (m *Model) SetContent(msg ContentInput) {
	m.title = msg.Title
	m.sideTitle = msg.SideTitle
	m.content = msg.Content
	m.viewport.SetContent(msg.Content)
	m.viewport.GotoTop()
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
