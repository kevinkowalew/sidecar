package main

import (
	"fmt"
	"gitdiff/git"
	"gitdiff/tui"
	"os"
	"strconv"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Println("Missing required screen width command argument")
		return
	}

	width, err := strconv.Atoi(os.Args[1])
	if err != nil {
		msg := fmt.Sprintf("Failed to parse screen width argument (%s): %s", os.Args[1], err.Error())
		fmt.Println(msg)
		return
	}

	// TODO: implement lazy hot load
	s, err := git.Current()
	if err != nil {
		panic(err)
	}

	t := tui.NewTUI(s.UnstagedFileDiffs, width-20)
	if err = t.Run(); err != nil {
		panic(err)
	}
}
