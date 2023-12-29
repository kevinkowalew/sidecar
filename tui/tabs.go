package tui

import (
	"gitdiff/git"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	inactiveTabBorder = tabBorderWithBottom("┴", "─", "┴")
	activeTabBorder   = tabBorderWithBottom("┘", " ", "└")
	docStyle          = lipgloss.NewStyle().Padding(0, 0, 0, 0)
	highlightColor    = lipgloss.AdaptiveColor{Light: "#874BFD", Dark: "#7D56F4"}
	inactiveTabStyle  = lipgloss.NewStyle().Border(inactiveTabBorder, true).BorderForeground(highlightColor).Padding(0, 1)
	activeTabStyle    = inactiveTabStyle.Copy().Border(activeTabBorder, true)
	windowStyle       = lipgloss.NewStyle().BorderForeground(highlightColor).Padding(0, 0).Border(lipgloss.NormalBorder()).UnsetBorderTop()
)

type Tabs struct {
	diffs                  []git.FileDiff
	scrollingText          *ScrollingText
	activeTab, screenWidth int
}

func NewTabs(diffs []git.FileDiff, screenWidth int) *Tabs {
	tabs := &Tabs{
		diffs:       diffs,
		screenWidth: screenWidth,
	}
	if len(tabs.diffs) > 0 {
		tabs.activeTab = 0
		tabs.scrollingText = NewScollingText(tabs.diffs[tabs.activeTab])
	}
	return tabs
}

func (t *Tabs) Init() tea.Cmd {
	return nil
}

func (t *Tabs) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch keypress := msg.String(); keypress {
		case "ctrl+c", "q":
			return t, tea.Quit
		case "right", "l", "n", "tab":
			t.activeTab = min(t.activeTab+1, len(t.diffs)-1)
			t.scrollingText = NewScollingText(t.diffs[t.activeTab])
			return t, nil
		case "left", "h", "p", "shift+tab":
			t.activeTab = max(t.activeTab-1, 0)
			t.scrollingText = NewScollingText(t.diffs[t.activeTab])
			return t, nil
		case "j":
			if t.scrollingText != nil {
				t.scrollingText.MoveCursorForward()
			}
		case "k":
			if t.scrollingText != nil {
				t.scrollingText.MoveCursorBackward()
			}
		case "d":
			if t.scrollingText != nil {
				t.scrollingText.Ignore()
			}
		case "c":
			if t.scrollingText != nil {
				t.scrollingText.Commit()
			}
		case "u":
			if t.scrollingText != nil {
				t.scrollingText.Undo()
			}
		}

	}

	return t, nil
}

func tabBorderWithBottom(left, middle, right string) lipgloss.Border {
	border := lipgloss.RoundedBorder()
	border.BottomLeft = left
	border.Bottom = middle
	border.BottomRight = right
	return border
}

func (t *Tabs) View() string {
	doc := strings.Builder{}

	var renderedTabs []string

	for i, diff := range t.diffs {
		var style lipgloss.Style
		isFirst, isLast, isActive := i == 0, i == len(t.diffs)-1, i == t.activeTab
		if isActive {
			style = activeTabStyle.Copy()
		} else {
			style = inactiveTabStyle.Copy()
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
		renderedTabs = append(renderedTabs, style.Render(padding+diff.Filename+padding))
	}

	row := lipgloss.JoinHorizontal(lipgloss.Top, renderedTabs...)
	doc.WriteString(row)
	doc.WriteString("\n")

	doc.WriteString(windowStyle.Width((lipgloss.Width(row) - windowStyle.GetHorizontalFrameSize())).Render(t.scrollingText.Render()))
	return docStyle.Render(doc.String())
}

func (t *Tabs) Run() error {
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
