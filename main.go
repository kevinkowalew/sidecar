package main

import (
	"gitdiff/git"
	"gitdiff/tui"

	"github.com/charmbracelet/lipgloss"
)

func main() {
	s, err := git.Current()
	if err != nil {
		panic(err)
	}

	selectedItemStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("170")).Bold(true)

	t := tui.NewTUI(s, selectedItemStyle)
	if err = t.Run(); err != nil {
		panic(err)
	}
}
