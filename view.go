package main

import (
	"fmt"
	"github.com/charmbracelet/lipgloss"
)

var (
	patternStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			Padding(0, 1)
	trackStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#cccccc"))
	highlightStyle = lipgloss.NewStyle().Inline(true).
			Background(lipgloss.Color("#cccccc")).
			Foreground(lipgloss.Color("#ffffff"))
	rowStyle = lipgloss.NewStyle().Inline(true).
			Background(lipgloss.Color("#222222")).
			Foreground(lipgloss.Color("#cccccc"))
	playRowStyle = lipgloss.NewStyle().Inline(true).
			Background(lipgloss.Color("#008000")).
			Foreground(lipgloss.Color("#ffffff"))
	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#ffffff")).
			Background(lipgloss.Color("#ff0000"))
)

const hexDigits = "0123456789ABCDEF"

func (m *model) HeaderView() string {
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
	return rb.String()
}

func (m *model) PatternView(patternHeight int) string {
	p := m.song.Patterns[m.editPattern]
	var rb RowBuilder
	numRows := patternHeight - 2 // borders
	if numRows <= 0 {
		return ""
	}
	if m.editRow < m.editRow0 {
		m.editRow0 = m.editRow
	}
	if m.editRow >= m.editRow0+numRows {
		m.editRow0 = m.editRow - numRows + 1
	}
	rowStrings := make([]string, 0, numRows)
	var oldStyle lipgloss.Style
	for y := m.editRow0; y < len(p) && y < m.editRow0+numRows; y++ {
		row := p[y]
		isCursorInRow := m.editRow == y
		rb.SetStyle(rowStyle)
		rb.WriteByte(hexDigits[y>>4])
		rb.WriteByte(hexDigits[y&0x0f])
		rb.WriteByte(' ')
		if y == m.playRow {
			rb.SetStyle(playRowStyle)
		}
		for x := m.editTrack0; x < len(row); x++ {
			if x > m.editTrack0 {
				rb.WriteByte(' ')
			}
			isCursorInTrack := m.editTrack == x
			msg := row[x]
			for i := 0; i < 3; i++ {
				for j := 0; j < 2; j++ {
					shallHighlight := isCursorInRow && isCursorInTrack && m.editColumn == i*2+j
					if shallHighlight {
						oldStyle = rb.style
						rb.SetStyle(highlightStyle)
					}
					b := msg[i]
					if b == 0 {
						rb.WriteRune('·')
					} else {
						shiftBits := (1 - j) * 4
						rb.WriteByte(hexDigits[(b>>shiftBits)&0x0f])
					}
					if shallHighlight {
						rb.SetStyle(oldStyle)
					}
				}
			}
		}
		rowStrings = append(rowStrings, rb.String())
		rb.Reset()
	}
	return patternStyle.Render(lipgloss.JoinVertical(0, rowStrings...))
}

func (m *model) CommandView() string {
	return m.commandModel.View()
}

func (m *model) ErrorView() string {
	return errorStyle.Render(fmt.Sprintf("%s", m.err))
}

func (m *model) View() string {
	patternHeight := m.windowHeight - 1
	if m.mode == CommandMode {
		patternHeight--
	}
	if m.err != nil {
		patternHeight--
	}
	if patternHeight <= 0 {
		return ""
	}
	var views []string
	views = append(views, m.HeaderView())
	views = append(views, m.PatternView(patternHeight))
	if m.mode == CommandMode {
		views = append(views, m.CommandView())
	}
	if m.err != nil {
		views = append(views, m.ErrorView())
	}
	return lipgloss.JoinVertical(lipgloss.Top, views...)
}
