package main

import (
	"slices"
)

func makePattern(rowCount, trackCount int) *Pattern {
	rows := make([]Row, rowCount)
	for i := range rowCount {
		rows[i] = make(Row, trackCount)
	}
	trackDefaults := make(Row, trackCount)
	return &Pattern{
		Rows:          rows,
		NumRows:       rowCount,
		NumTracks:     trackCount,
		TrackDefaults: trackDefaults,
	}
}

func makeDefaultPattern() *Pattern {
	return makePattern(64, 16)
}

func (p *Pattern) Width() int {
	return p.NumTracks * 6
}

func (p *Pattern) Height() int {
	return p.NumRows
}

func (p *Pattern) clone() *Pattern {
	clone := makePattern(p.NumRows, p.NumTracks)
	for rowIndex := 0; rowIndex < p.NumRows; rowIndex++ {
		clone.Rows[rowIndex] = slices.Clone(p.Rows[rowIndex])
	}
	clone.TrackDefaults = slices.Clone(p.TrackDefaults)
	return clone
}

func (p *Pattern) insertTracks(at, count int) *Pattern {
	rows := make([]Row, 0, p.NumRows)
	for _, srow := range p.Rows {
		drow := make(Row, p.NumTracks+count)
		drow = slices.Replace(drow, 0, at, srow[:at]...)
		drow = slices.Replace(drow, at+count, len(drow), srow[at:]...)
		rows = append(rows, drow)
	}
	trackDefaults := make(Row, p.NumTracks+count)
	trackDefaults = slices.Replace(trackDefaults, 0, at, p.TrackDefaults[:at]...)
	trackDefaults = slices.Replace(trackDefaults, at+count, len(trackDefaults), p.TrackDefaults[at:]...)
	return &Pattern{
		Rows:          rows,
		NumRows:       p.NumRows,
		NumTracks:     p.NumTracks + count,
		TrackDefaults: trackDefaults,
	}
}

func (p *Pattern) deleteTracks(at, count int) *Pattern {
	rows := make([]Row, 0, p.NumRows)
	for _, srow := range p.Rows {
		drow := make(Row, p.NumTracks-count)
		drow = slices.Replace(drow, 0, at, srow[:at]...)
		drow = slices.Replace(drow, at, len(drow), srow[at+count:]...)
		rows = append(rows, drow)
	}
	trackDefaults := make(Row, p.NumTracks-count)
	trackDefaults = slices.Replace(trackDefaults, 0, at, p.TrackDefaults[:at]...)
	trackDefaults = slices.Replace(trackDefaults, at, len(trackDefaults), p.TrackDefaults[at+count:]...)
	return &Pattern{
		Rows:          rows,
		NumRows:       p.NumRows,
		NumTracks:     p.NumTracks - count,
		TrackDefaults: trackDefaults,
	}
}

func (p *Pattern) withNumTracks(numTracks int) *Pattern {
	if numTracks == p.NumTracks {
		return p
	} else if numTracks > p.NumTracks {
		numTracksToAdd := numTracks - p.NumTracks
		return p.insertTracks(p.NumTracks, numTracksToAdd)
	} else {
		numTracksToRemove := p.NumTracks - numTracks
		return p.deleteTracks(numTracks, numTracksToRemove)
	}
}

func appendEmptyRows(rows []Row, count int, numTracks int) []Row {
	for range count {
		emptyRow := make(Row, numTracks)
		rows = append(rows, emptyRow)
	}
	return rows
}

func (p *Pattern) insertRows(at, count int) *Pattern {
	rows := make([]Row, 0, p.NumRows+count)
	for y, srow := range p.Rows {
		if at == y {
			rows = appendEmptyRows(rows, count, p.NumTracks)
		}
		rows = append(rows, slices.Clone(srow))
	}
	if at == p.NumRows {
		rows = appendEmptyRows(rows, count, p.NumTracks)
	}
	trackDefaults := slices.Clone(p.TrackDefaults)
	return &Pattern{
		Rows:          rows,
		NumRows:       p.NumRows + count,
		NumTracks:     p.NumTracks,
		TrackDefaults: trackDefaults,
	}
}

func (p *Pattern) deleteRows(at, count int) *Pattern {
	rows := make([]Row, 0, p.NumRows-count)
	for y := range at {
		srow := p.Rows[y]
		rows = append(rows, slices.Clone(srow))
	}
	for y := at; y < p.NumRows-count; y++ {
		srow := p.Rows[y+count]
		rows = append(rows, slices.Clone(srow))
	}
	trackDefaults := slices.Clone(p.TrackDefaults)
	return &Pattern{
		Rows:          rows,
		NumRows:       p.NumRows - count,
		NumTracks:     p.NumTracks,
		TrackDefaults: trackDefaults,
	}
}

func (p *Pattern) withNumRows(numRows int) *Pattern {
	if numRows == p.NumRows {
		return p
	} else if numRows > p.NumRows {
		numRowsToAdd := numRows - p.NumRows
		return p.insertRows(p.NumRows, numRowsToAdd)
	} else {
		numRowsToRemove := p.NumRows - numRows
		return p.deleteRows(numRows, numRowsToRemove)
	}
}

func (p *Pattern) getDigit(x, y int) byte {
	row := p.Rows[y]
	msg := &row[x/6]
	return msg.getDigit(x % 6)
}

func (p *Pattern) setDigit(x, y int, b byte) {
	row := p.Rows[y]
	msg := &row[x/6]
	msg.setDigit(x%6, b)
}

func (p *Pattern) getBlock(r Rect) Block {
	result := make(Block, r.H)
	for dy := 0; dy < r.H; dy++ {
		result[dy] = make([]byte, r.W)
		for dx := 0; dx < r.W; dx++ {
			result[dy][dx] = p.getDigit(r.X+dx, r.Y+dy)
		}
	}
	return result
}

func (p *Pattern) setBlock(r Rect, block Block) {
	for dy := 0; dy < r.H; dy++ {
		for dx := 0; dx < r.W; dx++ {
			p.setDigit(r.X+dx, r.Y+dy, block[dy][dx])
		}
	}
}

func (p *Pattern) zeroBlock(r Rect) {
	for dy := 0; dy < r.H; dy++ {
		for dx := 0; dx < r.W; dx++ {
			p.setDigit(r.X+dx, r.Y+dy, 0)
		}
	}
}

func (p *Pattern) copyBlock(r Rect, dx, dy int) {
	block := p.getBlock(r)
	p.setBlock(Rect{r.X + dx, r.Y + dy, r.W, r.H}, block)
}
