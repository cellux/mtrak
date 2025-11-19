package main

import (
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"math"
)

var defaultBrush = Brush{
	Rect:      Rect{0, 0, 1, 1},
	ExpandDir: Point{1, 1},
}

func drain[T any](ch chan T) {
	for {
		select {
		case <-ch:
			// keep draining
		default:
			return // channel is empty
		}
	}
}

func (m *Model) Reset() {
	m.err = nil
	//m.keymap = ?
	m.mode = EditMode
	m.prevmode = m.mode
	//m.windowSize
	//m.midiEngine
	//m.song
	m.brush = defaultBrush
	m.sel = m.brush.Rect
	//m.editPattern
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
	drain(m.pendingMidiMessages)
	//m.msgs ?
	m.undoableActions = nil
	m.undoneActions = nil
	m.clipboard = nil
	m.pasteOffset = 0
	m.usingTempBrush = false
}

func (m *Model) SetMode(newMode Mode) {
	m.prevmode = m.mode
	m.mode = newMode
}

func (m *Model) ResetMode() {
	m.mode = m.prevmode
}

func (m *Model) SetError(err error) {
	m.err = err
}

func (m *Model) CollapseBrush() {
	m.brush.X = m.editPos.X
	m.brush.Y = m.editPos.Y
	m.brush.W = 1
	m.brush.H = 1
}

func (m *Model) CollapseSelection() {
	m.sel = m.brush.Rect
}

func (m *Model) SetSong(song *Song) {
	m.Reset()
	m.song = song
	m.fix()
}

