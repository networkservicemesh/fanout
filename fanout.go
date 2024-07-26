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
	"sync"
	"time"

	"github.com/coredns/coredns/plugin"
	"github.com/coredns/coredns/plugin/debug"
	"github.com/coredns/coredns/plugin/dnstap"
	"github.com/coredns/coredns/plugin/metadata"
	clog "github.com/coredns/coredns/plugin/pkg/log"
	"github.com/coredns/coredns/request"
	"github.com/miekg/dns"
	"github.com/networkservicemesh/fanout/internal/selector"
	"github.com/pkg/errors"
)

var log = clog.NewWithPlugin("fanout")

// Fanout represents a plugin instance that can do async requests to list of DNS servers.
type Fanout struct {
	clients        []Client
	tlsConfig      *tls.Config
	excludeDomains Domain
	tlsServerName  string
	timeout        time.Duration
	race           bool
	net            string
	from           string
	attempts       int
	workerCount    int
	serverCount    int
	loadFactor     []int
	tapPlugin      *dnstap.Dnstap
	Next           plugin.Handler
}

// New returns reference to new Fanout plugin instance with default configs.
func New() *Fanout {
	return &Fanout{
		tlsConfig:      new(tls.Config),
		net:            "udp",
		attempts:       3,
		timeout:        defaultTimeout,
		excludeDomains: NewDomain(),
	}
}

func (f *Fanout) addClient(p Client) {
	f.clients = append(f.clients, p)
	f.loadFactor = append(f.loadFactor, maxLoadFactor)
	f.workerCount++
	f.serverCount++
}

// Name implements plugin.Handler.
func (f *Fanout) Name() string {
	return "fanout"
}

// ServeDNS implements plugin.Handler.
func (f *Fanout) ServeDNS(ctx context.Context, w dns.ResponseWriter, m *dns.Msg) (int, error) {
	req := request.Request{W: w, Req: m}
	if !f.match(&req) {
		return plugin.NextOrFailure(f.Name(), f.Next, ctx, w, m)
	}
	timeoutContext, cancel := context.WithTimeout(ctx, f.timeout)
	defer cancel()
	result := f.getFanoutResult(timeoutContext, f.runWorkers(timeoutContext, &req))
	if result == nil {
		return dns.RcodeServerFailure, timeoutContext.Err()
	}
	metadata.SetValueFunc(ctx, "fanout/upstream", func() string {
		return result.client.Endpoint()
	})
	if result.err != nil {
		return dns.RcodeServerFailure, result.err
	}
	if f.tapPlugin != nil {
		toDnstap(f, result.client.Endpoint(), &req, result.response, result.start)
	}
	if !req.Match(result.response) {
		debug.Hexdumpf(result.response, "Wrong reply for id: %d, %s %d", result.response.Id, req.QName(), req.QType())
		formerr := new(dns.Msg)
		formerr.SetRcode(req.Req, dns.RcodeFormatError)
		logErrIfNotNil(w.WriteMsg(formerr))
		return 0, nil
	}
	logErrIfNotNil(w.WriteMsg(result.response))
	return 0, nil
}

type clientSelector interface {
	Pick() Client
}

func (f *Fanout) runWorkers(ctx context.Context, req *request.Request) chan *response {
	var sel clientSelector
	if f.serverCount == len(f.clients) {
		sel = selector.NewSimpleSelector(f.clients)
	} else {
		sel = selector.NewWeightedRandSelector(f.clients, f.loadFactor)
	}

	workerChannel := make(chan Client, f.workerCount)
	responseCh := make(chan *response, f.serverCount)
	go func() {
		defer close(workerChannel)
		for i := 0; i < f.serverCount; i++ {
			select {
			case <-ctx.Done():
				return
			case workerChannel <- sel.Pick():
			}
		}
	}()

	go func() {
		var wg sync.WaitGroup
		wg.Add(f.workerCount)

		for i := 0; i < f.workerCount; i++ {
			go func() {
				defer wg.Done()
				for c := range workerChannel {
					select {
					case <-ctx.Done():
						return
					case responseCh <- f.processClient(ctx, c, &request.Request{W: req.W, Req: req.Req}):
					}
				}
			}()
		}

		wg.Wait()
		close(responseCh)
	}()

	return responseCh
}

func (f *Fanout) getFanoutResult(ctx context.Context, responseCh <-chan *response) *response {
	var result *response
	for {
		select {
		case <-ctx.Done():
			return result
		case r, ok := <-responseCh:
			if !ok {
				return result
			}
			if isBetter(result, r) {
				result = r
			}
			if r.err != nil {
				break
			}
			if f.race {
				return r
			}
			if r.response.Rcode != dns.RcodeSuccess {
				break
			}
			return r
		}
	}
}

func (f *Fanout) match(state *request.Request) bool {
	if !plugin.Name(f.from).Matches(state.Name()) || f.excludeDomains.Contains(state.Name()) {
		return false
	}
	return true
}

func (f *Fanout) processClient(ctx context.Context, c Client, r *request.Request) *response {
	start := time.Now()
	var err error
	for j := 0; j < f.attempts || f.attempts == 0; <-time.After(attemptDelay) {
		if ctx.Err() != nil {
			return &response{client: c, response: nil, start: start, err: ctx.Err()}
		}
		var msg *dns.Msg
		msg, err = c.Request(ctx, r)
		if err == nil {
			return &response{client: c, response: msg, start: start, err: err}
		}
		if f.attempts != 0 {
			j++
		}
	}
	return &response{client: c, response: nil, start: start, err: errors.Wrapf(err, "attempt limit has been reached")}
}
