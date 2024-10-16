package main

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"os"
	//"github.com/lucasb-eyer/go-colorful"
)

var program *tea.Program

func main() {
	m := &Model{}
	defer m.Close()
	program = tea.NewProgram(m, tea.WithAltScreen())
	if _, err := program.Run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if m.err != nil {
		fmt.Fprintln(os.Stderr, m.err)
		os.Exit(1)
	}
}