func (m *Model) ReplaceEditPattern(p *Pattern) {
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
	return int(m.midiEngine.client.GetSampleRate())
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
	outPort := m.midiEngine.outPort
	buf := outPort.MidiClearBuffer(nframes)
	var midiData MidiData
processPendingMidiMessages:
	for {
		select {
		case msg := <-m.pendingMidiMessages:
			status := msg[0]
			if status >= 0x80 {
				midiData.Time = 0
				midiData.Buffer = msg.bytes()
				outPort.MidiEventWrite(&midiData, buf)
			}
		default:
			break processPendingMidiMessages
		}
	}
	if !m.isPlaying {
		m.playFrame += uint64(nframes)
		return 0
	}
	framesPerTick := uint64(m.GetFramesPerTick())
	p := m.song.Patterns[m.playPattern]
	for i := range nframes {
		if m.playFrame%framesPerTick == 0 {
			if m.playTick == 0 {
				row := p.Rows[m.playRow]
				for numTrack, msg := range row {
					if msg[0] == 0 && (msg[1] != 0 || msg[2] != 0) {
						defaults := p.TrackDefaults[numTrack]
						for j := range 3 {
							if msg[j] == 0 {
								msg[j] = defaults[j]
							}
						}
					}
					if msg[0] >= 0x80 {
						midiData.Time = i
						midiData.Buffer = msg.bytes()
						outPort.MidiEventWrite(&midiData, buf)
						p.TrackDefaults[numTrack] = msg
					}
				}
			}
			m.playTick++
			if m.playTick >= m.song.TPL {
				m.playRow++
				if m.playRow == p.NumRows {
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

func (m *Model) Init() tea.Cmd {
	m.keymap = &defaultKeyMap
	m.midiEngine = &MidiEngine{}
	if err := m.midiEngine.Open(m.Process); err != nil {
		return m.QuitWithError(err)
	}
	m.song = &Song{
		BPM:      120,
		LPB:      4,
		TPL:      6,
		Patterns: make([]*Pattern, 1, 256),
	}
	FixSong(m.song)
	m.song.Patterns[0] = makeDefaultPattern()
	m.commandModel = textinput.New()
	m.pendingActions = make(chan Action, 64)
	m.pendingMidiMessages = make(chan MidiMessage, 64)
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

func (m *Model) getNoteByte() byte {
	p := m.song.Patterns[m.editPattern]
	noteOffset := m.editPos.X - m.editPos.X%6 + 2
	hi := p.getDigit(noteOffset, m.editPos.Y)
	lo := p.getDigit(noteOffset+1, m.editPos.Y)
	return hi<<4 + lo
}

func (m *Model) setNoteByte(midiNote byte) {
	p := m.song.Patterns[m.editPattern]
	noteOffset := m.editPos.X - m.editPos.X%6 + 2
	p.setDigit(noteOffset, m.editPos.Y, midiNote>>4)
	p.setDigit(noteOffset+1, m.editPos.Y, midiNote&0x0f)
}

type MessageHandler func(m *Model, msg tea.Msg) (cmds []tea.Cmd)

var modeSpecificMessageHandlers = map[Mode]MessageHandler{
	EditMode: func(m *Model, msg tea.Msg) (cmds []tea.Cmd) {
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
				case key.Matches(msg, m.keymap.InsertTrack):
					m.InsertTrack()
				case key.Matches(msg, m.keymap.DeleteTrack):
					m.DeleteTrack()
				case key.Matches(msg, m.keymap.IncBrushWidth):
					m.IncBrushWidth()
				case key.Matches(msg, m.keymap.DecBrushWidth):
					m.DecBrushWidth()
				case key.Matches(msg, m.keymap.IncBrushHeight):
					m.IncBrushHeight()
				case key.Matches(msg, m.keymap.DecBrushHeight):
					m.DecBrushHeight()
				case key.Matches(msg, m.keymap.IncSelectionWidth):
					m.SetMode(SelectMode)
					m.IncSelectionWidth()
				case key.Matches(msg, m.keymap.DecSelectionWidth):
					m.SetMode(SelectMode)
					m.DecSelectionWidth()
				case key.Matches(msg, m.keymap.IncSelectionHeight):
					m.applyTempBrush(6)
					m.SetMode(SelectMode)
					m.IncSelectionHeight()
				case key.Matches(msg, m.keymap.DecSelectionHeight):
					m.applyTempBrush(6)
					m.SetMode(SelectMode)
					m.DecSelectionHeight()
				case key.Matches(msg, m.keymap.InsertBlock):
					m.applyTempBrush(6)
					m.InsertBlock()
					m.revertTempBrush()
				case key.Matches(msg, m.keymap.DeleteBlock):
					m.applyTempBrush(6)
					m.DeleteBlock(false)
					m.revertTempBrush()
				case key.Matches(msg, m.keymap.BackspaceBlock):
					m.applyTempBrush(6)
					m.DeleteBlock(true)
					m.revertTempBrush()
				case key.Matches(msg, m.keymap.ZeroBlock):
					m.applyTempBrush(2)
					m.ZeroBlock()
					m.revertTempBrush()
				case key.Matches(msg, m.keymap.Cut):
					m.applyTempBrush(6)
					m.Cut()
					m.revertTempBrush()
				case key.Matches(msg, m.keymap.Copy):
					m.applyTempBrush(6)
					m.Copy()
					m.revertTempBrush()
				case key.Matches(msg, m.keymap.Paste):
					m.Paste()
				case key.Matches(msg, m.keymap.PlayOrStop):
					m.PlayOrStop()
				case key.Matches(msg, m.keymap.SetPlayFromRow):
					m.SetPlayFromRow()
				case key.Matches(msg, m.keymap.EnterCommandMode):
					m.EnterCommandMode()
				case key.Matches(msg, m.keymap.EnterNoteMode):
					m.EnterNoteMode()
				case key.Matches(msg, m.keymap.EnterChromaticMode):
					m.EnterChromaticMode()
				case key.Matches(msg, m.keymap.Undo):
					m.Undo()
				case key.Matches(msg, m.keymap.Redo):
					m.Redo()
				case key.Matches(msg, m.keymap.Save):
					m.SaveSong()
				}
			}
		}
		return cmds
	},
	NoteMode: func(m *Model, msg tea.Msg) (cmds []tea.Cmd) {
		switch msg := msg.(type) {
		case tea.KeyMsg:
			midiNote := m.KeyMsgToMidiNote(msg)
			if midiNote >= 0 {
				prevNote := m.getNoteByte()
				m.submitAction(
					func() {
						m.setNoteByte(byte(midiNote))
					},
					func() {
						m.setNoteByte(prevNote)
					},
				)
				msg := MidiMessage{0, byte(midiNote), 0}
				p := m.song.Patterns[m.editPattern]
				numTrack := m.editPos.X / 6
				for y := m.editPos.Y; y >= 0 && msg[0] == 0 && msg[2] == 0; y-- {
					ymsg := p.Rows[y][numTrack]
					if msg[0] == 0 && ymsg[0] != 0 {
						msg[0] = 0x90 + ymsg[0]&0x0f
					}
					if msg[2] == 0 && ymsg[2] != 0 {
						msg[2] = ymsg[2]
					}
				}
				defaults := p.TrackDefaults[numTrack]
				if msg[0] == 0 && defaults[0] != 0 {
					msg[0] = 0x90 + defaults[0]&0x0f
				}
				if msg[2] == 0 && defaults[2] != 0 {
					msg[2] = defaults[2]
				}
				m.pendingMidiMessages <- msg
			} else {
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
					m.PrevTrack()
				case key.Matches(msg, m.keymap.Right):
					m.NextTrack()
				case key.Matches(msg, m.keymap.NextTrack):
					m.NextTrack()
				case key.Matches(msg, m.keymap.PrevTrack):
					m.PrevTrack()
				case key.Matches(msg, m.keymap.InsertTrack):
					m.InsertTrack()
				case key.Matches(msg, m.keymap.DeleteTrack):
					m.DeleteTrack()
				case key.Matches(msg, m.keymap.ZeroBlock):
					m.setNoteByte(0)
				case key.Matches(msg, m.keymap.PlayOrStop):
					m.PlayOrStop()
				case key.Matches(msg, m.keymap.SetPlayFromRow):
					m.SetPlayFromRow()
				case key.Matches(msg, m.keymap.EnterChromaticMode):
					m.song.Chromatic = !m.song.Chromatic
				case key.Matches(msg, m.keymap.Undo):
					m.Undo()
				case key.Matches(msg, m.keymap.Redo):
					m.Redo()
				case key.Matches(msg, m.keymap.Save):
					m.SaveSong()
				}
			}
		}
		return cmds
	},
	SelectMode: func(m *Model, msg tea.Msg) (cmds []tea.Cmd) {
		leaveSelectMode := false
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch {
			case key.Matches(msg, m.keymap.IncSelectionWidth):
				m.IncSelectionWidth()
			case key.Matches(msg, m.keymap.DecSelectionWidth):
				m.DecSelectionWidth()
			case key.Matches(msg, m.keymap.IncSelectionHeight):
				m.IncSelectionHeight()
			case key.Matches(msg, m.keymap.DecSelectionHeight):
				m.DecSelectionHeight()
			case key.Matches(msg, m.keymap.InsertBlock):
				m.InsertBlock()
			case key.Matches(msg, m.keymap.DeleteBlock):
				m.DeleteBlock(false)
			case key.Matches(msg, m.keymap.BackspaceBlock):
				m.DeleteBlock(true)
			case key.Matches(msg, m.keymap.ZeroBlock):
				m.ZeroBlock()
			case key.Matches(msg, m.keymap.NextTrack):
				m.brush.Y = m.sel.Y
				m.editPos.Y = m.sel.Y
				leaveSelectMode = true
			case key.Matches(msg, m.keymap.PrevTrack):
				m.brush.Y = m.sel.Y
				m.editPos.Y = m.sel.Y
				leaveSelectMode = true
			case key.Matches(msg, m.keymap.Cut):
				m.Cut()
			case key.Matches(msg, m.keymap.Copy):
				m.Copy()
			case key.Matches(msg, m.keymap.Undo):
				m.Undo()
			case key.Matches(msg, m.keymap.Redo):
				m.Redo()
			default:
				leaveSelectMode = true
			}
		default:
			leaveSelectMode = true
		}
		if leaveSelectMode {
			m.revertTempBrush()
			m.CollapseSelection()
			m.ResetMode()
			m.msgs <- msg
		}
		return cmds
	},
	CommandMode: func(m *Model, msg tea.Msg) (cmds []tea.Cmd) {
		var cmd tea.Cmd
		m.commandModel, cmd = m.commandModel.Update(msg)
		cmds = append(cmds, cmd)
		if keyMsg, ok := msg.(tea.KeyMsg); ok {
			switch keyMsg.String() {
			case "enter":
				command := m.commandModel.Value()
				m.ExecuteCommand(command)
				fallthrough
			case "esc":
				m.commandModel.Blur()
				m.commandModel.Reset()
				m.ResetMode()
			}
		}
		return cmds
	},
}

func (m *Model) HandleMessage(msg tea.Msg) (cmds []tea.Cmd) {
	return modeSpecificMessageHandlers[m.mode](m, msg)
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
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
			m.ResetMode()
			m.CollapseBrush()
			m.CollapseSelection()
		}
	case tea.WindowSizeMsg:
		m.windowSize.W = msg.Width
		m.windowSize.H = msg.Height
	}
	cmds = append(cmds, m.HandleMessage(msg)...)
	return m, tea.Batch(cmds...)
}

func (m *Model) Close() error {
	if m.midiEngine != nil {
		m.midiEngine.Close()
		m.midiEngine = nil
	}
	return nil
}
