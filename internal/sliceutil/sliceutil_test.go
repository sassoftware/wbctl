// Copyright © 2026, SAS Institute Inc., Cary, NC, USA.  All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package sliceutil_test

import (
	"testing"

	"github.com/sas-institute-rnd-saas/wbctl/internal/sliceutil"
	"github.com/stretchr/testify/require"
)

const (
	expectedFirstNonZeroValue = "first-non-zero"
	expectedMultipleElements  = "multiple elements found"
	expectedNoElements        = "no elements found"
)

func TestCoalesce(t *testing.T) {
	tests := []struct {
		name     string
		values   []string
		expected string
	}{
		{
			name:     "returns zero value when all empty",
			values:   []string{"", ""},
			expected: "",
		},
		{
			name:     "returns first non-zero value",
			values:   []string{"", expectedFirstNonZeroValue, "ignored"},
			expected: expectedFirstNonZeroValue,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := sliceutil.Coalesce(tt.values...)
			require.Equal(t, tt.expected, actual)
		})
	}
}

func TestGetAndEnsureOnlyOne(t *testing.T) {
	tests := []struct {
		name        string
		values      []string
		expected    string
		expectedErr string
	}{
		{
			name:        "returns error when no elements",
			values:      nil,
			expectedErr: expectedNoElements,
		},
		{
			name:        "returns error when multiple elements",
			values:      []string{"first", "second"},
			expectedErr: expectedMultipleElements,
		},
		{
			name:     "returns only element",
			values:   []string{expectedFirstNonZeroValue},
			expected: expectedFirstNonZeroValue,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual, err := sliceutil.GetAndEnsureOnlyOne(tt.values)

			if tt.expectedErr != "" {
				require.EqualError(t, err, tt.expectedErr)
				return
			}

			require.NoError(t, err)
			require.Equal(t, tt.expected, actual)
		})
	}
}
