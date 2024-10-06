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
	InsertBlank      key.Binding
	DeleteLeft       key.Binding
	IncBrushWidth    key.Binding
	DecBrushWidth    key.Binding
	IncBrushHeight   key.Binding
	DecBrushHeight   key.Binding
	InsertBlockV     key.Binding
	DeleteBlockV     key.Binding
	InsertBlockH     key.Binding
	DeleteBlockH     key.Binding
	PlayOrStop       key.Binding
	SetPlayFromRow   key.Binding
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
	InsertBlank: key.NewBinding(
		key.WithKeys("."),
		key.WithHelp(".", "insert blank"),
	),
	DeleteLeft: key.NewBinding(
		key.WithKeys("backspace"),
		key.WithHelp("backspace", "delete left"),
	),
	IncBrushWidth: key.NewBinding(
		key.WithKeys("ctrl+right"),
		key.WithHelp("C-right", "increase brush width"),
	),
	DecBrushWidth: key.NewBinding(
		key.WithKeys("ctrl+left"),
		key.WithHelp("C-left", "decrease brush width"),
	),
	IncBrushHeight: key.NewBinding(
		key.WithKeys("ctrl+down"),
		key.WithHelp("C-down", "increase brush height"),
	),
	DecBrushHeight: key.NewBinding(
		key.WithKeys("ctrl+up"),
		key.WithHelp("C-up", "decrease brush height"),
	),
	InsertBlockV: key.NewBinding(
		key.WithKeys("insert", "ctrl+shift+down"),
		key.WithHelp("ins/C-S-down", "insert vertical block"),
	),
	DeleteBlockV: key.NewBinding(
		key.WithKeys("delete", "ctrl+shift+up"),
		key.WithHelp("del/C-S-up", "delete vertical block"),
	),
	InsertBlockH: key.NewBinding(
		key.WithKeys("ctrl+shift+right"),
		key.WithHelp("C-S-right", "insert horizontal block"),
	),
	DeleteBlockH: key.NewBinding(
		key.WithKeys("ctrl+shift+left"),
		key.WithHelp("C-S-left", "delete vertical block"),
	),
	PlayOrStop: key.NewBinding(
		key.WithKeys(" "),
		key.WithHelp(" ", "play/stop"),
	),
	SetPlayFromRow: key.NewBinding(
		key.WithKeys("s", "alt+ "),
		key.WithHelp("s/M-space", "set play from row"),
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
