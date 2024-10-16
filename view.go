package main

import (
	"fmt"
	"github.com/charmbracelet/lipgloss"
)

type Color struct {
	r, g, b int
}

func (c Color) LGC() lipgloss.Color {
	return lipgloss.Color(fmt.Sprintf("#%02x%02x%02x", c.r, c.g, c.b))
}

const (
	highlightBit = 1
	selectBit    = 2
	brushBit     = 4
	playBit      = 8
)

var palette [17]lipgloss.Style

const (
	BackgroundIndex = 16
)

func init() {
	for i := 0; i < 16; i++ {
		fg := Color{0xcc, 0xcc, 0xcc}
		bg := Color{0x20, 0x20, 0x20}
		if i&highlightBit > 0 {
			fg = Color{0xff, 0xff, 0xff}
			bg = Color{0x80, 0x80, 0x80}
		}
		if i&selectBit > 0 {
			bg.r += 0x40
		}
		if i&brushBit > 0 {
			bg.b += 0x40
		}
		if i&playBit > 0 {
			bg.g += 0x40
		}
		palette[i] = lipgloss.Style{}.Foreground(fg.LGC()).Background(bg.LGC())
	}
	palette[BackgroundIndex] = lipgloss.Style{}
}

const hexDigits = "0123456789ABCDEF"

func (m *AppModel) HeaderView() string {
	var rb RowBuilder
	rb.WriteString("BPM: ")
	rb.WriteString(fmt.Sprintf("%d", m.song.BPM))
	rb.WriteByte(' ')
	rb.WriteString("SR: ")
	rb.WriteString(fmt.Sprintf("%d", m.GetSampleRate()))
	rb.WriteByte(' ')
	rb.WriteString("FPB: ")
	rb.WriteString(fmt.Sprintf("%d", m.GetFramesPerBeat()))
	rb.WriteByte(' ')
	rb.WriteString("TPB: ")
	rb.WriteString(fmt.Sprintf("%d", m.GetTicksPerBeat()))
	rb.WriteByte(' ')
	rb.WriteString("FPT: ")
	rb.WriteString(fmt.Sprintf("%d", m.GetFramesPerTick()))
	rb.WriteByte(' ')
	rb.WriteString("LPB: ")
	rb.WriteString(fmt.Sprintf("%d", m.song.LPB))
	rb.WriteByte(' ')
	rb.WriteString("TPL: ")
	rb.WriteString(fmt.Sprintf("%d", m.song.TPL))
	rb.WriteByte(' ')
	rb.WriteString("isPlaying: ")
	rb.WriteString(fmt.Sprintf("%v", m.isPlaying))
	rb.WriteByte(' ')
	rb.WriteString("playTick: ")
	rb.WriteString(fmt.Sprintf("%d", m.playTick))
	rb.WriteByte(' ')
	rb.WriteString("playFrame: ")
	rb.WriteString(fmt.Sprintf("%v", m.playFrame))
	rb.WriteByte(' ')
	rb.WriteString("SEL: ")
	rb.WriteString(fmt.Sprintf("(%d,%d):(%d,%d)",
		m.selection.X,
		m.selection.Y,
		m.selection.W,
		m.selection.H,
	))
	rb.WriteByte(' ')
	rb.WriteString("BRU: ")
	rb.WriteString(fmt.Sprintf("(%d,%d):(%d,%d)",
		m.brush.X,
		m.brush.Y,
		m.brush.W,
		m.brush.H,
	))
	return rb.String()
}

