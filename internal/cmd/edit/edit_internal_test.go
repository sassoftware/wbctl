// Copyright © 2026, SAS Institute Inc., Cary, NC, USA.  All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package edit

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/sassoftware/wbctl/internal/constants"
	"github.com/sassoftware/wbctl/internal/osutil"
	"github.com/sassoftware/wbctl/internal/stubs"
	"github.com/spf13/afero"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
)

func TestGetEditor(t *testing.T) {
	tests := []struct {
		name           string
		editorEnv      string
		expectedEditor string
		expectedErr    string
	}{
		{
			name:           "returns default when EDITOR not set",
			editorEnv:      "",
			expectedEditor: osutil.Executable(osutil.DefaultBinDir(), osutil.DefaultEditor()),
		},
		{
			name:           "returns EDITOR env var when set",
			editorEnv:      osutil.ExecutableScript("testdata", "no-change"),
			expectedEditor: osutil.ExecutableScript("testdata", "no-change"),
		},
		{
			name:        "returns error when env set to missing file",
			editorEnv:   "does-not-exist",
			expectedErr: `exec: "does-not-exist": executable file not found in ` + osutil.PathEnvVarName(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.editorEnv != "" {
				t.Setenv("EDITOR", tt.editorEnv)
			} else {
				t.Setenv("EDITOR", "")
			}

			editor, err := getEditor()
			if tt.expectedErr != "" {
				require.EqualError(t, err, tt.expectedErr)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.expectedEditor, editor)
			}
		})
	}
}

type marshalSucceeder struct {
	OmitEmptyField             string `json:"omit_empty_field,omitempty"`
	EditableField              string `json:"editable_field,omitempty"           wbctl:"editable"`
	SeedField                  string `json:"seed_field,omitempty"`
	EditableFieldWithSeed      string `json:"editable_field_with_seed,omitempty" wbctl:"editable,seedfield:SeedField"`
	ExcludedBecauseNotEditable string `json:"excluded_because_not_editable"`
}

const (
	validJSON = `{
  "omit_empty_field": "omit empty field value",
  "editable_field": "editable field value",
  "seed_field": "seed field value",
  "editable_field_with_seed": "edit field with seed field value",
  "exclude_because_not_editable": "exclude because not editable"
}`

	expectedValidJSON = `{
    "editable_field": "editable field value",
    "editable_field_with_seed": "edit field with seed field value"
}`

	extraWhiteSpaceJSON = `
{

  "omit_empty_field": "omit empty field value",
  "editable_field": "editable field value",

  "seed_field": "seed field value",
  "editable_field_with_seed":        "edit field with seed field value",
  "exclude_because_not_editable": "exclude because not editable"
}
`
	expectedExtraWhiteSpaceJSON = `{
    "editable_field": "editable field value",
    "editable_field_with_seed": "edit field with seed field value"
}`

	validWithSeedJSON = `{
  "omit_empty_field": "omit empty field value",
  "editable_field": "editable field value",
  "seed_field": "seed field value",
  "exclude_because_not_editable": "exclude because not editable"
}`

	expectedValidWithSeedJSON = `{
    "editable_field": "editable field value",
    "editable_field_with_seed": "seed field value"
}`
)

func TestCleanJSON(t *testing.T) {
	tests := []struct {
		name       string
		inputJSON  []byte
		outputJSON []byte
	}{
		{
			name:       "cleans valid JSON",
			inputJSON:  []byte(validJSON),
			outputJSON: []byte(expectedValidJSON),
		},
		{
			name:       "returns original on invalid JSON",
			inputJSON:  []byte(`{invalid json}`),
			outputJSON: []byte(`{invalid json}`),
		},
		{
			name:       "cleans JSON with extra whitespace and newlines",
			inputJSON:  []byte(extraWhiteSpaceJSON),
			outputJSON: []byte(expectedExtraWhiteSpaceJSON),
		},
		{
			name:       "cleans JSON with extra whitespace and newlines",
			inputJSON:  []byte(validWithSeedJSON),
			outputJSON: []byte(expectedValidWithSeedJSON),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := cleanJSON[marshalSucceeder](tt.inputJSON)
			require.Equal(t, string(tt.outputJSON), string(result), string(result))
		})
	}
}

type marshalFailer struct {
	Unmarshalable marshalInnerFailer `wbctl:"editable"`
}

type marshalInnerFailer struct {
	Field string
}

type marshalFailers []marshalFailer

func (m marshalInnerFailer) MarshalJSON() ([]byte, error) {
	return nil, errors.New("failed to marshal JSON")
}

func TestCleanMarshalFailsJSON(t *testing.T) {
	result := cleanJSON[marshalFailer]([]byte(`{"Unmarshalable": {"Field": "value"}}`))
	require.JSONEq(t, `{"Unmarshalable": {"Field": "value"}}`, string(result), string(result))
}

func TestEditFailsJSON(t *testing.T) {
	f := marshalFailer{}
	sr := func(slice marshalFailers) []marshalFailer {
		return slice
	}
	_, err := edit(context.Background(), sr, afero.NewMemMapFs(), marshalFailers{f})
	require.Error(t, err)
}

type editInput struct {
	resourceType string
	resourceIDs  []string
	fs           afero.Fs
	command      string
}

type editExpected struct {
	err string
}

type editTestCase struct {
	name     string
	input    editInput
	expected editExpected
}

const (
	testResourceID1 = "resource-1"
	testResourceID2 = "resource-2"
)

func TestNewEditFunctionErrorPaths(t *testing.T) {
	supportedTypes := []string{
		constants.ImageType,
		constants.JobType,
		constants.SnapshotType,
		constants.VolumeType,
		constants.WorkbenchType,
	}
	supportedTypeErr := strings.Join(supportedTypes, "\n    ")

	tests := []editTestCase{
		{
			name: "missing resource type",
			input: editInput{
				resourceIDs: []string{testResourceID1},
			},
			expected: editExpected{
				err: "resource type is required\n  supported types:\n    " + supportedTypeErr,
			},
		},
		{
			name: "unknown resource type",
			input: editInput{
				resourceType: "unknown",
				resourceIDs:  []string{testResourceID1},
			},
			expected: editExpected{
				err: "unknown resource type: unknown\n  supported types:\n    " + supportedTypeErr,
			},
		},
		{
			name: "multiple ids",
			input: editInput{
				resourceType: constants.JobType,
				resourceIDs:  []string{testResourceID1, testResourceID2},
			},
			expected: editExpected{
				err: "Only one ID supported",
			},
		},
		{
			name: "no ids",
			input: editInput{
				resourceType: constants.JobType,
			},
			expected: editExpected{
				err: "ID required",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer viper.Reset()

			stubUI := &stubs.StubUI{}
			t.Setenv("EDITOR", tt.input.command)
			fs := afero.NewOsFs()
			if tt.input.fs != nil {
				fs = tt.input.fs
			}
			action := newEditFunction(context.Background(), fs, stubUI, tt.input.resourceType)
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
