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

// +build gofuzz

package fanout

import (
	"github.com/coredns/coredns/plugin/pkg/dnstest"
	"github.com/coredns/coredns/plugin/pkg/fuzz"

	"github.com/miekg/dns"
)

var f *Fanout

// abuse init to setup an environment to test against. This start another server to that will
// reflect responses.
func init() {
	f = New()
	s := dnstest.NewServer(r{}.reflectHandler)
	f.clients = append(f.clients, NewClient(s.Addr, "tcp"))
	f.clients = append(f.clients, NewClient(s.Addr, "udp"))
}

// Fuzz fuzzes fanaot.
func Fuzz(data []byte) int {
	return fuzz.Do(f, data)
}

type r struct{}

func (r r) reflectHandler(w dns.ResponseWriter, req *dns.Msg) {
	m := new(dns.Msg)
	m.SetReply(req)
	w.WriteMsg(m)
}
