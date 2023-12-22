package diff

type Snippet struct {
	filename string
	start    int
	length   int
	lines    []string
}

type DiffChunk struct {
	Old Snippet
	New Snippet
}

func ParseChunk(diff string) *DiffChunk {
	return nil
}

type Diff struct {
	Chunks []DiffChunk
}
