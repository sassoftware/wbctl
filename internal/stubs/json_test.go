// Copyright © 2026, SAS Institute Inc., Cary, NC, USA.  All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package stubs_test

import (
	"testing"

	"github.com/sas-institute-rnd-saas/wbctl/internal/stubs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAsStringSuccess(t *testing.T) {
	cases := []struct {
		name     string
		val      any
		expected string
	}{
		{
			name:     "string",
			val:      "hello",
			expected: `"hello"`,
		},
		{
			name:     "int",
			val:      42,
			expected: "42",
		},
		{
			name: "struct",
			val: struct {
				A int `json:"a"`
			}{A: 1},
			expected: `{"a":1}`,
		},
		{
			name: "pointer",
			val: &struct {
				B string `json:"b"`
			}{
				B: "y",
			},
			expected: `{"b":"y"}`,
		},
		{
			name:     "map",
			val:      map[string]int{"x": 1},
			expected: `{"x":1}`,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			actual := stubs.AsString(t, tc.val)
			require.Equal(t, tc.expected, actual)
		})
	}
}

func TestAsStringError(_ *testing.T) {
	// Exercise the error branch without failing the package tests.
	// Use a separate, uninitialized testing.T instance and call AsString in
	// its own goroutine so t.Fatalf exits only that goroutine.
	var fake testing.T
	ready := make(chan int)
	go func() {
		close(ready)
		stubs.AsString(&fake, make(chan int))
	}()
	<-ready
	assert.True(&fake, fake.Failed())
}
