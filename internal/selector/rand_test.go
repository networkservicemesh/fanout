package selector

import (
	"math/rand/v2"
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
			expected:   []string{"f", "d", "g", "e", "a", "c", "b"},
		},
		"pick_all_different_weight": {
			values:     []string{"a", "b", "c", "d", "e", "f", "g"},
			weights:    []int{100, 70, 10, 50, 100, 30, 50},
			picksCount: 7,
			expected:   []string{"e", "d", "f", "g", "a", "c", "b"},
		},
		"pick_some_same_weight": {
			values:     []string{"a", "b", "c", "d", "e", "f", "g"},
			weights:    []int{100, 100, 100, 100, 100, 100, 100},
			picksCount: 3,
			expected:   []string{"f", "d", "g"},
		},
		"pick_some_different_weight": {
			values:     []string{"a", "b", "c", "d", "e", "f", "g"},
			weights:    []int{100, 70, 10, 50, 100, 30, 50},
			picksCount: 3,
			expected:   []string{"e", "d", "f"},
		},
		"pick_more_than_available": {
			values:     []string{"a", "b", "c"},
			weights:    []int{100, 70, 10},
			picksCount: 4,
			expected:   []string{"b", "a", "c", ""},
		},
	}
	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			wrs := NewWeightedRandSelector(tc.values, tc.weights)
			// init rand with constant seed to get predefined result
			//nolint:gosec
			wrs.r = rand.New(rand.NewPCG(1, 2))

			actual := make([]string, 0, tc.picksCount)
			for range tc.picksCount {
				actual = append(actual, wrs.Pick())
			}

			assert.Equal(t, tc.expected, actual)
		})
	}
}
