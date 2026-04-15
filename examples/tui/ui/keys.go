package ui

import "github.com/charmbracelet/bubbles/key"

// Keys defines the global key bindings.
type Keys struct {
	Up       key.Binding
	Down     key.Binding
	Enter    key.Binding
	Back     key.Binding
	Quit     key.Binding
	Help     key.Binding
	Save     key.Binding
	Dismiss  key.Binding
	OpenURL  key.Binding
	CopyURL  key.Binding
	Search   key.Binding
	Logout   key.Binding
	NextPage key.Binding
	PrevPage key.Binding
	Radio    key.Binding
	Replay   key.Binding
	Prefs    key.Binding
	Tab      key.Binding
}

// DefaultKeys is the standard keybinding set.
var DefaultKeys = Keys{
	Up:       key.NewBinding(key.WithKeys("k", "up"), key.WithHelp("k/↑", "up")),
	Down:     key.NewBinding(key.WithKeys("j", "down"), key.WithHelp("j/↓", "down")),
	Enter:    key.NewBinding(key.WithKeys("enter", "l"), key.WithHelp("enter/l", "open")),
	Back:     key.NewBinding(key.WithKeys("esc", "h"), key.WithHelp("esc/h", "back")),
	Quit:     key.NewBinding(key.WithKeys("q", "ctrl+c"), key.WithHelp("q", "quit")),
	Help:     key.NewBinding(key.WithKeys("?"), key.WithHelp("?", "help")),
	Save:     key.NewBinding(key.WithKeys("s"), key.WithHelp("s", "save")),
	Dismiss:  key.NewBinding(key.WithKeys("d"), key.WithHelp("d", "dismiss")),
	OpenURL:  key.NewBinding(key.WithKeys("o"), key.WithHelp("o", "open in browser")),
	CopyURL:  key.NewBinding(key.WithKeys("c"), key.WithHelp("c", "copy share URL")),
	Search:   key.NewBinding(key.WithKeys("/"), key.WithHelp("/", "search")),
	Logout:   key.NewBinding(key.WithKeys("ctrl+l"), key.WithHelp("ctrl+l", "logout")),
	NextPage: key.NewBinding(key.WithKeys("n", "ctrl+f"), key.WithHelp("n", "next page")),
	PrevPage: key.NewBinding(key.WithKeys("p", "ctrl+b"), key.WithHelp("p", "prev page")),
	Radio:    key.NewBinding(key.WithKeys("r"), key.WithHelp("r", "radio")),
	Replay:   key.NewBinding(key.WithKeys("ctrl+r"), key.WithHelp("ctrl+r", "replay")),
	Prefs:    key.NewBinding(key.WithKeys("P"), key.WithHelp("P", "preferences")),
	Tab:      key.NewBinding(key.WithKeys("tab"), key.WithHelp("tab", "switch")),
}
