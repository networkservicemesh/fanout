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
	"crypto/rand"
	"math/big"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestDomainBasic(t *testing.T) {
	samples := []struct {
		child    string
		parent   string
		expected bool
	}{
		{".", ".", true},
		{"example.org.", ".", true},
		{"example.org.", "example.org.", true},
		{"example.org", "example.org", true},
		{"example.org.", "org.", true},
		{"org.", "example.org.", false},
	}

	for i, s := range samples {
		l := NewDomain()
		l.AddString(s.parent)
		require.Equal(t, s.expected, l.Contains(s.child), i)
	}
}

func TestDomainGet(t *testing.T) {
	d := NewDomain()
	d.AddString("google.com.")
	d.AddString("example.com.")
	require.True(t, d.Get(".").Get("com").Get("google").IsFinal())
}

func TestDomain_ContainsShouldWorkFast(t *testing.T) {
	var samples []string
	d := NewDomain()
	for i := 0; i < 100; i++ {
		for j := 0; j < 100; j++ {
			samples = append(samples, genSample(i+1))
			d.AddString(samples[len(samples)-1])
		}
	}
	start := time.Now()
	for i := 0; i < 10000; i++ {
		require.True(t, d.Contains(samples[i]))
	}
	require.True(t, time.Since(start) < time.Second/4)
}

func TestDomainFewEntries(t *testing.T) {
	d := NewDomain()
	d.AddString("google.com.")
	d.AddString("example.com.")
	require.True(t, d.Contains("google.com."))
	require.True(t, d.Contains("example.com."))
	require.False(t, d.Contains("com."))
}

func TestDomain_DoNotStoreExtraEntries(t *testing.T) {
	d := NewDomain()
	d.AddString("example.com.")
	d.AddString("advanced.example.com.")
	require.Nil(t, d.Get(".").Get("com").Get("example").Get("advanced"))
}

func BenchmarkDomain_ContainsMatch(b *testing.B) {
	d := NewDomain()
	var samples []string
	for i := 0; i < 100; i++ {
		for j := 0; j < 100; j++ {
			samples = append(samples, genSample(i+1))
			d.AddString(samples[len(samples)-1])
		}
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for j := 0; j < 10000; j++ {
			d.Contains(samples[j])
		}
	}
}

func BenchmarkDomain_AddString(b *testing.B) {
	d := NewDomain()
	var samples []string
	for i := 0; i < 100; i++ {
		for j := 0; j < 100; j++ {
			samples = append(samples, genSample(i+1))
		}
	}
	for i := 0; i < b.N; i++ {
		for j := 0; j < len(samples); j++ {
			d.AddString(samples[j])
		}
	}
}

func BenchmarkDomain_ContainsAny(b *testing.B) {
	d := NewDomain()
	var samples []string
	for i := 0; i < 100; i++ {
		for j := 0; j < 100; j++ {
			d.AddString(genSample(i + 1))
			samples = append(samples, genSample(i+1))
		}
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for j := 0; j < len(samples); j++ {
			d.Contains(samples[j])
		}
	}
}

func genSample(n int) string {
	randInt := func() int {
		r, err := rand.Int(rand.Reader, big.NewInt(100))
		if err != nil {
			panic(err.Error())
		}
		return int(r.Int64())
	}

	var sb strings.Builder
	for segment := 0; segment < n; segment++ {
		l := randInt()%9 + 1
		for i := 0; i < l; i++ {
			v := (randInt() % 26) + 97
			_, _ = sb.WriteRune(rune(v))
		}
		_, _ = sb.WriteRune('.')
	}
	return sb.String()
}
