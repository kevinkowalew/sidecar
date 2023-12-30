package git

import (
	"errors"
	"fmt"
	"strings"
)

type Diff struct {
	Filename string
	Chunks   []Chunk
}

func DiffSlicesEqual(ds, ods []Diff) bool {
	if len(ds) != len(ods) {
		return false
	}

	for i := 0; i < len(ds); i++ {
		if !ds[i].Equals(ods[i]) {
			return false
		}
	}

	return true
}

func (fd Diff) Equals(od Diff) bool {
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

func (fd Diff) Deleted() bool {
	return fd.Filename != "" && len(fd.Chunks) == 0
}

func ParseDiff(filename string, lines []string) (*Diff, error) {
	if len(lines) < 5 {
		return nil, errors.New("file diff has too few lines to parse")
	}

	partitions := make([][]string, 0)
	partition := make([]string, 0)
	for i := 4; i < len(lines); i++ {
		line := lines[i]

		if strings.HasPrefix(line, "@@") {
			if len(partition) > 0 {
				partitions = append(partitions, partition)
				partition = make([]string, 0)
			}
			partition = append(partition, line)
		} else {
			partition = append(partition, line)
		}
	}
	partitions = append(partitions, partition)

	chunks := make([]Chunk, 0)
	for _, partition := range partitions {
		chunk, err := parseChunk(filename, partition)
		if err != nil {
			return nil, fmt.Errorf("failed to parse paritition\n\n%s\n\n%v", err.Error(), partition)
		}
		chunks = append(chunks, *chunk)
	}
	return &Diff{
		Filename: filename,
		Chunks:   chunks,
	}, nil
}
