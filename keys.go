package main

import (
	"github.com/charmbracelet/bubbles/key"
)

type KeyMap struct {
	Quit                key.Binding
	Up                  key.Binding
	Down                key.Binding
	PageUp              key.Binding
	PageDown            key.Binding
	JumpToFirstRow      key.Binding
	JumpToLastRow       key.Binding
	JumpToTopLeft       key.Binding
	JumpToBottomRight   key.Binding
	Left                key.Binding
	Right               key.Binding
	NextTrack           key.Binding
	PrevTrack           key.Binding
	InsertTrack         key.Binding
	DeleteTrack         key.Binding
	IncBrushWidth       key.Binding
	DecBrushWidth       key.Binding
	IncBrushHeight      key.Binding
	DecBrushHeight      key.Binding
	IncSelectionWidth   key.Binding
	DecSelectionWidth   key.Binding
	IncSelectionHeight  key.Binding
	DecSelectionHeight  key.Binding
	InsertBlock         key.Binding
	DeleteBlock         key.Binding
	ZeroBlock           key.Binding
	BackspaceBlock      key.Binding
	PlayOrStop          key.Binding
	Cut                 key.Binding
	Copy                key.Binding
	Paste               key.Binding
	SetPlayFromRow      key.Binding
	EnterCommandMode    key.Binding
	EnterNoteMode       key.Binding
	ToggleChromaticMode key.Binding
	Undo                key.Binding
	Redo                key.Binding
	Save                key.Binding
}

var defaultKeyMap = KeyMap{
	Quit: key.NewBinding(
		key.WithKeys("ctrl+q"),
		key.WithHelp("C-q", "quit"),
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
		key.WithKeys("home"),
		key.WithHelp("home", "first row"),
	),
	JumpToLastRow: key.NewBinding(
		key.WithKeys("end"),
		key.WithHelp("end", "last row"),
	),
	JumpToTopLeft: key.NewBinding(
		key.WithKeys("ctrl+home"),
		key.WithHelp("C-home", "top left"),
	),
	JumpToBottomRight: key.NewBinding(
		key.WithKeys("ctrl+end"),
		key.WithHelp("C-end", "bottom right"),
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
	InsertTrack: key.NewBinding(
		key.WithKeys("ctrl+shift+right"),
		key.WithHelp("C-S-right", "insert track"),
	),
	DeleteTrack: key.NewBinding(
		key.WithKeys("ctrl+shift+left"),
		key.WithHelp("C-S-left", "delete track"),
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
	IncSelectionWidth: key.NewBinding(
		key.WithKeys("shift+right"),
		key.WithHelp("S-right", "increase selection width"),
	),
	DecSelectionWidth: key.NewBinding(
		key.WithKeys("shift+left"),
		key.WithHelp("S-left", "decrease selection width"),
	),
	IncSelectionHeight: key.NewBinding(
		key.WithKeys("shift+down"),
		key.WithHelp("S-down", "increase selection height"),
	),
	DecSelectionHeight: key.NewBinding(
		key.WithKeys("shift+up"),
		key.WithHelp("S-up", "decrease selection height"),
	),
	InsertBlock: key.NewBinding(
		key.WithKeys("insert", "ctrl+shift+down"),
		key.WithHelp("ins/C-S-down", "insert block"),
	),
	DeleteBlock: key.NewBinding(
		key.WithKeys("delete", "ctrl+shift+up"),
		key.WithHelp("del/C-S-up", "delete block"),
	),
	ZeroBlock: key.NewBinding(
		key.WithKeys("."),
		key.WithHelp(".", "fill selection with zeroes"),
	),
	BackspaceBlock: key.NewBinding(
		key.WithKeys("backspace"),
		key.WithHelp("backspace", "delete block above"),
	),
	Cut: key.NewBinding(
		key.WithKeys("ctrl+x"),
		key.WithHelp("C-x", "cut block"),
	),
	Copy: key.NewBinding(
		key.WithKeys("ctrl+c"),
		key.WithHelp("C-c", "copy block"),
	),
	Paste: key.NewBinding(
		key.WithKeys("ctrl+v"),
		key.WithHelp("C-v", "paste block"),
	),
	PlayOrStop: key.NewBinding(
		key.WithKeys(" "),
		key.WithHelp(" ", "play/stop"),
	),
	SetPlayFromRow: key.NewBinding(
		key.WithKeys("s"),
		key.WithHelp("s", "set play from row"),
	),
	EnterCommandMode: key.NewBinding(
		key.WithKeys(":"),
		key.WithHelp(":", "enter command"),
	),
	EnterNoteMode: key.NewBinding(
		key.WithKeys("ctrl+n"),
		key.WithHelp("C-n", "note mode"),
	),
	ToggleChromaticMode: key.NewBinding(
		key.WithKeys("N"),
		key.WithHelp("S-n", "toggle chromatic mode"),
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
