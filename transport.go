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
	"context"
	"crypto/tls"
	"net"

	"github.com/miekg/dns"
	ot "github.com/opentracing/opentracing-go"
)

// Transport represent a solution to connect to remote DNS endpoint with specific network
type Transport interface {
	Dial(ctx context.Context, net string) (*dns.Conn, error)
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
func (t *transportImpl) Dial(ctx context.Context, network string) (*dns.Conn, error) {
	if t.tlsConfig != nil {
		network = tcptlc
	}
	if network == tcptlc {
		return t.dial(ctx, &dns.Client{Net: network, Dialer: &net.Dialer{Timeout: maxTimeout}, TLSConfig: t.tlsConfig})
	}
	return t.dial(ctx, &dns.Client{Net: network, Dialer: &net.Dialer{Timeout: maxTimeout}})
}

func (t *transportImpl) dial(ctx context.Context, c *dns.Client) (*dns.Conn, error) {
	span := ot.SpanFromContext(ctx)
	if span != nil {
		childSpan := span.Tracer().StartSpan("connect", ot.ChildOf(span.Context()))
		ctx = ot.ContextWithSpan(ctx, childSpan)
		defer childSpan.Finish()
	}
	var d net.Dialer
	if c.Dialer == nil {
		d = net.Dialer{Timeout: maxTimeout}
	} else {
		d = *c.Dialer
	}
	network := c.Net
	if network == "" {
		network = "udp"
	}
	var conn = new(dns.Conn)
	var err error
	if network == tcptlc {
		conn.Conn, err = tls.DialWithDialer(&d, network, t.addr, c.TLSConfig)
	} else {
		conn.Conn, err = d.DialContext(ctx, network, t.addr)
	}
	if err != nil {
		return nil, err
	}
	return conn, nil
}
