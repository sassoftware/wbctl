// Copyright © 2026, SAS Institute Inc., Cary, NC, USA.  All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package wbctljson_test

import (
	"errors"
	"strings"
	"testing"
	"testing/iotest"

	"github.com/sas-institute-rnd-saas/wbctl/internal/types"
	"github.com/sas-institute-rnd-saas/wbctl/internal/wbctljson"
	"github.com/stretchr/testify/require"
)

// Expected output constants for TestMarshalIndentEditOnlyFields.
const (
	expectedInvalidSeedfieldJSON = `{}`
	expectedJobSingle            = `{
    "name": "test-job"
}`
	expectedJobMultiple = `{
    "application_id": "app-1",
    "context": "ctx-1",
    "description": "test description",
    "labels": {
        "env": "test"
    },
    "name": "test-job",
    "nodepool": "pool-1",
    "properties": {
        "key": "value"
    },
    "schedule": "0 0 * * *",
    "sub_path": "/sub"
}`
	expectedJobPointer = `{
    "application_id": "app-2",
    "name": "ptr-job"
}`
	expectedJobOmitempty = `{
    "name": "only-name"
}`
	expectedWorkbenchZero      = `{}`
	expectedWorkbenchSeedfield = `{
    "name": "test-wb",
    "state": "Running"
}`
	expectedWorkbenchState = `{
    "name": "state-wb",
    "state": "Stopped"
}`
	expectedJobServices = `{
    "name": "job-with-services",
    "services": [
        {
            "id": "svc-1",
            "image": "img-1"
        },
        {
            "id": "svc-2",
            "image": "img-2"
        }
    ]
}`
	expectedJobWorkspaces = `{
    "name": "job-with-workspaces",
    "workspaces": [
        {
            "id": "ws-1",
            "mountName": "mount-1"
        }
    ]
}`
	expectedJobEmptyLabels = `{
    "name": "job-with-empty-labels"
}`
	expectedWorkbenchMultiple = `{
    "labels": {
        "env": "prod",
        "tier": "premium"
    },
    "name": "multi-field-wb",
    "nodepool": "high-mem",
    "properties": {
        "cpu": 4,
        "memory": 8
    }
}`
	expectedUntaggedField = `{
    "UntaggedField": "untagged-field"
}`
	expectedMapJSON  = `{    "a":"b"}`
	expectedValidMap = `{
    "ids": [
        "a",
        "b"
    ],
    "param": {
        "description": "test",
        "name": "test"
    }
}`
	expectedValidNestedMap = `{
    "ids": [
        "a",
        "b"
    ],
    "nest1": {
        "nest2": {
            "nest3": {
                "param": {
                    "description": "test",
                    "name": "test"
                }
            }
        }
    }
}`
	expectedSliceJSON  = `[    "slice-item"]`
	expectedStructJSON = `{    "value":"struct-value"}`
	expectedValidSlice = `[
    {
        "description": "test",
        "name": "slice-image"
    },
    "b"
]`
	expectedIgnoredField = `{
    "visible": "keep"
}`
)

// Input fixtures for TestMarshalIndentEditOnlyFields.
var (
	testJobSingle = types.Job{
		ID:   "non-editable-id",
		Name: "test-job",
	}
	testJobMultiple = types.Job{
		ID:            "non-editable-id",
		Name:          "test-job",
		ApplicationID: "app-1",
		Context:       "ctx-1",
		Description:   "test description",
		Labels:        map[string]string{"env": "test"},
		Nodepool:      "pool-1",
		Schedule:      "0 0 * * *",
		Properties:    map[string]interface{}{"key": "value"},
		SubPath:       "/sub",
	}
	testJobPointer = &types.Job{
		Name:          "ptr-job",
		ApplicationID: "app-2",
	}
	testJobOmitempty = types.Job{
		Name: "only-name",
	}
	testWorkbenchZero = types.Workbench{
		ID: "wb-1",
	}
	testWorkbenchSeedfield = types.Workbench{
		ID:           "wb-1",
		Name:         "test-wb",
		DesiredState: "Running",
	}
	testWorkbenchState = types.Workbench{
		ID:           "wb-2",
		Name:         "state-wb",
		DesiredState: "Running",
		State:        "Stopped",
	}
	testJobServices = types.Job{
		Name: "job-with-services",
		Services: []types.ServiceRequest{
			{ID: "svc-1", Image: "img-1"},
			{ID: "svc-2", Image: "img-2"},
		},
	}
	testJobWorkspaces = types.Job{
		Name: "job-with-workspaces",
		Workspaces: []types.WorkspaceRequest{
			{ID: "ws-1", MountName: "mount-1"},
		},
	}
	testJobEmptyLabels = types.Job{
		Name:   "job-with-empty-labels",
		Labels: map[string]string{},
	}
	testWorkbenchMultiple = types.Workbench{
		ID:         "wb-3",
		Name:       "multi-field-wb",
		Labels:     map[string]string{"tier": "premium", "env": "prod"},
		Nodepool:   "high-mem",
		Properties: map[string]interface{}{"cpu": 4, "memory": 8},
	}
	testInvalidSeedfieldStructInput = testInvalidSeedfieldStruct{}
	testSimpleMapInput              = map[string]any{"a": "b"}
	testSimpleSliceInput            = []any{"slice-item"}
	testSimpleStructInput           = testMarshalStruct{
		Value: "struct-value",
	}
	testIgnoredFieldInput = struct {
		Visible string `json:"visible,omitempty" wbctl:"editable"`
		Hidden  string `json:"hidden,omitempty"  wbctl:"-"`
	}{
		Visible: "keep",
		Hidden:  "drop",
	}
)

