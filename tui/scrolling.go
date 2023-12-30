package tui

import (
	"fmt"
	"gitdiff/git"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	// TODO: make thes
	oldLineStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF0000"))
	newLineStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#00FF00"))

	menuItemStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(lipgloss.Color("#6F8FAF")).
			Bold(false).
			PaddingLeft(1).
			PaddingRight(1)
)

type ScrollingText struct {
	diff          git.FileDiff
	selectedIndex int

	deleted bool
	commit  map[*git.FileDiffChunk]bool
	remove  map[*git.FileDiffChunk]bool
	skip    map[*git.FileDiffChunk]bool
	history []git.FileDiffChunk
}

func NewScollingText(diff git.FileDiff) *ScrollingText {
	t := &ScrollingText{
		diff:          diff,
		selectedIndex: -1,
		commit:        make(map[*git.FileDiffChunk]bool, 0),
		remove:        make(map[*git.FileDiffChunk]bool, 0),
		skip:          make(map[*git.FileDiffChunk]bool, 0),
		history:       make([]git.FileDiffChunk, 0),
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
		s += createMenuItem("Undo", "u")
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

func createMenuItem(title, key string) string {
	return menuItemStyle.Render(fmt.Sprintf("%s (%s)", title, key)) + " "
}

func (st *ScrollingText) createMenu() string {
	s := ""
	if len(st.history) < len(st.diff.Chunks) {
		s += fmt.Sprintf("%d/%d ", st.selectedIndex+1-len(st.history), len(st.diff.Chunks)-len(st.history))
		s += createMenuItem("Commit", "c")
	}

	if len(st.history) > 0 {
		s += createMenuItem("Undo", "u")
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

func (st *ScrollingText) process(targetMap map[*git.FileDiffChunk]bool) {
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

func (st *ScrollingText) reducedChunks() []git.FileDiffChunk {
	chunks := make([]git.FileDiffChunk, 0)
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
		st.history = make([]git.FileDiffChunk, 0)
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

func (st *ScrollingText) getStyledChunkLines(index int, chunk git.FileDiffChunk) []string {
	return append(
		st.styleSnippet(chunk.Old, oldLineStyle),
		st.styleSnippet(chunk.New, newLineStyle)...,
	)
}

func (st *ScrollingText) styleSnippet(snippet git.Snippet, style lipgloss.Style) []string {
	lines := []string{}
	for _, line := range snippet.Lines {
		lines = append(lines, style.Render(line))
	}
	return lines
}
