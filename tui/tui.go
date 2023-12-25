package tui

import (
	"fmt"
	"gitdiff/git"
	"sort"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	diffSelectMode = "diffSelect"
	diffMode       = "diffMode"
)

var (
	oldLineStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF0000")).Bold(false)
	newLineStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#00FF00")).Bold(false)
)

type TUI struct {
	status        *git.Status
	selectedStyle lipgloss.Style

	selectedFileDiffIndex int
	selectedDiffLineIndex int

	deletedDiff        map[int]bool
	deletedDiffHistory []int

	mode string
}

func NewTUI(status *git.Status, selectedStyle lipgloss.Style) *TUI {
	return &TUI{
		status:                status,
		selectedStyle:         selectedStyle,
		selectedFileDiffIndex: 0,
		selectedDiffLineIndex: 0,
		mode:                  diffSelectMode,
		deletedDiff:           make(map[int]bool, 0),
	}
}

func (t TUI) SelectedFileDiff() *git.FileDiff {
	if t.selectedFileDiffIndex < len(t.status.StagedFileDiffs) {
		return &t.status.StagedFileDiffs[t.selectedFileDiffIndex]
	}

	adjustedIndex := t.selectedFileDiffIndex - len(t.status.UnstagedFileDiffs)
	if adjustedIndex < len(t.status.UnstagedFileDiffs) {
		return &t.status.UnstagedFileDiffs[t.selectedFileDiffIndex]
	}

	return nil
}

func (t TUI) AdjustedSelectedFileDiffIndex() int {
	if t.selectedFileDiffIndex < len(t.status.StagedFileDiffs) {
		return t.selectedFileDiffIndex
	}
	return t.selectedFileDiffIndex - len(t.status.StagedFileDiffs)
}

func (t TUI) Init() tea.Cmd {
	return nil
}

func (t TUI) DiffSelectMode() bool {
	return t.mode == diffSelectMode
}

func (t TUI) DiffMode() bool {
	return t.mode == diffMode
}

func (t TUI) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.KeyMsg:
		switch msg.String() {

		case "ctrl+c":
			return t, tea.Quit
		case "d":
			if t.DiffMode() {
				t.deletedDiff[t.selectedDiffLineIndex] = true
				t.deletedDiffHistory = append(t.deletedDiffHistory, t.selectedDiffLineIndex)
			}
		case "u":
			if t.DiffMode() {
				if len(t.deletedDiffHistory) > 0 {
					delete(t.deletedDiff, t.deletedDiffHistory[len(t.deletedDiffHistory)-1])
					t.deletedDiffHistory = t.deletedDiffHistory[:len(t.deletedDiffHistory)-1]
				}
			}
		case "c":
			if t.DiffSelectMode() {
				t.toggleSelectedFileDiff()
			}
		case "q":
			if t.DiffSelectMode() {
				return t, tea.Quit
			}
		case "esc":
			if t.DiffSelectMode() {
				return t, tea.Quit
			} else if t.DiffMode() {
				t.mode = diffSelectMode
			}
		case "up", "k":
			if t.DiffSelectMode() && t.selectedFileDiffIndex > 0 {
				t.selectedFileDiffIndex--
			} else if t.DiffMode() && t.selectedDiffLineIndex > 0 {
				t.selectedDiffLineIndex--
			}
		case "down", "j":
			if t.DiffSelectMode() && t.selectedFileDiffIndex < len(t.status.StagedFileDiffs)+len(t.status.UnstagedFileDiffs)-1 {
				t.selectedFileDiffIndex++
			} else if t.DiffMode() {
				t.selectedDiffLineIndex++
			}
		case "enter":
			if t.DiffSelectMode() && !t.SelectedFileDiff().Deleted() {
				t.mode = diffMode
				t.selectedDiffLineIndex = 0
			}
		}
	}

	return t, nil
}

func printDiffs(diffs []git.FileDiff) string {
	s := "["
	for _, diff := range diffs {
		s += diff.Filename + ", "
	}
	s += "]"
	return s
}

