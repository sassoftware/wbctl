// Copyright © 2026, SAS Institute Inc., Cary, NC, USA.  All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package get

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/sassoftware/wbctl/internal/constants"
	"github.com/sassoftware/wbctl/internal/stubs"
	"github.com/sassoftware/wbctl/internal/timeutil"
	"github.com/spf13/afero"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
)

type getInput struct {
	resourceIDs  []string
	resourceType string
}

type getExpected struct {
	err string
}

type getTestCase struct {
	name     string
	input    getInput
	expected getExpected
}

func TestNewGetFunctionErrorPaths(t *testing.T) {
	fixedNow := time.Date(2025, time.May, 8, 0, 0, 0, 0, time.UTC)
	timeutil.SetNow(func() time.Time { return fixedNow })
	t.Cleanup(func() { timeutil.SetNow(time.Now) })

	supportedTypes := []string{
		constants.ApplicationsType,
		constants.DataplanesType,
		constants.EventsType,
		constants.ImagesType,
		constants.ImageBuildsType,
		constants.InitializersType,
		constants.JobsType,
		constants.JobSubmissionsType,
		constants.NodepoolsType,
		constants.PingsType,
		constants.ServicesType,
		constants.SnapshotsType,
		constants.StorageclassesType,
		constants.UsersType,
		constants.VolumesType,
		constants.VolumeConflictsType,
		constants.WorkbenchesType,
	}

	tests := []getTestCase{
		{
			name: "missing resource type",
			expected: getExpected{
				err: "resource type is required\n  supported types:\n    " + strings.Join(supportedTypes, "\n    "),
			},
		},
		{
			name: "unknown resource type",
			input: getInput{
				resourceType: "unknown",
				resourceIDs:  []string{"conflict-1"},
			},
			expected: getExpected{
				err: "unknown resource type: unknown\n  supported types:\n    " + strings.Join(supportedTypes, "\n    "),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer viper.Reset()

			stubUI := &stubs.StubUI{}
			action := newGetFunction(context.Background(), afero.NewMemMapFs(), stubUI, tt.input.resourceType)
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
