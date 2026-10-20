// Copyright © 2026, SAS Institute Inc., Cary, NC, USA.  All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package auth_test

import (
	"testing"

	"github.com/sassoftware/wbctl/internal/auth"
	"github.com/sassoftware/wbctl/internal/constants"
	"github.com/stretchr/testify/require"
)

func TestNormalizedAuthMode(t *testing.T) {
	testCases := []struct {
		name     string
		input    auth.GetAccessTokenParams
		expected string
	}{
		{
			name:     "empty",
			input:    auth.GetAccessTokenParams{},
			expected: constants.AuthModeDevice,
		},
		{
			name:     "unknown",
			input:    auth.GetAccessTokenParams{AuthMode: "unknown"},
			expected: "unknown",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual := tc.input.NormalizedAuthMode()
			require.Equal(t, tc.expected, actual)
		})
	}
}
