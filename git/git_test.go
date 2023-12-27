package git

import "testing"

func TestSnippet_Equals(t *testing.T) {
	type testCase struct {
		name     string
		one, two Snippet
		expected bool
	}

	testCases := []testCase{
		{
			name:     "requires identical start",
			one:      Snippet{Start: 0},
			two:      Snippet{Start: 1},
			expected: false,
		},
		{
			name:     "requires identical length",
			one:      Snippet{Start: 0, Length: 1},
			two:      Snippet{Start: 0, Length: 2},
			expected: false,
		},
		{
			name:     "requires requires Lines to be equivalent",
			one:      Snippet{Start: 0, Length: 2, Lines: []string{"one", "twoh"}},
			two:      Snippet{Start: 0, Length: 2, Lines: []string{"one", "two"}},
			expected: false,
		},
		{
			name:     "happy path",
			one:      Snippet{Start: 0, Length: 2, Lines: []string{"one", "two"}},
			two:      Snippet{Start: 0, Length: 2, Lines: []string{"one", "two"}},
			expected: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual := tc.one.Equals(tc.two)
			if actual != tc.expected {
				t.Fail()
			}
		})
	}

}

func TestFileDiff_Equals(t *testing.T) {
	type testCase struct {
		name     string
		one, two FileDiff
		expected bool
	}

	testCases := []testCase{
		{
			name: "requires identical diff filename",
			one: FileDiff{
				Filename: "one",
			},
			two: FileDiff{
				Filename: "two",
			},
			expected: false,
		},
		{
			name: "requires identical chunks",
			one: FileDiff{
				Filename: "one",
				Chunks: []FileDiffChunk{
					{
						Old: Snippet{Start: 0, Length: 0, Lines: []string{"one"}},
						New: Snippet{Start: 0, Length: 0, Lines: []string{"two"}},
					},
				},
			},
			two: FileDiff{
				Filename: "one",
				Chunks: []FileDiffChunk{
					{
						Old: Snippet{Start: 0, Length: 0, Lines: []string{}},
						New: Snippet{Start: 0, Length: 0, Lines: []string{}},
					},
				},
			},
			expected: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual := tc.one.Equals(tc.two)
			if actual != tc.expected {
				t.Fail()
			}
		})
	}
}
