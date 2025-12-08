package tui

import "github.com/charmbracelet/bubbles/key"

type KeyMap struct {
	NextItem   key.Binding
	PrevItem   key.Binding
	Select     key.Binding
	Cancel     key.Binding
	Backspace  key.Binding
	ClearQuery key.Binding
}

func DefaultKeyMap() KeyMap {
	return KeyMap{
		NextItem: key.NewBinding(
			key.WithKeys("down", "ctrl+n", "ctrl+j"),
			key.WithHelp("ctrl+n", "next"),
		),
		PrevItem: key.NewBinding(
			key.WithKeys("up", "ctrl+p", "ctrl+k"),
			key.WithHelp("ctrl+p", "prev"),
		),
		Select: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "select"),
		),
		Cancel: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "cancel"),
		),
		Backspace: key.NewBinding(
			key.WithKeys("backspace"),
			key.WithHelp("backspace", "delete"),
		),
		ClearQuery: key.NewBinding(
			key.WithKeys("ctrl+c"),
			key.WithHelp("ctrl+c", "clear query"),
		),
	}
}
