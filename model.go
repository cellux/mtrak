package main

import (
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"math"
)

var defaultBrush = Area{
	Rect:      Rect{0, 0, 1, 1},
	ExpandDir: Point{1, 1},
}

func (m *Model) Reset() {
	m.err = nil
	//m.keymap = ?
	m.mode = EditMode
	m.prevmode = m.mode
	//m.windowSize
	//m.me
	//m.song
	m.brush = defaultBrush
	m.sel = m.brush.Rect
	m.editPattern = 0
	m.editPos.X = 0
	m.editPos.Y = 0
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
	m.clipboard = nil
	m.pasteOffset = 0
	m.defaultBrush = false
}

func (m *Model) SetError(err error) {
	m.err = err
}

func (m *Model) CollapseBrush() {
	m.brush.X = m.editPos.X
	m.brush.Y = m.editPos.Y
	m.brush.W = 1
	m.brush.H = 1
	m.CollapseSelection()
}

func (m *Model) CollapseSelection() {
	m.sel = m.brush.Rect
}

func (m *Model) SetSong(song *Song) {
	m.Reset()
	m.song = song
}

func (m *Model) ReplaceEditPattern(p Pattern) {
	m.song.Patterns[m.editPattern] = p
	m.fix()
}

func (m *Model) Play() {
	m.playTick = 0
	m.isPlaying = true
}

func (m *Model) Stop() {
	m.isPlaying = false
	m.playTick = 0
}

func (m *Model) QuitWithError(err error) tea.Cmd {
	m.err = err
	return tea.Quit
}

func (m *Model) GetSampleRate() int {
	return int(m.me.client.GetSampleRate())
}

func (m *Model) GetBeatsPerSecond() float64 {
	return float64(m.song.BPM) / 60.0
}

func (m *Model) GetFramesPerBeat() int {
	sr := float64(m.GetSampleRate())
	bps := m.GetBeatsPerSecond()
	return int(math.Round(sr / bps))
}

func (m *Model) GetTicksPerBeat() int {
	return m.song.TPL * m.song.LPB
}

func (m *Model) GetFramesPerTick() int {
	sr := float64(m.GetSampleRate())
	bps := m.GetBeatsPerSecond()
	tpb := float64(m.GetTicksPerBeat())
	return int(math.Round(sr / bps / tpb))
}

func (m *Model) processPendingActions() {
	for {
		select {
		case action := <-m.pendingActions:
			action.doFn()
			m.msgs <- action
		default:
			return
		}
	}
}

func (m *Model) Process(nframes uint32) int {
	m.processPendingActions()
	outPort := m.me.outPort
	buf := outPort.MidiClearBuffer(nframes)
	if !m.isPlaying {
		m.playFrame += uint64(nframes)
		return 0
	}
	framesPerTick := uint64(m.GetFramesPerTick())
	var midiData MidiData
	p := m.song.Patterns[m.playPattern]
	for i := range nframes {
		if m.playFrame%framesPerTick == 0 {
			if m.playTick == 0 {
				row := p[m.playRow]
				for _, msg := range row {
					status := msg[0]
					if status >= 0x80 {
						midiData.Time = i
						midiData.Buffer = msg.bytes()
						outPort.MidiEventWrite(&midiData, buf)
					}
				}
			}
			m.playTick++
			if m.playTick >= m.song.TPL {
				m.playRow++
				if m.playRow == len(p) {
					// TODO: advance to next pattern in sequence
					m.playRow = 0
				}
				m.playTick = 0
				m.msgs <- redrawMsg{}
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

func (m *Model) Init() tea.Cmd {
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
		for msg := range m.msgs {
			program.Send(msg)
		}
	}()
	if m.filename != "" {
		m.LoadSong()
	}
	return nil
}

func (m *Model) getDigit() byte {
	p := m.song.Patterns[m.editPattern]
	return p.getDigit(m.editPos.X, m.editPos.Y)
}

func (m *Model) setDigit(b byte) {
	p := m.song.Patterns[m.editPattern]
	p.setDigit(m.editPos.X, m.editPos.Y, b)
}

func (m *Model) insertDigit(b byte) {
	m.setDigit(b)
	m.Right()
}

func (m *Model) getBlock() Block {
	p := m.song.Patterns[m.editPattern]
	return p.getBlock(m.brush.Rect)
}

func (m *Model) setBlock(block Block) {
	p := m.song.Patterns[m.editPattern]
	p.setBlock(m.brush.Rect, block)
}

func (m *Model) zeroBlock() {
	p := m.song.Patterns[m.editPattern]
	p.zeroBlock(m.brush.Rect)
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
		}
	case tea.WindowSizeMsg:
		m.windowSize.W = msg.Width
		m.windowSize.H = msg.Height
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
				case key.Matches(msg, m.keymap.JumpToTopLeft):
					m.JumpToTopLeft()
				case key.Matches(msg, m.keymap.JumpToBottomRight):
					m.JumpToBottomRight()
				case key.Matches(msg, m.keymap.Left):
					m.Left()
				case key.Matches(msg, m.keymap.Right):
					m.Right()
				case key.Matches(msg, m.keymap.NextTrack):
					m.NextTrack()
				case key.Matches(msg, m.keymap.PrevTrack):
					m.PrevTrack()
				case key.Matches(msg, m.keymap.DeleteBrush):
					m.DeleteBrush()
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
				case key.Matches(msg, m.keymap.IncSelectionWidth):
					m.IncSelectionWidth()
				case key.Matches(msg, m.keymap.DecSelectionWidth):
					m.DecSelectionWidth()
				case key.Matches(msg, m.keymap.IncSelectionHeight):
					m.IncSelectionHeight()
				case key.Matches(msg, m.keymap.DecSelectionHeight):
					m.DecSelectionHeight()
				case key.Matches(msg, m.keymap.InsertBlockV):
					m.InsertBlockV()
				case key.Matches(msg, m.keymap.DeleteBlockV):
					m.DeleteBlockV()
				case key.Matches(msg, m.keymap.InsertBlockH):
					m.InsertBlockH()
				case key.Matches(msg, m.keymap.DeleteBlockH):
					m.DeleteBlockH()
				case key.Matches(msg, m.keymap.Cut):
					m.Cut()
				case key.Matches(msg, m.keymap.Copy):
					m.Copy()
				case key.Matches(msg, m.keymap.Paste):
					m.Paste()
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
				m.mode = m.prevmode
				m.commandModel.Blur()
				m.commandModel.Reset()
			}
		}
	}
	return m, tea.Batch(cmds...)
}

func (m *Model) Close() error {
	if m.me != nil {
		m.me.Close()
		m.me = nil
	}
	return nil
}
