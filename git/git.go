package git

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"sync"
)

func UnstagedDiffs() ([]Diff, error) {
	cmds := []string{"diff", "--name-only"}

	filenames, err := git(cmds...)
	if err != nil {
		return nil, err
	}

	type result struct {
		filename string
		diff     *Diff
		err      error
	}

	var wg sync.WaitGroup
	resultChan := make(chan result, len(filenames))

	for _, name := range filenames {
		if len(name) == 0 || strings.Contains(name, ".diff") {
			continue
		}
		wg.Add(1)

		go func(filename string) {
			diff, err := diffFile(filename)
			resultChan <- result{
				filename: filename,
				diff:     diff,
				err:      err,
			}
			wg.Done()
		}(name)
	}

	wg.Wait()
	close(resultChan)

	var diffs []Diff
	for result := range resultChan {
		if result.err != nil {
			return nil, fmt.Errorf("failed to generate diff for file (%s): %s", result.filename, result.err.Error())
		}
		diffs = append(diffs, *result.diff)
	}

	return diffs, err
}

func diffFile(filename string) (*Diff, error) {
	_, err := os.Stat(filename)
	if err != nil {
		return &Diff{
			Filename: filename,
		}, nil
	}

	var lines []string
	lines, err = git("diff", "--unified=0", filename)

	if err != nil {
		return nil, fmt.Errorf("failed to get diff for file (%s): %s", filename, err.Error())
	}

	return ParseDiff(filename, lines)
}

func git(cmds ...string) ([]string, error) {
	output, err := exec.Command("git", cmds...).Output()
	if err != nil {
		return nil, err
	}

	lines := strings.Split(string(output), "\n")
	return lines, nil
}

func parseError(filename, msg string) error {
	return fmt.Errorf("unable to parse Chunk (%s): %s", filename, msg)
}
