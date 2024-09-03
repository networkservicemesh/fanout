// Copyright (c) 2024 MWS and/or its affiliates.
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

package selector

// Sequential selector acts like a queue and picks elements one-by-one starting from the first element
type Sequential[T any] struct {
	values []T
	idx    int
}

// NewSequentialSelector inits Sequential selector with default starting index 0
func NewSequentialSelector[T any](values []T) *Sequential[T] {
	return &Sequential[T]{
		values: values,
		idx:    0,
	}
}

// Pick returns next available element from values array if exists.
// Returns default value of type T otherwise
func (s *Sequential[T]) Pick() T {
	var result T
	if s.idx >= len(s.values) {
		return result
	}

	result = s.values[s.idx]
	s.idx++

	return result
}
