package main

import (
	"fmt"
	"github.com/charmbracelet/lipgloss"
	"github.com/lucasb-eyer/go-colorful"
)

type Colors struct {
	chromeFill  colorful.Color
	chromeText  colorful.Color
	headerLabel colorful.Color
	headerValue colorful.Color
	border      colorful.Color
	trackLabel  colorful.Color
	patternNum  colorful.Color
	patternText colorful.Color

	cursorFill colorful.Color
	cursorText colorful.Color
	brushFill  colorful.Color
	brushText  colorful.Color
	selectFill colorful.Color
	selectText colorful.Color

	playFill colorful.Color
	playText colorful.Color
	beatFill colorful.Color
	beatText colorful.Color

	errorFill colorful.Color
	errorText colorful.Color
}

var colors Colors

const (
	cursorBit = 1
	brushBit  = 2
	selectBit = 4
	playBit   = 8
	beatBit   = 16
)

var patternPalette [32]lipgloss.Style

type Styles struct {
	chrome      lipgloss.Style
	headerLabel lipgloss.Style
	headerValue lipgloss.Style
	border      lipgloss.Style
	trackLabel  lipgloss.Style
	patternNum  lipgloss.Style
	error       lipgloss.Style
}

var styles Styles

func modColor(color colorful.Color, dc, dl float64) colorful.Color {
	h, c, l := color.Hcl()
	return colorful.Hcl(h, c+dc, l+dl).Clamped()
}

func LGC(color colorful.Color) lipgloss.Color {
	return lipgloss.Color(color.Hex())
}

func init() {
	colors.chromeFill = colorful.Hcl(250, 0.03, 0.14)
	colors.chromeText = colorful.Hcl(250, 0.02, 0.88)
	colors.headerLabel = colorful.Hcl(250, 0.01, 0.70)
	colors.headerValue = colorful.Hcl(250, 0.02, 0.88)
	colors.border = modColor(colors.chromeFill, 0, 0.10)
	colors.trackLabel = modColor(colors.headerLabel, 0, -0.20)
	colors.patternNum = colors.trackLabel
	colors.patternText = colors.chromeText

	colors.cursorFill = colorful.Hcl(220, 0.12, 0.40)
	colors.cursorText = colorful.Hcl(220, 0.10, 0.88)
	colors.brushFill = modColor(colors.cursorFill, -0.02, 0)
	colors.brushText = colors.cursorText
	colors.selectFill = modColor(colors.brushFill, -0.02, 0)
	colors.selectText = colors.cursorText

	colors.playFill = colorful.Hcl(40, 0.12, 0.40)
	colors.playText = colors.cursorText
	colors.beatFill = colorful.Hcl(40, 0.06, 0.20)
	colors.beatText = colors.cursorText

	colors.errorFill = colorful.Hcl(25, 0.14, 0.40)
	colors.errorText = colorful.Hcl(25, 0.14, 0.66)

	for i := range len(patternPalette) {
		text := colors.patternText
		fill := colors.chromeFill
		if i&cursorBit > 0 {
			text = text.BlendHcl(colors.cursorText, 0.5)
			fill = fill.BlendHcl(colors.cursorFill, 0.5)
		}
		if i&selectBit > 0 {
			text = text.BlendHcl(colors.selectText, 0.5)
			fill = fill.BlendHcl(colors.selectFill, 0.5)
		}
		if i&brushBit > 0 {
			text = text.BlendHcl(colors.brushText, 0.5)
			fill = fill.BlendHcl(colors.brushFill, 0.5)
		}
		if i&playBit > 0 {
			text = text.BlendHcl(colors.playText, 0.5)
			fill = fill.BlendHcl(colors.playFill, 0.5)
		}
		if i&beatBit > 0 {
			text = text.BlendHcl(colors.beatText, 0.5)
			fill = fill.BlendHcl(colors.beatFill, 0.5)
		}
		patternPalette[i] = lipgloss.Style{}.Foreground(LGC(text)).Background(LGC(fill))
	}

	styles.chrome = lipgloss.NewStyle().
		Foreground(LGC(colors.chromeText)).
		Background(LGC(colors.chromeFill))

	styles.headerLabel = lipgloss.NewStyle().
		Foreground(LGC(colors.headerLabel))

	styles.headerValue = lipgloss.NewStyle().
		Foreground(LGC(colors.headerValue))

	styles.border = lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		Foreground(LGC(colors.border)).
		Background(LGC(colors.chromeFill)).
		BorderForeground(LGC(colors.border)).
		BorderBackground(LGC(colors.chromeFill))

	styles.trackLabel = lipgloss.NewStyle().
		Foreground(LGC(colors.trackLabel)).
		Background(LGC(colors.chromeFill))

	styles.patternNum = lipgloss.NewStyle().
		Foreground(LGC(colors.patternNum)).
		Background(LGC(colors.chromeFill))

	styles.error = lipgloss.NewStyle().
		Foreground(LGC(colors.errorText)).
		Background(LGC(colors.errorFill))
}

