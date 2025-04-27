package viewer

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
)

type Styles struct {
	Main      lipgloss.Style
	Idented   lipgloss.Style
	OkHealth  lipgloss.Style
	BadHealth lipgloss.Style
	Metadata  lipgloss.Style
}

func DefaultStyles() Styles {
	return Styles{
		Main:      lipgloss.NewStyle().UnsetBackground().UnsetForeground(),
		Idented:   lipgloss.NewStyle().MarginLeft(2),
		OkHealth:  lipgloss.NewStyle().Bold(true).Foreground(lipgloss.ANSIColor(ansi.Green)),
		BadHealth: lipgloss.NewStyle().Bold(true).Foreground(lipgloss.ANSIColor(ansi.Red)),
		Metadata:  lipgloss.NewStyle().Bold(true),
	}
}
