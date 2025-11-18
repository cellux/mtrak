package main

import (
	tea "github.com/charmbracelet/bubbletea"
)

func buildScale(steps ...int) Scale {
	scale := make(Scale, len(steps))
	offset := 0
	for i, step := range steps {
		scale[i] = offset
		offset += step
	}
	return scale
}

var scales = func() ScaleRegistry {
	scales := ScaleRegistry{
		"major":            buildScale(2, 2, 1, 2, 2, 2, 1),
		"natural minor":    buildScale(2, 1, 2, 2, 1, 2, 2),
		"harmonic minor":   buildScale(2, 1, 2, 2, 1, 3, 1),
		"melodic minor":    buildScale(2, 1, 2, 2, 2, 2, 1),
		"major pentatonic": buildScale(2, 2, 3, 2, 3),
		"whole tone":       buildScale(2, 2, 2, 2, 2, 2),
		"octatonic":        buildScale(2, 1, 2, 1, 2, 1, 2, 1),
	}
	scales["M"] = scales["major"]
	scales["minor"] = scales["natural minor"]
	scales["min"] = scales["natural minor"]
	scales["m"] = scales["natural minor"]
	scales["hmin"] = scales["harmonic minor"]
	scales["hm"] = scales["harmonic minor"]
	scales["mmin"] = scales["melodic minor"]
	scales["mm"] = scales["melodic minor"]
	scales["pentatonic"] = scales["major pentatonic"]
	scales["penta"] = scales["major pentatonic"]
	scales["p"] = scales["major pentatonic"]
	return scales
}()

var keysToScaleDegreesInNoteMode = map[string]int{
	"z": 0,
	"x": 1,
	"c": 2,
	"v": 3,
	"b": 4,
	"n": 5,
	"m": 6,
	"q": 12,
	"w": 13,
	"e": 14,
	"r": 15,
	"t": 16,
	"y": 17,
	"u": 18,
}

var keysToScaleDegreesInChromaticMode = map[string]int{
	"z": 0,
	"s": 1,
	"x": 2,
	"d": 3,
	"c": 4,
	"v": 5,
	"g": 6,
	"b": 7,
	"h": 8,
	"n": 9,
	"j": 10,
	"m": 11,
	"q": 12,
	"2": 13,
	"w": 14,
	"3": 15,
	"e": 16,
	"r": 17,
	"5": 18,
	"t": 19,
	"6": 20,
	"y": 21,
	"7": 22,
	"u": 23,
}

func (m *Model) DegreeToMidiNote(degree int) int {
	scale := m.song.Scale
	scaleSize := len(scale)
	degree += m.song.Mode
	return min(m.song.Root+(degree/scaleSize)*12+scale[degree%scaleSize], 127)
}

func (m *Model) KeyMsgToMidiNote(msg tea.KeyMsg) int {
	if m.song.Chromatic {
		if degree, ok := keysToScaleDegreesInChromaticMode[msg.String()]; ok {
			return min(m.song.Root + degree)
		}
	} else {
		if degree, ok := keysToScaleDegreesInNoteMode[msg.String()]; ok {
			scaleSize := len(m.song.Scale)
			if degree >= 12 {
				degree = degree - 12 + scaleSize
			}
			return m.DegreeToMidiNote(degree)
		}
	}
	return -1
}
