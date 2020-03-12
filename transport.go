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
	"crypto/tls"

	"github.com/miekg/dns"
)

// Transport represent a solution to connect to remote DNS endpoint with specific network
type Transport interface {
	Dial(net string) (*dns.Conn, error)
	SetTLSConfig(*tls.Config)
}

// NewTransport creates new transport with address
func NewTransport(addr string) Transport {
	return &transportImpl{
		addr: addr,
	}
}

type transportImpl struct {
	tlsConfig *tls.Config
	addr      string
}

// SetTLSConfig sets tls config for transport
func (t *transportImpl) SetTLSConfig(c *tls.Config) {
	t.tlsConfig = c
}

// Dial dials the address configured in transportImpl, potentially reusing a connection or creating a new one.
func (t *transportImpl) Dial(net string) (*dns.Conn, error) {
	if t.tlsConfig != nil {
		net = tcptlc
	}
	if net == tcptlc {
		return dns.DialTimeoutWithTLS("tcp", t.addr, t.tlsConfig, defaultTimeout)
	}
	return dns.DialTimeout(net, t.addr, defaultTimeout)
}
