package diff

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseChunk(t *testing.T) {
	type testCase struct {
		name     string
		filename string
		expected DiffChunk
	}

	testCases := []testCase{
		{
			name:     "deletions and additions",
			filename: "one.diff",
			expected: DiffChunk{
				Old: Snippet{
					start:  19,
					length: 3,
					lines: []string{
						"-               state:         state,",
						"-               selectedStyle: selectedStyle,",
						"-               selectedIndex: 0,",
					},
				},
				New: Snippet{
					start:  40,
					length: 4,
					lines: []string{
						"+               state:             state,",
						"+               selectedStyle:     selectedStyle,",
						"+               selectedFileIndex: 0,",
						"+               mode:              fileSelectMode,`,",
					},
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			//actual := ParseChunk(tc.file)
			//assert.Equal(tc.expected, actual)
		})
	}
}
