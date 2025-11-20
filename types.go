package main

import (
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type (
	MidiMessage [3]byte
	Row         []MidiMessage
)

type Pattern struct {
	Rows          []Row `json:"rows"`
	NumRows       int   `json:"numRows"`
	NumTracks     int   `json:"numTracks"`
	TrackDefaults Row   `json:"trackDefaults"`
}

type Song struct {
	BPM       int        `json:"bpm"` // beats per minute
	LPB       int        `json:"lpb"` // lines per beat
	TPL       int        `json:"tpl"` // ticks per line
	Patterns  []*Pattern `json:"patterns"`
	Root      int        `json:"root"`      // root note
	Scale     ScaleId    `json:"scale"`     // scale id
	Mode      int        `json:"mode"`      // offset of degree 0 within the scale
	Chromatic bool       `json:"chromatic"` // note mode uses chromatic scale?
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

type Brush struct {
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
	SelectMode  Mode = 1
	NoteMode    Mode = 2
	CommandMode Mode = 3
)

type Model struct {
	err                 error
	keymap              *KeyMap
	mode                Mode
	prevModes           []Mode
	windowSize          Size
	midiEngine          *MidiEngine
	song                *Song
	brush               Brush
	sel                 Rect
	editPattern         int
	editPos             Point
	firstVisibleRow     int
	firstVisibleTrack   int
	playPattern         int
	playRow             int
	playTick            int
	playFrame           uint64
	isPlaying           bool
	playFromRow         int
	commandModel        textinput.Model
	filename            string
	pendingActions      chan Action
	pendingMidiMessages chan MidiMessage
	msgs                chan tea.Msg
	undoableActions     []Action
	undoneActions       []Action
	clipboard           Block
	pasteOffset         int
	usingTempBrush      bool
}

type (
	redrawMsg struct{}
)
