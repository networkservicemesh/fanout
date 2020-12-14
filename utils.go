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
	"net"
	"strconv"
	"time"

	"github.com/coredns/coredns/plugin/dnstap/msg"
	"github.com/coredns/coredns/request"

	tap "github.com/dnstap/golang-dnstap"
	"github.com/miekg/dns"
)

func logErrIfNotNil(err error) {
	if err == nil {
		return
	}
	log.Error(err)
}

func toDnstap(f *Fanout, host string, state *request.Request, reply *dns.Msg, start time.Time) {
	// Query
	q := new(tap.Message)
	msg.SetQueryTime(q, start)
	h, p, _ := net.SplitHostPort(host)      // this is preparsed and can't err here
	port, _ := strconv.ParseUint(p, 10, 32) // same here
	ip := net.ParseIP(h)

	var ta net.Addr = &net.UDPAddr{IP: ip, Port: int(port)}
	t := f.net

	if t == "tcp" {
		ta = &net.TCPAddr{IP: ip, Port: int(port)}
	}

	var _ = msg.SetQueryAddress(q, ta)

	if f.tapPlugin.IncludeRawMessage {
		buf, _ := state.Req.Pack()
		q.QueryMessage = buf
	}
	msg.SetType(q, tap.Message_FORWARDER_QUERY)
	f.tapPlugin.TapMessage(q)

	// Response
	if reply != nil {
		r := new(tap.Message)

		if f.tapPlugin.IncludeRawMessage {
			buf, _ := reply.Pack()
			r.ResponseMessage = buf
		}
		msg.SetQueryTime(r, start)
		var _ = msg.SetQueryAddress(r, ta)
		msg.SetResponseTime(r, time.Now())
		msg.SetType(r, tap.Message_FORWARDER_RESPONSE)
		f.tapPlugin.TapMessage(r)
	}
}
