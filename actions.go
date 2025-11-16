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
	numPatterns := len(patterns)

	if m.editPattern < 0 {
		m.editPattern = 0
	} else if m.editPattern >= numPatterns {
		m.editPattern = numPatterns - 1
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

	// fix selection
	if m.sel.W > patternWidth {
		m.sel.W = patternWidth
	}
	if m.sel.X+m.sel.W > patternWidth {
		m.sel.X = patternWidth - m.sel.W
	}
	if m.sel.X < 0 || m.sel.X >= patternWidth {
		m.CollapseSelection()
	}
	if m.sel.H > patternHeight {
		m.sel.H = patternHeight
	}
	if m.sel.Y+m.sel.H > patternHeight {
		m.sel.Y = patternHeight - m.sel.H
	}
	if m.sel.Y < 0 || m.sel.Y >= patternHeight {
		m.CollapseSelection()
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
	m.revertWideBrush()
	m.CollapseSelection()
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

func (m *Model) DeleteBrush() {
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

func (m *Model) stepBrushWidth(expandDir int) {
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
	m.CollapseSelection()
}

func (m *Model) IncBrushWidth() {
	m.stepBrushWidth(1)
}

func (m *Model) DecBrushWidth() {
	m.stepBrushWidth(-1)
}

func (m *Model) stepBrushHeight(expandDir int) {
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
	m.CollapseSelection()
}

func (m *Model) IncBrushHeight() {
	m.stepBrushHeight(1)
}

func (m *Model) DecBrushHeight() {
	m.stepBrushHeight(-1)
}

func (m *Model) IncSelectionWidth() {
	sel := m.sel
	if m.sel.X+m.sel.W == m.brush.X+m.brush.W {
		// brush is at right side of current selection
		m.Right()
		m.sel = sel
		m.sel.W = m.brush.X + m.brush.W - m.sel.X
	} else {
		m.Right()
		m.sel = sel
		m.sel.X = m.brush.X
		m.sel.W = sel.X + sel.W - m.sel.X
	}
}

func (m *Model) DecSelectionWidth() {
	sel := m.sel
	if m.sel.X == m.brush.X {
		// brush is at left side of current selection
		m.Left()
		m.sel = sel
		m.sel.X = m.brush.X
		m.sel.W = sel.X + sel.W - m.sel.X
	} else {
		m.Left()
		m.sel = sel
		m.sel.W = m.brush.X + m.brush.W - m.sel.X
	}
}

func (m *Model) applyWideBrush() {
	if m.brush.W == 1 && m.sel.W == 1 {
		for m.brush.W < 6 {
			m.stepBrushWidth(1)
		}
		m.usingWideBrush = true
	}
}

func (m *Model) revertWideBrush() {
	if m.usingWideBrush {
		m.CollapseBrush()
		m.usingWideBrush = false
	}
}

func (m *Model) IncSelectionHeight() {
	m.applyWideBrush()
	sel := m.sel
	if m.sel.Y+m.sel.H == m.brush.Y+m.brush.H {
		// brush is at bottom side of current selection
		m.Down()
		m.sel = sel
		m.sel.H = m.brush.Y + m.brush.H - m.sel.Y
	} else {
		m.Down()
		m.sel = sel
		m.sel.Y = m.brush.Y
		m.sel.H = sel.Y + sel.H - m.sel.Y
	}
}

func (m *Model) DecSelectionHeight() {
	m.applyWideBrush()
	sel := m.sel
	if m.sel.Y == m.brush.Y {
		// brush os at top side of current selection
		m.Up()
		m.sel = sel
		m.sel.Y = m.brush.Y
		m.sel.H = sel.Y + sel.H - m.sel.Y
	} else {
		m.Up()
		m.sel = sel
		m.sel.H = m.brush.Y + m.brush.H - m.sel.Y
	}
}

func (m *Model) InsertBlock() {
	m.applyWideBrush()
	p := m.song.Patterns[m.editPattern]
	clone := p.clone()
	patternHeight := len(p)
	blockToMove := Rect{
		m.sel.X,
		m.sel.Y,
		m.sel.W,
		patternHeight - m.sel.Y - m.sel.H,
	}
	clone.copyBlock(blockToMove, 0, m.sel.H)
	clone.zeroBlock(m.sel)
	m.submitAction(
		func() {
			m.ReplaceEditPattern(clone)
		},
		func() {
			m.ReplaceEditPattern(p)
		},
	)
}

func (m *Model) DeleteBlock(backspace bool) {
	if backspace && m.sel.Y < m.sel.H {
		return
	}
	m.applyWideBrush()
	p := m.song.Patterns[m.editPattern]
	clone := p.clone()
	patternHeight := len(p)
	oldSel := m.sel
	sel := oldSel
	oldBrush := m.brush
	brush := oldBrush
	oldEditPos := m.editPos
	editPos := oldEditPos
	if backspace {
		sel.Y -= sel.H
		brush.Y -= sel.H
		editPos.Y -= sel.H
	}
	blockToMove := Rect{
		sel.X,
		sel.Y + sel.H,
		sel.W,
		patternHeight - sel.Y - sel.H,
	}
	clone.copyBlock(blockToMove, 0, -sel.H)
	blockToZero := Rect{
		sel.X,
		patternHeight - sel.H,
		sel.W,
		sel.H,
	}
	clone.zeroBlock(blockToZero)
	m.submitAction(
		func() {
			m.sel = sel
			m.brush = brush
			m.editPos = editPos
			m.ReplaceEditPattern(clone)
		},
		func() {
			m.ReplaceEditPattern(p)
			m.sel = oldSel
			m.brush = oldBrush
			m.editPos = oldEditPos
		},
	)
}

func (m *Model) CurrentTrack() int {
	return m.editPos.X / 6
}

func (m *Model) InsertTrack() {
	p := m.song.Patterns[m.editPattern]
	clone := p.insertTrack(m.CurrentTrack())
	m.submitAction(
		func() {
			m.ReplaceEditPattern(clone)
		},
		func() {
			m.ReplaceEditPattern(p)
		},
	)
}

func (m *Model) DeleteTrack() {
	p := m.song.Patterns[m.editPattern]
	editRow := p[m.editPos.Y]
	numTracks := len(editRow)
	if numTracks == 1 {
		return
	}
	clone := p.deleteTrack(m.CurrentTrack())
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
	m.applyWideBrush()
	p := m.song.Patterns[m.editPattern]
	sel := m.sel
	block := p.getBlock(sel)
	m.submitAction(
		func() {
			m.clipboard = block
			m.pasteOffset = sel.X % 6
			p.zeroBlock(sel)
		},
		func() {
			p := m.song.Patterns[m.editPattern]
			p.setBlock(sel, block)
		},
	)
}

func (m *Model) Copy() {
	m.applyWideBrush()
	p := m.song.Patterns[m.editPattern]
	sel := m.sel
	block := p.getBlock(sel)
	m.submitAction(
		func() {
			m.clipboard = block
			m.pasteOffset = sel.X % 6
		},
		nil,
	)
}

func (m *Model) pasteBlock(pos Point, block Block) (prevBlock Block) {
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
	editRow := p[pos.Y]
	patternWidth := len(editRow) * 6
	rect := Rect{
		X: pos.X - pos.X%6 + m.pasteOffset,
		Y: pos.Y,
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
	pos := m.editPos
	m.submitAction(
		func() {
			prevBlock = m.pasteBlock(pos, m.clipboard)
		},
		func() {
			m.pasteBlock(pos, prevBlock)
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
	m.submitAction(
		func() {
			m.playFromRow = m.editPos.Y
			if !m.isPlaying {
				m.playRow = m.playFromRow
			}
		},
		nil,
	)
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
	m.submitAction(
		func() {
			m.SetSong(song)
		},
		nil,
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
