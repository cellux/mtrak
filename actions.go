package main

func (m *model) Up() {
	if m.editRow > 0 {
		m.editRow--
	}
}

func (m *model) Down() {
	p := m.patterns[m.editPattern]
	if m.editRow < len(p)-1 {
		m.editRow++
	}
}

func (m *model) PageUp() {
	if m.editRow%16 != 0 {
		m.editRow -= m.editRow % 16
	} else {
		m.editRow -= 16
	}
	if m.editRow < 0 {
		m.editRow = 0
	}
}

func (m *model) PageDown() {
	p := m.patterns[m.editPattern]
	m.editRow += 16
	if m.editRow%16 != 0 {
		m.editRow -= m.editRow % 16
	}
	if m.editRow >= len(p) {
		m.editRow = len(p) - 1
	}
}

func (m *model) Left() {
	if m.editColumn > 0 {
		m.editColumn--
	} else if m.editTrack > 0 {
		m.editTrack--
		m.editColumn = 5
	}
}

func (m *model) Right() {
	p := m.patterns[m.editPattern]
	row := p[m.editRow]
	if m.editColumn < 5 {
		m.editColumn++
	} else if m.editTrack < len(row)-1 {
		m.editTrack++
		m.editColumn = 0
	}
}

func (m *model) NextTrack() {
	p := m.patterns[m.editPattern]
	row := p[m.editRow]
	if m.editTrack < len(row)-1 {
		m.editTrack++
		m.editColumn = 0
	}
}

func (m *model) PrevTrack() {
	if m.editTrack > 0 {
		m.editTrack--
		m.editColumn = 0
	}
}

func (m *model) DeleteLeft() {
	if m.editTrack > 0 || m.editColumn > 0 {
		m.Left()
		m.setByte(0)
	}
}

func (m *model) DeleteUnder() {
	m.setByte(0)
}

func (m *model) InsertBlank() {
	m.insertByte(0)
}

func (m *model) PlayOrStop() {
	if m.isPlaying {
		m.Stop()
	} else {
		m.playRow = m.startRow
		m.Play()
	}
}

func (m *model) SetStartRow() {
	m.startRow = m.editRow
}
