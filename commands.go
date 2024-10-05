package main

import (
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
	}
}
