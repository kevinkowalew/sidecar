package git

import "testing"

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
				Chunks: []Chunk{
					{
						Old: Snippet{Start: 0, Length: 0, Lines: []string{"one"}},
						New: Snippet{Start: 0, Length: 0, Lines: []string{"two"}},
					},
				},
			},
			two: FileDiff{
				Filename: "one",
				Chunks: []Chunk{
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
