package main

import (
	"encoding/json"
	"fmt"
	"os"
)

func (m *Model) submitAction(doFn ActionFunction, undoFn ActionFunction) {
	m.pendingActions <- Action{doFn, undoFn}
}

func (m *Model) fix() {
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
	if m.brush.W > patternWidth {
		m.brush.W = patternWidth
	}
	if m.brush.X+m.brush.W > patternWidth {
		m.brush.X = patternWidth - m.brush.W
	}
	if m.brush.X < 0 || m.brush.X >= patternWidth {
		m.CollapseBrush()
	}
	if m.brush.H > patternHeight {
		m.brush.H = patternHeight
	}
	if m.brush.Y+m.brush.H > patternHeight {
		m.brush.Y = patternHeight - m.brush.H
	}
	if m.brush.Y < 0 || m.brush.Y >= patternHeight {
		m.CollapseBrush()
	}
}

func (m *Model) moveBrush(dx, dy int) {
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

func (m *Model) Up() {
	m.moveBrush(0, -m.brush.H)
}

func (m *Model) Down() {
	m.moveBrush(0, m.brush.H)
}

func (m *Model) PageUp() {
	m.moveBrush(0, -max(m.song.LPB, m.brush.H))
}

func (m *Model) PageDown() {
	m.moveBrush(0, max(m.song.LPB, m.brush.H))
}

func (m *Model) JumpToFirstRow() {
	m.moveBrush(0, -m.brush.Y)
}

func (m *Model) JumpToLastRow() {
	p := m.song.Patterns[m.editPattern]
	patternHeight := len(p)
	m.moveBrush(0, patternHeight-m.brush.H-m.brush.Y)
}

func (m *Model) JumpToTopLeft() {
	m.moveBrush(-m.brush.X, -m.brush.Y)
}

func (m *Model) JumpToBottomRight() {
	p := m.song.Patterns[m.editPattern]
	patternHeight := len(p)
	editRow := p[m.editPos.Y]
	patternWidth := len(editRow) * 6
	m.moveBrush(
		patternWidth-m.brush.W-m.brush.X,
		patternHeight-m.brush.H-m.brush.Y,
	)
}

func (m *Model) Left() {
	m.moveBrush(-m.brush.W, 0)
}

func (m *Model) Right() {
	m.moveBrush(m.brush.W, 0)
}

func (m *Model) NextTrack() {
	m.moveBrush(6, 0)
}

func (m *Model) PrevTrack() {
	m.moveBrush(-6, 0)
}

func (m *Model) InsertBlank() {
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

func (m *Model) DeleteLeft() {
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

func (m *Model) stepBrushWidthExp(expandDir int) {
	p := m.song.Patterns[m.editPattern]
	editRow := p[m.editPos.Y]
	patternWidth := len(editRow) * 6
	if expandDir == m.brush.ExpandDir.X {
		switch {
		case m.brush.W < 2:
			m.brush.W = 2
		case m.brush.W < 6:
			m.brush.W = 6
		default:
			m.brush.W = patternWidth
		}
	} else {
		switch {
		case m.brush.W > 6:
			m.brush.W = 6
		case m.brush.W > 2:
			m.brush.W = 2
		case m.brush.W == 2:
			m.brush.W = 1
		default:
			m.brush.W = 2
			m.brush.ExpandDir.X = expandDir
		}
	}
	m.brush.X = m.editPos.X - (m.editPos.X % m.brush.W)
}

func (m *Model) IncBrushWidthExp() {
	m.stepBrushWidthExp(1)
}

func (m *Model) DecBrushWidthExp() {
	m.stepBrushWidthExp(-1)
}

func (m *Model) stepBrushHeightExp(expandDir int) {
	p := m.song.Patterns[m.editPattern]
	patternHeight := len(p)
	if expandDir == m.brush.ExpandDir.Y {
		switch {
		case m.brush.H < m.song.LPB:
			m.brush.H = m.song.LPB
		default:
			m.brush.H = patternHeight
		}
	} else {
		switch {
		case m.brush.H > m.song.LPB:
			m.brush.H = m.song.LPB
		case m.brush.H > 1:
			m.brush.H = 1
		default:
			m.brush.H = m.song.LPB
			m.brush.ExpandDir.Y = expandDir
		}
	}
	m.brush.Y = m.editPos.Y - (m.editPos.Y % m.brush.H)
}

func (m *Model) IncBrushHeightExp() {
	m.stepBrushHeightExp(1)
}

func (m *Model) DecBrushHeightExp() {
	m.stepBrushHeightExp(-1)
}

func (m *Model) IncBrushWidthLin() {
	p := m.song.Patterns[m.editPattern]
	editRow := p[m.editPos.Y]
	patternWidth := len(editRow) * 6
	brushWidthAdd := 0
	switch {
	case m.brush.W < 2:
		brushWidthAdd = 1
	case m.brush.W < 6:
		brushWidthAdd = 2
	default:
		brushWidthAdd = 6
	}
	if m.brush.X+m.brush.W+brushWidthAdd <= patternWidth {
		m.brush.W += brushWidthAdd
	}
}

func (m *Model) DecBrushWidthLin() {
	brushWidthAdd := 0
	switch {
	case m.brush.W > 6:
		brushWidthAdd = -6
	case m.brush.W > 2:
		brushWidthAdd = -2
	case m.brush.W == 2:
		brushWidthAdd = -1
	}
	m.brush.W += brushWidthAdd
}

func (m *Model) IncBrushHeightLin() {
	p := m.song.Patterns[m.editPattern]
	patternHeight := len(p)
	brushHeightAdd := 0
	switch {
	case m.brush.H < m.song.LPB:
		brushHeightAdd = 1
	default:
		brushHeightAdd = m.song.LPB
	}
	if m.brush.Y+m.brush.H+brushHeightAdd <= patternHeight {
		m.brush.H += brushHeightAdd
	}
}

func (m *Model) DecBrushHeightLin() {
	brushHeightAdd := 0
	switch {
	case m.brush.H > m.song.LPB:
		brushHeightAdd = -m.song.LPB
	case m.brush.H > 1:
		brushHeightAdd = -1
	}
	m.brush.H += brushHeightAdd
}

func (m *Model) InsertBlockV() {
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

func (m *Model) DeleteBlockV() {
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

func (m *Model) InsertBlockH() {
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

func (m *Model) DeleteBlockH() {
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

func (m *Model) Cut() {
	m.submitAction(
		func() {
			m.clipboard = m.getBlock()
			m.zeroBlock()
		},
		func() {
			m.setBlock(m.clipboard)
		},
	)
}

func (m *Model) Copy() {
	m.submitAction(
		func() {
			m.clipboard = m.getBlock()
		},
		nil,
	)
}

func (m *Model) pasteBlock(block Block) (prevBlock Block) {
	if block == nil {
		return
	}
	blockH := len(block)
	if blockH == 0 {
		return
	}
	blockW := len(block[0])
	if blockW == 0 {
		return
	}
	p := m.song.Patterns[m.editPattern]
	patternHeight := len(p)
	editRow := p[m.editPos.Y]
	patternWidth := len(editRow) * 6
	rect := Rect{
		X: m.editPos.X,
		Y: m.editPos.Y,
		W: blockW,
		H: blockH,
	}
	if rect.X+rect.W > patternWidth {
		rect.W = patternWidth - rect.X
	}
	if rect.Y+rect.H > patternHeight {
		rect.H = patternHeight - rect.Y
	}
	prevBlock = p.getBlock(rect)
	p.setBlock(rect, block)
	return prevBlock
}

func (m *Model) Paste() {
	var prevBlock Block
	m.submitAction(
		func() {
			prevBlock = m.pasteBlock(m.clipboard)
		},
		func() {
			m.pasteBlock(prevBlock)
		},
	)
}

func (m *Model) PlayOrStop() {
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

func (m *Model) SetPlayFromRow() {
	m.playFromRow = m.editPos.Y
}

func (m *Model) EnterCommand() {
	m.prevmode = m.mode
	m.mode = CommandMode
	m.commandModel.Focus()
}

func (m *Model) LoadSong() {
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

func (m *Model) SaveSong() {
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

func (m *Model) Undo() {
	if len(m.undoableActions) == 0 {
		return
	}
	lastAction := m.undoableActions[len(m.undoableActions)-1]
	m.undoableActions = m.undoableActions[:len(m.undoableActions)-1]
	m.undoneActions = append(m.undoneActions, lastAction)
	m.submitAction(lastAction.undoFn, nil)
}

func (m *Model) Redo() {
	if len(m.undoneActions) == 0 {
		return
	}
	lastAction := m.undoneActions[len(m.undoneActions)-1]
	m.undoneActions = m.undoneActions[:len(m.undoneActions)-1]
	m.submitAction(lastAction.doFn, lastAction.undoFn)
}
