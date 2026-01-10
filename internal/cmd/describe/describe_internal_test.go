// Copyright © 2026, SAS Institute Inc., Cary, NC, USA.  All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package describe

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/sas-institute-rnd-saas/wbctl/internal/constants"
	"github.com/sas-institute-rnd-saas/wbctl/internal/stubs"
	"github.com/sas-institute-rnd-saas/wbctl/internal/timeutil"
	"github.com/spf13/afero"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
)

type describeInput struct {
	resourceIDs  []string
	resourceType string
}

type describeExpected struct {
	err string
}

type describeTestCase struct {
	name     string
	input    describeInput
	expected describeExpected
}

func TestNewDescribeFunctionErrorPaths(t *testing.T) {
	fixedNow := time.Date(2025, time.May, 8, 0, 0, 0, 0, time.UTC)
	timeutil.SetNow(func() time.Time { return fixedNow })
	t.Cleanup(func() { timeutil.SetNow(time.Now) })

	supportedTypes := []string{
		constants.DataplanesType,
		constants.ImagesType,
		constants.ImageBuildsType,
		constants.JobsType,
		constants.JobSubmissionsType,
		constants.SnapshotsType,
		constants.UsersType,
		constants.VolumesType,
		constants.VolumeConflictsType,
		constants.WorkbenchesType,
	}

	tests := []describeTestCase{
		{
			name: "missing resource type",
			expected: describeExpected{
				err: "resource type is required\n  supported types:\n    " +
					strings.Join(supportedTypes, "\n    "),
			},
		},
		{
			name: "unknown resource type",
			input: describeInput{
				resourceType: "unknown",
				resourceIDs:  []string{"conflict-1"},
			},
			expected: describeExpected{
				err: "unknown resource type: unknown\n  supported types:\n    " +
					strings.Join(supportedTypes, "\n    "),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer viper.Reset()

			stubUI := &stubs.StubUI{}
			action := newDescribeFunction(context.Background(), afero.NewMemMapFs(), stubUI, tt.input.resourceType)
			err := action(nil, nil)
			if tt.expected.err != "" {
				expectedErr := tt.expected.err
				require.EqualError(t, err, expectedErr)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
