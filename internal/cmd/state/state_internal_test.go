// Copyright © 2026, SAS Institute Inc., Cary, NC, USA.  All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package state

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

type stateInput struct {
	resourceIDs  []string
	resourceType string
}

type stateExpected struct {
	err string
}

type stateTestCase struct {
	name     string
	input    stateInput
	expected stateExpected
}

func TestNewUpdateStateFunctionErrorPaths(t *testing.T) {
	resourceTypesMessage := strings.Join([]string{constants.WorkbenchesType}, "\n    ")

	actionTests := []struct {
		actionName   string
		desiredState string
	}{
		{
			actionName:   "start",
			desiredState: "RUNNING",
		},
		{
			actionName:   "stop",
			desiredState: "STOPPED",
		},
	}

	tests := []stateTestCase{
		{
			name: "missing resource type",
			input: stateInput{
				resourceIDs: []string{"controller-1"},
			},
			expected: stateExpected{
				err: "resource type is required\n  supported types:\n    " + resourceTypesMessage,
			},
		},
		{
			name: "missing resource ids",
			input: stateInput{
				resourceType: constants.WorkbenchType,
			},
			expected: stateExpected{
				err: "at least one resource ID must be provided for ${action}",
			},
		},
		{
			name: "unknown resource type",
			input: stateInput{
				resourceType: "unknown",
				resourceIDs:  []string{"controller-1"},
			},
			expected: stateExpected{
				err: "unknown resource type: unknown\n  supported types:\n    " + resourceTypesMessage,
			},
		},
	}

	for _, actionTest := range actionTests {
		for _, tt := range tests {
			t.Run(actionTest.actionName+" "+tt.name, func(t *testing.T) {
				defer viper.Reset()

				stubUI := &stubs.StubUI{}
				action := newUpdateStateFunction(
					context.Background(),
					afero.NewMemMapFs(),
					stubUI,
					tt.input.resourceType,
					actionTest.desiredState,
				)
				err := action(nil, tt.input.resourceIDs)
				if tt.expected.err != "" {
					expectedErr := strings.ReplaceAll(tt.expected.err, "${action}", actionTest.actionName)
					require.EqualError(t, err, expectedErr)
				} else {
					require.NoError(t, err)
				}
			})
		}
	}
}
