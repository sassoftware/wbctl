// Copyright © 2026, SAS Institute Inc., Cary, NC, USA.  All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package cancel

import (
	"context"
	"testing"

	"github.com/sassoftware/wbctl/internal/constants"
	"github.com/sassoftware/wbctl/internal/stubs"
	"github.com/spf13/afero"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
)

const (
	cancelTestControllerID1       = "controller-1"
	cancelTestUnknownResourceType = "unknown"
	cancelTestMissingIDsErr       = "at least one resource ID must be provided for cancel"
	cancelTestUnknownTypeErr      = "unknown resource type: unknown" +
		"\n  supported types:" +
		"\n    " + constants.JobSubmissionsType
	cancelTestMissingResourceTypeErr = "resource type is required" +
		"\n  supported types:" +
		"\n    " + constants.JobSubmissionsType
)

type cancelInput struct {
	resourceType string
	resourceIDs  []string
}

type cancelExpected struct {
	err string
}

type cancelTestCase struct {
	name     string
	input    cancelInput
	expected cancelExpected
}

func TestNewCancelFunctionErrorPaths(t *testing.T) {
	tests := []cancelTestCase{
		{
			name:     "missing resource type",
			input:    cancelInput{resourceIDs: []string{cancelTestControllerID1}},
			expected: cancelExpected{err: cancelTestMissingResourceTypeErr},
		},
		{
			name:     "missing resource ids",
			input:    cancelInput{resourceType: constants.JobSubmissionType},
			expected: cancelExpected{err: cancelTestMissingIDsErr},
		},
		{
			name: "unknown resource type",
			input: cancelInput{
				resourceType: cancelTestUnknownResourceType,
				resourceIDs:  []string{cancelTestControllerID1},
			},
			expected: cancelExpected{err: cancelTestUnknownTypeErr},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer viper.Reset()

			stubUI := &stubs.StubUI{}
			action := newCancelFunction(context.Background(), afero.NewMemMapFs(), stubUI, tt.input.resourceType)
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
