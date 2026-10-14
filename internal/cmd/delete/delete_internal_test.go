// Copyright © 2026, SAS Institute Inc., Cary, NC, USA.  All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Package delete provides the delete command.
// This package name collides with the builtin delete function but context is
// enough to distinguish it.
// nolint: predeclared
package delete

import (
	"context"
	"strings"
	"testing"

	"github.com/sassoftware/wbctl/internal/constants"
	"github.com/sassoftware/wbctl/internal/stubs"
	"github.com/spf13/afero"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
)

type deleteInput struct {
	resourceIDs  []string
	resourceType string
}

type deleteExpected struct {
	err string
}

type deleteTestCase struct {
	name     string
	input    deleteInput
	expected deleteExpected
}

func TestNewDeleteFunctionErrorPaths(t *testing.T) {
	supportedTypes := []string{
		constants.ImagesType,
		constants.JobsType,
		constants.JobSubmissionsType,
		constants.SnapshotsType,
		constants.VolumesType,
		constants.WorkbenchesType,
	}
	supportedTypeErr := strings.Join(supportedTypes, "\n    ")
	tests := []deleteTestCase{
		{
			name: "missing resource type",
			input: deleteInput{
				resourceIDs: []string{"resource-1"},
			},
			expected: deleteExpected{
				err: "resource type is required\n  supported types:\n    " + supportedTypeErr,
			},
		},
		{
			name: "unknown resource type",
			input: deleteInput{
				resourceType: "unknown",
				resourceIDs:  []string{"resource-1"},
			},
			expected: deleteExpected{
				err: "unknown resource type: unknown\n  supported types:\n    " + supportedTypeErr,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer viper.Reset()

			stubUI := &stubs.StubUI{}
			action := newDeleteFunction(context.Background(), afero.NewMemMapFs(), stubUI, tt.input.resourceType)
			err := action(nil, nil)
			if tt.expected.err != "" {
				if strings.Contains(tt.expected.err, "%s") {
					expectedErr := tt.expected.err
					require.EqualError(t, err, expectedErr)
				} else {
					require.EqualError(t, err, tt.expected.err)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}
