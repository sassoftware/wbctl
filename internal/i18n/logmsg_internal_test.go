// Copyright © 2026, SAS Institute Inc., Cary, NC, USA.  All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package i18n

import (
	"sync"
	"testing"

	"github.com/sas-institute-rnd-saas/wbctl/internal/exiter"
	"github.com/stretchr/testify/require"
)

func TestInitialize(t *testing.T) {
	testCases := []struct {
		name             string
		messages         string
		culture          string
		expectedExitCode int
	}{
		{
			name: "success",
		},
		{
			name:             "properties failure",
			messages:         "a=${unbalanced",
			expectedExitCode: 1,
		},
		{
			name:             "parser failure",
			culture:          "unknown",
			expectedExitCode: 1,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			origMessages := wbctlMessages
			t.Cleanup(
				func() {
					wbctlMessages = origMessages
					once = sync.Once{}
					culture = "en"
				},
			)
			exitCode := 0
			exiter.OsExiter = func(code int) {
				exitCode = code
			}
			if tc.messages != "" {
				wbctlMessages = tc.messages
			}
			if tc.culture != "" {
				culture = tc.culture
			}
			initialize()
			require.Equal(t, tc.expectedExitCode, exitCode)
		})
	}
}

func TestFormatError(t *testing.T) {
	testCases := []struct {
		name     string
		messages string
		key      string
		expected string
		args     []A
	}{
		{
			name:     "parse failure",
			messages: "a={unbalanced",
			key:      "a",
			expected: "ParseError: `UnbalancedBraces` at 11",
		},
		{
			name:     "format failure",
			messages: "a={x}",
			key:      "a",
			args:     []A{{"x": []int{1}}},
			expected: "toString: Unsupported type: []int",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			origMessages := wbctlMessages
			t.Cleanup(
				func() {
					wbctlMessages = origMessages
					once = sync.Once{}
				},
			)
			if tc.messages != "" {
				wbctlMessages = tc.messages
			}
			actual := Format(tc.key, tc.args...)
			require.Equal(t, tc.expected, actual)
		})
	}
}
