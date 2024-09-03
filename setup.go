// Copyright (c) 2020 Doc.ai and/or its affiliates.
//
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

package fanout

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/coredns/caddy"
	"github.com/coredns/caddy/caddyfile"
	"github.com/coredns/coredns/core/dnsserver"
	"github.com/coredns/coredns/plugin"
	"github.com/coredns/coredns/plugin/dnstap"
	"github.com/coredns/coredns/plugin/pkg/parse"
	"github.com/coredns/coredns/plugin/pkg/tls"
	"github.com/coredns/coredns/plugin/pkg/transport"
	"github.com/pkg/errors"
)

func init() {
	caddy.RegisterPlugin("fanout", caddy.Plugin{
		ServerType: "dns",
		Action:     setup,
	})
}

func setup(c *caddy.Controller) error {
	f, err := parseFanout(c)
	if err != nil {
		return plugin.Error("fanout", err)
	}
	l := len(f.clients)
	if len(f.clients) > maxIPCount {
		return plugin.Error("fanout", errors.Errorf("more than %d TOs configured: %d", maxIPCount, l))
	}

	dnsserver.GetConfig(c).AddPlugin(func(next plugin.Handler) plugin.Handler {
		f.Next = next
		return f
	})

	c.OnStartup(func() error {
		if taph := dnsserver.GetConfig(c).Handler("dnstap"); taph != nil {
			if tapPlugin, ok := taph.(*dnstap.Dnstap); ok {
				f.tapPlugin = tapPlugin
			}
		}
		return f.OnStartup()
	})
	c.OnShutdown(f.OnShutdown)

	return nil
}

// OnStartup starts a goroutines for all clients.
func (f *Fanout) OnStartup() (err error) {
	return nil
}

// OnShutdown stops all configured clients.
func (f *Fanout) OnShutdown() error {
	return nil
}

func parseFanout(c *caddy.Controller) (*Fanout, error) {
	var (
		f   *Fanout
		err error
		i   int
	)
	for c.Next() {
		if i > 0 {
			return nil, plugin.ErrOnce
		}
		i++
		f, err = parsefanoutStanza(&c.Dispenser)
		if err != nil {
			return nil, err
		}
	}

	return f, nil
}

func parsefanoutStanza(c *caddyfile.Dispenser) (*Fanout, error) {
	f := New()
	if !c.Args(&f.from) {
		return f, c.ArgErr()
	}

	normalized := plugin.Host(f.from).NormalizeExact()
	if len(normalized) == 0 {
		return nil, errors.Errorf("unable to normalize '%s'", f.from)
	}
	f.from = normalized[0]

	to := c.RemainingArgs()
	if len(to) == 0 {
		return f, c.ArgErr()
	}
	toHosts, err := parse.HostPortOrFile(to...)
	if err != nil {
		return f, err
	}
	for c.NextBlock() {
		err = parseValue(strings.ToLower(c.Val()), f, c)
		if err != nil {
			return nil, err
		}
	}
	initClients(f, toHosts)
	err = initServerSelectionPolicy(f)
	if err != nil {
		return nil, err
	}

	if f.workerCount > len(f.clients) || f.workerCount == 0 {
		f.workerCount = len(f.clients)
	}

	return f, nil
}

func initClients(f *Fanout, hosts []string) {
	transports := make([]string, len(hosts))
	for i, host := range hosts {
		trans, h := parse.Transport(host)
		f.clients = append(f.clients, NewClient(h, f.net))
		transports[i] = trans
	}

	f.tlsConfig.ServerName = f.tlsServerName
	for i := range f.clients {
		if transports[i] == transport.TLS {
			f.clients[i].SetTLSConfig(f.tlsConfig)
		}
	}
}

func initServerSelectionPolicy(f *Fanout) error {
	if f.serverCount > len(f.clients) || f.serverCount == 0 {
		f.serverCount = len(f.clients)
	}

	loadFactor := f.loadFactor
	if len(loadFactor) == 0 {
		for i := 0; i < len(f.clients); i++ {
			loadFactor = append(loadFactor, maxLoadFactor)
		}
	}
	if len(loadFactor) != len(f.clients) {
		return errors.New("load-factor params count must be the same as the number of hosts")
	}

	f.serverSelectionPolicy = &sequentialPolicy{}
	if f.policyType == policyWeightedRandom {
		f.serverSelectionPolicy = &weightedPolicy{loadFactor: loadFactor}
	}

	return nil
}

