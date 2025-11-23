package main

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"regexp"
	"strings"
)

type (
	Scale   []int
	ScaleId = int
)

const (
	ScaleMajor ScaleId = iota
	ScaleNaturalMinor
	ScaleHarmonicMinor
	ScaleMelodicMinor
	ScalePentatonic
	ScaleWholeTone
	ScaleOctatonic
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

var scales []Scale
var scaleNameById []string
var scaleCodeById []string
var scaleIdByName map[string]ScaleId

var scaleRegex *regexp.Regexp

func init() {
	scales = []Scale{
		buildScale(2, 2, 1, 2, 2, 2, 1),    // major
		buildScale(2, 1, 2, 2, 1, 2, 2),    // natural minor
		buildScale(2, 1, 2, 2, 1, 3, 1),    // harmonic minor
		buildScale(2, 1, 2, 2, 2, 2, 1),    // melodic minor
		buildScale(2, 2, 3, 2, 3),          // pentatonic
		buildScale(2, 2, 2, 2, 2, 2),       // whole tone
		buildScale(2, 1, 2, 1, 2, 1, 2, 1), // octatonic
	}
	scaleNameById = []string{
		"major",
		"natural minor",
		"harmonic minor",
		"melodic minor",
		"pentatonic",
		"whole tone",
		"octatonic",
	}
	scaleCodeById = []string{
		"M",
		"m",
		"hm",
		"mm",
		"p",
		"w",
		"o",
	}
	scaleIdByName = map[string]ScaleId{
		"M":              ScaleMajor,
		"major":          ScaleMajor,
		"natural minor":  ScaleNaturalMinor,
		"minor":          ScaleNaturalMinor,
		"min":            ScaleNaturalMinor,
		"m":              ScaleNaturalMinor,
		"harmonic minor": ScaleHarmonicMinor,
		"hmin":           ScaleHarmonicMinor,
		"hm":             ScaleHarmonicMinor,
		"melodic minor":  ScaleMelodicMinor,
		"mmin":           ScaleMelodicMinor,
		"mm":             ScaleMelodicMinor,
		"pentatonic":     ScalePentatonic,
		"penta":          ScalePentatonic,
		"p":              ScalePentatonic,
		"whole tone":     ScaleWholeTone,
		"whole":          ScaleWholeTone,
		"octatonic":      ScaleOctatonic,
		"octa":           ScaleOctatonic,
		"o":              ScaleOctatonic,
	}
	scaleNames := make([]string, 0, len(scaleIdByName))
	for name := range scaleIdByName {
		scaleNames = append(scaleNames, name)
	}
	scaleRegex = regexp.MustCompile(`^(` + strings.Join(scaleNames, `|`) + `)\b`)
}

var keysToScaleDegreesInNoteMode = map[string]int{
	"z": 0,
	"x": 1,
	"c": 2,
	"v": 3,
	"b": 4,
	"n": 5,
	"m": 6,
	",": 7,
	"q": 12,
	"w": 13,
	"e": 14,
	"r": 15,
	"t": 16,
	"y": 17,
	"u": 18,
	"i": 19,
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
	",": 12,
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
	"i": 24,
}

func (m *Model) DegreeToMidiNote(degree int) int {
	scaleId := m.song.Scale
	scale := scales[scaleId]
	scaleSize := len(scale)
	degree += m.song.Mode
	return min(m.song.Root+(degree/scaleSize)*12+scale[degree%scaleSize], 127)
}

func (m *Model) KeyMsgToMidiNote(msg tea.KeyMsg) int {
	if m.song.Chromatic {
		if degree, ok := keysToScaleDegreesInChromaticMode[msg.String()]; ok {
			return min(m.song.Root+degree, 127)
		}
	} else {
		if degree, ok := keysToScaleDegreesInNoteMode[msg.String()]; ok {
			scaleId := m.song.Scale
			scale := scales[scaleId]
			scaleSize := len(scale)
			if degree >= 12 {
				degree = degree - 12 + scaleSize
			}
			return m.DegreeToMidiNote(degree)
		}
	}
	return -1
}

var noteRegex = regexp.MustCompile(`^([A-G])(#|-)([0-9])`)

var degreeByNoteName = map[byte]int{
	'C': 0,
	'D': 2,
	'E': 4,
	'F': 5,
	'G': 7,
	'A': 9,
	'B': 11,
}

func (m *Model) parseNote(s string) (note int, consumed int, err error) {
	parts := noteRegex.FindStringSubmatch(strings.ToUpper(s))
	if parts == nil {
		return -1, 0, fmt.Errorf("invalid note: %s", s)
	}
	degree := degreeByNoteName[parts[1][0]]
	octave := int(parts[3][0] - '0')
	note = octave*12 + degree
	if parts[2][0] == '#' {
		note++
	}
	return min(note, 127), len(parts[0]), nil
}

func (m *Model) parseScale(s string) (root int, scale ScaleId, mode int, err error) {
	root = -1
	scale = -1
	mode = -1
	for {
		s = strings.TrimSpace(s)
		if s == "" {
			return root, scale, mode, nil
		}
		_scaleName := scaleRegex.FindString(s)
		if _scaleName != "" {
			scale = scaleIdByName[_scaleName]
			s = s[len(_scaleName):]
			continue
		}
		_note, consumed, err := m.parseNote(s)
		if err == nil {
			root = _note
			s = s[consumed:]
			continue
		}
		_mode, consumed, err := m.parseMode(s)
		if err == nil {
			mode = _mode
			s = s[consumed:]
			continue
		}
		return -1, -1, -1, fmt.Errorf("invalid scale: %s", s)
	}
}

var modeRegex = regexp.MustCompile(`^\+?[0-9]+`)

func (m *Model) parseMode(s string) (mode int, consumed int, err error) {
	modeString := modeRegex.FindString(s)
	if modeString == "" {
		return -1, 0, fmt.Errorf("invalid mode: %s", s)
	}
	mode, _ = parseInt(modeString)
	scale := scales[m.song.Scale]
	if mode < 0 || mode >= len(scale) {
		return -1, 0, fmt.Errorf("invalid mode: %d: scale has only %d degrees", mode, len(scale))
	}
	return mode, len(modeString), nil
}
