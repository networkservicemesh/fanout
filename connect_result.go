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

package fanout

import (
	"time"

	"github.com/miekg/dns"
)

type response struct {
	client   Client
	response *dns.Msg
	start    time.Time
	err      error
}

func isBetter(left, right *response) bool {
	if right == nil {
		return false
	}
	if left == nil {
		return true
	}
	if right.err != nil {
		return false
	}
	if left.err != nil {
		return true
	}
	if right.response == nil {
		return false
	}
	if left.response == nil {
		return true
	}
	return left.response.MsgHdr.Rcode != dns.RcodeSuccess &&
		right.response.MsgHdr.Rcode == dns.RcodeSuccess
}
