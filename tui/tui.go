package tui

import (
	"fmt"
	"gitdiff/git"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	fileSelectMode = "fileSelect"
	diffMode       = "diffMode"
)

var (
	titleStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFFDF5")).
		Background(lipgloss.Color("#25A065")).
		Padding(0, 1)
)

type TUI struct {
	state         *git.State
	selectedStyle lipgloss.Style

	selectedFileIndex int

	diff              []string
	selectedDiffIndex int
	deletedMap        map[int]bool
	deleteStack       []int

	mode string
}

func NewTUI(state *git.State, selectedStyle lipgloss.Style) *TUI {
	return &TUI{
		state:             state,
		selectedStyle:     selectedStyle,
		selectedFileIndex: 0,
		mode:              fileSelectMode,
	}
}

func (t TUI) SelectedChange() *git.File {
	return &t.state.Changes[t.selectedFileIndex]
}

func (t TUI) Init() tea.Cmd {
	return nil
}

func (t TUI) FileSelectMode() bool {
	return t.mode == fileSelectMode
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
				if t.selectedDiffIndex != -1 {
					t.deletedMap[t.selectedDiffIndex] = true
					t.deleteStack = append(t.deleteStack, t.selectedDiffIndex)
				}
			}
		case "u":
			if t.DiffMode() {
				if len(t.deleteStack) > 0 {
					delete(t.deletedMap, t.deleteStack[len(t.deleteStack)-1])
					t.deleteStack = t.deleteStack[:len(t.deleteStack)-1]
				}
			}
		case "q":
			if t.FileSelectMode() {
				return t, tea.Quit
			}
		case "esc":
			if t.FileSelectMode() {
				return t, tea.Quit
			} else if t.DiffMode() {
				t.mode = fileSelectMode
			}
		case "up", "k":
			if t.FileSelectMode() {
				if t.selectedFileIndex > 0 {
					t.selectedFileIndex--
				}
			} else if t.DiffMode() {
				prev := t.findPrevDiffLine(t.selectedDiffIndex)
				if prev != -1 {
					t.selectedDiffIndex = prev
				}
			}
		case "down", "j":
			if t.FileSelectMode() {
				if t.selectedFileIndex < len(t.state.Changes)-1 {
					t.selectedFileIndex++
				}
			} else if t.DiffMode() {
				next := t.findNextDiffLine(t.selectedDiffIndex)
				if next != -1 {
					t.selectedDiffIndex = next
				}
			}
		case "enter":
			if t.FileSelectMode() {
				lines, err := t.SelectedChange().Diff()
				if err != nil {
					return t, nil
				}
				t.diff = lines
				t.mode = diffMode
				t.selectedDiffIndex = t.findNextDiffLine(-1)
				t.deleteStack = make([]int, 0)
				t.deletedMap = make(map[int]bool, 0)
			} else if t.DiffMode() {
				// TODO: implement menu with options to back out line
			}
		}
	}

	return t, nil
}

func (t TUI) View() string {
	s := ""
	if t.FileSelectMode() {
		s += "\nChanges to be commited:\n"
		s += t.generateSectionString(true)

		s += "\nChanges not staged for commit:\n"
		s += t.generateSectionString(false)
	} else if t.DiffMode() {

		for i, line := range t.diff {
			_, ok := t.deletedMap[i]
			if ok {
				continue
			}

			if i == t.selectedDiffIndex {
				s += t.styleSelectedLine(line)
			} else {
				s += line
			}
			s += "\n"
		}
	}

	return s
}
func (t TUI) generateSectionString(staged bool) string {
	s := ""
	for i, f := range t.state.Changes {
		if f.Staged() == staged {
			continue
		}

		if i == t.selectedFileIndex {
			s += t.styleSelectedLine(f.Name())
		} else {
			s += fmt.Sprintf("%s", f.Name())
		}
		s += "\n"
	}
	return s
}

func (t TUI) findNextDiffLine(startingIndex int) int {
	for i := startingIndex + 1; i < len(t.diff); i++ {
		s := t.diff[i]
		if strings.HasPrefix(s, "-") && !strings.HasPrefix(s, "---") {
			return i
		}

		if strings.HasPrefix(s, "+") && !strings.HasPrefix(s, "+++") {
			return i
		}
	}

	return -1
}

func (t TUI) findPrevDiffLine(startingIndex int) int {
	for i := startingIndex - 1; i > -1; i-- {
		s := t.diff[i]
		if strings.HasPrefix(s, "-") && !strings.HasPrefix(s, "---") {
			return i
		}

		if strings.HasPrefix(s, "+") && !strings.HasPrefix(s, "+++") {
			return i
		}
	}

	return -1
}

func (t TUI) styleSelectedLine(s string) string {
	return t.selectedStyle.Render(s)
}

func (tui TUI) Run() error {
	_, err := tea.NewProgram(tui).Run()
	return err
}
