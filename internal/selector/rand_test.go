package selector

import (
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWeightedRand_Pick(t *testing.T) {
	testCases := map[string]struct {
		values  []string
		weights []int

		picksCount int

		expected []string
	}{
		"pick_all_same_weight": {
			values:     []string{"a", "b", "c", "d", "e", "f", "g"},
			weights:    []int{100, 100, 100, 100, 100, 100, 100},
			picksCount: 7,
			expected:   []string{"b", "a", "d", "f", "g", "c", "e"},
		},
		"pick_all_different_weight": {
			values:     []string{"a", "b", "c", "d", "e", "f", "g"},
			weights:    []int{100, 70, 10, 50, 100, 30, 50},
			picksCount: 7,
			expected:   []string{"d", "a", "e", "b", "g", "c", "f"},
		},
		"pick_some_same_weight": {
			values:     []string{"a", "b", "c", "d", "e", "f", "g"},
			weights:    []int{100, 100, 100, 100, 100, 100, 100},
			picksCount: 3,
			expected:   []string{"b", "a", "d"},
		},
		"pick_some_different_weight": {
			values:     []string{"a", "b", "c", "d", "e", "f", "g"},
			weights:    []int{100, 70, 10, 50, 100, 30, 50},
			picksCount: 3,
			expected:   []string{"d", "a", "e"},
		},
		"pick_more_than_available": {
			values:     []string{"a", "b", "c"},
			weights:    []int{70, 10, 100},
			picksCount: 4,
			expected:   []string{"a", "c", "b", ""},
		},
	}
	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			wrs := NewWeightedRandSelector(tc.values, tc.weights)
			// init rand with constant seed to get predefined result
			//nolint:gosec
			wrs.r = rand.New(rand.NewSource(1))

			actual := make([]string, 0, tc.picksCount)
			for i := 0; i < tc.picksCount; i++ {
				actual = append(actual, wrs.Pick())
			}

			assert.Equal(t, tc.expected, actual)
		})
	}
}
