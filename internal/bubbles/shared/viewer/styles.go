package viewer

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
)

type Styles struct {
	Title     lipgloss.Style
	SideTitle lipgloss.Style
	Viewport  lipgloss.Style
	Footer    lipgloss.Style
}

func DefaultStyles() Styles {
	return Styles{
		Title: lipgloss.NewStyle().
			Bold(true).
			Background(lipgloss.ANSIColor(ansi.BrightBlack)).
			Foreground(lipgloss.ANSIColor(ansi.White)).
			Padding(0, 1, 0, 1).
			Margin(1, 0, 0, 1),
		SideTitle: lipgloss.NewStyle().
			Bold(true).
			Background(lipgloss.ANSIColor(ansi.Green)).
			Foreground(lipgloss.ANSIColor(ansi.Black)).
			Padding(0, 1, 0, 1).
			Margin(1, 0, 0, 1),
		Viewport: lipgloss.NewStyle().
			// Border(lipgloss.NormalBorder(), true, true, true, true).
			Margin(1, 0, 0, 1).
			Padding(0, 1, 0, 1),
		Footer: lipgloss.NewStyle().
			Padding(0, 1, 0, 1),
	}
}
