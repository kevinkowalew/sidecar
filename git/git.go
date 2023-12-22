package git

import (
	"os/exec"
	"strings"
)

type State struct {
	Changes []File
}

func LoadCurrentState() (*State, error) {
	unstagedFiles, err := git("diff", "--name-only")
	if err != nil {
		return nil, err
	}

	files := make([]File, 0)
	files = append(files, createFiles(unstagedFiles, UnstagedStatus)...)

	stagedFiles, err := git("diff", "--staged", "--name-only")
	if err != nil {
		return nil, err
	}
	files = append(files, createFiles(stagedFiles, StagedStatus)...)

	return &State{files}, nil
}

func createFiles(filenames []string, status string) []File {
	files := make([]File, 0)
	for _, name := range filenames {
		if len(name) == 0 {
			continue
		}

		files = append(files, File{
			name:   name,
			status: status,
		})
	}
	return files
}

func git(cmds ...string) ([]string, error) {
	output, err := exec.Command("git", cmds...).Output()
	if err != nil {
		return nil, err
	}

	lines := strings.Split(string(output), "\n")
	return lines, nil
}
