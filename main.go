package main

import (
	"gitdiff/git"
	"gitdiff/tui"
)

func main() {
	s, err := git.Current()
	if err != nil {
		panic(err)
	}

	t := tui.NewTabs(s.UnstagedFileDiffs, 100)
	if err = t.Run(); err != nil {
		panic(err)
	}
}