func parseValue(v string, f *Fanout, c *caddyfile.Dispenser) error {
	switch v {
	case "tls":
		return parseTLS(f, c)
	case "network":
		return parseProtocol(f, c)
	case "tls-server":
		return parseTLSServer(f, c)
	case "worker-count":
		return parseWorkerCount(f, c)
	case "policy":
		return parsePolicy(f, c)
	case "weighted-random-server-count":
		serverCount, err := parsePositiveInt(c)
		f.serverCount = serverCount
		return err
	case "weighted-random-load-factor":
		return parseLoadFactor(f, c)
	case "timeout":
		return parseTimeout(f, c)
	case "race":
		return parseRace(f, c)
	case "except":
		return parseIgnored(f, c)
	case "except-file":
		return parseIgnoredFromFile(f, c)
	case "attempt-count":
		num, err := parsePositiveInt(c)
		f.attempts = num
		return err
	default:
		return errors.Errorf("unknown property %v", v)
	}
}

func parsePolicy(f *Fanout, c *caddyfile.Dispenser) error {
	if !c.NextArg() {
		return c.ArgErr()
	}

	policyType := strings.ToLower(c.Val())
	if policyType != policyWeightedRandom && policyType != policySequential {
		return errors.Errorf("unknown policy %q", c.Val())
	}
	f.policyType = policyType

	return nil
}

func parseTimeout(f *Fanout, c *caddyfile.Dispenser) error {
	if !c.NextArg() {
		return c.ArgErr()
	}
	var err error
	val := c.Val()
	f.timeout, err = time.ParseDuration(val)
	return err
}

func parseRace(f *Fanout, c *caddyfile.Dispenser) error {
	if c.NextArg() {
		return c.ArgErr()
	}
	f.race = true
	return nil
}

func parseIgnoredFromFile(f *Fanout, c *caddyfile.Dispenser) error {
	args := c.RemainingArgs()
	if len(args) != 1 {
		return c.ArgErr()
	}
	b, err := os.ReadFile(filepath.Clean(args[0]))
	if err != nil {
		return err
	}
	names := strings.Split(string(b), "\n")
	for i := 0; i < len(names); i++ {
		normalized := plugin.Host(names[i]).NormalizeExact()
		if len(normalized) == 0 {
			return errors.Errorf("unable to normalize '%s'", names[i])
		}
		f.excludeDomains.AddString(normalized[0])
	}
	return nil
}

func parseIgnored(f *Fanout, c *caddyfile.Dispenser) error {
	ignore := c.RemainingArgs()
	if len(ignore) == 0 {
		return c.ArgErr()
	}
	for i := 0; i < len(ignore); i++ {
		normalized := plugin.Host(ignore[i]).NormalizeExact()
		if len(normalized) == 0 {
			return errors.Errorf("unable to normalize '%s'", ignore[i])
		}
		f.excludeDomains.AddString(normalized[0])
	}
	return nil
}

func parseWorkerCount(f *Fanout, c *caddyfile.Dispenser) error {
	var err error
	f.workerCount, err = parsePositiveInt(c)
	if err == nil {
		if f.workerCount < minWorkerCount {
			return errors.New("worker count should be more or equal 2. Consider to use Forward plugin")
		}
		if f.workerCount > maxWorkerCount {
			return errors.Errorf("worker count more then max value: %v", maxWorkerCount)
		}
	}
	return err
}

func parseLoadFactor(f *Fanout, c *caddyfile.Dispenser) error {
	args := c.RemainingArgs()
	if len(args) == 0 {
		return c.ArgErr()
	}

	for _, arg := range args {
		loadFactor, err := strconv.Atoi(arg)
		if err != nil {
			return c.ArgErr()
		}

		if loadFactor < minLoadFactor {
			return errors.New("load-factor should be more or equal 1")
		}
		if loadFactor > maxLoadFactor {
			return errors.Errorf("load-factor %d should be less than %d", loadFactor, maxLoadFactor)
		}

		f.loadFactor = append(f.loadFactor, loadFactor)
	}

	return nil
}

func parsePositiveInt(c *caddyfile.Dispenser) (int, error) {
	if !c.NextArg() {
		return -1, c.ArgErr()
	}
	v := c.Val()
	num, err := strconv.Atoi(v)
	if err != nil {
		return -1, c.ArgErr()
	}
	if num < 0 {
		return -1, c.ArgErr()
	}
	return num, nil
}

func parseTLSServer(f *Fanout, c *caddyfile.Dispenser) error {
	if !c.NextArg() {
		return c.ArgErr()
	}
	f.tlsServerName = c.Val()
	return nil
}

func parseProtocol(f *Fanout, c *caddyfile.Dispenser) error {
	if !c.NextArg() {
		return c.ArgErr()
	}
	net := strings.ToLower(c.Val())
	if net != tcp && net != udp && net != tcptls {
		return errors.New("unknown network protocol")
	}
	f.net = net
	return nil
}

func parseTLS(f *Fanout, c *caddyfile.Dispenser) error {
	args := c.RemainingArgs()
	if len(args) > 3 {
		return c.ArgErr()
	}

	tlsConfig, err := tls.NewTLSConfigFromArgs(args...)
	if err != nil {
		return err
	}
	f.tlsConfig = tlsConfig
	return nil
}
