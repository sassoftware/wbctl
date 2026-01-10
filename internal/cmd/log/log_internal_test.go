// Copyright © 2026, SAS Institute Inc., Cary, NC, USA.  All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package log

import (
	"context"
	"strings"
	"testing"

	"github.com/sas-institute-rnd-saas/wbctl/internal/constants"
	"github.com/sas-institute-rnd-saas/wbctl/internal/stubs"
	"github.com/spf13/afero"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
)

type logInput struct {
	resourceIDs  []string
	resourceType string
}

type logExpected struct {
	err string
}

type logTestCase struct {
	name     string
	input    logInput
	expected logExpected
}

func TestNewLogFunctionErrorPaths(t *testing.T) {
	supportedTypes := []string{
		constants.ImageBuildsType,
		constants.JobSubmissionsType,
	}
	supportedTypeMessage := strings.Join(supportedTypes, "\n    ")

	tests := []logTestCase{
		{
			name: "new log function missing resource type",
			input: logInput{
				resourceIDs: []string{"controller-1"},
			},
			expected: logExpected{
				err: "resource type is required\n  supported types:\n    " + supportedTypeMessage,
			},
		},
		{
			name: "new log function unknown resource type",
			input: logInput{
				resourceType: "unknown",
				resourceIDs:  []string{"controller-1"},
			},
			expected: logExpected{
				err: "unknown resource type: unknown\n  supported types:\n    " + supportedTypeMessage,
			},
		},
		{
			name:     "new log function errors when given no ids",
			input:    logInput{resourceType: constants.ImageBuildType},
			expected: logExpected{err: "at least one resource ID must be provided for logs"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer viper.Reset()

			stubUI := &stubs.StubUI{}
			action := newLogFunction(context.Background(), afero.NewMemMapFs(), stubUI, tt.input.resourceType)
			err := action(nil, tt.input.resourceIDs)
			if tt.expected.err != "" {
				expectedErr := tt.expected.err
				require.EqualError(t, err, expectedErr)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
