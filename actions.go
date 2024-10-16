package main

import (
	"encoding/json"
	"fmt"
	"os"
)

func (m *AppModel) submitAction(doFn ActionFunction, undoFn ActionFunction) {
	m.pendingActions <- Action{doFn, undoFn}
}

func (m *AppModel) fix() {
	patterns := m.song.Patterns
	if m.editPattern < 0 || m.editPattern >= len(patterns) {
		m.editPattern = 0
	}

	p := patterns[m.editPattern]
	patternHeight := len(p)

	if m.editPos.Y < 0 {
		m.editPos.Y = 0
	} else if m.editPos.Y >= patternHeight {
		m.editPos.Y = patternHeight - 1
	}

	if m.playFromRow >= patternHeight {
		m.playFromRow = 0
	}

	editRow := p[m.editPos.Y]
	patternWidth := len(editRow) * 6

	if m.editPos.X < 0 {
		m.editPos.X = 0
	} else if m.editPos.X >= patternWidth {
		m.editPos.X = patternWidth - 1
	}

	// fix brush
	if m.brush.X < 0 || m.brush.X >= patternWidth {
		m.CollapseBrush()
	}
	if m.brush.Y < 0 || m.brush.Y >= patternHeight {
		m.CollapseBrush()
	}
	if m.brush.W > patternWidth {
		m.brush.W = patternWidth
	}
	if m.brush.X+m.brush.W > patternWidth {
		m.brush.X = patternWidth - m.brush.W
	}
	if m.brush.H > patternHeight {
		m.brush.W = patternHeight
	}
	if m.brush.Y+m.brush.H > patternHeight {
		m.brush.Y = patternHeight - m.brush.H
	}

	// fix selection
	if m.selection.X < 0 || m.selection.X >= patternWidth {
		m.SelectNone()
	}
	if m.selection.Y < 0 || m.selection.Y >= patternHeight {
		m.SelectNone()
	}
	if m.selection.W > patternWidth {
		m.selection.W = patternWidth
	}
	if m.selection.X+m.selection.W > patternWidth {
		m.selection.X = patternWidth - m.selection.W
	}
	if m.selection.H > patternHeight {
		m.selection.H = patternHeight
	}
	if m.selection.Y+m.selection.H > patternHeight {
		m.selection.Y = patternHeight - m.selection.H
	}
}

func (m *AppModel) moveBrush(dx, dy int) {
	p := m.song.Patterns[m.editPattern]
	if dx != 0 {
		editRow := p[m.editPos.Y]
		patternWidth := len(editRow) * 6
		if dx > 0 {
			if m.brush.X+dx+m.brush.W <= patternWidth {
				m.brush.X += dx
				m.editPos.X += dx
			} else {
				m.moveBrush(patternWidth-m.brush.W-m.brush.X, 0)
			}
		} else {
			if m.brush.X+dx >= 0 {
				m.brush.X += dx
				m.editPos.X += dx
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
				m.editPos.Y += dy
			} else {
				m.moveBrush(0, patternHeight-m.brush.H-m.brush.Y)
			}
		} else {
			if m.brush.Y+dy >= 0 {
				m.brush.Y += dy
				m.editPos.Y += dy
			} else {
				m.moveBrush(0, -m.brush.Y)
			}
		}
	}
}

func (m *AppModel) Up() {
	m.moveBrush(0, -m.brush.H)
}

func (m *AppModel) Down() {
	m.moveBrush(0, m.brush.H)
}

func (m *AppModel) PageUp() {
	m.moveBrush(0, -max(m.song.LPB, m.brush.H))
}

func (m *AppModel) PageDown() {
	m.moveBrush(0, max(m.song.LPB, m.brush.H))
}

func (m *AppModel) JumpToFirstRow() {
	m.moveBrush(0, -m.brush.Y)
}

func (m *AppModel) JumpToLastRow() {
	p := m.song.Patterns[m.editPattern]
	patternHeight := len(p)
	m.moveBrush(0, patternHeight-m.brush.H-m.brush.Y)
}

func (m *AppModel) JumpToTopLeft() {
	m.moveBrush(-m.brush.X, -m.brush.Y)
}

func (m *AppModel) JumpToBottomRight() {
	p := m.song.Patterns[m.editPattern]
	patternHeight := len(p)
	editRow := p[m.editPos.Y]
	patternWidth := len(editRow) * 6
	m.moveBrush(
		patternWidth-m.brush.W-m.brush.X,
		patternHeight-m.brush.H-m.brush.Y,
	)
}

func (m *AppModel) Left() {
	m.moveBrush(-m.brush.W, 0)
}

func (m *AppModel) Right() {
	m.moveBrush(m.brush.W, 0)
}

func (m *AppModel) NextTrack() {
	m.moveBrush(6, 0)
}

func (m *AppModel) PrevTrack() {
	m.moveBrush(-6, 0)
}

