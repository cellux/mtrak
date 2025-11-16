package main

import (
	"slices"
)

func (p *Pattern) clone() Pattern {
	clone := make(Pattern, len(*p))
	for rowIndex := 0; rowIndex < len(*p); rowIndex++ {
		clone[rowIndex] = slices.Clone((*p)[rowIndex])
	}
	return clone
}

func (p *Pattern) insertTrack(at int) Pattern {
	clone := make(Pattern, len(*p))
	for rowIndex := 0; rowIndex < len(*p); rowIndex++ {
		srow := (*p)[rowIndex]
		drow := make(Row, len(srow)+1)
		drow = slices.Replace(drow, 0, at, srow[:at]...)
		drow = slices.Replace(drow, at+1, len(drow), srow[at:]...)
		clone[rowIndex] = drow
	}
	return clone
}

func (p *Pattern) deleteTrack(at int) Pattern {
	clone := make(Pattern, len(*p))
	for rowIndex := 0; rowIndex < len(*p); rowIndex++ {
		srow := (*p)[rowIndex]
		drow := make(Row, len(srow)-1)
		drow = slices.Replace(drow, 0, at, srow[:at]...)
		drow = slices.Replace(drow, at, len(drow), srow[at+1:]...)
		clone[rowIndex] = drow
	}
	return clone
}

func (p *Pattern) getDigit(x, y int) byte {
	row := (*p)[y]
	msg := &row[x/6]
	return msg.getDigit(x % 6)
}

func (p *Pattern) setDigit(x, y int, b byte) {
	row := (*p)[y]
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
