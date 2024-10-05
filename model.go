package main

import (
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"math"
)

const (
	MaxUndoableActions = 64
)

type Row = []MidiMessage
type Pattern = []Row

const (
	EditMode    = 0
	CommandMode = 1
)

type Song struct {
	BPM      int       `json:"bpm"` // beats per minute
	LPB      int       `json:"lpb"` // lines per beat
	TPL      int       `json:"tpl"` // ticks per line
	Patterns []Pattern `json:"patterns"`
}

type model struct {
	err             error
	keymap          *KeyMap
	mode            int
	windowWidth     int
	windowHeight    int
	me              *MidiEngine
	song            *Song
	editPattern     int
	editRow         int
	editRow0        int
	editTrack       int
	editTrack0      int
	editColumn      int
	playPattern     int
	playRow         int
	playTick        int
	playFrame       uint64
	isPlaying       bool
	startRow        int
	commandModel    textinput.Model
	filename        string
	pendingActions  chan Action
	undoableActions []Action
	undoneActions   []Action
}

func (m *model) SetError(err error) {
	m.err = err
}

func (m *model) SetSong(song *Song) {
	m.song = song
	m.editPattern = 0
	m.editRow = 0
	m.editRow0 = 0
	m.editTrack = 0
	m.editTrack0 = 0
	m.editColumn = 0
	m.playPattern = 0
	m.playRow = 0
	m.playTick = 0
	m.startRow = 0
}

func (m *model) Play() {
	m.playTick = 0
	m.isPlaying = true
}

func (m *model) Stop() {
	m.playTick = 0
	m.isPlaying = false
}

func (m *model) QuitWithError(err error) tea.Cmd {
	m.err = err
	return tea.Quit
}

func (m *model) GetSampleRate() int {
	return int(m.me.client.GetSampleRate())
}

func (m *model) GetBeatsPerSecond() float64 {
	return float64(m.song.BPM) / 60.0
}

func (m *model) GetFramesPerBeat() int {
	sr := float64(m.GetSampleRate())
	bps := m.GetBeatsPerSecond()
	return int(math.Round(sr / bps))
}

func (m *model) GetTicksPerBeat() int {
	return m.song.TPL * m.song.LPB
}

func (m *model) GetFramesPerTick() int {
	sr := float64(m.GetSampleRate())
	bps := m.GetBeatsPerSecond()
	tpb := float64(m.GetTicksPerBeat())
	return int(math.Round(sr / bps / tpb))
}

type redrawMsg struct{}

func (m *model) Process(nframes uint32) int {
loop:
	for {
		select {
		case action := <-m.pendingActions:
			action.doFn()
			program.Send(action)
		default:
			break loop
		}
	}
	outPort := m.me.outPort
	buf := outPort.MidiClearBuffer(nframes)
	if !m.isPlaying {
		m.playFrame += uint64(nframes)
		return 0
	}
	framesPerTick := uint64(m.GetFramesPerTick())
	var midiData MidiData
	for i := range nframes {
		if m.playFrame%framesPerTick == 0 {
			if m.playTick == 0 {
				p := m.song.Patterns[m.playPattern]
				row := p[m.playRow]
				for _, msg := range row {
					status := msg[0]
					if status >= 0x80 {
						midiData.Time = i
						midiData.Buffer = msg[0:MidiMessageLength(status)]
						outPort.MidiEventWrite(&midiData, buf)
					}
				}
				m.playRow++
				if m.playRow == len(p) {
					m.playRow = 0
				}
				program.Send(redrawMsg{})
			}
			m.playTick++
			if m.playTick == m.song.TPL {
				m.playTick = 0
			}
		}
		m.playFrame++
	}
	return 0
}

func makePattern(rowCount, columnCount int) Pattern {
	rows := make([]Row, rowCount)
	for i := range rowCount {
		rows[i] = make([]MidiMessage, columnCount)
	}
	return rows
}

func (m *model) Init() tea.Cmd {
	m.pendingActions = make(chan Action, 64)
	m.keymap = &defaultKeyMap
	m.commandModel = textinput.New()
	m.song = &Song{
		BPM:      120,
		LPB:      4,
		TPL:      6,
		Patterns: make([]Pattern, 1, 256),
	}
	m.song.Patterns[0] = makePattern(64, 16)
	m.me = &MidiEngine{}
	if err := m.me.Open(m.Process); err != nil {
		return m.QuitWithError(err)
	}
	return nil
}

func (m *model) getByte() byte {
	p := m.song.Patterns[m.editPattern]
	row := p[m.editRow]
	msg := &row[m.editTrack]
	switch m.editColumn {
	case 0:
		return msg[0] & 0xf0 >> 4
	case 1:
		return msg[0] & 0x0f
	case 2:
		return msg[1] & 0xf0 >> 4
	case 3:
		return msg[1] & 0x0f
	case 4:
		return msg[2] & 0xf0 >> 4
	case 5:
		return msg[2] & 0x0f
	}
	return 0
}

