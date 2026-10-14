// Copyright © 2026, SAS Institute Inc., Cary, NC, USA.  All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package rebuild

import (
	"context"
	"testing"

	"github.com/sassoftware/wbctl/internal/constants"
	"github.com/sassoftware/wbctl/internal/stubs"
	"github.com/spf13/afero"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
)

type rebuildInput struct {
	resourceIDs  []string
	resourceType string
}

type rebuildExpected struct {
	err string
}

type rebuildTestCase struct {
	name     string
	input    rebuildInput
	expected rebuildExpected
}

func TestNewRebuildFunctionErrorPaths(t *testing.T) {
	tests := []rebuildTestCase{
		{
			name:     "missing resource type",
			input:    rebuildInput{resourceIDs: []string{"controller-1"}},
			expected: rebuildExpected{err: "resource type is required\n  supported types:\n    " + constants.ImagesType},
		},
		{
			name:     "missing resource ids",
			input:    rebuildInput{resourceType: constants.ImageType},
			expected: rebuildExpected{err: "at least one resource ID must be provided for rebuild"},
		},
		{
			name:  "unknown resource type",
			input: rebuildInput{resourceType: "unknown", resourceIDs: []string{"controller-1"}},
			expected: rebuildExpected{
				err: "unknown resource type: unknown\n  supported types:\n    " + constants.ImagesType,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer viper.Reset()

			stubUI := &stubs.StubUI{}
			action := newRebuildFunction(context.Background(), afero.NewMemMapFs(), stubUI, tt.input.resourceType)
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
