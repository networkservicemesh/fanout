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
	"io/ioutil"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/coredns/caddy"
)

func TestSetup(t *testing.T) {
	tests := []struct {
		input            string
		expectedFrom     string
		expectedTo       []string
		expectedIgnored  []string
		expectedWorkers  int
		expectedAttempts int
		expectedTimeout  time.Duration
		expectedNetwork  string
		expectedErr      string
	}{
		// positive
		{input: "fanout . 127.0.0.1", expectedFrom: ".", expectedAttempts: 3, expectedWorkers: 1, expectedTimeout: defaultTimeout, expectedNetwork: "udp"},
		{input: "fanout . 127.0.0.1 {\nexcept a b\nworker-count 3\n}", expectedFrom: ".", expectedTimeout: defaultTimeout, expectedAttempts: 3, expectedWorkers: 1, expectedIgnored: []string{"a.", "b."}, expectedNetwork: "udp"},
		{input: "fanout . 127.0.0.1 127.0.0.2 {\nnetwork tcp\n}", expectedFrom: ".", expectedTimeout: defaultTimeout, expectedAttempts: 3, expectedWorkers: 2, expectedNetwork: "tcp", expectedTo: []string{"127.0.0.1:53", "127.0.0.2:53"}},
		{input: "fanout . 127.0.0.1 127.0.0.2 127.0.0.3 127.0.0.4 {\nworker-count 3\ntimeout 1m\n}", expectedTimeout: time.Minute, expectedAttempts: 3, expectedFrom: ".", expectedWorkers: 3, expectedNetwork: "udp"},
		{input: "fanout . 127.0.0.1 127.0.0.2 127.0.0.3 127.0.0.4 {\nattempt-count 2\n}", expectedTimeout: defaultTimeout, expectedFrom: ".", expectedAttempts: 2, expectedWorkers: 4, expectedNetwork: "udp"},

		// negative
		{input: "fanout . aaa", expectedErr: "not an IP address or file"},
		{input: "fanout . 127.0.0.1 {\nexcept a b\nworker-count 1\n}", expectedErr: "use Forward plugin"},
		{input: "fanout . 127.0.0.1 {\nexcept a b\nworker-count ten\n}", expectedErr: "'ten'"},
		{input: "fanout . 127.0.0.1 127.0.0.2 {\nnetwork XXX\n}", expectedErr: "unknown network protocol"},
	}

	for i, test := range tests {
		c := caddy.NewTestController("dns", test.input)
		f, err := parseFanout(c)
		if test.expectedErr != "" && err == nil {
			t.Fatalf("Test %d: expected error but not found errors", i)
		}

		if err != nil {
			if !strings.Contains(err.Error(), test.expectedErr) {
				t.Fatalf("Test %d: expected error to contain: %v, found error: %v, input: %s", i, test.expectedErr, err, test.input)
			}
			continue
		}
		if f.timeout != test.expectedTimeout {
			t.Fatalf("Test %d: expected: %d, got: %d", i, test.expectedTimeout, f.timeout)
		}
		if f.attempts != test.expectedAttempts {
			t.Fatalf("Test %d: expected: %d, got: %d", i, test.expectedAttempts, f.attempts)
		}
		if f.from != test.expectedFrom && test.expectedFrom != "" {
			t.Fatalf("Test %d: expected: %s, got: %s", i, test.expectedFrom, f.from)
		}
		if test.expectedIgnored != nil {
			for _, expected := range test.expectedIgnored {
				if !f.excludeDomains.Contains(expected) {
					t.Fatalf("Test %d: missed exclude domain name: %v", i, test.expectedIgnored)
				}
			}
		}
		if test.expectedTo != nil {
			var to []string
			for j := 0; j < len(f.clients); j++ {
				to = append(to, f.clients[j].Endpoint())
			}
			if !reflect.DeepEqual(to, test.expectedTo) {
				t.Fatalf("Test %d: expected: %q, actual: %q", i, test.expectedTo, to)
			}
		}
		if f.workerCount != test.expectedWorkers {
			t.Fatalf("Test %d: expected: %d, got: %d", i, test.expectedWorkers, f.workerCount)
		}
		if f.net != test.expectedNetwork {
			t.Fatalf("Test %d: expected: %v, got: %v", i, test.expectedNetwork, f.net)
		}
	}
}

func TestSetupResolvconf(t *testing.T) {
	const resolv = "resolv.conf"
	if err := ioutil.WriteFile(resolv,
		[]byte(`nameserver 10.10.255.252
nameserver 10.10.255.253`), 0666); err != nil {
		t.Fatalf("Failed to write resolv.conf file: %s", err)
	}
	defer func() {
		logErrIfNotNil(os.Remove(resolv))
	}()

	tests := []struct {
		input         string
		shouldErr     bool
		expectedErr   string
		expectedNames []string
	}{
		{`fanout . ` + resolv, false, "", []string{"10.10.255.252:53", "10.10.255.253:53"}},
	}

	for i, test := range tests {
		c := caddy.NewTestController("dns", test.input)
		f, err := parseFanout(c)

		if test.shouldErr && err == nil {
			t.Errorf("Test %d: expected error but found %s for input %s", i, err, test.input)
			continue
		}

		if err != nil {
			if !test.shouldErr {
				t.Errorf("Test %d: expected no error but found one for input %s, got: %v", i, test.input, err)
			}

			if !strings.Contains(err.Error(), test.expectedErr) {
				t.Errorf("Test %d: expected error to contain: %v, found error: %v, input: %s", i, test.expectedErr, err, test.input)
			}
		}

		if !test.shouldErr {
			for j, n := range test.expectedNames {
				addr := f.clients[j].Endpoint()
				if n != addr {
					t.Errorf("Test %d, expected %q, got %q", j, n, addr)
				}
			}
		}
	}
}
