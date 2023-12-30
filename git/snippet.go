package git

import (
	"errors"
	"strconv"
	"strings"
)

type Snippet struct {
	Start  int
	Length int
	Lines  []string
}

func (s Snippet) Equals(os Snippet) bool {
	if s.Start != os.Start {
		return false
	}
	return s.Start == os.Start && s.Length == os.Length && slicesEqual(s.Lines, os.Lines)
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

func parseSnippet(s string) (*Snippet, error) {
	p := strings.Split(s, ",")

	if len(p[0]) < 2 {
		return nil, errors.New("snippet prefix invalid length: " + p[0])
	}

	start, err := strconv.Atoi(p[0][1:])
	if err != nil {
		return nil, errors.New("failed to parse start: " + err.Error())
	}

	snippet := &Snippet{Start: start}
	if len(p) == 1 {
		snippet.Length = 1
		return snippet, nil
	}

	length, err := strconv.Atoi(p[1])
	if err != nil {
		return nil, errors.New("failed to parse length: " + err.Error())
	}
	snippet.Length = length
	return snippet, nil
}
