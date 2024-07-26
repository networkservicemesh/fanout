// Copyright (c) 2020 Doc.ai and/or its affiliates.
//
// SPDX-License-Identifier: Apache-2.0
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at:
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package selector implements weighted random selection algorithm
package selector

import "math/rand/v2"

// WeightedRand selector picks elements randomly based on their weights
type WeightedRand[T any] struct {
	values      []T
	weights     []int
	totalWeight int
	r           *rand.Rand
}

// NewWeightedRandSelector inits WeightedRand by copying source values and calculating total weight
func NewWeightedRandSelector[T any](values []T, weights []int) *WeightedRand[T] {
	wrs := &WeightedRand[T]{
		values:      make([]T, len(values)),
		weights:     make([]int, len(weights)),
		totalWeight: 0,
		//nolint:gosec // it's overhead to use crypto/rand here
		r: rand.New(rand.NewPCG(rand.Uint64(), rand.Uint64())),
	}
	// copy the underlying array values as we're going to modify content of slices
	copy(wrs.values, values)
	copy(wrs.weights, weights)

	for _, w := range weights {
		wrs.totalWeight += w
	}

	return wrs
}

// Pick returns randomly chose element from values based on its weight if any exists
func (wrs *WeightedRand[T]) Pick() T {
	var defaultVal T
	if len(wrs.values) == 0 {
		return defaultVal
	}

	rNum := wrs.r.IntN(wrs.totalWeight) + 1

	sum := 0
	for i := range len(wrs.values) {
		sum += wrs.weights[i]
		if sum >= rNum {
			wrs.totalWeight -= wrs.weights[i]
			result := wrs.values[i]

			// remove picked element and its weight
			wrs.values[i] = wrs.values[len(wrs.values)-1]
			wrs.values = wrs.values[:len(wrs.values)-1]
			wrs.weights[i] = wrs.weights[len(wrs.weights)-1]
			wrs.weights = wrs.weights[:len(wrs.weights)-1]
			return result
		}
	}

	return defaultVal
}
