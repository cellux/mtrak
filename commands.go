package main

import (
	"fmt"
	"strconv"
	"strings"
)

func parseInt(s string) (int, error) {
	i, err := strconv.ParseInt(s, 0, 0)
	return int(i), err
}

func (m *Model) ExecuteCommand(command string) {
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
			bpm, err := parseInt(items[1])
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
			lpb, err := parseInt(items[1])
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
			tpl, err := parseInt(items[1])
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
			numRows, err := parseInt(items[1])
			if err != nil {
				m.SetError(err)
				return
			}
			p := m.song.Patterns[m.editPattern]
			if numRows == p.NumRows {
				return
			}
			if numRows < 1 {
				m.SetError(fmt.Errorf("invalid row count: %d", numRows))
				return
			}
			clone := p.withNumRows(numRows)
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
			numTracks, err := parseInt(items[1])
			if err != nil {
				m.SetError(err)
				return
			}
			p := m.song.Patterns[m.editPattern]
			if numTracks == p.NumTracks {
				return
			}
			if numTracks < 1 {
				m.SetError(fmt.Errorf("invalid track count: %d", numTracks))
				return
			}
			clone := p.withNumTracks(numTracks)
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
