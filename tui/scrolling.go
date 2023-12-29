package tui

import (
	"fmt"
	"gitdiff/git"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type ScrollingText struct {
	diff          git.FileDiff
	selectedIndex int

	commitMap map[*git.FileDiffChunk]bool
	ignoreMap map[*git.FileDiffChunk]bool
	history   []git.FileDiffChunk
}

func NewScollingText(diff git.FileDiff) *ScrollingText {
	//TODO: add runtime validation of number of chunks
	//TODO: make colors injectable
	t := &ScrollingText{
		diff:          diff,
		selectedIndex: -1,
		commitMap:     make(map[*git.FileDiffChunk]bool, 0),
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
	s += "\n\n"
	s += st.createMenu()
	s += "\n"
	return s
}

func (st *ScrollingText) Complete() bool {
	return len(st.history) == len(st.diff.Chunks)
}

func (st *ScrollingText) renderChunks() string {
	if len(st.history) == len(st.diff.Chunks) {
		return "No remaining changes."
	}

	s := ""
	for index, chunk := range st.diff.Chunks {
		if index != st.selectedIndex {
			continue
		}

		_, ok := st.commitMap[&chunk]
		if ok {
			continue
		}

		s += strings.Join(st.getStyledChunkLines(index, chunk), "\n")
		break
	}
	return s
}

func (st *ScrollingText) createMenu() string {
	createMenuItem := func(title, key string) string {
		return menuItemStyle.Render(fmt.Sprintf("%s (%s)", title, key)) + " "
	}

	s := ""
	if len(st.history) < len(st.diff.Chunks) {
		s += fmt.Sprintf("%d/%d ", st.selectedIndex+1-len(st.history), len(st.diff.Chunks)-len(st.history))
		s += createMenuItem("Commit", "c")
		s += createMenuItem("Ignore", "d")
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

func (st *ScrollingText) Ignore() {
	st.process(st.commitMap)
}

func (st *ScrollingText) Commit() {
	st.process(st.commitMap)
}

func (st *ScrollingText) process(targetMap map[*git.FileDiffChunk]bool) {
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
		_, ok := st.commitMap[&chunk]
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
		delete(st.commitMap, &st.history[len(st.history)-1])
		st.history = make([]git.FileDiffChunk, 0)
	default:
		delete(st.commitMap, &st.history[len(st.history)-1])
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