func (t TUI) View() string {
	s := fmt.Sprint(t.selectedFileDiffIndex)
	s += "\n\n"
	s += printDiffs(t.status.StagedFileDiffs)
	s += "\n\n"
	s += printDiffs(t.status.UnstagedFileDiffs)
	s += "\nCommited:\n"
	if t.DiffSelectMode() {
		for i, diff := range t.status.StagedFileDiffs {
			s += t.styleFileDiffLine(i, diff, true)
		}

		s += "\nUncommitted:\n"
		for i, diff := range t.status.UnstagedFileDiffs {
			s += t.styleFileDiffLine(i, diff, false)
		}

	} else if t.DiffMode() {
		index := 0
		for _, chunk := range t.SelectedFileDiff().Chunks {
			for _, l := range chunk.Old.Lines {
				s += t.styleDiffDetailLine(index, l)
			}
			index++

			for _, l := range chunk.New.Lines {
				s += t.styleDiffDetailLine(index, l)
			}

			index++
			s += "\n"
		}
	}

	s += "\n"
	s += t.createMenuString()

	return s
}

func (t TUI) styleFileDiffLine(index int, diff git.FileDiff, staged bool) string {
	txt := ""
	if diff.Deleted() {
		txt += "deleted: "
	} else {
		txt += "modified: "
	}
	txt += diff.Filename

	if t.isSelectedDiffIndex(index, staged) {
		return t.selectedStyle.Render(txt) + "\n"
	}
	return txt + "\n"
}

func (t TUI) styleDiffDetailLine(index int, l string) string {
	if t.isDeletedDiffIndex(index) {
		return ""
	}
	if index == t.selectedDiffLineIndex {
		return t.selectedStyle.Render(l) + "\n"
	}
	return newLineStyle.Render(l) + "\n"
}

func (t TUI) toggleSelectedFileDiff() {
	diff := *t.SelectedFileDiff()
	if t.selectedFileDiffIndex < len(t.status.StagedFileDiffs) {
		t.status.StagedFileDiffs = deleteElement(t.status.StagedFileDiffs, t.selectedFileDiffIndex)
		t.status.UnstagedFileDiffs = append(t.status.UnstagedFileDiffs, diff)
		git.SortFileDiffs(t.status.UnstagedFileDiffs)
	} else {
		t.status.UnstagedFileDiffs = deleteElement(t.status.UnstagedFileDiffs, t.AdjustedSelectedFileDiffIndex())
		t.status.StagedFileDiffs = append(t.status.StagedFileDiffs, diff)
		git.SortFileDiffs(t.status.StagedFileDiffs)
	}

	for i, d := range t.status.StagedFileDiffs {
		if d.Equals(diff) {
			t.selectedFileDiffIndex = i
			return
		}
	}

	for i, d := range t.status.UnstagedFileDiffs {
		if d.Equals(diff) {
			t.selectedFileDiffIndex = i + len(t.status.StagedFileDiffs)
			return
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

func (t TUI) sortDiffsByName() {
	sort.Slice(t.status.StagedFileDiffs, func(i, j int) bool {
		return t.status.StagedFileDiffs[i].Filename < t.status.StagedFileDiffs[j].Filename
	})
	sort.Slice(t.status.UnstagedFileDiffs, func(i, j int) bool {
		return t.status.UnstagedFileDiffs[i].Filename < t.status.UnstagedFileDiffs[j].Filename
	})
}

func (t TUI) createMenuString() string {
	if t.DiffMode() {
		return ""
	}

	type menuItem struct {
		key  string
		name string
	}

	var items []menuItem
	items = append(items, menuItem{key: "c", name: "Toggle"})
	items = append(items, menuItem{key: "enter", name: "Enter"})
	items = append(items, menuItem{key: "d", name: "Delete"})
	items = append(items, menuItem{key: "u", name: "Undo"})
	items = append(items, menuItem{key: "esc", name: "Apply"})

	s := "    "
	style := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF")).Background(lipgloss.Color("#6F8FAF")).Bold(false).PaddingLeft(1).PaddingRight(1)
	for i, item := range items {
		msg := ""
		if i > 0 {
			s += "  "
		}
		msg += fmt.Sprintf("%s (%s)", item.name, item.key)
		s += style.Render(msg)
	}
	return s
}

func (t TUI) isDeletedDiffIndex(index int) bool {
	_, ok := t.deletedDiff[index]
	return ok
}

func (t TUI) isSelectedDiffIndex(index int, stagedDiff bool) bool {
	if stagedDiff {
		return index == t.selectedFileDiffIndex
	} else {
		return index == t.AdjustedSelectedFileDiffIndex()
	}
}

func (tui TUI) Run() error {
	_, err := tea.NewProgram(tui).Run()
	return err
}
