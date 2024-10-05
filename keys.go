package main

import (
	"github.com/charmbracelet/bubbles/key"
)

type KeyMap struct {
	Quit             key.Binding
	Up               key.Binding
	Down             key.Binding
	PageUp           key.Binding
	PageDown         key.Binding
	JumpToFirstRow   key.Binding
	JumpToLastRow    key.Binding
	Left             key.Binding
	Right            key.Binding
	NextTrack        key.Binding
	PrevTrack        key.Binding
	JumpToFirstTrack key.Binding
	JumpToLastTrack  key.Binding
	DeleteLeft       key.Binding
	DeleteUnder      key.Binding
	InsertBlank      key.Binding
	PlayOrStop       key.Binding
	SetStartRow      key.Binding
	EnterCommand     key.Binding
	Undo             key.Binding
	Redo             key.Binding
	Save             key.Binding
}

var defaultKeyMap = KeyMap{
	Quit: key.NewBinding(
		key.WithKeys("ctrl+c", "ctrl+q"),
		key.WithHelp("C-c|C-q", "quit"),
	),
	Up: key.NewBinding(
		key.WithKeys("up"),
		key.WithHelp("up", "up"),
	),
	Down: key.NewBinding(
		key.WithKeys("down"),
		key.WithHelp("down", "down"),
	),
	PageUp: key.NewBinding(
		key.WithKeys("pgup"),
		key.WithHelp("pgup", "page up"),
	),
	PageDown: key.NewBinding(
		key.WithKeys("pgdown"),
		key.WithHelp("pgdown", "page down"),
	),
	JumpToFirstRow: key.NewBinding(
		key.WithKeys("ctrl+home"),
		key.WithHelp("C-home", "first row"),
	),
	JumpToLastRow: key.NewBinding(
		key.WithKeys("ctrl+end"),
		key.WithHelp("C-end", "last row"),
	),
	Left: key.NewBinding(
		key.WithKeys("left"),
		key.WithHelp("left", "left"),
	),
	Right: key.NewBinding(
		key.WithKeys("right"),
		key.WithHelp("right", "right"),
	),
	NextTrack: key.NewBinding(
		key.WithKeys("tab"),
		key.WithHelp("tab", "next track"),
	),
	PrevTrack: key.NewBinding(
		key.WithKeys("shift+tab"),
		key.WithHelp("S+tab", "previous track"),
	),
	JumpToFirstTrack: key.NewBinding(
		key.WithKeys("home"),
		key.WithHelp("home", "first track"),
	),
	JumpToLastTrack: key.NewBinding(
		key.WithKeys("end"),
		key.WithHelp("end", "last track"),
	),
	DeleteLeft: key.NewBinding(
		key.WithKeys("backspace"),
		key.WithHelp("backspace", "delete left"),
	),
	DeleteUnder: key.NewBinding(
		key.WithKeys("delete"),
		key.WithHelp("delete", "delete under"),
	),
	InsertBlank: key.NewBinding(
		key.WithKeys("."),
		key.WithHelp(".", "insert blank"),
	),
	PlayOrStop: key.NewBinding(
		key.WithKeys(" "),
		key.WithHelp(" ", "play/stop"),
	),
	SetStartRow: key.NewBinding(
		key.WithKeys("s", "alt+ "),
		key.WithHelp("s/M-space", "set start row"),
	),
	EnterCommand: key.NewBinding(
		key.WithKeys(":"),
		key.WithHelp(":", "enter command"),
	),
	Undo: key.NewBinding(
		key.WithKeys("ctrl+z"),
		key.WithHelp("C-z", "undo"),
	),
	Redo: key.NewBinding(
		key.WithKeys("ctrl+r"),
		key.WithHelp("C-r", "redo"),
	),
	Save: key.NewBinding(
		key.WithKeys("ctrl+s"),
		key.WithHelp("C-s", "save"),
	),
}