func (m *AppModel) PatternView(r Rect) string {
	p := m.song.Patterns[m.editPattern]
	patternHeight := len(p)
	editRow := p[m.editPos.Y]
	patternWidth := len(editRow)
	numRows := r.H
	numRows -= 2 // borders
	numCols := r.W
	numCols -= 2 // borders
	numCols -= 2 // padding
	numCols -= 4 // row index
	numCols -= 1 // row index gap
	var maxVisibleTracks int
	if numCols%7 == 6 {
		maxVisibleTracks = (numCols + 1) / 7
	} else {
		maxVisibleTracks = numCols / 7
	}
	if numRows <= 0 || maxVisibleTracks < 1 {
		return ""
	}
	visibleTracks := maxVisibleTracks
	if visibleTracks > patternWidth {
		visibleTracks = patternWidth
	}
	numCols = visibleTracks * 6
	if m.brush.X < m.firstVisibleTrack*6 {
		m.firstVisibleTrack = m.brush.X / 6
	} else if m.brush.X+m.brush.W > m.firstVisibleTrack*6+numCols {
		m.firstVisibleTrack = (m.brush.X + m.brush.W - numCols + 6) / 6
	}
	if m.firstVisibleTrack+visibleTracks > patternWidth {
		m.firstVisibleTrack = patternWidth - visibleTracks
	}
	if m.firstVisibleTrack < 0 {
		m.firstVisibleTrack = 0
	}
	if m.brush.Y < m.firstVisibleRow {
		m.firstVisibleRow = m.brush.Y
	} else if m.brush.Y+m.brush.H > m.firstVisibleRow+numRows {
		m.firstVisibleRow = m.brush.Y + m.brush.H - numRows
	}
	if m.firstVisibleRow+numRows > patternHeight {
		m.firstVisibleRow = patternHeight - numRows
	}
	if m.firstVisibleRow < 0 {
		m.firstVisibleRow = 0
	}
	var rb RowBuilder
	currentStyleIndex := BackgroundIndex
	setStyle := func(index int) {
		if index != currentStyleIndex {
			currentStyleIndex = index
			rb.SetStyle(palette[currentStyleIndex])
		}
	}
	rowStrings := make([]string, 0, numRows)
	for y := m.firstVisibleRow; y < min(patternHeight, m.firstVisibleRow+numRows); y++ {
		row := p[y]
		setStyle(BackgroundIndex)
		rb.WriteString(fmt.Sprintf("%04X", y))
		rb.WriteByte(' ')
		rowStyleIndex := 0
		if y == m.playRow {
			rowStyleIndex |= playBit
		}
		x := m.firstVisibleTrack * 6
		for t := m.firstVisibleTrack; t < min(patternWidth, m.firstVisibleTrack+visibleTracks); t++ {
			setStyle(rowStyleIndex)
			if t > m.firstVisibleTrack {
				rb.WriteByte(' ')
			}
			msg := row[t]
			for i := 0; i < 3; i++ {
				for j := 0; j < 2; j++ {
					cellStyleIndex := rowStyleIndex
					if y == m.editPos.Y && x == m.editPos.X {
						cellStyleIndex |= highlightBit
					}
					insideBrush := x >= m.brush.X &&
						y >= m.brush.Y &&
						x < (m.brush.X+m.brush.W) &&
						y < (m.brush.Y+m.brush.H)
					if insideBrush {
						cellStyleIndex |= brushBit
					}
					insideSelection := m.hasSelection() &&
						x >= m.selection.X &&
						y >= m.selection.Y &&
						x < (m.selection.X+m.selection.W) &&
						y < (m.selection.Y+m.selection.H)
					if insideSelection {
						cellStyleIndex |= selectBit
					}
					setStyle(cellStyleIndex)
					b := msg[i]
					if b == 0 {
						rb.WriteRune('·')
					} else {
						shiftBits := (1 - j) * 4
						rb.WriteByte(hexDigits[(b>>shiftBits)&0x0f])
					}
					x++
				}
			}
		}
		rowStrings = append(rowStrings, rb.String())
		rb.Reset()
	}
	patternStyle := lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderTop(false).
		BorderRight(true).
		BorderBottom(true).
		BorderLeft(true).
		Padding(0, 1)
	patternWithoutTopBorder := patternStyle.Render(lipgloss.JoinVertical(0, rowStrings...))
	setStyle(BackgroundIndex)
	roundedBorder := lipgloss.RoundedBorder()
	rb.WriteString(roundedBorder.TopLeft)
	for i := 0; i < 1+4+1; i++ {
		// padding + row index + gap
		rb.WriteString(roundedBorder.Top)
	}
	for t := m.firstVisibleTrack; t < min(patternWidth, m.firstVisibleTrack+visibleTracks); t++ {
		if t > m.firstVisibleTrack {
			rb.WriteString(roundedBorder.Top)
		}
		rb.WriteString(roundedBorder.Top)
		rb.WriteString("╴")
		rb.WriteString(fmt.Sprintf("%02X", t))
		rb.WriteString("╶")
		rb.WriteString(roundedBorder.Top)
	}
	rb.WriteString(roundedBorder.Top)
	rb.WriteString(roundedBorder.TopRight)
	topBorder := rb.String()
	rb.Reset()
	return lipgloss.JoinVertical(0, topBorder, patternWithoutTopBorder)
}

func (m *AppModel) CommandView() string {
	return m.commandModel.View()
}

func (m *AppModel) ErrorView() string {
	errorStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#ffffff")).
		Background(lipgloss.Color("#ff0000"))
	return errorStyle.Render(fmt.Sprintf("%s", m.err))
}

func (m *AppModel) View() string {
	patternViewWidth := m.windowSize.W
	patternViewHeight := m.windowSize.H - 1
	if m.mode == CommandMode {
		patternViewHeight--
	}
	if m.err != nil {
		patternViewHeight--
	}
	if patternViewWidth <= 0 || patternViewHeight <= 0 {
		return ""
	}
	var views []string
	views = append(views, m.HeaderView())
	views = append(views, m.PatternView(Rect{0, 0, patternViewWidth, patternViewHeight}))
	if m.mode == CommandMode {
		views = append(views, m.CommandView())
	}
	if m.err != nil {
		views = append(views, m.ErrorView())
	}
	return lipgloss.JoinVertical(lipgloss.Top, views...)
}