// testStructWithInvalidType has a field that can't be marshaled to JSON.
type testStructWithInvalidType struct {
	Name string      `json:"name" wbctl:"editable"`
	Chan chan string `json:"chan" wbctl:"editable"`
}

type testInvalidSeedfieldStruct struct {
	Value string `json:"value,omitempty" wbctl:"editable,seedfield:MissingField"`
}

type testMarshalStruct struct {
	Value string `json:"value"`
}

func TestMarshalIndentEditOnlyFields(t *testing.T) {
	tests := []struct {
		name        string
		input       interface{}
		expected    string
		expectedErr string
	}{
		{
			name:     "job with single editable field",
			input:    testJobSingle,
			expected: expectedJobSingle,
		},
		{
			name:     "job with multiple editable fields",
			input:    testJobMultiple,
			expected: expectedJobMultiple,
		},
		{
			name:     "job pointer with editable fields",
			input:    testJobPointer,
			expected: expectedJobPointer,
		},
		{
			name:     "job with empty editable fields uses omitempty",
			input:    testJobOmitempty,
			expected: expectedJobOmitempty,
		},
		{
			name:     "workbench with all zero fields (omitempty skips empty fields)",
			input:    testWorkbenchZero,
			expected: expectedWorkbenchZero,
		},
		{
			name:     "workbench state field populated with seedfield",
			input:    testWorkbenchSeedfield,
			expected: expectedWorkbenchSeedfield,
		},
		{
			name:     "workbench with state set directly",
			input:    testWorkbenchState,
			expected: expectedWorkbenchState,
		},
		{
			name:     "job with services slice",
			input:    testJobServices,
			expected: expectedJobServices,
		},
		{
			name:     "job with workspaces slice",
			input:    testJobWorkspaces,
			expected: expectedJobWorkspaces,
		},
		{
			name:     "job with empty labels map",
			input:    testJobEmptyLabels,
			expected: expectedJobEmptyLabels,
		},
		{
			name:     "workbench with multiple editable fields including map",
			input:    testWorkbenchMultiple,
			expected: expectedWorkbenchMultiple,
		},
		{
			name: "field without json tag",
			input: struct {
				UntaggedField string `wbctl:"editable"`
			}{
				UntaggedField: "untagged-field",
			},
			expected: expectedUntaggedField,
		},
		{
			name:     "ignored field is skipped",
			input:    testIgnoredFieldInput,
			expected: expectedIgnoredField,
		},
		{
			name: "valid map",
			input: map[string]interface{}{
				"ids": []string{"a", "b"},
				"param": types.Image{
					Description: "test",
					Name:        "test",
				},
			},
			expected: expectedValidMap,
		},
		{
			name: "nested valid map",
			input: map[string]interface{}{
				"ids": []string{"a", "b"},
				"nest1": map[string]interface{}{
					"nest2": map[string]interface{}{
						"nest3": map[string]interface{}{
							"param": types.Image{
								Description: "test",
								Name:        "test",
							},
						},
					},
				},
			},
			expected: expectedValidNestedMap,
		},
		{
			name: "invalid map keys",
			input: map[int]string{
				1: "test",
			},
			expectedErr: "unsupported map key kind: int",
		},
		{
			name: "error when field contains unmarshalable type",
			input: testStructWithInvalidType{
				Name: "test",
				Chan: make(chan string),
			},
			expectedErr: "json: unsupported type: chan string",
		},
		{
			name: "error when field in map contains unmarshalable type",
			input: map[string]interface{}{
				"test": testStructWithInvalidType{
					Name: "test",
					Chan: make(chan string),
				},
			},
			expectedErr: "json: unsupported type: chan string",
		},
		{
			name: "valid slice",
			input: []interface{}{
				types.Image{
					Description: "test",
					Name:        "slice-image",
				},
				"b",
			},
			expected: expectedValidSlice,
		},
		{
			name:        "invalid channel",
			input:       make(chan string),
			expectedErr: "unsupported type: chan",
		},
		{
			name:        "invalid slice of channel",
			input:       []chan string{make(chan string)},
			expectedErr: "json: unsupported type: chan string",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual, err := wbctljson.MakeMarshaler[any](wbctljson.EditOptions)(tt.input)

			// The only error cases are handled in types_internal_test.go
			if tt.expectedErr != "" {
				require.EqualError(t, err, tt.expectedErr)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.expected, string(actual), string(actual))
			}
		})
	}
}

