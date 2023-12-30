package git

import "strings"

type Chunk struct {
	Old Snippet
	New Snippet
}

func (c Chunk) Equals(oc Chunk) bool {
	return c.Old.Equals(oc.Old) && c.New.Equals(oc.New)
}

func parseChunk(filename string, lines []string) (*Chunk, error) {
	if len(lines) == 0 {
		return nil, parseError("empty input", filename)
	}

	headerParts := strings.Split(lines[0], " ")
	if len(headerParts) < 4 {
		return nil, parseError(filename, "invalid header: "+lines[0])
	}

	oldSnippet, err := parseSnippet(headerParts[1])
	if err != nil {
		return nil, parseError(filename, "unable to parse old snippet: "+err.Error())
	}

	newSnippet, err := parseSnippet(headerParts[2])
	if err != nil {
		return nil, parseError(filename, "unable to parse new snippet: "+headerParts[2])
	}

	for i := 1; i < len(lines); i++ {
		if i <= oldSnippet.Length {
			oldSnippet.Lines = append(oldSnippet.Lines, lines[i])
		} else {
			newSnippet.Lines = append(newSnippet.Lines, lines[i])
		}
	}

	return &Chunk{*oldSnippet, *newSnippet}, nil
}
