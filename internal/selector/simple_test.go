package selector

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSimple_Pick(t *testing.T) {
	testCases := map[string]struct {
		values []string

		picksCount int

		expected []string
	}{
		"pick_all": {
			values:     []string{"a", "b", "c", "d", "e", "f", "g"},
			picksCount: 7,
			expected:   []string{"a", "b", "c", "d", "e", "f", "g"},
		},
		"pick_some": {
			values:     []string{"a", "b", "c", "d", "e", "f", "g"},
			picksCount: 3,
			expected:   []string{"a", "b", "c"},
		},
		"pick_more_than_available": {
			values:     []string{"a", "b", "c"},
			picksCount: 4,
			expected:   []string{"a", "b", "c", ""},
		},
	}
	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			wrs := NewSimpleSelector(tc.values)

			actual := make([]string, 0, tc.picksCount)
			for range tc.picksCount {
				actual = append(actual, wrs.Pick())
			}

			assert.Equal(t, tc.expected, actual)
		})
	}
}
