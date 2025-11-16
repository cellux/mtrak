package main

import (
	"flag"
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"os"
)

var program *tea.Program

func main() {
	m := &Model{}
	defer m.Close()
	flag.Parse()
	args := flag.Args()
	if len(args) > 0 {
		if len(args) > 1 {
			fmt.Fprintln(os.Stderr, "Usage: mtrak [filename]")
			os.Exit(1)
		}
		m.filename = args[0]
	}
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
