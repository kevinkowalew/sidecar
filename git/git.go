package git

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"sort"
	"strconv"
	"strings"
	"sync"
)

const (
	StagedStatus   = "staged"
	UnstagedStatus = "unstaged"
)

type (
	Snippet struct {
		start  int
		length int
		Lines  []string
	}

	FileDiffChunk struct {
		Old Snippet
		New Snippet
	}

	FileDiff struct {
		Filename string
		Chunks   []FileDiffChunk
	}

	Status struct {
		UnstagedFileDiffs []FileDiff
		StagedFileDiffs   []FileDiff
	}
)

func (fd FileDiff) Equals(od FileDiff) bool {
	if fd.Filename != od.Filename {
		return false
	}

	if len(fd.Chunks) != len(od.Chunks) {
		return false
	}

	for i := range fd.Chunks {
		if !fd.Chunks[i].Equals(od.Chunks[i]) {
			return false
		}
	}

	return true
}

func (c FileDiffChunk) Equals(oc FileDiffChunk) bool {
	return c.Old.Equals(oc.Old) && c.New.Equals(oc.New)
}

func (s Snippet) Equals(os Snippet) bool {
	return s.start == os.start && s.length == os.length && slicesEqual(s.Lines, os.Lines)
}

func slicesEqual(s1, s2 []string) bool {
	if len(s1) != len(s2) {
		return false
	}

	for i := range s1 {
		if s1[i] != s2[i] {
			return false
		}
	}

	return true
}

func (fd FileDiff) Deleted() bool {
	return fd.Filename != "" && len(fd.Chunks) == 0
}

func Current() (*Status, error) {
	type result struct {
		staged bool
		diffs  []FileDiff
		err    error
	}

	var wg sync.WaitGroup
	resultChan := make(chan result, 2)

	wg.Add(2)
	runner := func(staged bool) {
		diffs, err := getFileDiffs(staged)
		resultChan <- result{staged: staged, diffs: diffs, err: err}
		wg.Done()
	}
	go runner(false)
	go runner(true)

	wg.Wait()
	close(resultChan)

	status := &Status{}
	for result := range resultChan {
		if result.err != nil {
			diffType := "unstaged"
			if result.staged {
				diffType = "staged"
			}
			if result.staged {
				return nil, fmt.Errorf("failed to load %s diffs: %s", diffType, result.err.Error())
			}
		}

		if result.staged {
			status.StagedFileDiffs = result.diffs
		} else {
			status.UnstagedFileDiffs = result.diffs
		}
	}
	SortFileDiffs(status.StagedFileDiffs)
	SortFileDiffs(status.UnstagedFileDiffs)
	return status, nil
}

func SortFileDiffs(diffs []FileDiff) {
	sort.Slice(diffs, func(i, j int) bool {
		return diffs[i].Filename < diffs[j].Filename
	})
}
func getFileDiffs(staged bool) ([]FileDiff, error) {
	cmds := []string{"diff", "--name-only"}
	if staged {
		cmds = append(cmds, "--staged")
	}

	filenames, err := git(cmds...)
	if err != nil {
		return nil, err
	}

	type result struct {
		filename string
		diff     *FileDiff
		err      error
	}

	var wg sync.WaitGroup
	resultChan := make(chan result, len(filenames))

	for _, name := range filenames {
		// TODO: update parsers to accept .diffs files
		if len(name) == 0 || strings.Contains(name, ".diff") {
			continue
		}
		wg.Add(1)

		go func(filename string, staged bool) {
			diff, err := diffFile(filename, staged)
			resultChan <- result{
				filename: filename,
				diff:     diff,
				err:      err,
			}
			wg.Done()
		}(name, staged)

	}

	wg.Wait()
	close(resultChan)

	var diffs []FileDiff
	for result := range resultChan {
		if result.err != nil {
			return nil, fmt.Errorf("failed to generate diff for file (%s): %s", result.filename, result.err.Error())
		}
		diffs = append(diffs, *result.diff)
	}

	return diffs, err
}

func diffFile(filename string, staged bool) (*FileDiff, error) {
	_, err := os.Stat(filename)
	if err != nil {
		return &FileDiff{
			Filename: filename,
		}, nil
	}

	var lines []string
	if staged {
		lines, err = git("diff", "--staged", "--unified=0", filename)
	} else {
		lines, err = git("diff", "--unified=0", filename)
	}

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

func ParseDiff(filename string, lines []string) (*FileDiff, error) {
	if len(lines) < 5 {
		return nil, errors.New("file diff has too few lines to parse")
	}

	partitions := make([][]string, 0)
	partition := make([]string, 0)
	for i := 4; i < len(lines); i++ {
		line := lines[i]

		if strings.HasPrefix(line, "@@") && len(partition) == 0 {
			partition = append(partition, line)
		} else if strings.HasPrefix(line, "@@") && len(partition) > 0 {
			partitions = append(partitions, partition)
			partition = make([]string, 0)
			partition = append(partition, line)
		} else {
			partition = append(partition, line)
		}
	}

	chunks := make([]FileDiffChunk, 0)
	for _, partition := range partitions {
		chunk, err := parseFileDiffChunk(filename, partition)
		if err != nil {
			return nil, fmt.Errorf("failed to parse paritition\n\n%s\n\n%v", err.Error(), partition)
		}
		chunks = append(chunks, *chunk)
	}
	return &FileDiff{
		Filename: filename,
		Chunks:   chunks,
	}, nil
}

func parseFileDiffChunk(filename string, lines []string) (*FileDiffChunk, error) {
	if len(lines) == 0 {
		return nil, parseError("empty input")
	}

	headerParts := strings.Split(lines[0], " ")
	if len(headerParts) < 5 {
		return nil, parseError("invalid header: " + lines[0])
	}

	oldSnippet, err := parseSnippet(headerParts[1])
	if err != nil {
		return nil, parseError("unable to parse old snippet: " + err.Error())
	}

	newSnippet, err := parseSnippet(headerParts[2])
	if err != nil {
		return nil, parseError("unable to parse new snippet: " + headerParts[2])
	}

	for i := 1; i < len(lines); i++ {
		if i <= oldSnippet.length {
			oldSnippet.Lines = append(oldSnippet.Lines, lines[i])
		} else {
			newSnippet.Lines = append(newSnippet.Lines, lines[i])
		}
	}

	return &FileDiffChunk{*oldSnippet, *newSnippet}, nil
}

func parseSnippet(s string) (*Snippet, error) {
	p := strings.Split(s, ",")

	if len(p[0]) < 2 {
		return nil, errors.New("snippet prefix invalid length: " + p[0])
	}

	start, err := strconv.Atoi(p[0][1:])
	if err != nil {
		return nil, errors.New("failed to parse start: " + err.Error())
	}

	snippet := &Snippet{start: start}
	if len(p) == 1 {
		snippet.length = 1
		return snippet, nil
	}

	length, err := strconv.Atoi(p[1])
	if err != nil {
		return nil, errors.New("failed to parse length: " + err.Error())
	}
	snippet.length = length
	return snippet, nil
}

func parseError(msg string) error {
	return fmt.Errorf("unable to parse FileDiffChunk: %s", msg)
}
