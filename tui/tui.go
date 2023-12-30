package tui

import (
	"sidecar/git"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type TUI struct {
	Diffs                      []git.Diff
	textViews                  []*ScrollingText
	selectedIndex, screenWidth int

	docStyle, inselectedIndexStyle, selectedIndexStyle, windowStyle lipgloss.Style
	program                                                         *tea.Program
}

func NewTUI(diffs []git.Diff, screenWidth int, highlightColor lipgloss.AdaptiveColor) *TUI {
	inselectedIndexBorder := tabBorderWithBottom("┴", "─", "┴")
	selectedIndexBorder := tabBorderWithBottom("┘", " ", "└")
	inselectedIndexStyle := lipgloss.NewStyle().Border(inselectedIndexBorder, true).BorderForeground(highlightColor).Padding(0, 1)

	tabs := &TUI{
		Diffs:                diffs,
		screenWidth:          screenWidth,
		docStyle:             lipgloss.NewStyle().Padding(0, 0, 0, 0),
		inselectedIndexStyle: inselectedIndexStyle,
		selectedIndexStyle:   inselectedIndexStyle.Copy().Border(selectedIndexBorder, true),
		windowStyle:          lipgloss.NewStyle().BorderForeground(highlightColor).Padding(0, 0).Border(lipgloss.NormalBorder()).UnsetBorderTop(),
	}

	for _, diff := range diffs {
		tabs.textViews = append(tabs.textViews, NewScollingText(diff, highlightColor))
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

func (t *TUI) UpdateDiffs(diffs []git.Diff) {
	if !git.DiffSlicesEqual(t.Diffs, diffs) {
		t.Diffs = diffs
	}
}

func (t *TUI) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch keypress := msg.String(); keypress {
		case "ctrl+c", "q":
			return t, tea.Quit
		case "right", "l", "n", "tab":
			if t.selectedIndex < len(t.Diffs)-1 {
				t.selectedIndex++
			} else {
				t.selectedIndex = 0
			}
			return t, nil
		case "left", "h", "p", "shift+tab":
			if t.selectedIndex > 0 {
				t.selectedIndex--
			} else {
				t.selectedIndex = len(t.Diffs) - 1
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
	if len(t.Diffs) == 1 {
		t.Diffs = make([]git.Diff, 0)
		t.textViews = make([]*ScrollingText, 0)
		return
	}

	if t.selectedIndex >= 0 && t.selectedIndex < len(t.textViews) {
		t.Diffs = append(t.Diffs[:t.selectedIndex], t.Diffs[t.selectedIndex+1:]...)
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
	if len(t.Diffs) == 0 {
		return "Scanning for changes..."
	}
	doc := strings.Builder{}

	var renderedTUI []string

	for i, diff := range t.Diffs {
		var style lipgloss.Style
		isFirst, isLast, isActive := i == 0, i == len(t.Diffs)-1, i == t.selectedIndex
		if isActive {
			style = t.selectedIndexStyle.Copy()
		} else {
			style = t.inselectedIndexStyle.Copy()
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
		tabWidth := t.screenWidth / len(t.Diffs)
		paddingCount := (tabWidth - len(diff.Filename)) / 2
		padding := strings.Repeat(" ", paddingCount)
		renderedTUI = append(renderedTUI, style.Render(padding+diff.Filename+padding))
	}

	row := lipgloss.JoinHorizontal(lipgloss.Top, renderedTUI...)
	doc.WriteString(row)
	doc.WriteString("\n")
	textView := t.textViews[t.selectedIndex]

	if textView.Complete() {
		t.windowStyle = t.windowStyle.Align(lipgloss.Center)
	} else {
		t.windowStyle = t.windowStyle.Align(lipgloss.Left)
	}
	doc.WriteString(t.windowStyle.Width((lipgloss.Width(row) - t.windowStyle.GetHorizontalFrameSize())).Render(textView.Render()))
	return t.docStyle.Render(doc.String())
}

func (t *TUI) Run() error {
	t.program = tea.NewProgram(t)
	_, err := t.program.Run()
	return err
}

func (t *TUI) Kill() {
	if t.program != nil {
		t.program.Kill()
	}
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
