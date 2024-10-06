package main

import (
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"math"
	"slices"
)

const (
	MaxUndoableActions = 64
)

type Row []MidiMessage
type Pattern []Row

func (p *Pattern) clone() Pattern {
	clone := make(Pattern, len(*p))
	for rowIndex := 0; rowIndex < len(*p); rowIndex++ {
		clone[rowIndex] = slices.Clone((*p)[rowIndex])
	}
	return clone
}

func (p *Pattern) getDigit(x, y int) byte {
	row := (*p)[y]
	msg := &row[x/6]
	switch x % 6 {
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

func (p *Pattern) setDigit(x, y int, b byte) {
	row := (*p)[y]
	msg := &row[x/6]
	switch x % 6 {
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

type Rect struct {
	X int
	Y int
	W int
	H int
}

type Block [][]byte

func (p *Pattern) getBlock(r Rect) Block {
	result := make(Block, r.H)
	for dy := 0; dy < r.H; dy++ {
		result[dy] = make([]byte, r.W)
		for dx := 0; dx < r.W; dx++ {
			result[dy][dx] = p.getDigit(r.X+dx, r.Y+dy)
		}
	}
	return result
}

func (p *Pattern) setBlock(r Rect, block Block) {
	for dy := 0; dy < r.H; dy++ {
		for dx := 0; dx < r.W; dx++ {
			p.setDigit(r.X+dx, r.Y+dy, block[dy][dx])
		}
	}
}

func (p *Pattern) zeroBlock(r Rect) {
	for dy := 0; dy < r.H; dy++ {
		for dx := 0; dx < r.W; dx++ {
			p.setDigit(r.X+dx, r.Y+dy, 0)
		}
	}
}

func (p *Pattern) copyBlock(r Rect, dx, dy int) {
	block := p.getBlock(r)
	p.setBlock(Rect{r.X + dx, r.Y + dy, r.W, r.H}, block)
}

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

var defaultBrush = Rect{0, 0, 1, 1}

type model struct {
	err               error
	keymap            *KeyMap
	mode              int
	windowWidth       int
	windowHeight      int
	me                *MidiEngine
	song              *Song
	brush             Rect
	sel               Rect
	editPattern       int
	editX             int
	editY             int
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
}

func (m *model) Reset() {
	m.err = nil
	//m.keymap = ?
	m.mode = EditMode
	//m.windowWidth
	//m.windowHeight
	//m.me
	//m.song
	m.brush = defaultBrush
	m.sel = Rect{}
	m.editPattern = 0
	m.editX = 0
	m.editY = 0
	m.firstVisibleRow = 0
	m.firstVisibleTrack = 0
	m.playPattern = 0
	m.playRow = 0
	m.playTick = 0
	//m.playFrame
	m.isPlaying = false
	m.playFromRow = 0
	m.commandModel.Reset()
	//m.filename
	//m.pendingActions ?
	//m.msgs ?
	m.undoableActions = nil
	m.undoneActions = nil
}

func (m *model) SetError(err error) {
	m.err = err
}

func (m *model) CollapseBrush() {
	m.brush.X = m.editX
	m.brush.Y = m.editY
	m.brush.W = 1
	m.brush.H = 1
}

func (m *model) SetSong(song *Song) {
	m.Reset()
	m.song = song
}

func (m *model) ReplaceEditPattern(p Pattern) {
	m.song.Patterns[m.editPattern] = p
	m.fix()
}

func (m *model) Play() {
	m.playTick = 0
	m.isPlaying = true
}

func (m *model) Stop() {
	m.isPlaying = false
	m.playTick = 0
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
			m.msgs <- action
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
					// TODO: advance to next pattern in sequence
					m.playRow = 0
				}
				m.msgs <- redrawMsg{}
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

func makePattern(rowCount, trackCount int) Pattern {
	rows := make([]Row, rowCount)
	for i := range rowCount {
		rows[i] = make([]MidiMessage, trackCount)
	}
	return rows
}

func (m *model) Init() tea.Cmd {
	m.keymap = &defaultKeyMap
	m.me = &MidiEngine{}
	if err := m.me.Open(m.Process); err != nil {
		return m.QuitWithError(err)
	}
	m.song = &Song{
		BPM:      120,
		LPB:      4,
		TPL:      6,
		Patterns: make([]Pattern, 1, 256),
	}
	m.song.Patterns[0] = makePattern(64, 16)
	m.commandModel = textinput.New()
	m.pendingActions = make(chan Action, 64)
	m.msgs = make(chan tea.Msg, 64)
	m.Reset()
	go func() {
		for {
			select {
			case msg := <-m.msgs:
				program.Send(msg)
			}
		}
	}()
	return nil
}

func (m *model) getDigit() byte {
	p := m.song.Patterns[m.editPattern]
	return p.getDigit(m.editX, m.editY)
}

func (m *model) setDigit(b byte) {
	p := m.song.Patterns[m.editPattern]
	p.setDigit(m.editX, m.editY, b)
}

func (m *model) insertDigit(b byte) {
	m.setDigit(b)
	m.Right()
}

func (m *model) getBlock() Block {
	p := m.song.Patterns[m.editPattern]
	return p.getBlock(m.brush)
}

func (m *model) setBlock(block Block) {
	p := m.song.Patterns[m.editPattern]
	p.setBlock(m.brush, block)
}

func (m *model) zeroBlock() {
	p := m.song.Patterns[m.editPattern]
	p.zeroBlock(m.brush)
}

func (m *model) hasSelection() bool {
	return m.sel.W > 0 && m.sel.H > 0
}

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case Action:
		if msg.undoFn != nil {
			m.undoableActions = append(m.undoableActions, msg)
			if len(m.undoableActions) > MaxUndoableActions {
				m.undoableActions = m.undoableActions[len(m.undoableActions)-MaxUndoableActions:]
			}
		}
		return m, nil
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			m.err = nil
			m.CollapseBrush()
			m.SelectNone()
		}
	case tea.WindowSizeMsg:
		m.windowWidth = msg.Width
		m.windowHeight = msg.Height
	}
	switch m.mode {
	case EditMode:
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "0", "1", "2", "3", "4", "5", "6", "7", "8", "9", "a", "b", "c", "d", "e", "f":
				value := byte(msg.Runes[0])
				if value >= 0x61 {
					value = value - 0x61 + 0x3a
				}
				value -= 0x30
				prevDigit := m.getDigit()
				m.submitAction(
					func() {
						m.insertDigit(value)
					},
					func() {
						m.Left()
						m.setDigit(prevDigit)
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
				case key.Matches(msg, m.keymap.InsertBlank):
					m.InsertBlank()
				case key.Matches(msg, m.keymap.DeleteLeft):
					m.DeleteLeft()
				case key.Matches(msg, m.keymap.IncBrushWidth):
					m.IncBrushWidth()
				case key.Matches(msg, m.keymap.DecBrushWidth):
					m.DecBrushWidth()
				case key.Matches(msg, m.keymap.IncBrushHeight):
					m.IncBrushHeight()
				case key.Matches(msg, m.keymap.DecBrushHeight):
					m.DecBrushHeight()
				case key.Matches(msg, m.keymap.InsertBlockV):
					m.InsertBlockV()
				case key.Matches(msg, m.keymap.DeleteBlockV):
					m.DeleteBlockV()
				case key.Matches(msg, m.keymap.InsertBlockH):
					m.InsertBlockH()
				case key.Matches(msg, m.keymap.DeleteBlockH):
					m.DeleteBlockH()
				case key.Matches(msg, m.keymap.PlayOrStop):
					m.PlayOrStop()
				case key.Matches(msg, m.keymap.SetPlayFromRow):
					m.SetPlayFromRow()
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
	case CommandMode:
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
