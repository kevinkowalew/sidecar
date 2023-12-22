package main

import (
	"gitdiff/git"
	"gitdiff/tui"

	"github.com/charmbracelet/lipgloss"
)

func main() {
	s, err := git.LoadCurrentState()
	if err != nil {
		panic(err)
	}

	selectedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("177")).Bold(true)

	t := tui.NewTUI(s, selectedStyle)
	err = t.Run()
	if err != nil {
		panic(err)
	}
}
