package tui

import (
	"fmt"
	"sidecar/git"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type ScrollingText struct {
	diff          git.Diff
	selectedIndex int

	deleted bool
	commit  map[*git.Chunk]bool
	remove  map[*git.Chunk]bool
	skip    map[*git.Chunk]bool
	history []git.Chunk

	menuItemStyle, oldLineStyle, newLineStyle lipgloss.Style
}

func NewScollingText(diff git.Diff, buttonColor lipgloss.AdaptiveColor) *ScrollingText {
	menuItemStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFFFFF")).
		Background(buttonColor).
		Bold(false).
		PaddingLeft(1).
		PaddingRight(1)

	t := &ScrollingText{
		diff:          diff,
		selectedIndex: -1,
		commit:        make(map[*git.Chunk]bool, 0),
		remove:        make(map[*git.Chunk]bool, 0),
		skip:          make(map[*git.Chunk]bool, 0),
		history:       make([]git.Chunk, 0),
		menuItemStyle: menuItemStyle,
		oldLineStyle:  lipgloss.NewStyle().Foreground(lipgloss.Color("#FF0000")),
		newLineStyle:  lipgloss.NewStyle().Foreground(lipgloss.Color("#00FF00")),
	}

	if len(diff.Chunks) > 0 {
		t.selectedIndex = 0
	}

	return t
}

func (st *ScrollingText) Init() tea.Cmd {
	return nil
}

func (st *ScrollingText) Render() string {
	s := st.renderChunks()
	s += st.createMenu()
	s += "\n"
	return s
}

func (st *ScrollingText) Complete() bool {
	return len(st.history) == len(st.diff.Chunks) || st.deleted
}

func (st *ScrollingText) renderChunks() string {
	s := ""
	if st.deleted {
		s += "Deleted\n\n"
		s += st.createMenuItem("Undo", "u")
		return s
	}
	if len(st.history) == len(st.diff.Chunks) {
		return "\nNo remaining changes.\n\n"
	}

	for index, chunk := range st.diff.Chunks {
		if index != st.selectedIndex {
			continue
		}

		_, ok := st.commit[&chunk]
		if ok {
			continue
		}

		s += strings.Join(st.getStyledChunkLines(index, chunk), "\n")
		break
	}
	s += "\n\n"
	return s
}

func (st *ScrollingText) createMenuItem(title, key string) string {
	return st.menuItemStyle.Render(fmt.Sprintf("%s (%s)", title, key)) + " "
}

func (st *ScrollingText) createMenu() string {
	s := ""
	if len(st.history) < len(st.diff.Chunks) {
		s += fmt.Sprintf("%d/%d ", st.selectedIndex+1-len(st.history), len(st.diff.Chunks)-len(st.history))
		s += st.createMenuItem("Commit", "c")
	}

	if len(st.history) > 0 {
		s += st.createMenuItem("Undo", "u")
	}

	return s
}

func (st *ScrollingText) MoveCursorForward() {
	if st.selectedIndex < len(st.diff.Chunks)-1 {
		st.selectedIndex++
	}
}

func (st *ScrollingText) MoveCursorBackward() {
	if st.selectedIndex > 0 {
		st.selectedIndex--
	}
}

func (st *ScrollingText) Skip() {
	st.process(st.skip)
}

func (st *ScrollingText) Commit() {
	st.process(st.commit)
}

func (st *ScrollingText) process(targetMap map[*git.Chunk]bool) {
	if len(st.history) == len(st.diff.Chunks) {
		return
	}
	chunk := st.diff.Chunks[st.selectedIndex]
	targetMap[&chunk] = true
	st.history = append(st.history, chunk)

	original := st.selectedIndex
	st.MoveCursorForward()
	if original == st.selectedIndex {
		st.MoveCursorBackward()
	}
}

func (st *ScrollingText) reducedChunks() []git.Chunk {
	chunks := make([]git.Chunk, 0)
	for _, chunk := range st.diff.Chunks {
		_, ok := st.commit[&chunk]
		if ok {
			continue
		}
		chunks = append(chunks, chunk)
	}
	return chunks
}

func (st *ScrollingText) Undo() {
	switch len(st.history) {
	case 0:
		return
	case 1:
		delete(st.commit, &st.history[len(st.history)-1])
		st.history = make([]git.Chunk, 0)
	default:
		delete(st.commit, &st.history[len(st.history)-1])
		st.history = st.history[:len(st.history)-1]
	}

	original := st.selectedIndex
	st.MoveCursorBackward()
	if original == st.selectedIndex {
		st.MoveCursorForward()
	}
}

func (st *ScrollingText) getStyledChunkLines(index int, chunk git.Chunk) []string {
	return append(
		st.styleSnippet(chunk.Old, st.oldLineStyle),
		st.styleSnippet(chunk.New, st.newLineStyle)...,
	)
}

func (st *ScrollingText) styleSnippet(snippet git.Snippet, style lipgloss.Style) []string {
	lines := []string{}
	for _, line := range snippet.Lines {
		lines = append(lines, style.Render(line))
	}
	return lines
}
