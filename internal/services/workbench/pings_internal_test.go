// Copyright © 2026, SAS Institute Inc., Cary, NC, USA.  All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package workbench

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestSafeTimeDiff(t *testing.T) {
	now := time.Date(2026, time.May, 8, 0, 0, 0, 0, time.UTC)
	later := now.Add(2 * time.Second)

	tests := []struct {
		name     string
		start    *time.Time
		end      *time.Time
		expected string
	}{
		{name: "nil start", start: nil, end: &later, expected: "<unknown>"},
		{name: "nil end", start: &now, end: nil, expected: "<unknown>"},
		{name: "valid diff", start: &now, end: &later, expected: "2s"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := safeTimeDiff(tt.start, tt.end)
			require.Equal(t, tt.expected, actual)
		})
	}
}
