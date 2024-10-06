package main

import (
	"slices"
	"strconv"
	"strings"
)

func (m *model) ExecuteCommand(command string) {
	items := strings.Fields(command)
	if len(items) == 0 {
		return
	}
	switch items[0] {
	case "r", "read":
		if len(items) > 1 {
			m.filename = items[1]
		}
		m.LoadSong()
	case "w", "write":
		if len(items) > 1 {
			m.filename = items[1]
		}
		m.SaveSong()
	case "bpm":
		if len(items) > 1 {
			bpm, err := strconv.Atoi(items[1])
			if err != nil {
				m.SetError(err)
				return
			}
			m.song.BPM = bpm
		}
	case "lpb":
		if len(items) > 1 {
			lpb, err := strconv.Atoi(items[1])
			if err != nil {
				m.SetError(err)
				return
			}
			m.song.LPB = lpb
		}
	case "tpl":
		if len(items) > 1 {
			tpl, err := strconv.Atoi(items[1])
			if err != nil {
				m.SetError(err)
				return
			}
			m.song.TPL = tpl
		}
	case "rows":
		if len(items) > 1 {
			rows, err := strconv.Atoi(items[1])
			if err != nil {
				m.SetError(err)
				return
			}
			p := m.song.Patterns[m.editPattern]
			if rows == len(p) {
				return
			}
			clone := p.clone()
			if rows < len(clone) {
				clone = clone[:rows]
			} else {
				clone = slices.Grow(clone, rows-len(p))
				trackCount := len(p[0])
				for i := 0; i < rows-len(p); i++ {
					clone = append(clone, make([]MidiMessage, trackCount))
				}
			}
			m.submitAction(
				func() {
					m.ReplaceEditPattern(clone)
				},
				func() {
					m.ReplaceEditPattern(p)
				},
			)
		}
	}
}
