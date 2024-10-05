package main

import (
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"math"
)

type Row = []MidiMessage
type Pattern = []Row

type model struct {
	err          error
	keymap       *KeyMap
	windowWidth  int
	windowHeight int
	me           *MidiEngine
	bpm          int // beats per minute
	lpb          int // lines per beat
	tpl          int // ticks per line
	patterns     []Pattern
	editPattern  int
	editRow      int
	editRow0     int
	editTrack    int
	editTrack0   int
	editColumn   int
	playPattern  int
	playRow      int
	playTick     int
	playFrame    uint64
	isPlaying    bool
	startRow     int
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
	return float64(m.bpm) / 60.0
}

func (m *model) GetFramesPerBeat() int {
	sr := float64(m.GetSampleRate())
	bps := m.GetBeatsPerSecond()
	return int(math.Round(sr / bps))
}

func (m *model) GetTicksPerBeat() int {
	return m.tpl * m.lpb
}

func (m *model) GetFramesPerTick() int {
	sr := float64(m.GetSampleRate())
	bps := m.GetBeatsPerSecond()
	tpb := float64(m.GetTicksPerBeat())
	return int(math.Round(sr / bps / tpb))
}

type redrawMsg struct{}

func (m *model) Process(nframes uint32) int {
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
				p := m.patterns[m.playPattern]
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
			if m.playTick == m.tpl {
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
	m.keymap = &defaultKeyMap
	m.bpm = 120
	m.lpb = 4
	m.tpl = 6
	m.patterns = make([]Pattern, 256)
	m.patterns[0] = makePattern(64, 16)
	m.me = &MidiEngine{}
	if err := m.me.Open(m.Process); err != nil {
		return m.QuitWithError(err)
	}
	return nil
}

func (m *model) setByte(b byte) {
	p := m.patterns[m.editPattern]
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
	var cmd tea.Cmd = nil
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "0", "1", "2", "3", "4", "5", "6", "7", "8", "9", "a", "b", "c", "d", "e", "f":
			value := byte(msg.Runes[0])
			if value >= 0x61 {
				value = value - 0x61 + 0x3a
			}
			m.insertByte(value - 0x30)
		default:
			switch {
			case key.Matches(msg, m.keymap.Quit):
				cmd = tea.Quit
			case key.Matches(msg, m.keymap.Up):
				m.Up()
			case key.Matches(msg, m.keymap.Down):
				m.Down()
			case key.Matches(msg, m.keymap.PageUp):
				m.PageUp()
			case key.Matches(msg, m.keymap.PageDown):
				m.PageDown()
			case key.Matches(msg, m.keymap.Left):
				m.Left()
			case key.Matches(msg, m.keymap.Right):
				m.Right()
			case key.Matches(msg, m.keymap.NextTrack):
				m.NextTrack()
			case key.Matches(msg, m.keymap.PrevTrack):
				m.PrevTrack()
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
			}
		}
	case tea.WindowSizeMsg:
		m.windowWidth = msg.Width
		m.windowHeight = msg.Height
	}
	return m, cmd
}

func (m *model) Close() error {
	if m.me != nil {
		m.me.Close()
		m.me = nil
	}
	return nil
}