var testOptions = wbctljson.MarshalerOptions{
	EditOnly: false,
	Pretty:   false,
	Indent:   "    ",
}

func TestMarshalWbctlTypeCoverageBranches(t *testing.T) {
	testCases := []struct {
		name     string
		input    any
		opts     wbctljson.MarshalerOptions
		expected string
	}{
		{
			name:     "map with scalar value marshals compact",
			input:    testSimpleMapInput,
			opts:     testOptions,
			expected: expectedMapJSON,
		},
		{
			name:     "slice with scalar value marshals compact",
			input:    testSimpleSliceInput,
			opts:     testOptions,
			expected: expectedSliceJSON,
		},
		{
			name:     "struct marshals compact",
			input:    testSimpleStructInput,
			opts:     testOptions,
			expected: expectedStructJSON,
		},
		{
			name:     "invalid seedfield reference is skipped",
			input:    testInvalidSeedfieldStructInput,
			opts:     wbctljson.EditOptions,
			expected: expectedInvalidSeedfieldJSON,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			actual, err := wbctljson.MakeMarshaler[any](testCase.opts)(testCase.input)
			require.NoError(t, err)
			require.Equal(t, testCase.expected, string(actual))
		})
	}
}

func TestMakeWbctlUnmarshaler(t *testing.T) {
	tests := []struct {
		name string
		run  func(*testing.T)
	}{
		{
			name: "returns invalid unmarshal error for nil target pointer",
			run: func(t *testing.T) {
				unmarshal := wbctljson.MakeUnmarshaler[types.Conflict]()
				var actual *types.Conflict

				err := unmarshal(strings.NewReader(`{"volume_name":"vol-1"}`), actual)

				require.ErrorContains(t, err, "json: Unmarshal(nil *types.Conflict)")
			},
		},
		{
			name: "returns json error for invalid payload",
			run: func(t *testing.T) {
				unmarshal := wbctljson.MakeUnmarshaler[types.Conflict]()
				actual := types.Conflict{}

				err := unmarshal(strings.NewReader(`{"volume_name":`), &actual)

				require.ErrorContains(t, err, "unexpected end of JSON input")
			},
		},
		{
			name: "returns type error when struct target payload is not object",
			run: func(t *testing.T) {
				unmarshal := wbctljson.MakeUnmarshaler[types.Conflict]()
				actual := types.Conflict{}

				err := unmarshal(strings.NewReader(`"not-an-object"`), &actual)

				require.ErrorContains(t, err, "cannot unmarshal string into Go value of type types.Conflict")
			},
		},
		{
			name: "returns type error when slice target payload is not array",
			run: func(t *testing.T) {
				unmarshal := wbctljson.MakeUnmarshaler[[]types.Conflict]()
				actual := []types.Conflict{}

				err := unmarshal(strings.NewReader(`{"volume_name":"vol-1"}`), &actual)

				require.ErrorContains(t, err, "cannot unmarshal object into Go value of type []types.Conflict")
			},
		},
		{
			name: "returns type error when map target payload is not object",
			run: func(t *testing.T) {
				unmarshal := wbctljson.MakeUnmarshaler[map[string]types.Conflict]()
				actual := map[string]types.Conflict{}

				err := unmarshal(strings.NewReader(`[{"volume_name":"vol-1"}]`), &actual)

				require.ErrorContains(t, err, "cannot unmarshal array into Go value of type map[string]types.Conflict")
			},
		},
		{
			name: "returns error when reader fails",
			run: func(t *testing.T) {
				unmarshal := wbctljson.MakeUnmarshaler[types.Conflict]()
				actual := types.Conflict{}
				err := unmarshal(iotest.ErrReader(errors.New("read error")), &actual)
				require.ErrorContains(t, err, "reading input: read error")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.run(t)
		})
	}
}
