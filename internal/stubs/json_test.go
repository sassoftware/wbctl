// Copyright © 2026, SAS Institute Inc., Cary, NC, USA.  All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package stubs_test

import (
	"testing"

	"github.com/sassoftware/wbctl/internal/stubs"
	"github.com/stretchr/testify/assert"
)

func TestAsStringSuccess(t *testing.T) {
	cases := []struct {
		name     string
		val      any
		expected string
		err      bool
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
		{
			name:     "channel",
			val:      make(chan int),
			expected: `FAILED TO MARSHAL: `,
			err:      true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			actual := stubs.AsString(tc.val)
			if tc.err {
				assert.Contains(t, actual, tc.expected)
			} else {
				assert.Equal(t, tc.expected, actual)
			}
		})
	}
}
