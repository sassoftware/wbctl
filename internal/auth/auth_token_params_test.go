// Copyright © 2026, SAS Institute Inc., Cary, NC, USA.  All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package auth_test

import (
	"testing"

	"github.com/sas-institute-rnd-saas/wbctl/internal/auth"
	"github.com/sas-institute-rnd-saas/wbctl/internal/constants"
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
			expected: constants.AuthModePKCE,
		},
		{
			name:     "unknown",
			input:    auth.GetAccessTokenParams{AuthMode: "unknown"},
			expected: "unknown",
		},
		{
			name: "auto with client secret",
			input: auth.GetAccessTokenParams{
				AuthMode:     constants.AuthModeAuto,
				ClientSecret: "client-secret",
			},
			expected: constants.AuthModeClientCredentials,
		},
		{
			name: "auto with username",
			input: auth.GetAccessTokenParams{
				AuthMode: constants.AuthModeAuto,
				Username: "username",
			},
			expected: constants.AuthModePassword,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual := tc.input.NormalizedAuthMode()
			require.Equal(t, tc.expected, actual)
		})
	}
}
