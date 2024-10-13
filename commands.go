package main

import (
	"fmt"
	"slices"
	"strconv"
	"strings"
)

func (m *AppModel) ExecuteCommand(command string) {
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
			if bpm < 1 {
				m.SetError(fmt.Errorf("invalid BPM: %d", bpm))
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
			if lpb < 1 {
				m.SetError(fmt.Errorf("invalid LPB: %d", lpb))
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
			if tpl < 1 {
				m.SetError(fmt.Errorf("invalid TPL: %d", tpl))
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
			if rows < 1 {
				m.SetError(fmt.Errorf("invalid row count: %d", rows))
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
	case "tracks":
		if len(items) > 1 {
			tracks, err := strconv.Atoi(items[1])
			if err != nil {
				m.SetError(err)
				return
			}
			p := m.song.Patterns[m.editPattern]
			currentTrackCount := len(p[0])
			if tracks == currentTrackCount {
				return
			}
			if tracks < 1 {
				m.SetError(fmt.Errorf("invalid track count: %d", tracks))
				return
			}
			clone := p.clone()
			if tracks < currentTrackCount {
				for i := 0; i < len(clone); i++ {
					clone[i] = clone[i][:tracks]
				}
			} else {
				for i := 0; i < len(clone); i++ {
					clone[i] = slices.Grow(clone[i], tracks-currentTrackCount)
					for j := 0; j < tracks-currentTrackCount; j++ {
						clone[i] = append(clone[i], MidiMessage{})
					}
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
