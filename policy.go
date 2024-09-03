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

package fanout

import "github.com/networkservicemesh/fanout/internal/selector"

type policy interface {
	selector(clients []Client) clientSelector
}

type clientSelector interface {
	Pick() Client
}

// sequentialPolicy is used to select clients based on its sequential order
type sequentialPolicy struct {
}

// creates new sequential selector of provided clients
func (p *sequentialPolicy) selector(clients []Client) clientSelector {
	return selector.NewSequentialSelector(clients)
}

// weightedPolicy is used to select clients randomly based on its loadFactor (weights)
type weightedPolicy struct {
	loadFactor []int
}

// creates new weighted random selector of provided clients based on loadFactor
func (p *weightedPolicy) selector(clients []Client) clientSelector {
	return selector.NewWeightedRandSelector(clients, p.loadFactor)
}
