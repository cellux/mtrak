package main

import (
	"encoding/json"
	"fmt"
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

func (m *model) fix() {
	patterns := m.song.Patterns
	if m.editPattern < 0 || m.editPattern >= len(patterns) {
		m.editPattern = 0
	}

	p := patterns[m.editPattern]
	patternHeight := len(p)

	if m.editY < 0 {
		m.editY = 0
	} else if m.editY >= patternHeight {
		m.editY = patternHeight - 1
	}

	if m.playFromRow >= patternHeight {
		m.playFromRow = 0
	}

	editRow := p[m.editY]
	patternWidth := len(editRow) * 6

	if m.editX < 0 {
		m.editX = 0
	} else if m.editX >= patternWidth {
		m.editX = patternWidth - 1
	}

	// fix brush
	if m.brush.X < 0 || m.brush.X >= patternWidth {
		m.CollapseBrush()
	}
	if m.brush.Y < 0 || m.brush.Y >= patternHeight {
		m.CollapseBrush()
	}
	if m.brush.X+m.brush.W > patternWidth {
		m.brush.W = patternWidth - m.brush.X
	}
	if m.brush.Y+m.brush.H > patternHeight {
		m.brush.H = patternHeight - m.brush.Y
	}

	// fix selection
	if m.selection.X < 0 || m.selection.X >= patternWidth {
		m.SelectNone()
	}
	if m.selection.Y < 0 || m.selection.Y >= patternHeight {
		m.SelectNone()
	}
	if m.selection.X+m.selection.W > patternWidth {
		m.selection.W = patternWidth - m.selection.X
	}
	if m.selection.Y+m.selection.H > patternHeight {
		m.selection.H = patternHeight - m.selection.Y
	}
}

func (m *model) moveBrush(dx, dy int) {
	p := m.song.Patterns[m.editPattern]
	if dx != 0 {
		editRow := p[m.editY]
		patternWidth := len(editRow) * 6
		if dx > 0 {
			if m.brush.X+dx+m.brush.W <= patternWidth {
				m.brush.X += dx
				m.editX += dx
			} else {
				m.moveBrush(patternWidth-m.brush.W-m.brush.X, 0)
			}
		} else {
			if m.brush.X+dx >= 0 {
				m.brush.X += dx
				m.editX += dx
			} else {
				m.moveBrush(-m.brush.X, 0)
			}
		}
	}
	if dy != 0 {
		patternHeight := len(p)
		if dy > 0 {
			if m.brush.Y+dy+m.brush.H <= patternHeight {
				m.brush.Y += dy
				m.editY += dy
			} else {
				m.moveBrush(0, patternHeight-m.brush.H-m.brush.Y)
			}
		} else {
			if m.brush.Y+dy >= 0 {
				m.brush.Y += dy
				m.editY += dy
			} else {
				m.moveBrush(0, -m.brush.Y)
			}
		}
	}
}

func (m *model) Up() {
	m.moveBrush(0, -m.brush.H)
}

func (m *model) Down() {
	m.moveBrush(0, m.brush.H)
}

func (m *model) PageUp() {
	m.moveBrush(0, -max(m.song.LPB, m.brush.H))
}

func (m *model) PageDown() {
	m.moveBrush(0, max(m.song.LPB, m.brush.H))
}

func (m *model) JumpToFirstRow() {
	m.moveBrush(0, -m.brush.Y)
}

func (m *model) JumpToLastRow() {
	p := m.song.Patterns[m.editPattern]
	patternHeight := len(p)
	m.moveBrush(0, patternHeight-m.brush.H-m.brush.Y)
}

func (m *model) Left() {
	m.moveBrush(-m.brush.W, 0)
}

func (m *model) Right() {
	m.moveBrush(m.brush.W, 0)
}

func (m *model) NextTrack() {
	m.moveBrush(6, 0)
}

func (m *model) PrevTrack() {
	m.moveBrush(-6, 0)
}

func (m *model) JumpToFirstTrack() {
	m.moveBrush(-m.brush.X, 0)
}

func (m *model) JumpToLastTrack() {
	p := m.song.Patterns[m.editPattern]
	editRow := p[m.editY]
	patternWidth := len(editRow) * 6
	m.moveBrush(patternWidth-m.brush.W-m.brush.X, 0)
}

func (m *model) InsertBlank() {
	p := m.song.Patterns[m.editPattern]
	editRow := p[m.editY]
	patternWidth := len(editRow) * 6
	var prevBlock Block
	if m.brush.X+m.brush.W <= patternWidth {
		m.submitAction(
			func() {
				prevBlock = m.getBlock()
				m.zeroBlock()
				m.Right()
			},
			func() {
				m.Left()
				m.setBlock(prevBlock)
			},
		)
	}
}

func (m *model) DeleteLeft() {
	var prevBlock Block
	if m.brush.X-m.brush.W >= 0 {
		m.submitAction(
			func() {
				m.Left()
				prevBlock = m.getBlock()
				m.zeroBlock()
			},
			func() {
				m.setBlock(prevBlock)
				m.Right()
			},
		)
	}
}

func (m *model) SelectNone() {
	m.selection = Rect{}
}

func (m *model) SelectAll() {
	p := m.song.Patterns[m.editPattern]
	patternHeight := len(p)
	editRow := p[m.editY]
	patternWidth := len(editRow) * 6
	m.selection = Rect{0, 0, patternWidth, patternHeight}
}

func (m *model) IncBrushWidth() {
	p := m.song.Patterns[m.editPattern]
	editRow := p[m.editY]
	patternWidth := len(editRow) * 6
	switch m.brush.W {
	case 1:
		m.brush.X = m.brush.X - (m.brush.X % 2)
		m.brush.W = 2
	case 2:
		m.brush.X = m.brush.X - (m.brush.X % 6)
		m.brush.W = 6
	default:
		m.brush.X = 0
		m.brush.W = patternWidth
	}
	m.SelectNone()
}

func (m *model) DecBrushWidth() {
	switch m.brush.W {
	case 1:
	case 2:
		m.brush.X = m.editX
		m.brush.W = 1
	case 6:
		m.brush.X = m.editX - (m.editX % 2)
		m.brush.W = 2
	default:
		m.brush.X = m.editX - (m.editX % 6)
		m.brush.W = 6
	}
	m.SelectNone()
}

func (m *model) IncBrushHeight() {
	p := m.song.Patterns[m.editPattern]
	patternHeight := len(p)
	switch m.brush.H {
	case 1:
		m.brush.Y = m.brush.Y - (m.brush.Y % m.song.LPB)
		m.brush.H = m.song.LPB
	default:
		m.brush.Y = 0
		m.brush.H = patternHeight
	}
	m.SelectNone()
}

func (m *model) DecBrushHeight() {
	switch m.brush.H {
	case 1:
	case m.song.LPB:
		m.brush.Y = m.editY
		m.brush.H = 1
	default:
		m.brush.Y = m.editY - (m.editY % m.song.LPB)
		m.brush.H = m.song.LPB
	}
	m.SelectNone()
}

func (m *model) InsertBlockV() {
	p := m.song.Patterns[m.editPattern]
	clone := p.clone()
	patternHeight := len(p)
	blockToMove := Rect{
		m.brush.X,
		m.brush.Y,
		m.brush.W,
		patternHeight - m.brush.Y - m.brush.H,
	}
	clone.copyBlock(blockToMove, 0, m.brush.H)
	clone.zeroBlock(m.brush)
	m.submitAction(
		func() {
			m.song.Patterns[m.editPattern] = clone
		},
		func() {
			m.song.Patterns[m.editPattern] = p
		},
	)
}

func (m *model) DeleteBlockV() {
	p := m.song.Patterns[m.editPattern]
	clone := p.clone()
	patternHeight := len(p)
	blockToMove := Rect{
		m.brush.X,
		m.brush.Y + m.brush.H,
		m.brush.W,
		patternHeight - m.brush.Y - m.brush.H,
	}
	clone.copyBlock(blockToMove, 0, -m.brush.H)
	blockToZero := Rect{
		m.brush.X,
		patternHeight - m.brush.H,
		m.brush.W,
		m.brush.H,
	}
	clone.zeroBlock(blockToZero)
	m.submitAction(
		func() {
			m.song.Patterns[m.editPattern] = clone
		},
		func() {
			m.song.Patterns[m.editPattern] = p
		},
	)
}

func (m *model) InsertBlockH() {
	p := m.song.Patterns[m.editPattern]
	clone := p.clone()
	editRow := p[m.editY]
	patternWidth := len(editRow) * 6
	blockToMove := Rect{
		m.brush.X,
		m.brush.Y,
		patternWidth - m.brush.X - m.brush.W,
		m.brush.H,
	}
	clone.copyBlock(blockToMove, m.brush.W, 0)
	clone.zeroBlock(m.brush)
	m.submitAction(
		func() {
			m.song.Patterns[m.editPattern] = clone
		},
		func() {
			m.song.Patterns[m.editPattern] = p
		},
	)
}

func (m *model) DeleteBlockH() {
	p := m.song.Patterns[m.editPattern]
	clone := p.clone()
	editRow := p[m.editY]
	patternWidth := len(editRow) * 6
	blockToMove := Rect{
		m.brush.X + m.brush.W,
		m.brush.Y,
		patternWidth - m.brush.X - m.brush.W,
		m.brush.H,
	}
	clone.copyBlock(blockToMove, -m.brush.W, 0)
	blockToZero := Rect{
		patternWidth - m.brush.W,
		m.brush.Y,
		m.brush.W,
		m.brush.H,
	}
	clone.zeroBlock(blockToZero)
	m.submitAction(
		func() {
			m.song.Patterns[m.editPattern] = clone
		},
		func() {
			m.song.Patterns[m.editPattern] = p
		},
	)
}

func (m *model) PlayOrStop() {
	m.submitAction(
		func() {
			if m.isPlaying {
				m.Stop()
			} else {
				m.playRow = m.playFromRow
				m.Play()
			}
		},
		nil,
	)
}

func (m *model) SetPlayFromRow() {
	m.playFromRow = m.editY
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
		m.SetError(fmt.Errorf("no filename"))
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
