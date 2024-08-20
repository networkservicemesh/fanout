// Copyright (c) 2024 MWS and/or its affiliates.
//
// SPDX-License-Identifier: Apache-2.0
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at:
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
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
			for i := 0; i < tc.picksCount; i++ {
				actual = append(actual, wrs.Pick())
			}

			assert.Equal(t, tc.expected, actual)
		})
	}
}
