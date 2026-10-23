// Copyright © 2026, SAS Institute Inc., Cary, NC, USA.  All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package logutil_test

import (
	"log/slog"
	"testing"

	"github.com/sassoftware/wbctl/internal/logutil"
	"github.com/stretchr/testify/require"
)

func TestUpdateLevel(t *testing.T) {
	testCases := []struct {
		name     string
		initial  string
		input    string
		expected slog.Level
	}{
		{
			name:     "increment first",
			input:    "true",
			expected: 0,
		},
		{
			name:     "increment second",
			input:    "true",
			expected: -4,
		},
		{
			name:     "increment third",
			input:    "true",
			expected: -8,
		},
		{
			name:     "increment from error",
			initial:  "error",
			input:    "true",
			expected: 0,
		},
		{
			name:     "increment from warn",
			initial:  "warn",
			input:    "true",
			expected: 0,
		},
		{
			name:     "reset",
			input:    "false",
			expected: 12,
		},
		{
			name:     "trace",
			input:    "trace",
			expected: -8,
		},
		{
			name:     "debug",
			input:    "debug",
			expected: -4,
		},
		{
			name:     "info",
			input:    "info",
			expected: 0,
		},
		{
			name:     "warn",
			input:    "warn",
			expected: 4,
		},
		{
			name:     "error",
			input:    "error",
			expected: 8,
		},
		{
			name:     "silent",
			input:    "silent",
			expected: 12,
		},
		{
			name:     "unknown",
			input:    "unknown",
			expected: 12,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.initial != "" {
				err := logutil.UpdateLevel(tc.initial)
				require.NoError(t, err)
			}
			err := logutil.UpdateLevel(tc.input)
			require.NoError(t, err)
			actual := logutil.GetLevel()
			require.Equal(t, tc.expected, actual)
		})
	}
}

func TestSetLevel(t *testing.T) {
	testCases := []struct {
		name     string
		initial  string
		input    any
		expected slog.Level
	}{
		{
			name:     "Trace level",
			input:    logutil.LevelTrace,
			expected: logutil.LevelTrace,
		},
		{
			name:     "Trace int",
			input:    -8,
			expected: logutil.LevelTrace,
		},
		{
			name:     "Trace string",
			input:    "Trace",
			expected: logutil.LevelTrace,
		},
		{
			name:     "Debug level",
			input:    logutil.LevelDebug,
			expected: logutil.LevelDebug,
		},
		{
			name:     "Debug int",
			input:    -4,
			expected: logutil.LevelDebug,
		},
		{
			name:     "Debug string",
			input:    "Debug",
			expected: logutil.LevelDebug,
		},
		{
			name:     "Info level",
			input:    logutil.LevelInfo,
			expected: logutil.LevelInfo,
		},
		{
			name:     "Info int",
			input:    0,
			expected: logutil.LevelInfo,
		},
		{
			name:     "Info string",
			input:    "Info",
			expected: logutil.LevelInfo,
		},
		{
			name:     "Warn level",
			input:    logutil.LevelWarn,
			expected: logutil.LevelWarn,
		},
		{
			name:     "Warn int",
			input:    4,
			expected: logutil.LevelWarn,
		},
		{
			name:     "Warn string",
			input:    "Warn",
			expected: logutil.LevelWarn,
		},
		{
			name:     "Error level",
			input:    logutil.LevelError,
			expected: logutil.LevelError,
		},
		{
			name:     "Error int",
			input:    8,
			expected: logutil.LevelError,
		},
		{
			name:     "Error string",
			input:    "Error",
			expected: logutil.LevelError,
		},
		{
			name:     "Silent level",
			input:    logutil.LevelSilent,
			expected: logutil.LevelSilent,
		},
		{
			name:     "Silent int",
			input:    12,
			expected: logutil.LevelSilent,
		},
		{
			name:     "Silent string",
			input:    "Silent",
			expected: logutil.LevelSilent,
		},
		{
			name:     "Unknown type",
			input:    3.14,
			expected: logutil.LevelSilent,
		},
		{
			name:     "Unknown int",
			input:    9999,
			expected: 9999,
		},
		{
			name:     "Unknown string",
			input:    "Unknown",
			expected: logutil.LevelSilent,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			logutil.SetLevel(tc.input)
			actual := logutil.GetLevel()
			require.Equal(t, tc.expected, actual)
		})
	}
}

func TestGetLogger(t *testing.T) {
	logger := logutil.GetLogger()
	require.NotNil(t, logger)
}
