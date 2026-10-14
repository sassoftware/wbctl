// Copyright © 2026, SAS Institute Inc., Cary, NC, USA.  All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package maputil_test

import (
	"fmt"
	"testing"

	"github.com/sassoftware/wbctl/internal/maputil"
	"github.com/stretchr/testify/require"
)

func TestFromEqualDelimitedLines(t *testing.T) {
	tests := []struct {
		name        string
		input       []string
		expected    map[string]string
		expectPanic bool
	}{
		{
			name:  "parses multiple key value pairs",
			input: []string{"a=1", "b=2"},
			expected: map[string]string{
				"a": "1",
				"b": "2",
			},
		},
		{
			name:  "keeps equal signs in value",
			input: []string{"a=1=2"},
			expected: map[string]string{
				"a": "1=2",
			},
		},
		{
			name:  "does not panic for line without delimiter",
			input: []string{"only-key"},
			expected: map[string]string{
				"only-key": "",
			},
		},
		{
			name:     "returns nil for no lines",
			input:    []string{},
			expected: nil,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			if testCase.expectPanic {
				require.Panics(t, func() {
					maputil.FromEqualDelimitedLines(testCase.input)
				})
				return
			}

			actual := maputil.FromEqualDelimitedLines(testCase.input)
			require.Equal(t, testCase.expected, actual)
		})
	}
}

func TestMapNoError(t *testing.T) {
	tests := []struct {
		name     string
		input    []int
		expected []string
	}{
		{
			name:     "maps all values",
			input:    []int{1, 2, 3},
			expected: []string{"v-1", "v-2", "v-3"},
		},
		{
			name:     "handles empty source",
			input:    []int{},
			expected: nil,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			actual := maputil.MapNoError(func(value int) string {
				return fmt.Sprintf("v-%d", value)
			}, testCase.input)
			require.Equal(t, testCase.expected, actual)
		})
	}
}

func TestSafeValueAsString(t *testing.T) {
	tests := []struct {
		name     string
		input    map[string]interface{}
		key      string
		expected string
	}{
		{
			name: "returns string value directly",
			input: map[string]interface{}{
				"name": "alice",
			},
			key:      "name",
			expected: "alice",
		},
		{
			name: "returns marshaled json for non-string value",
			input: map[string]interface{}{
				"meta": map[string]interface{}{
					"enabled": true,
				},
			},
			key:      "meta",
			expected: `{"enabled":true}`,
		},
		{
			name: "returns empty string when key is missing",
			input: map[string]interface{}{
				"other": "value",
			},
			key:      "missing",
			expected: "",
		},
		{
			name: "returns empty string when value cannot be marshaled",
			input: map[string]interface{}{
				"bad": func() {},
			},
			key:      "bad",
			expected: "",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			actual := maputil.SafeValueAsString(testCase.input, testCase.key)
			require.Equal(t, testCase.expected, actual)
		})
	}
}
