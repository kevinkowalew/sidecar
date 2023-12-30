package sidecar

import (
	"sidecar/git"
	"sidecar/tui"
	"time"

	"github.com/charmbracelet/lipgloss"
)

type Sidecar struct {
	tui            *tui.TUI
	refreshTicker  *time.Ticker
	highlightColor lipgloss.AdaptiveColor
	screenWidth    int
}

func NewSidecar(screenWidth, refreshRateInSeconds int) *Sidecar {
	return &Sidecar{
		screenWidth:    screenWidth,
		refreshTicker:  time.NewTicker(time.Duration(refreshRateInSeconds) * time.Second),
		highlightColor: lipgloss.AdaptiveColor{Light: "#874BFD", Dark: "#7D56F4"},
	}
}

func (s *Sidecar) Run() error {
	diffs, err := git.UnstagedDiffs()
	if err != nil {
		return err
	}
	s.tui = tui.NewTUI(diffs, s.screenWidth, s.highlightColor)

	go func() {
		for {
			select {
			case <-s.refreshTicker.C:
				diffs, err := git.UnstagedDiffs()
				if err != nil {
					return
				}
				s.tui.UpdateDiffs(diffs)
			}
		}
	}()

	return s.tui.Run()
}
