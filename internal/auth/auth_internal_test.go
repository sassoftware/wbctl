// Copyright © 2026, SAS Institute Inc., Cary, NC, USA.  All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package auth

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

const (
	authInternalEmptyRefreshTokenErr = "provided refresh token is empty"
)

func TestIsDeviceGrantErrorType(t *testing.T) {
	tests := []struct {
		err      error
		name     string
		expected bool
		types    []string
	}{
		{
			name:     "nil error returns false",
			err:      nil,
			types:    []string{"slow_down"},
			expected: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.expected, isDeviceGrantErrorType(tt.err, tt.types))
		})
	}
}

func TestPerformRefreshForToken(t *testing.T) {
	tests := []struct {
		name         string
		refreshToken string
		expectedErr  string
	}{
		{
			name:         "empty refresh token returns error",
			refreshToken: "",
			expectedErr:  authInternalEmptyRefreshTokenErr,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := performRefreshForToken(
				context.Background(),
				"http://localhost:20000/",
				"https://example.test/token",
				tt.refreshToken,
				"http://ingress.example.com/",
			)
			require.EqualError(t, err, tt.expectedErr)
		})
	}
}

func TestGetWellKnownConfigAtPathInvalidContext(t *testing.T) {
	// This test is to ensure that the function returns an error when a nil
	// context is passed so ignore lint error
	// nolint: staticcheck
	_, err := getWellKnownConfigAtPath(nil, "https://example.com", "endpoint")
	require.ErrorContains(t, err, "net/http: nil Context")
}
