// Copyright © 2026, SAS Institute Inc., Cary, NC, USA.  All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package workbench

import (
	"context"
	"errors"
	"testing"

	"github.com/sassoftware/wbctl/internal/constants"
	"github.com/sassoftware/wbctl/internal/i18n"
	"github.com/stretchr/testify/require"
)

type testResource struct {
	ID           string
	ControllerID string
	Name         string
}

type testResources []testResource

func (r testResource) GetID() string           { return r.ID }
func (r testResource) GetControllerID() string { return r.ControllerID }
func (r testResource) GetName() string         { return r.Name }

func TestSelectByIDThenName(t *testing.T) {
	a := testResource{ID: "id-a", ControllerID: "ctrl-a", Name: "name-a"}
	b := testResource{ID: "id-b", ControllerID: "ctrl-b", Name: "name-b"}
	// c has a name equal to a's ID, to exercise ID > name precedence.
	c := testResource{ID: "id-c", ControllerID: "ctrl-c", Name: "id-a"}
	// d shares a name with b.
	d := testResource{ID: "id-d", ControllerID: "ctrl-d", Name: "name-b"}

	tests := []struct {
		name      string
		args      []string
		resources testResources
		expected  testResources
	}{
		{
			name:      "no args returns all resources",
			args:      nil,
			resources: testResources{a, b},
			expected:  testResources{a, b},
		},
		{
			name:      "empty string args are skipped",
			args:      []string{"", "id-a"},
			resources: testResources{a, b},
			expected:  testResources{a},
		},
		{
			name:      "match by id",
			args:      []string{"id-a"},
			resources: testResources{a, b},
			expected:  testResources{a},
		},
		{
			name:      "match by controller id",
			args:      []string{"ctrl-b"},
			resources: testResources{a, b},
			expected:  testResources{b},
		},
		{
			name:      "match by name",
			args:      []string{"name-a"},
			resources: testResources{a, b},
			expected:  testResources{a},
		},
		{
			name:      "id takes precedence over name",
			args:      []string{"id-a"},
			resources: testResources{a, c},
			expected:  testResources{a},
		},
		{
			name:      "name used when no id matches",
			args:      []string{"name-b"},
			resources: testResources{a, b, d},
			expected:  testResources{b, d},
		},
		{
			name:      "no match returns empty",
			args:      []string{"missing"},
			resources: testResources{a, b},
			expected:  testResources{},
		},
		{
			name:      "multiple args are resolved independently",
			args:      []string{"id-a", "name-b"},
			resources: testResources{a, b, c, d},
			expected:  testResources{a, b, d},
		},
		{
			name:      "resource matched by several args is returned once",
			args:      []string{"id-a", "name-a", "ctrl-a"},
			resources: testResources{a, b},
			expected:  testResources{a},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := selectByIDThenName(
				tt.args,
				tt.resources,
			)
			require.Equal(t, tt.expected, actual)
		})
	}
}

func TestResolveIDs(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		getter      func(context.Context, *WorkbenchService, []string, string) (testResources, error)
		expected    []string
		expectedErr string
	}{
		{
			name: "empty args returns args without calling getter",
			args: nil,
			getter: func(context.Context, *WorkbenchService, []string, string) (testResources, error) {
				t.Error("getter must not be called")
				return nil, nil
			},
			expected: nil,
		},
		{
			name: "resolves to ids",
			args: []string{"name-a", "ctrl-b"},
			getter: func(_ context.Context, _ *WorkbenchService, args []string, sortSpec string) (testResources, error) {
				require.Equal(t, []string{"name-a", "ctrl-b"}, args)
				require.Empty(t, sortSpec)
				return testResources{{ID: "id-a"}, {ID: "id-b"}}, nil
			},
			expected: []string{"id-a", "id-b"},
		},
		{
			name: "getter error is returned",
			args: []string{"name-a"},
			getter: func(context.Context, *WorkbenchService, []string, string) (testResources, error) {
				return nil, errors.New("getter err")
			},
			expectedErr: "getter err",
		},
		{
			name: "no matches returns not found error",
			args: []string{"ws-1", "ws-2"},
			getter: func(context.Context, *WorkbenchService, []string, string) (testResources, error) {
				return testResources{}, nil
			},
			expectedErr: i18n.Format(
				i18n.ResourceNotFoundWithID,
				i18n.A{
					constants.I18nArgTypeKey: constants.VolumeType,
					constants.I18nArgKeyID:   "ws-1, ws-2",
				},
			),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual, err := resolveIDs(
				context.Background(), nil, tt.args, tt.getter, testResource.GetID, constants.VolumeType,
			)
			if tt.expectedErr != "" {
				require.EqualError(t, err, tt.expectedErr)
				require.Nil(t, actual)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.expected, actual)
		})
	}
}
