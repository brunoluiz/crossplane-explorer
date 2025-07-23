package viewer

import "github.com/charmbracelet/bubbles/key"

// KeyMap defines the keybindings for the viewport.
type KeyMap struct {
	LineUp        key.Binding
	LineDown      key.Binding
	PageUp        key.Binding
	PageDown      key.Binding
	HalfPageUp    key.Binding
	HalfPageDown  key.Binding
	GotoTop       key.Binding
	GotoBottom    key.Binding
	Help          key.Binding
	CloseFullHelp key.Binding
	Quit          key.Binding

	Search         key.Binding
	SearchConfirm  key.Binding
	SearchQuit     key.Binding
	SearchNext     key.Binding
	SearchPrevious key.Binding
}

// DefaultKeyMap returns a set of default keybindings.
func DefaultKeyMap() KeyMap {
	return KeyMap{
		LineUp: key.NewBinding(
			key.WithKeys("up", "k"),
			key.WithHelp("↑/k", "scroll up"),
		),
		LineDown: key.NewBinding(
			key.WithKeys("down", "j"),
			key.WithHelp("↓/j", "scroll down"),
		),
		PageUp: key.NewBinding(
			key.WithKeys("pgup", "ctrl+b"),
			key.WithHelp("b/pgup", "page up"),
		),
		PageDown: key.NewBinding(
			key.WithKeys("pgdown", " ", "ctrl+f"),
			key.WithHelp("f/pgdn", "page down"),
		),
		HalfPageUp: key.NewBinding(
			key.WithKeys("ctrl+u"),
			key.WithHelp("u", "½ page up"),
		),
		HalfPageDown: key.NewBinding(
			key.WithKeys("ctrl+d"),
			key.WithHelp("d", "½ page down"),
		),
		GotoTop: key.NewBinding(
			key.WithKeys("g", "home"),
			key.WithHelp("g/home", "go to top"),
		),
		GotoBottom: key.NewBinding(
			key.WithKeys("G", "end"),
			key.WithHelp("G/end", "go to bottom"),
		),
		Help: key.NewBinding(
			key.WithKeys("?", "h"),
			key.WithHelp("?", "toggle help"),
		),
		CloseFullHelp: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "close help"),
		),
		Quit: key.NewBinding(
			key.WithKeys("q", "esc"),
			key.WithHelp("q", "quit"),
		),
		Search: key.NewBinding(
			key.WithKeys("/"),
			key.WithHelp("/", "search"),
		),
		SearchConfirm: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "confirm search"),
		),
		SearchQuit: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "quit search"),
		),
		SearchNext: key.NewBinding(
			key.WithKeys("n"),
			key.WithHelp("n", "next search"),
		),
		SearchPrevious: key.NewBinding(
			key.WithKeys("N"),
			key.WithHelp("N", "previous search"),
		),
	}
}

// ShortHelp returns keybindings to show in the short help view. It doesn't
// include less frequently used keybindings.
func (k KeyMap) ShortHelp() []key.Binding {
	return []key.Binding{
		k.Help,
		k.Quit,
		k.Search,
	}
}

// FullHelp returns keybindings for the expanded help view.
func (k KeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.LineUp, k.LineDown, k.GotoTop, k.GotoBottom},
		{k.PageUp, k.PageDown, k.HalfPageUp, k.HalfPageDown},
		{k.Search, k.SearchConfirm, k.SearchQuit},
		{k.SearchNext, k.SearchPrevious},
		{k.Help, k.Quit},
	}
}