func (m *AppModel) InsertBlank() {
	p := m.song.Patterns[m.editPattern]
	editRow := p[m.editPos.Y]
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

func (m *AppModel) DeleteLeft() {
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

func (m *AppModel) SelectNone() {
	m.selection = Rect{}
}

func (m *AppModel) SelectAll() {
	p := m.song.Patterns[m.editPattern]
	patternHeight := len(p)
	editRow := p[m.editPos.Y]
	patternWidth := len(editRow) * 6
	m.selection = Rect{0, 0, patternWidth, patternHeight}
}

func (m *AppModel) stepBrushWidth(expandDir int) {
	p := m.song.Patterns[m.editPattern]
	editRow := p[m.editPos.Y]
	patternWidth := len(editRow) * 6
	if expandDir == m.brush.ExpandDir.X {
		switch m.brush.W {
		case 1:
			m.brush.X = m.editPos.X - (m.editPos.X % 2)
			m.brush.W = 2
		case 2:
			m.brush.X = m.editPos.X - (m.editPos.X % 6)
			m.brush.W = 6
		case 6:
			m.brush.X = 0
			m.brush.W = patternWidth
		}
	} else {
		switch m.brush.W {
		case patternWidth:
			m.brush.X = m.editPos.X - (m.editPos.X % 6)
			m.brush.W = 6
		case 6:
			m.brush.X = m.editPos.X - (m.editPos.X % 2)
			m.brush.W = 2
		case 2:
			m.brush.X = m.editPos.X
			m.brush.W = 1
		case 1:
			m.brush.X = m.editPos.X - (m.editPos.X % 2)
			m.brush.W = 2
			m.brush.ExpandDir.X = expandDir
		}
	}
	m.SelectNone()
}

func (m *AppModel) IncBrushWidth() {
	m.stepBrushWidth(1)
}

func (m *AppModel) DecBrushWidth() {
	m.stepBrushWidth(-1)
}

func (m *AppModel) stepBrushHeight(expandDir int) {
	p := m.song.Patterns[m.editPattern]
	patternHeight := len(p)
	if expandDir == m.brush.ExpandDir.Y {
		switch m.brush.H {
		case 1:
			m.brush.Y = m.editPos.Y - (m.editPos.Y % m.song.LPB)
			m.brush.H = m.song.LPB
		case m.song.LPB:
			m.brush.Y = 0
			m.brush.H = patternHeight
		}
	} else {
		switch m.brush.H {
		case patternHeight:
			m.brush.Y = m.editPos.Y - (m.editPos.Y % m.song.LPB)
			m.brush.H = m.song.LPB
		case m.song.LPB:
			m.brush.Y = m.editPos.Y
			m.brush.H = 1
		case 1:
			m.brush.Y = m.editPos.Y - (m.editPos.Y % m.song.LPB)
			m.brush.H = m.song.LPB
			m.brush.ExpandDir.Y = expandDir
		}
	}
	m.SelectNone()
}

func (m *AppModel) IncBrushHeight() {
	m.stepBrushHeight(1)
}

func (m *AppModel) DecBrushHeight() {
	m.stepBrushHeight(-1)
}

func (m *AppModel) InsertBlockV() {
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
	clone.zeroBlock(m.brush.Rect)
	m.submitAction(
		func() {
			m.ReplaceEditPattern(clone)
		},
		func() {
			m.ReplaceEditPattern(p)
		},
	)
}

func (m *AppModel) DeleteBlockV() {
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
			m.ReplaceEditPattern(clone)
		},
		func() {
			m.ReplaceEditPattern(p)
		},
	)
}

func (m *AppModel) InsertBlockH() {
	p := m.song.Patterns[m.editPattern]
	clone := p.clone()
	editRow := p[m.editPos.Y]
	patternWidth := len(editRow) * 6
	blockToMove := Rect{
		m.brush.X,
		m.brush.Y,
		patternWidth - m.brush.X - m.brush.W,
		m.brush.H,
	}
	clone.copyBlock(blockToMove, m.brush.W, 0)
	clone.zeroBlock(m.brush.Rect)
	m.submitAction(
		func() {
			m.ReplaceEditPattern(clone)
		},
		func() {
			m.ReplaceEditPattern(p)
		},
	)
}

func (m *AppModel) DeleteBlockH() {
	p := m.song.Patterns[m.editPattern]
	clone := p.clone()
	editRow := p[m.editPos.Y]
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
			m.ReplaceEditPattern(clone)
		},
		func() {
			m.ReplaceEditPattern(p)
		},
	)
}

func (m *AppModel) PlayOrStop() {
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

func (m *AppModel) SetPlayFromRow() {
	m.playFromRow = m.editPos.Y
}

func (m *AppModel) EnterCommand() {
	m.mode = CommandMode
	m.commandModel.Focus()
}

func (m *AppModel) LoadSong() {
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

func (m *AppModel) SaveSong() {
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

func (m *AppModel) Undo() {
	if len(m.undoableActions) == 0 {
		return
	}
	lastAction := m.undoableActions[len(m.undoableActions)-1]
	m.undoableActions = m.undoableActions[:len(m.undoableActions)-1]
	m.undoneActions = append(m.undoneActions, lastAction)
	m.submitAction(lastAction.undoFn, nil)
}

func (m *AppModel) Redo() {
	if len(m.undoneActions) == 0 {
		return
	}
	lastAction := m.undoneActions[len(m.undoneActions)-1]
	m.undoneActions = m.undoneActions[:len(m.undoneActions)-1]
	m.submitAction(lastAction.doFn, lastAction.undoFn)
}
