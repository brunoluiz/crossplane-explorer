package tree

import "github.com/charmbracelet/lipgloss"

type Styles struct {
	Help lipgloss.Style
}

func DefaultStyles() Styles {
	return Styles{
		Help: lipgloss.NewStyle().Margin(0, 0, 0, 0).Foreground(lipgloss.AdaptiveColor{Light: "#333", Dark: "#eee"}),
	}
}
