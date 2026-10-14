// Copyright © 2026, SAS Institute Inc., Cary, NC, USA.  All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package timeutil_test

import (
	"testing"
	"time"

	"github.com/sassoftware/wbctl/internal/timeutil"
	"github.com/stretchr/testify/require"
)

const (
	expectedDurationFiveSeconds = "5s"
	expectedEmptyDuration       = ""
)

func TestSetNowAndSince(t *testing.T) {
	fixedNow := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)
	zeroTime := time.Time{}
	fiveSecondsAgo := fixedNow.Add(-5 * time.Second)

	tests := []struct {
		name     string
		at       *time.Time
		expected string
	}{
		{
			name:     "returns empty string for nil input",
			at:       nil,
			expected: expectedEmptyDuration,
		},
		{
			name:     "returns empty string for zero time",
			at:       &zeroTime,
			expected: expectedEmptyDuration,
		},
		{
			name:     "returns human duration",
			at:       &fiveSecondsAgo,
			expected: expectedDurationFiveSeconds,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			timeutil.SetNow(func() time.Time { return fixedNow })
			defer timeutil.SetNow(time.Now)

			actual := timeutil.Since(tt.at)
			require.Equal(t, tt.expected, actual)
		})
	}
}
