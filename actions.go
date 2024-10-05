package main

import (
	"encoding/json"
	"os"
)

type ActionFunction func()

type Action struct {
	doFn   ActionFunction
	undoFn ActionFunction
}

func (m *model) submitAction(doFn ActionFunction, undoFn ActionFunction) {
	m.pendingActions <- Action{doFn, undoFn}
}

func (m *model) Up() {
	if m.editRow > 0 {
		m.editRow--
	}
}

func (m *model) Down() {
	p := m.song.Patterns[m.editPattern]
	if m.editRow < len(p)-1 {
		m.editRow++
	}
}

func (m *model) PageUp() {
	if m.editRow%16 != 0 {
		m.editRow -= m.editRow % 16
	} else {
		m.editRow -= 16
	}
	if m.editRow < 0 {
		m.editRow = 0
	}
}

func (m *model) PageDown() {
	p := m.song.Patterns[m.editPattern]
	m.editRow += 16
	if m.editRow%16 != 0 {
		m.editRow -= m.editRow % 16
	}
	if m.editRow >= len(p) {
		m.editRow = len(p) - 1
	}
}

func (m *model) Left() {
	if m.editColumn > 0 {
		m.editColumn--
	} else if m.editTrack > 0 {
		m.editTrack--
		m.editColumn = 5
	}
}

func (m *model) Right() {
	p := m.song.Patterns[m.editPattern]
	row := p[m.editRow]
	if m.editColumn < 5 {
		m.editColumn++
	} else if m.editTrack < len(row)-1 {
		m.editTrack++
		m.editColumn = 0
	}
}

func (m *model) NextTrack() {
	p := m.song.Patterns[m.editPattern]
	row := p[m.editRow]
	if m.editTrack < len(row)-1 {
		m.editTrack++
		m.editColumn = 0
	}
}

func (m *model) PrevTrack() {
	if m.editTrack > 0 {
		m.editTrack--
		m.editColumn = 0
	}
}

func (m *model) DeleteLeft() {
	var prevByte byte
	m.submitAction(
		func() {
			if m.editTrack > 0 || m.editColumn > 0 {
				m.Left()
				prevByte = m.getByte()
				m.setByte(0)
			}
		},
		func() {
			m.setByte(prevByte)
			m.Right()
		},
	)
}

func (m *model) DeleteUnder() {
	var prevByte byte
	m.submitAction(
		func() {
			prevByte = m.getByte()
			m.setByte(0)
		},
		func() {
			m.setByte(prevByte)
		},
	)
}

func (m *model) InsertBlank() {
	var prevByte byte
	m.submitAction(
		func() {
			prevByte = m.getByte()
			m.insertByte(0)
		},
		func() {
			m.Left()
			m.setByte(prevByte)
		},
	)
}

func (m *model) PlayOrStop() {
	if m.isPlaying {
		m.Stop()
	} else {
		m.playRow = m.startRow
		m.Play()
	}
}

func (m *model) SetStartRow() {
	m.startRow = m.editRow
}

func (m *model) EnterCommand() {
	m.mode = CommandMode
	m.commandModel.Focus()
}

func (m *model) LoadSong() {
	if m.filename == "" {
		return
	}
	b, err := os.ReadFile(m.filename)
	if err != nil {
		m.SetError(err)
		return
	}
	song := &Song{}
	if err := json.Unmarshal(b, song); err != nil {
		m.SetError(err)
		return
	}
	prevSong := m.song
	m.submitAction(
		func() {
			m.SetSong(song)
		},
		func() {
			m.SetSong(prevSong)
		},
	)
}

func (m *model) SaveSong() {
	if m.filename == "" {
		return
	}
	b, err := json.Marshal(m.song)
	if err != nil {
		m.SetError(err)
		return
	}
	if err := os.WriteFile(m.filename, b, 0o644); err != nil {
		m.SetError(err)
		return
	}
}

func (m *model) Undo() {
	if len(m.undoableActions) == 0 {
		return
	}
	lastAction := m.undoableActions[len(m.undoableActions)-1]
	m.undoableActions = m.undoableActions[:len(m.undoableActions)-1]
	m.undoneActions = append(m.undoneActions, lastAction)
	m.submitAction(lastAction.undoFn, nil)
}

func (m *model) Redo() {
	if len(m.undoneActions) == 0 {
		return
	}
	lastAction := m.undoneActions[len(m.undoneActions)-1]
	m.undoneActions = m.undoneActions[:len(m.undoneActions)-1]
	m.submitAction(lastAction.doFn, lastAction.undoFn)
}
