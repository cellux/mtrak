package main

import (
	"github.com/charmbracelet/lipgloss"
	"strings"
)

type RowBuilder struct {
	sb    strings.Builder
	group strings.Builder
	style *lipgloss.Style
}

func (rb *RowBuilder) WriteByte(b byte) error {
	return rb.group.WriteByte(b)
}

func (rb *RowBuilder) WriteRune(r rune) (int, error) {
	return rb.group.WriteRune(r)
}

func (rb *RowBuilder) WriteString(s string) (int, error) {
	return rb.group.WriteString(s)
}

func (rb *RowBuilder) flush() {
	if rb.group.Len() > 0 {
		rb.sb.WriteString(rb.style.Render(rb.group.String()))
		rb.group.Reset()
	}
}

func (rb *RowBuilder) SetStyle(style *lipgloss.Style) {
	if rb.style != style {
		rb.flush()
		rb.style = style
	}
}

func (rb *RowBuilder) String() string {
	rb.flush()
	return rb.sb.String()
}

func (rb *RowBuilder) Reset() {
	rb.sb.Reset()
	rb.group.Reset()
}
