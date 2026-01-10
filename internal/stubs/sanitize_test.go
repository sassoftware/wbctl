// Copyright © 2026, SAS Institute Inc., Cary, NC, USA.  All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package stubs_test

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sassoftware/wbctl/internal/stubs"
	"github.com/stretchr/testify/assert"
)

func TestReplace(t *testing.T) {
	tests := []struct {
		name string
		src  string
		reps map[string]string
		want string
	}{
		{
			name: "basic",
			src:  "hello {X}",
			reps: map[string]string{"{X}": "world"},
			want: "hello world",
		},
		{
			name: "noop",
			src:  "nochange",
			want: "nochange",
		},
		{
			name: "multi",
			src:  "{A}-{B}",
			reps: map[string]string{
				"{A}": "1",
				"{B}": "2",
			},
			want: "1-2",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := stubs.Replace(tc.src, tc.reps)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestSanitize(t *testing.T) {
	tests := []struct {
		name     string
		expected string
		actual   string
		reps     map[string]string
		wantExp  string
		wantAct  string
	}{
		{
			name:     "single",
			expected: "id-{SANITIZE:\\d+:<ID>}",
			actual:   "id-123",
			wantExp:  "id-<ID>",
			wantAct:  "id-<ID>",
		},
		{
			name:     "with_replace",
			expected: "foo-{SANITIZE:\\d+:X}-{ENV}",
			actual:   "foo-1-ignored",
			reps:     map[string]string{"{ENV}": "prod"},
			wantExp:  "foo-X-prod",
			wantAct:  "foo-X-ignored",
		},
		{
			name:     "multiple",
			expected: "a-{SANITIZE:\\d+:N}-{SANITIZE:[a-z][a-z]+:S}",
			actual:   "a-7-abc",
			wantExp:  "a-N-S",
			wantAct:  "a-N-S",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			gotExp, gotAct := stubs.Sanitize(tc.expected, tc.actual, tc.reps)
			assert.Equal(t, tc.wantExp, gotExp)
			assert.Equal(t, tc.wantAct, gotAct)
		})
	}
}

func TestSanitizeSlice(t *testing.T) {
	tests := []struct {
		name     string
		expected []string
		actual   []string
		reps     map[string]string
		wantExp  []string
		wantAct  []string
	}{
		{
			name:     "basic",
			expected: []string{"u-{SANITIZE:\\d+:X}", "v-{SANITIZE:foo:bar}"},
			actual:   []string{"u-1", "v-foo"},
			wantExp:  []string{"u-X", "v-bar"},
			wantAct:  []string{"u-X", "v-bar"},
		},
		{
			name:     "unchanged when slice lengths differ",
			expected: []string{"u-{SANITIZE:\\d+:X}", "v-{SANITIZE:foo:bar}"},
			actual:   []string{"u-1", "v-foo", "unexpected extra element"},
			wantExp:  []string{"u-{SANITIZE:\\d+:X}", "v-{SANITIZE:foo:bar}"},
			wantAct:  []string{"u-1", "v-foo", "unexpected extra element"},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			gotExp, gotAct := stubs.SanitizeSlice(tc.expected, tc.actual, tc.reps)
			assert.Equal(t, tc.wantExp, gotExp)
			assert.Equal(t, tc.wantAct, gotAct)
		})
	}
}

func TestPortHost(t *testing.T) {
	tests := []struct {
		name     string
		url      string
		wantPort string
		wantHost string
	}{
		{
			name:     "with port",
			url:      "http://example.com:8080/path",
			wantPort: "8080",
			wantHost: "example.com",
		},
		{
			name:     "no port",
			url:      "https://example.com/path",
			wantHost: "example.com",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			p := stubs.Port(t, tc.url)
			h := stubs.Host(t, tc.url)
			assert.Equal(t, tc.wantPort, p)
			assert.Equal(t, tc.wantHost, h)
		})
	}
}

func TestExampleSanitizeUsage(t *testing.T) {
	expected := "/authorization?" +
		"{SANITIZE:&code_challenge=.*?&:&code_challenge=SANITIZED&}" +
		"ingressHost=${HOST}%3A${PORT}" +
		"{SANITIZE:&state=.*:&state=SANITIZED}"
	actual := ""
	replacements := map[string]string{}
	srv := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		_ = r.Body.Close()
		expected, actual = stubs.Sanitize(expected, r.URL.String(), replacements)
		assert.Equal(t, expected, actual)
	}))
	replacements = map[string]string{
		"${PORT}": stubs.Port(t, srv.URL),
		"${HOST}": stubs.Host(t, srv.URL),
	}
	b := make([]byte, 32)
	// Read some random values that would be impossible to benchmark. Read is
	// documented to never return an error.
	_, _ = rand.Read(b)
	code := base64.RawURLEncoding.EncodeToString(b)
	_, _ = rand.Read(b)
	state := base64.RawURLEncoding.EncodeToString(b)
	req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, srv.URL+
		"/authorization?&code_challenge="+code+"&ingressHost="+
		stubs.Host(t, srv.URL)+"%3A"+stubs.Port(t, srv.URL)+
		"&state="+state, nil)
	res, _ := http.DefaultClient.Do(req)
	_ = res.Body.Close()
}
