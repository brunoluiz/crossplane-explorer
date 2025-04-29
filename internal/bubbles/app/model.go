package app

import (
	"fmt"
	"log/slog"

	navigatorpane "github.com/brunoluiz/crossplane-explorer/internal/bubbles/layout/navigator"
	viewerpane "github.com/brunoluiz/crossplane-explorer/internal/bubbles/layout/viewer"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Pane string

const (
	PaneIrrecoverableError Pane = "error"
	PaneNavigator          Pane = "tree"
	PaneViewer             Pane = "summary"
)

type Model struct {
	keyMap    KeyMap
	viewer    viewerpane.Model
	navigator navigatorpane.Model
	width     int
	height    int
	logger    *slog.Logger

	pane Pane
	err  error
}

type WithOpt func(*Model)

func New(
	logger *slog.Logger,
	navigatorModel navigatorpane.Model,
	viewerModel viewerpane.Model,
	opts ...WithOpt,
) *Model {
	m := &Model{
		keyMap:    DefaultKeyMap(),
		logger:    logger,
		navigator: navigatorModel,
		viewer:    viewerModel,
		width:     0,
		height:    0,
		pane:      PaneNavigator,
	}

	for _, opt := range opts {
		opt(m)
	}

	return m
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(
		m.navigator.Init(),
		m.viewer.Init(),
	)
}

func (m Model) View() string {
	switch m.pane {
	case PaneIrrecoverableError:
		return fmt.Sprintf("There was a fatal error: %s\nPress q to exit", m.err.Error())
	case PaneNavigator:
		return lipgloss.JoinVertical(
			lipgloss.Left,
			m.navigator.View(),
		)
	case PaneViewer:
		return m.viewer.View()
	default:
		return "No pane selected"
	}
}

type ColumnLayout int

func (m *Model) setIrrecoverableError(err error) {
	m.err = err
	m.pane = PaneIrrecoverableError
}
