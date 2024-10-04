package slice_test

import (
	"testing"

	"github.com/sourcenetwork/defradb/internal/utils/slice"
	"github.com/sourcenetwork/immutable"
	"github.com/stretchr/testify/assert"
)

func TestRemoveFirstIf(t *testing.T) {
	tests := []struct {
		name      string
		input     []int
		predicate func(int) bool
		expected  []int
		found     immutable.Option[int]
	}{
		{
			name:      "remove in the middle",
			input:     []int{1, 3, 4, 5, 6},
			predicate: func(n int) bool { return n%2 == 0 },
			expected:  []int{1, 3, 6, 5},
			found:     immutable.Some(4),
		},
		{
			name:      "nothing removed",
			input:     []int{1, 3, 4, 5, 6},
			predicate: func(n int) bool { return n > 10 },
			expected:  []int{1, 3, 4, 5, 6},
			found:     immutable.None[int](),
		},
		{
			name:      "empty slice",
			input:     []int{},
			predicate: func(n int) bool { return n == 5 },
			expected:  []int{},
			found:     immutable.None[int](),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, found := slice.RemoveFirstIf(tt.input, tt.predicate)
			assert.Equal(t, tt.expected, result, "expected %v, got %v", tt.expected, result)
			assert.Equal(t, tt.found, found, "expected found %v, got %v", tt.found, found)
		})
	}
}

func TestRemoveDuplicates(t *testing.T) {
	tests := []struct {
		name     string
		input    []int
		expected []int
	}{
		{
			name:     "no duplicates",
			input:    []int{1, 2, 3, 4, 5},
			expected: []int{1, 2, 3, 4, 5},
		},
		{
			name:     "all duplicates",
			input:    []int{1, 1, 1, 1, 1},
			expected: []int{1},
		},
		{
			name:     "some duplicates",
			input:    []int{1, 2, 4, 2, 3, 4, 4, 5},
			expected: []int{1, 2, 3, 4, 5},
		},
		{
			name:     "empty slice",
			input:    []int{},
			expected: []int{},
		},
		{
			name:     "single element",
			input:    []int{1},
			expected: []int{1},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := slice.RemoveDuplicates(tt.input)
			assert.ElementsMatch(t, tt.expected, result, "expected %v, got %v", tt.expected, result)
		})
	}
}
