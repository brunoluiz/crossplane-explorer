package statusbar

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
	"github.com/mistakenelf/teacup/statusbar"
)

type Model struct {
	statusbar      statusbar.Model
	path           []string
	pathSeparator  string
	rootSymbol     string
	primaryColor   statusbar.ColorConfig
	secondaryColor statusbar.ColorConfig
	neutralColor   statusbar.ColorConfig
}

type config struct {
	path           []string
	pathSeparator  string
	rootSymbol     string
	primaryColor   statusbar.ColorConfig
	secondaryColor statusbar.ColorConfig
	neutralColor   statusbar.ColorConfig
}

type WithOpt func(*config)

func WithPrimaryStatusColor(cl statusbar.ColorConfig) func(c *config) {
	return func(c *config) { c.primaryColor = cl }
}

func WithSecondaryStatusColor(cl statusbar.ColorConfig) func(c *config) {
	return func(c *config) { c.secondaryColor = cl }
}

func WithNeutralStatusColor(cl statusbar.ColorConfig) func(c *config) {
	return func(c *config) { c.neutralColor = cl }
}

func WithPathSeparator(p string) func(c *config) {
	return func(c *config) { c.pathSeparator = p }
}

func WithInitialPath(p []string) func(c *config) {
	return func(c *config) { c.path = p }
}

func itoa(c ansi.BasicColor) string {
	return fmt.Sprintf("%d", c)
}

func New(opts ...WithOpt) Model {
	cfg := config{
		path:          []string{},
		pathSeparator: "\ueab6 ",
		primaryColor: statusbar.ColorConfig{
			Foreground: lipgloss.AdaptiveColor{Dark: itoa(ansi.White), Light: itoa(ansi.White)},
			Background: lipgloss.AdaptiveColor{Light: itoa(ansi.Magenta), Dark: itoa(ansi.Magenta)},
		},
		secondaryColor: statusbar.ColorConfig{
			Foreground: lipgloss.AdaptiveColor{Dark: itoa(ansi.Black), Light: itoa(ansi.Black)},
			Background: lipgloss.AdaptiveColor{Light: itoa(ansi.Blue), Dark: itoa(ansi.Blue)},
		},
		neutralColor: statusbar.ColorConfig{
			Foreground: lipgloss.AdaptiveColor{Dark: itoa(ansi.White), Light: itoa(ansi.White)},
			Background: lipgloss.AdaptiveColor{Light: itoa(ansi.BrightBlack), Dark: itoa(ansi.BrightBlack)},
		},
	}
	for _, opt := range opts {
		opt(&cfg)
	}

	s := statusbar.New(
		cfg.primaryColor,
		cfg.neutralColor,
		cfg.neutralColor,
		cfg.neutralColor,
	)
	s.FirstColumn = "$"
	s.SecondColumn = strings.Join(cfg.path, cfg.pathSeparator)

	return Model{
		path:           cfg.path,
		pathSeparator:  cfg.pathSeparator,
		rootSymbol:     cfg.rootSymbol,
		statusbar:      s,
		primaryColor:   cfg.primaryColor,
		secondaryColor: cfg.secondaryColor,
		neutralColor:   cfg.neutralColor,
	}
}

func (m Model) Init() tea.Cmd { return nil }
func (m *Model) View() string { return m.statusbar.View() }

func (m *Model) GetHeight() int { return statusbar.Height }

func (m *Model) SetPath(path []string) {
	m.path = path
	m.statusbar.SecondColumn = strings.Join(m.path, m.pathSeparator)
}
