// Copyright © 2026, SAS Institute Inc., Cary, NC, USA.  All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package jsonutil_test

import (
	"errors"
	"os"
	"testing"

	"github.com/sas-institute-rnd-saas/wbctl/internal/jsonutil"
	"github.com/sas-institute-rnd-saas/wbctl/internal/stubs"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"
)

type jsonInput struct {
	Name string `json:"name"`
}

func TestGetFromFileOrStdin(t *testing.T) {
	tests := []struct {
		name        string
		fs          afero.Fs
		inputPath   string
		fileContent string
		stdinJSON   string
		expected    jsonInput
		expectedErr string
	}{
		{
			name:        "reads from file path",
			fs:          afero.NewMemMapFs(),
			inputPath:   "input.json",
			fileContent: `{"name":"from-file"}`,
			expected:    jsonInput{Name: "from-file"},
		},
		{
			name:      "reads from stdin when path is dash",
			fs:        afero.NewMemMapFs(),
			inputPath: "-",
			stdinJSON: `{"name":"from-stdin"}`,
			expected:  jsonInput{Name: "from-stdin"},
		},
		{
			name: "returns error when fs open fails",
			fs: stubs.StubFS{
				Fs: afero.NewMemMapFs(),
				OpenFunc: func(_ afero.Fs, _ string) (afero.File, error) {
					return nil, errors.New("forced open error")
				},
			},
			inputPath:   "input.json",
			expectedErr: "forced open error",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			fsys := testCase.fs
			if testCase.fileContent != "" {
				err := afero.WriteFile(fsys, testCase.inputPath, []byte(testCase.fileContent), 0o600)
				require.NoError(t, err)
			}

			cleanup := setStdin(t, testCase.stdinJSON)
			defer cleanup()

			res, err := jsonutil.GetFromFileOrStdin[jsonInput](fsys, testCase.inputPath)
			if testCase.expectedErr != "" {
				require.EqualError(t, err, testCase.expectedErr)
				return
			}

			require.NoError(t, err)
			require.Equal(t, testCase.expected, res)
		})
	}
}

func setStdin(t *testing.T, stdinJSON string) func() {
	t.Helper()
	if stdinJSON == "" {
		return func() {}
	}

	reader, writer, err := os.Pipe()
	require.NoError(t, err)
	_, err = writer.WriteString(stdinJSON)
	require.NoError(t, err)
	require.NoError(t, writer.Close())

	previousStdin := os.Stdin
	os.Stdin = reader

	return func() {
		os.Stdin = previousStdin
		require.NoError(t, reader.Close())
	}
}
