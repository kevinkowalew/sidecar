package tui

import (
	"gitdiff/git"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	inselectedIndexBorder = tabBorderWithBottom("┴", "─", "┴")
	selectedIndexBorder   = tabBorderWithBottom("┘", " ", "└")
	docStyle              = lipgloss.NewStyle().Padding(0, 0, 0, 0)
	highlightColor        = lipgloss.AdaptiveColor{Light: "#874BFD", Dark: "#7D56F4"}
	inselectedIndexStyle  = lipgloss.NewStyle().Border(inselectedIndexBorder, true).BorderForeground(highlightColor).Padding(0, 1)
	selectedIndexStyle    = inselectedIndexStyle.Copy().Border(selectedIndexBorder, true)
	windowStyle           = lipgloss.NewStyle().BorderForeground(highlightColor).Padding(0, 0).Border(lipgloss.NormalBorder()).UnsetBorderTop()
)

type TUI struct {
	diffs                      []git.FileDiff
	textViews                  []*ScrollingText
	selectedIndex, screenWidth int
}

func NewTUI(diffs []git.FileDiff, screenWidth int) *TUI {
	tabs := &TUI{
		diffs:       diffs,
		screenWidth: screenWidth,
	}

	for _, diff := range diffs {
		tabs.textViews = append(tabs.textViews, NewScollingText(diff))
	}

	return tabs
}

func (t *TUI) Init() tea.Cmd {
	return nil
}

func (t *TUI) textView() *ScrollingText {
	if t.selectedIndex < len(t.textViews) {
		return t.textViews[t.selectedIndex]
	}
	return nil
}

func (t *TUI) allSelectionsComplete() bool {
	for _, textView := range t.textViews {
		if !textView.Complete() {
			return false
		}
	}
	return true
}

func (t *TUI) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch keypress := msg.String(); keypress {
		case "ctrl+c", "q":
			return t, tea.Quit
		case "right", "l", "n", "tab":
			if t.selectedIndex < len(t.diffs)-1 {
				t.selectedIndex++
			} else {
				t.selectedIndex = 0
			}
			return t, nil
		case "left", "h", "p", "shift+tab":
			if t.selectedIndex > 0 {
				t.selectedIndex--
			} else {
				t.selectedIndex = len(t.diffs) - 1
			}
			return t, nil
		case "j":
			if t.textView() != nil {
				t.textView().MoveCursorForward()
			}
		case "k":
			if t.textView() != nil {
				t.textView().MoveCursorBackward()
			}
		case "c":
			if t.textView() != nil {
				t.textView().Commit()
				if t.textView().Complete() {
					t.removeDiff()
				}
			}
		case "u":
			if t.textView() != nil {
				t.textView().Undo()
			}
		}

	}

	return t, nil
}

func (t *TUI) removeDiff() {
	if len(t.diffs) == 1 {
		t.diffs = make([]git.FileDiff, 0)
		t.textViews = make([]*ScrollingText, 0)
		return
	}

	if t.selectedIndex >= 0 && t.selectedIndex < len(t.textViews) {
		t.diffs = append(t.diffs[:t.selectedIndex], t.diffs[t.selectedIndex+1:]...)
		t.textViews = append(t.textViews[:t.selectedIndex], t.textViews[t.selectedIndex+1:]...)

		if t.selectedIndex == len(t.textViews) {
			t.selectedIndex--
		}
	}
}

func tabBorderWithBottom(left, middle, right string) lipgloss.Border {
	border := lipgloss.RoundedBorder()
	border.BottomLeft = left
	border.Bottom = middle
	border.BottomRight = right
	return border
}

func (t *TUI) View() string {
	if len(t.diffs) == 0 {
		return "Scanning for changes..."
	}
	doc := strings.Builder{}

	var renderedTUI []string

	for i, diff := range t.diffs {
		var style lipgloss.Style
		isFirst, isLast, isActive := i == 0, i == len(t.diffs)-1, i == t.selectedIndex
		if isActive {
			style = selectedIndexStyle.Copy()
		} else {
			style = inselectedIndexStyle.Copy()
		}
		border, _, _, _, _ := style.GetBorder()
		if isFirst && isActive {
			border.BottomLeft = "│"
		} else if isFirst && !isActive {
			border.BottomLeft = "├"
		} else if isLast && isActive {
			border.BottomRight = "│"
		} else if isLast && !isActive {
			border.BottomRight = "┤"
		}

		style = style.Border(border)
		tabWidth := t.screenWidth / len(t.diffs)
		paddingCount := (tabWidth - len(diff.Filename)) / 2
		padding := strings.Repeat(" ", paddingCount)
		renderedTUI = append(renderedTUI, style.Render(padding+diff.Filename+padding))
	}

	row := lipgloss.JoinHorizontal(lipgloss.Top, renderedTUI...)
	doc.WriteString(row)
	doc.WriteString("\n")
	textView := t.textViews[t.selectedIndex]

	if textView.Complete() {
		windowStyle = windowStyle.Align(lipgloss.Center)
	} else {
		windowStyle = windowStyle.Align(lipgloss.Left)
	}
	doc.WriteString(windowStyle.Width((lipgloss.Width(row) - windowStyle.GetHorizontalFrameSize())).Render(textView.Render()))
	return docStyle.Render(doc.String())
}

func (t *TUI) Run() error {
	_, err := tea.NewProgram(t).Run()
	return err
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