const hexDigits = "0123456789ABCDEF"

func (m *Model) HeaderView() string {
	var rb RowBuilder
	rb.SetStyle(&styles.headerLabel)
	rb.WriteString("BPM:")
	rb.SetStyle(&styles.headerValue)
	rb.WriteString(fmt.Sprintf("%d", m.song.BPM))
	rb.WriteByte(' ')
	rb.SetStyle(&styles.headerLabel)
	rb.WriteString("SR:")
	rb.SetStyle(&styles.headerValue)
	rb.WriteString(fmt.Sprintf("%d", m.GetSampleRate()))
	rb.WriteByte(' ')
	rb.SetStyle(&styles.headerLabel)
	rb.WriteString("FPB:")
	rb.SetStyle(&styles.headerValue)
	rb.WriteString(fmt.Sprintf("%d", m.GetFramesPerBeat()))
	rb.WriteByte(' ')
	rb.SetStyle(&styles.headerLabel)
	rb.WriteString("TPB:")
	rb.SetStyle(&styles.headerValue)
	rb.WriteString(fmt.Sprintf("%d", m.GetTicksPerBeat()))
	rb.WriteByte(' ')
	rb.SetStyle(&styles.headerLabel)
	rb.WriteString("FPT:")
	rb.SetStyle(&styles.headerValue)
	rb.WriteString(fmt.Sprintf("%d", m.GetFramesPerTick()))
	rb.WriteByte(' ')
	rb.SetStyle(&styles.headerLabel)
	rb.WriteString("LPB:")
	rb.SetStyle(&styles.headerValue)
	rb.WriteString(fmt.Sprintf("%d", m.song.LPB))
	rb.WriteByte(' ')
	rb.SetStyle(&styles.headerLabel)
	rb.WriteString("TPL:")
	rb.SetStyle(&styles.headerValue)
	rb.WriteString(fmt.Sprintf("%d", m.song.TPL))
	rb.WriteByte(' ')
	rb.SetStyle(&styles.headerLabel)
	rb.WriteString("BRU:")
	rb.SetStyle(&styles.headerValue)
	rb.WriteString(fmt.Sprintf("(%d,%d)",
		m.brush.W,
		m.brush.H,
	))
	return rb.String()
}

