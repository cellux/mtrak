package main

import (
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type (
	MidiMessage [3]byte
	Row         []MidiMessage
	Pattern     []Row
)

type Song struct {
	BPM      int       `json:"bpm"` // beats per minute
	LPB      int       `json:"lpb"` // lines per beat
	TPL      int       `json:"tpl"` // ticks per line
	Patterns []Pattern `json:"patterns"`
}

type Point struct {
	X int
	Y int
}

type Size struct {
	W int
	H int
}

type Rect struct {
	X int
	Y int
	W int
	H int
}

type Area struct {
	Rect
	ExpandDir Point
}

type Block [][]byte

type ActionFunction func()

type Action struct {
	doFn   ActionFunction
	undoFn ActionFunction
}

type Mode int

const (
	EditMode    Mode = 0
	CommandMode Mode = 1
)

type Model struct {
	err               error
	keymap            *KeyMap
	mode              Mode
	prevmode          Mode
	windowSize        Size
	me                *MidiEngine
	song              *Song
	brush             Area
	sel               Rect
	editPattern       int
	editPos           Point
	firstVisibleRow   int
	firstVisibleTrack int
	playPattern       int
	playRow           int
	playTick          int
	playFrame         uint64
	isPlaying         bool
	playFromRow       int
	commandModel      textinput.Model
	filename          string
	pendingActions    chan Action
	msgs              chan tea.Msg
	undoableActions   []Action
	undoneActions     []Action
	clipboard         Block
	pasteOffset       int
	defaultBrush      bool
}

type (
	redrawMsg struct{}
)
