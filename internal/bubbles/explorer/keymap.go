package explorer

import "github.com/charmbracelet/bubbles/key"

type KeyMap struct {
	Quit key.Binding
}

// DefaultKeyMap returns a default set of keybindings.
func DefaultKeyMap() KeyMap {
	return KeyMap{
		Quit: key.NewBinding(
			key.WithKeys("ctrl+c", "ctrl+d"),
			key.WithHelp("q/esc", "quit"),
		),
	}
}