func (m *Model) PatternView(r Rect) string {
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
	visibleTracks := min(maxVisibleTracks, patternWidth)
	numCols = visibleTracks * 6
	if m.brush.X+m.brush.W > m.firstVisibleTrack*6+numCols {
		m.firstVisibleTrack = (m.brush.X + m.brush.W - numCols + 6) / 6
	}
	if m.brush.X < m.firstVisibleTrack*6 {
		m.firstVisibleTrack = m.brush.X / 6
	}
	if m.firstVisibleTrack+visibleTracks > patternWidth {
		m.firstVisibleTrack = patternWidth - visibleTracks
	}
	if m.firstVisibleTrack < 0 {
		m.firstVisibleTrack = 0
	}
	if m.brush.Y+m.brush.H > m.firstVisibleRow+numRows {
		m.firstVisibleRow = m.brush.Y + m.brush.H - numRows
	}
	if m.brush.Y < m.firstVisibleRow {
		m.firstVisibleRow = m.brush.Y
	}
	if m.firstVisibleRow+numRows > patternHeight {
		m.firstVisibleRow = patternHeight - numRows
	}
	if m.firstVisibleRow < 0 {
		m.firstVisibleRow = 0
	}
	var rb RowBuilder
	rowStrings := make([]string, 0, numRows)
	for y := m.firstVisibleRow; y < min(patternHeight, m.firstVisibleRow+numRows); y++ {
		row := p[y]
		rb.SetStyle(&styles.patternNum)
		rb.WriteString(fmt.Sprintf("%04X", y))
		rb.WriteByte(' ')
		rowStyleIndex := 0
		if y == m.playRow {
			rowStyleIndex |= playBit
		}
		x := m.firstVisibleTrack * 6
		for t := m.firstVisibleTrack; t < min(patternWidth, m.firstVisibleTrack+visibleTracks); t++ {
			rb.SetStyle(&patternPalette[rowStyleIndex])
			if t > m.firstVisibleTrack {
				rb.WriteByte(' ')
			}
			msg := row[t]
			for i := range 3 {
				for j := range 2 {
					cellStyleIndex := rowStyleIndex
					if y == m.editPos.Y && x == m.editPos.X {
						cellStyleIndex |= cursorBit
					}
					insideBrush := x >= m.brush.X &&
						y >= m.brush.Y &&
						x < (m.brush.X+m.brush.W) &&
						y < (m.brush.Y+m.brush.H)
					if insideBrush {
						cellStyleIndex |= brushBit
					}
					rb.SetStyle(&patternPalette[cellStyleIndex])
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
	withoutTopBorderStyle := styles.border.
		BorderTop(false).
		BorderRight(true).
		BorderBottom(true).
		BorderLeft(true).
		Padding(0, 1)
	patternWithoutTopBorder := withoutTopBorderStyle.Render(lipgloss.JoinVertical(0, rowStrings...))
	topBorderStyle := styles.border.Border(lipgloss.Border{}, false)
	rb.SetStyle(&topBorderStyle)
	roundedBorder := lipgloss.RoundedBorder()
	rb.WriteString(roundedBorder.TopLeft)
	for range 1 + 4 + 1 {
		// padding + row index + gap
		rb.WriteString(roundedBorder.Top)
	}
	for t := m.firstVisibleTrack; t < min(patternWidth, m.firstVisibleTrack+visibleTracks); t++ {
		if t > m.firstVisibleTrack {
			rb.WriteString(roundedBorder.Top)
		}
		rb.WriteString(roundedBorder.Top)
		rb.WriteString("╴")
		rb.SetStyle(&styles.trackLabel)
		rb.WriteString(fmt.Sprintf("%02X", t))
		rb.SetStyle(&topBorderStyle)
		rb.WriteString("╶")
		rb.WriteString(roundedBorder.Top)
	}
	rb.WriteString(roundedBorder.Top)
	rb.WriteString(roundedBorder.TopRight)
	topBorder := rb.String()
	rb.Reset()
	return lipgloss.JoinVertical(0, topBorder, patternWithoutTopBorder)
}

func (m *Model) CommandView() string {
	return m.commandModel.View()
}

func (m *Model) ErrorView() string {
	return styles.error.Render(fmt.Sprintf("%s", m.err))
}

func (m *Model) View() string {
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
	return lipgloss.JoinVertical(lipgloss.Left, views...)
}
