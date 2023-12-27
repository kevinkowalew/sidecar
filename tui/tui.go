package tui

import (
	"fmt"
	"gitdiff/git"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	fileDiffMode    = "fileDiffMode"
	viewChangesMode = "viewChanges"
)

var (
	oldLineStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF0000")).
			Bold(false)
	newLineStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#00FF00")).
			Bold(false)
	menuItemStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(lipgloss.Color("#6F8FAF")).
			Bold(false).
			PaddingLeft(1).
			PaddingRight(1)
)

type TUI struct {
	status        *git.Status
	selectedStyle lipgloss.Style

	selectedDiff  *git.FileDiff
	selectedChunk *git.FileDiffChunk
	mode          string
}

func NewTUI(status *git.Status, selectedStyle lipgloss.Style) *TUI {
	tui := &TUI{
		status:        status,
		selectedStyle: selectedStyle,
		mode:          fileDiffMode,
	}

	if len(status.StagedFileDiffs) > 0 {
		tui.selectedDiff = &status.StagedFileDiffs[0]
	} else if len(status.UnstagedFileDiffs) > 0 {
		tui.selectedDiff = &status.UnstagedFileDiffs[0]
	}
	return tui
}

func (t *TUI) Init() tea.Cmd {
	return nil
}

func (t *TUI) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.KeyMsg:
		switch msg.String() {

		case "ctrl+c":
			return t, tea.Quit
		case "c":
			t.toggleSelectedFileDiff()
		case "q":
			if t.mode == fileDiffMode {
				return t, tea.Quit
			}
		case "esc":
			if t.mode == fileDiffMode {
				return t, tea.Quit
			} else if t.mode == viewChangesMode {
				t.mode = fileDiffMode
			}
		case "enter":
			if t.mode == fileDiffMode {
				t.mode = viewChangesMode
			}
		case "up", "k":
			if t.mode == fileDiffMode {
				t.moveCursorBack()
			}
		case "down", "j":
			if t.mode == fileDiffMode {
				t.moveCursorForward()
			}
		}
	}

	return t, nil
}

func (t *TUI) moveCursorForward() {
	if t.selectedDiff == nil {
		return
	}

	next := false
	for _, diff := range t.status.StagedFileDiffs {
		if diff.Equals(*t.selectedDiff) {
			next = true
		} else if next {
			t.selectedDiff = &diff
			return
		}
	}
	for _, diff := range t.status.UnstagedFileDiffs {
		if diff.Equals(*t.selectedDiff) {
			next = true
		} else if next {
			t.selectedDiff = &diff
			return
		}
	}
}

func (t *TUI) moveCursorBack() {
	if t.selectedDiff == nil {
		return
	}

	next := false
	for i := len(t.status.UnstagedFileDiffs) - 1; i >= 0; i-- {
		diff := t.status.UnstagedFileDiffs[i]
		if diff.Equals(*t.selectedDiff) {
			next = true
			continue
		} else if next {
			t.selectedDiff = &diff
			return
		}
	}

	for i := len(t.status.StagedFileDiffs) - 1; i >= 0; i-- {
		diff := t.status.StagedFileDiffs[i]
		if diff.Equals(*t.selectedDiff) {
			next = true
			continue
		} else if next {
			t.selectedDiff = &diff
			return
		}
	}
}

func (t *TUI) View() string {
	if t.mode == fileDiffMode {
		s := "Commited:\n"
		for _, diff := range t.status.StagedFileDiffs {
			s += t.styleFileDiffLine(diff)
		}
		s += "\n"

		s += "Uncommited:\n"
		for _, diff := range t.status.UnstagedFileDiffs {
			s += t.styleFileDiffLine(diff)
		}
		s += "\n"

		s += t.createMenu()
		return s
	}

	s := ""
	for _, chunk := range t.selectedDiff.Chunks {
		s += t.styleChunk(chunk)
		s += "\n\n"
	}
	return s
}

func (t *TUI) styleChunk(chunk git.FileDiffChunk) string {
	return t.styleSnippet(chunk.Old, oldLineStyle) + t.styleSnippet(chunk.New, newLineStyle)
}

func (t *TUI) styleSnippet(snippet git.Snippet, style lipgloss.Style) string {
	s := ""
	for _, line := range snippet.Lines {
		s += style.Render(line)
		s += "\n"
	}
	return s
}

func (t *TUI) styleFileDiffLine(diff git.FileDiff) string {
	txt := ""
	if diff.Deleted() {
		txt += "  deleted: "
	} else {
		txt += "  modified: "
	}
	txt += diff.Filename

	if t.selectedDiff != nil {
		if diff.Equals(*t.selectedDiff) {
			return t.selectedStyle.Render(txt) + "\n"
		}
	}
	return txt + "\n"
}

func (t *TUI) toggleSelectedFileDiff() {
	if t.selectedDiff == nil {
		return
	}

	for i, diff := range t.status.StagedFileDiffs {
		if diff.Equals(*t.selectedDiff) {
			t.status.StagedFileDiffs = deleteElement(t.status.StagedFileDiffs, i)
			t.status.UnstagedFileDiffs = append(t.status.UnstagedFileDiffs, diff)
			git.SortFileDiffs(t.status.UnstagedFileDiffs)
			return
		}
	}
	for i, diff := range t.status.UnstagedFileDiffs {
		if diff.Equals(*t.selectedDiff) {
			t.status.UnstagedFileDiffs = deleteElement(t.status.UnstagedFileDiffs, i)
			t.status.StagedFileDiffs = append(t.status.StagedFileDiffs, diff)
			git.SortFileDiffs(t.status.StagedFileDiffs)
			break
		}
	}
}

func maxInt(x, y int) int {
	if x > y {
		return x
	}
	return y
}

func deleteElement(slice []git.FileDiff, index int) []git.FileDiff {
	if len(slice) <= 1 {
		return []git.FileDiff{}
	}

	if index == 0 {
		return slice[1:]
	}
	if index == len(slice)-1 {
		return slice[:len(slice)-1]
	}
	return append(slice[:index], slice[index+1:]...)
}

func (t *TUI) createMenu() string {
	type menuItem struct {
		key  string
		name string
	}

	var items []menuItem
	items = append(items, menuItem{key: "c", name: "Toggle"})
	items = append(items, menuItem{key: "v", name: "View"})
	items = append(items, menuItem{key: "esc", name: "Quity"})

	s := "    "
	for i, item := range items {
		msg := ""
		if i > 0 {
			s += "  "
		}
		msg += fmt.Sprintf("%s (%s)", item.name, item.key)
		s += menuItemStyle.Render(msg)
	}
	return s
}

func (tui *TUI) Run() error {
	_, err := tea.NewProgram(tui).Run()

	return err
}
