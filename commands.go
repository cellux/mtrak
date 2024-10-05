package main

import (
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
	}
}