func (m *model) setByte(b byte) {
	p := m.song.Patterns[m.editPattern]
	row := p[m.editRow]
	msg := &row[m.editTrack]
	switch m.editColumn {
	case 0:
		msg[0] = msg[0]&0x0f | (b << 4)
	case 1:
		msg[0] = msg[0]&0xf0 | (b & 0x0f)
	case 2:
		msg[1] = msg[1]&0x0f | (b << 4)
	case 3:
		msg[1] = msg[1]&0xf0 | (b & 0x0f)
	case 4:
		msg[2] = msg[2]&0x0f | (b << 4)
	case 5:
		msg[2] = msg[2]&0xf0 | (b & 0x0f)
	}
}

func (m *model) insertByte(b byte) {
	m.setByte(b)
	m.Right()
}

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if action, ok := msg.(Action); ok {
		if action.undoFn != nil {
			m.undoableActions = append(m.undoableActions, action)
			if len(m.undoableActions) > MaxUndoableActions {
				m.undoableActions = m.undoableActions[len(m.undoableActions)-MaxUndoableActions:]
			}
		}
		return m, nil
	}
	var cmds []tea.Cmd
	var cmd tea.Cmd
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		switch keyMsg.String() {
		case "esc":
			m.err = nil
		}
	}
	if windowSizeMsg, ok := msg.(tea.WindowSizeMsg); ok {
		m.windowWidth = windowSizeMsg.Width
		m.windowHeight = windowSizeMsg.Height
	} else if m.mode == EditMode {
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "0", "1", "2", "3", "4", "5", "6", "7", "8", "9", "a", "b", "c", "d", "e", "f":
				value := byte(msg.Runes[0])
				if value >= 0x61 {
					value = value - 0x61 + 0x3a
				}
				value -= 0x30
				prevByte := m.getByte()
				m.submitAction(
					func() {
						m.insertByte(value)
					},
					func() {
						m.Left()
						m.setByte(prevByte)
					},
				)
			default:
				switch {
				case key.Matches(msg, m.keymap.Quit):
					cmds = append(cmds, tea.Quit)
				case key.Matches(msg, m.keymap.Up):
					m.Up()
				case key.Matches(msg, m.keymap.Down):
					m.Down()
				case key.Matches(msg, m.keymap.PageUp):
					m.PageUp()
				case key.Matches(msg, m.keymap.PageDown):
					m.PageDown()
				case key.Matches(msg, m.keymap.JumpToFirstRow):
					m.JumpToFirstRow()
				case key.Matches(msg, m.keymap.JumpToLastRow):
					m.JumpToLastRow()
				case key.Matches(msg, m.keymap.Left):
					m.Left()
				case key.Matches(msg, m.keymap.Right):
					m.Right()
				case key.Matches(msg, m.keymap.NextTrack):
					m.NextTrack()
				case key.Matches(msg, m.keymap.PrevTrack):
					m.PrevTrack()
				case key.Matches(msg, m.keymap.JumpToFirstTrack):
					m.JumpToFirstTrack()
				case key.Matches(msg, m.keymap.JumpToLastTrack):
					m.JumpToLastTrack()
				case key.Matches(msg, m.keymap.DeleteLeft):
					m.DeleteLeft()
				case key.Matches(msg, m.keymap.DeleteUnder):
					m.DeleteUnder()
				case key.Matches(msg, m.keymap.InsertBlank):
					m.InsertBlank()
				case key.Matches(msg, m.keymap.PlayOrStop):
					m.PlayOrStop()
				case key.Matches(msg, m.keymap.SetStartRow):
					m.SetStartRow()
				case key.Matches(msg, m.keymap.EnterCommand):
					m.EnterCommand()
				case key.Matches(msg, m.keymap.Undo):
					m.Undo()
				case key.Matches(msg, m.keymap.Redo):
					m.Redo()
				case key.Matches(msg, m.keymap.Save):
					m.SaveSong()
				}
			}
		}
	} else if m.mode == CommandMode {
		m.commandModel, cmd = m.commandModel.Update(msg)
		cmds = append(cmds, cmd)
		if keyMsg, ok := msg.(tea.KeyMsg); ok {
			switch keyMsg.String() {
			case "enter":
				command := m.commandModel.Value()
				m.ExecuteCommand(command)
				fallthrough
			case "esc":
				m.mode = EditMode
				m.commandModel.Blur()
				m.commandModel.Reset()
			}
		}
	}
	return m, tea.Batch(cmds...)
}

func (m *model) Close() error {
	if m.me != nil {
		m.me.Close()
		m.me = nil
	}
	return nil
}
