package viewer

import "github.com/charmbracelet/bubbles/key"

type KeyMap struct {
	PageUp   key.Binding
	PageDown key.Binding
	Quit     key.Binding
}

// DefaultKeyMap returns a default set of keybindings.
func DefaultKeyMap() KeyMap {
	return KeyMap{
		PageUp: key.NewBinding(
			key.WithKeys("ctrl+b"),
			key.WithHelp("ctrl+b", "page up"),
		),
		PageDown: key.NewBinding(
			key.WithKeys("ctrl+f"),
			key.WithHelp("ctrl+f", "page down"),
		),
		Quit: key.NewBinding(
			key.WithKeys("q", "esc"),
			key.WithHelp("q/esc", "quit"),
		),
	}
}
