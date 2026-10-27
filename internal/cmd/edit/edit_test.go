// Copyright © 2026, SAS Institute Inc., Cary, NC, USA.  All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package edit_test

import (
	"context"
	"errors"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/sassoftware/wbctl/internal/cmd"
	"github.com/sassoftware/wbctl/internal/cmd/edit"
	"github.com/sassoftware/wbctl/internal/cmd/flags"
	"github.com/sassoftware/wbctl/internal/constants"
	"github.com/sassoftware/wbctl/internal/osutil"
	"github.com/sassoftware/wbctl/internal/postgrest"
	"github.com/sassoftware/wbctl/internal/stubs"
	"github.com/spf13/afero"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
)

const (
	testResourceID1 = "resource-1"
	oneImageJSON    = `[
    {
        "image_name": "test-image",
        "image_registry": "example.com",
        "image_tag": "latest",
        "name": "replace-me"
    }
]`
	twoImagesJSON = `[
    {
        "image_name": "test-image",
        "image_registry": "example.com",
        "image_tag": "latest",
        "name": "resource-1"
    },
    {
        "image_name": "test-image",
        "image_registry": "example.com",
        "image_tag": "latest",
        "name": "resource-2"
    }
]`
	oneUpdatedImageJSON = `{"ids":["resource-1"],"param":{"image_name":"test-image",` +
		`"image_registry":"example.com","image_tag":"latest","name":"UPDATED"}}`
	oneWorkbenchJSON = `[
    {
        "id": "resource-1",
        "application_id": "vscode",
        "controller_id": "sas-workbench-eb29ajr40lrt9gpicxjy5qdy8",
        "created_at": "2026-06-03T15:14:34.040941Z",
        "created_by": "c6139205-3b7e-406a-b5c5-f777f6c851c9",
        "current_state": "STOPPED",
        "dataplane_id": "86232878-4412-418b-a768-22038a1c4821",
        "desired_state": "STOPPED",
        "email": "pupadmin1@sas-test-mickey.com",
        "modified_at": "2026-06-08T14:02:03.536715Z",
        "modified_by": "c6139205-3b7e-406a-b5c5-f777f6c851c9",
        "name": "updatedname",
        "nodepool": "C0M0",
        "services": [
            {
                "id": "compute",
                "env": {
                    "DEFAULT_MOUNTNAME": "edit"
                }
            },
            {
                "id": "vscode",
                "env": {
                    "DEFAULT_MOUNTNAME": "edit"
                }
            }
        ],
        "state": "STOPPED",
        "url": "https://wbdev-data-plane.rnd-dev.workbench.sas.com/sas-workbench-eb29ajr40lrt9gpicxjy5qdy8",
        "username": "pupadmin1 Test User",
        "workspace_names": {
            "sas-workbench-v0fq4c4ff9xqlv70j76amcqe": "edit"
        },
        "workspaces": [
            {
                "id": "sas-workbench-v0fq4c4ff9xqlv70j76amcqe",
                "mountName": "replace-me"
            }
        ]
    }
]`
	oneWorkbenchUpdateResponseJSON = `{
        "id": "resource-1",
        "application_id": "vscode",
        "controller_id": "sas-workbench-eb29ajr40lrt9gpicxjy5qdy8",
        "created_at": "2026-06-03T15:14:34.040941Z",
        "created_by": "c6139205-3b7e-406a-b5c5-f777f6c851c9",
        "current_state": "STOPPED",
        "dataplane_id": "86232878-4412-418b-a768-22038a1c4821",
        "desired_state": "STOPPED",
        "email": "pupadmin1@sas-test-mickey.com",
        "modified_at": "2026-06-08T14:02:03.536715Z",
        "modified_by": "c6139205-3b7e-406a-b5c5-f777f6c851c9",
        "name": "updatedname",
        "nodepool": "C0M0",
        "services": [
            {
                "id": "compute",
                "env": {
                    "DEFAULT_MOUNTNAME": "edit"
                }
            },
            {
                "id": "vscode",
                "env": {
                    "DEFAULT_MOUNTNAME": "edit"
                }
            }
        ],
        "state": "STOPPED",
        "url": "https://wbdev-data-plane.rnd-dev.workbench.sas.com/sas-workbench-eb29ajr40lrt9gpicxjy5qdy8",
        "username": "pupadmin1 Test User",
        "workspace_names": {
            "sas-workbench-v0fq4c4ff9xqlv70j76amcqe": "edit"
        },
        "workspaces": [
            {
                "id": "sas-workbench-v0fq4c4ff9xqlv70j76amcqe",
                "mountName": "replace-me"
            }
        ]
    }`
	oneUpdatedWorkbenchJSON = `{"param":{` +
		`"id":"resource-1",` +
		`"application_id":"vscode",` +
		`"dataplane_id":"86232878-4412-418b-a768-22038a1c4821",` +
		`"name":"updatedname",` +
		`"nodepool":"C0M0",` +
		`"services":[{"id":"compute","env":{"DEFAULT_MOUNTNAME":"edit"}},` +
		`{"id":"vscode","env":{"DEFAULT_MOUNTNAME":"edit"}}],` +
		`"state":"STOPPED",` +
		`"workspaces":[{"id":"sas-workbench-v0fq4c4ff9xqlv70j76amcqe","mountName":"UPDATED"}]` +
		`}}`
	oneJobJSON = `[{
    "application_id": "workbench-environment-builder",
    "dataplane_id": "ae0d1d17-c239-443e-ab80-70731b9e15de",
    "name": "Job-2",
    "nodepool": "C0M0",
    "services": [
        {
            "id": "workbench-environment-builder",
            "env": {
                "DEFAULT_MOUNTNAME": "containers",
                "WB_LOCALE": "en-US,en"
            }
        }
    ],
    "spec": {
        "tasks": {
            "default": {
                "cmds": [
                    "echo \"test123\""
                ]
            }
        },
        "version": "3"
    },
    "workspaces": [
        {
            "id": "sas-workbench-5tcc2a32w0wu293go4nu8ldvj",
            "mountName": "replace-me"
        }
    ]
}]`
	oneUpdatedJobJSON = `{"ids":["resource-1"],` +
		`"param":{"application_id":"workbench-environment-builder",` +
		`"dataplane_id":"ae0d1d17-c239-443e-ab80-70731b9e15de",` +
		`"name":"Job-2",` +
		`"nodepool":"C0M0",` +
		`"services":[{"id":"workbench-environment-builder",` +
		`"env":{"DEFAULT_MOUNTNAME":"containers","WB_LOCALE":"en-US,en"}}],` +
		`"spec":{"tasks":{"default":{"cmds":["echo \"test123\""]}},"version":"3"},` +
		`"workspaces":[{"id":"sas-workbench-5tcc2a32w0wu293go4nu8ldvj","mountName":"UPDATED"}]}}`
	expectedInvalidEndpointErr = "failed to get device token" +
		"\n Caused by:" +
		"\n\t* parse \"a\\nb\": net/url: invalid control character in URL"
	expectedInvalidResourceErr = "" +
		"GET ${URL}/postgrest/jobs?id=in.%28resource-1%29 failed" +
		"\n Caused by:" +
		"\n\t* invalid character 'i' looking for beginning of value"
	expectedEditErr = "" +
		"POST ${URL}/postgrest/rpc/job_bulk_update failed" +
		"\n Caused by:" +
		"\n\t* 400 error"
)

func makeOpenFileFuncFailAfter(
	after int,
	err error,
) func(
	afero.Fs,
	string,
	int,
	os.FileMode,
) (afero.File, error) {
	a := after
	return func(
		fs afero.Fs,
		s string,
		i int,
		mode os.FileMode,
	) (afero.File, error) {
		if a < 1 {
			return nil, err
		}
		a--
		return fs.OpenFile(s, i, mode)
	}
}

func makeOpenFuncFailAfter(
	after int,
	err error,
) func(
	afero.Fs,
	string,
) (afero.File, error) {
	a := after
	return func(
		fs afero.Fs,
		s string,
	) (afero.File, error) {
		if a < 1 {
			return nil, err
		}
		a--
		return fs.Open(s)
	}
}

type editInput struct {
	endpoint     string
	resourceType string
	resourceIDs  []string
	fs           afero.Fs
	command      string
}

type editExpected struct {
	err       string
	getCalls  int
	postCalls int
	displays  []string
	warns     []string
}

type editTestCase struct {
	name          string
	input         editInput
	mockEndpoints []stubs.StubHTTPEndpoint
	expected      editExpected
}

func TestEditCommand(t *testing.T) {
	tests := []editTestCase{
		{
			name: "no resources",
			input: editInput{
				resourceType: constants.ImageType,
				resourceIDs:  []string{testResourceID1},
			},
			mockEndpoints: getEditEndpoints(editConfig{
				getURL: postgrest.Endpoint(constants.WorkbenchEndpointImages),
				getConfig: stubs.StubHTTPPartialGetConfig{
					ExpectedQuery:  "controller_id=in.(resource-1)",
					ResponseStatus: http.StatusOK,
					ResponseBody:   "[]",
				},
			}),
			expected: editExpected{
				displays: []string{stubs.ExpectedDeviceAuthDisplay},
				getCalls: 1,
				err:      "no elements found",
			},
		},
		{
			name:  "multiple resources",
			input: editInput{resourceType: constants.ImageType, resourceIDs: []string{testResourceID1}},
			mockEndpoints: getEditEndpoints(editConfig{
				getURL: postgrest.Endpoint(constants.WorkbenchEndpointImages),
				getConfig: stubs.StubHTTPPartialGetConfig{
					ExpectedQuery:  "controller_id=in.(resource-1)",
					ResponseStatus: http.StatusOK,
					ResponseBody:   twoImagesJSON,
				},
			}),
			expected: editExpected{
				displays: []string{stubs.ExpectedDeviceAuthDisplay},
				getCalls: 1, err: "multiple elements found",
			},
		},
		{
			name: "tempfile create fails",
			input: editInput{
				resourceType: constants.ImageType,
				resourceIDs:  []string{testResourceID1},
				fs: stubs.StubFS{
					Fs:           afero.NewMemMapFs(),
					OpenFileFunc: makeOpenFileFuncFailAfter(2, errors.New("open file test error")),
				},
			},
			mockEndpoints: getEditEndpoints(editConfig{
				getURL: postgrest.Endpoint(constants.WorkbenchEndpointImages),
				getConfig: stubs.StubHTTPPartialGetConfig{
					ExpectedQuery:  "controller_id=in.(resource-1)",
					ResponseStatus: http.StatusOK,
					ResponseBody:   oneImageJSON,
				},
			}),
			expected: editExpected{
				displays: []string{stubs.ExpectedDeviceAuthDisplay},
				getCalls: 1, err: "open file test error",
			},
		},
		{
			name: "tempfile write fails",
			input: editInput{
				resourceType: constants.ImageType,
				resourceIDs:  []string{testResourceID1},
				fs: stubs.StubFS{
					Fs:           afero.NewMemMapFs(),
					OpenFileFunc: makeOpenFileFuncFailAfter(3, errors.New("open file test error")),
				},
			},
			mockEndpoints: getEditEndpoints(editConfig{
				getURL: postgrest.Endpoint(constants.WorkbenchEndpointImages),
				getConfig: stubs.StubHTTPPartialGetConfig{
					ExpectedQuery:  "controller_id=in.(resource-1)",
					ResponseStatus: http.StatusOK,
					ResponseBody:   oneImageJSON,
				},
			}),
			expected: editExpected{
				displays: []string{stubs.ExpectedDeviceAuthDisplay},
				getCalls: 1, err: "open file test error",
			},
		},
		{
			name: "read edited file fails",
			input: editInput{
				resourceType: constants.ImageType,
				resourceIDs:  []string{testResourceID1},
				fs: stubs.StubFS{
					Fs:       afero.NewMemMapFs(),
					OpenFunc: makeOpenFuncFailAfter(1, errors.New("open test error")),
				},
				command: osutil.ExecutableScript("testdata", "no-change"),
			},
			mockEndpoints: getEditEndpoints(editConfig{
				getURL: postgrest.Endpoint(constants.WorkbenchEndpointImages),
				getConfig: stubs.StubHTTPPartialGetConfig{
					ExpectedQuery:  "controller_id=in.(resource-1)",
					ResponseStatus: http.StatusOK,
					ResponseBody:   oneImageJSON,
				},
			}),
			expected: editExpected{
				displays: []string{stubs.ExpectedDeviceAuthDisplay},
				warns: []string{
					"failed to save device token: failed to save token cache\n" +
						" Caused by:\n" +
						"\t* failed to load token cache\n" +
						"\t* open test error",
				},
				getCalls: 1,
				err:      "open test error",
			},
		},
		{
			name: "editor not found",
			input: editInput{
				resourceType: constants.ImageType,
				resourceIDs:  []string{testResourceID1},
				command:      "no-such-editor",
			},
			mockEndpoints: getEditEndpoints(editConfig{
				getURL: postgrest.Endpoint(constants.WorkbenchEndpointImages),
				getConfig: stubs.StubHTTPPartialGetConfig{
					ExpectedQuery:  "controller_id=in.(resource-1)",
					ResponseStatus: http.StatusOK,
					ResponseBody:   oneImageJSON,
				},
			}),
			expected: editExpected{
				displays: []string{stubs.ExpectedDeviceAuthDisplay},
				getCalls: 1,
				err:      "exec: \"no-such-editor\": executable file not found in " + osutil.PathEnvVarName(),
			},
		},
		{
			name: "editor invalid",
			input: editInput{
				resourceType: constants.ImageType,
				resourceIDs:  []string{testResourceID1},
				command:      osutil.ExecutableScript("testdata", "invalid"),
			},
			mockEndpoints: getEditEndpoints(editConfig{
				getURL: postgrest.Endpoint(constants.WorkbenchEndpointImages),
				getConfig: stubs.StubHTTPPartialGetConfig{
					ExpectedQuery:  "controller_id=in.(resource-1)",
					ResponseStatus: http.StatusOK,
					ResponseBody:   oneImageJSON,
				},
			}),
			expected: editExpected{
				displays: []string{stubs.ExpectedDeviceAuthDisplay},
				getCalls: 1,
				err: "failed to start editor\n" +
					" Caused by:\n" +
					"\t* fork/exec " + osutil.ExecutableScript("testdata", "invalid") + ": " + osutil.ExecFormatError(),
			},
		},
		{
			name: "editor fails",
			input: editInput{
				resourceType: constants.ImageType,
				resourceIDs:  []string{testResourceID1},
				command:      osutil.ExecutableScript("testdata", "fail"),
			},
			mockEndpoints: getEditEndpoints(editConfig{
				getURL: postgrest.Endpoint(constants.WorkbenchEndpointImages),
				getConfig: stubs.StubHTTPPartialGetConfig{
					ExpectedQuery:  "controller_id=in.(resource-1)",
					ResponseStatus: http.StatusOK,
					ResponseBody:   oneImageJSON,
				},
			}),
			expected: editExpected{
				displays: []string{stubs.ExpectedDeviceAuthDisplay},
				getCalls: 1,
				err: "failed to wait for editor\n" +
					" Caused by:\n" +
					"\t* exit status 1",
			},
		},
		{
			name: "edit image",
			input: editInput{
				resourceType: constants.ImageType,
				resourceIDs:  []string{testResourceID1},
				command:      osutil.ExecutableScript("testdata", "edit"),
			},
			mockEndpoints: getEditEndpoints(editConfig{
				getURL: postgrest.Endpoint(constants.WorkbenchEndpointImages),
				getConfig: stubs.StubHTTPPartialGetConfig{
					ExpectedQuery:  "controller_id=in.(resource-1)",
					ResponseStatus: http.StatusOK,
					ResponseBody:   oneImageJSON,
				},
				postURL: postgrest.Endpoint(constants.WorkbenchEndpointRPCImageBulkUpdate),
				postConfig: stubs.StubHTTPPartialPostConfig{
					ExpectedRequestBody: oneUpdatedImageJSON,
					ResponseStatus:      http.StatusOK,
					ResponseBody:        oneImageJSON,
				},
			}),
			expected: editExpected{
				getCalls:  1,
				postCalls: 1,
				displays:  []string{stubs.ExpectedDeviceAuthDisplay, "resource-1"},
			},
		},
		{
			name: "edit image no change",
			input: editInput{
				resourceType: constants.ImageType,
				resourceIDs:  []string{testResourceID1},
				command:      osutil.ExecutableScript("testdata", "no-change"),
			},
			mockEndpoints: getEditEndpoints(editConfig{
				getURL: postgrest.Endpoint(constants.WorkbenchEndpointImages),
				getConfig: stubs.StubHTTPPartialGetConfig{
					ExpectedQuery:  "controller_id=in.(resource-1)",
					ResponseStatus: http.StatusOK,
					ResponseBody:   oneImageJSON,
				},
				postURL: postgrest.Endpoint(constants.WorkbenchEndpointRPCImageBulkUpdate),
				postConfig: stubs.StubHTTPPartialPostConfig{
					ExpectedRequestBody: oneImageJSON,
					ResponseStatus:      http.StatusOK,
					ResponseBody:        oneImageJSON,
				},
			}),
			expected: editExpected{
				displays: []string{stubs.ExpectedDeviceAuthDisplay},
				getCalls: 1,
				err:      "edit canceled: file unchanged",
			},
		},
		{
			name: "edit image empty file",
			input: editInput{
				resourceType: constants.ImageType,
				resourceIDs:  []string{testResourceID1},
				command:      osutil.ExecutableScript("testdata", "empty"),
			},
			mockEndpoints: getEditEndpoints(editConfig{
				getURL: postgrest.Endpoint(constants.WorkbenchEndpointImages),
				getConfig: stubs.StubHTTPPartialGetConfig{
					ExpectedQuery:  "controller_id=in.(resource-1)",
					ResponseStatus: http.StatusOK,
					ResponseBody:   oneImageJSON,
				},
			}),
			expected: editExpected{
				displays: []string{stubs.ExpectedDeviceAuthDisplay},
				getCalls: 1,
				err:      "edit canceled: file empty",
			},
		},
		{
			name: "edit workbench volumes",
			input: editInput{
				resourceType: constants.WorkbenchType,
				resourceIDs:  []string{testResourceID1},
				command:      osutil.ExecutableScript("testdata", "edit"),
			},
			mockEndpoints: getEditEndpoints(editConfig{
				getURL: postgrest.Endpoint(constants.WorkbenchEndpointWorkbenches),
				getConfig: stubs.StubHTTPPartialGetConfig{
					ExpectedQuery:  "controller_id=in.(resource-1)",
					ResponseStatus: http.StatusOK,
					ResponseBody:   oneWorkbenchJSON,
				},
				postURL: postgrest.Endpoint(constants.WorkbenchEndpointRPCWorkbenchUpdate),
				postConfig: stubs.StubHTTPPartialPostConfig{
					ExpectedRequestBody: oneUpdatedWorkbenchJSON,
					ResponseStatus:      http.StatusOK,
					ResponseBody:        oneWorkbenchUpdateResponseJSON,
				},
			}),
			expected: editExpected{
				getCalls:  1,
				postCalls: 1,
				displays:  []string{stubs.ExpectedDeviceAuthDisplay, "resource-1"},
			},
		},
		{
			name: "edit jobs",
			input: editInput{
				resourceType: constants.JobType,
				resourceIDs:  []string{testResourceID1},
				command:      osutil.ExecutableScript("testdata", "edit"),
			},
			mockEndpoints: getEditEndpoints(editConfig{
				getURL: postgrest.Endpoint(constants.WorkbenchEndpointJobs),
				getConfig: stubs.StubHTTPPartialGetConfig{
					ExpectedQuery:  "id=in.(resource-1)",
					ResponseStatus: http.StatusOK,
					ResponseBody:   oneJobJSON,
				},
				postURL: postgrest.Endpoint(constants.WorkbenchEndpointRPCJobBulkUpdate),
				postConfig: stubs.StubHTTPPartialPostConfig{
					ExpectedRequestBody: oneUpdatedJobJSON,
					ResponseStatus:      http.StatusOK,
					ResponseBody:        oneJobJSON,
				},
			}),
			expected: editExpected{
				getCalls:  1,
				postCalls: 1,
				displays:  []string{stubs.ExpectedDeviceAuthDisplay, "resource-1"},
			},
		},
		{
			name: "invalid endpoint",
			input: editInput{
				endpoint:     "a\nb",
				resourceType: constants.JobType,
				resourceIDs:  []string{testResourceID1},
			},
			expected: editExpected{
				err: expectedInvalidEndpointErr,
			},
		},
		{
			name: "invalid get resource",
			input: editInput{
				resourceType: constants.JobType,
				resourceIDs:  []string{testResourceID1},
			},
			mockEndpoints: getEditEndpoints(editConfig{
				getURL: postgrest.Endpoint(constants.WorkbenchEndpointJobs),
				getConfig: stubs.StubHTTPPartialGetConfig{
					ExpectedQuery:  "id=in.(resource-1)",
					ResponseStatus: http.StatusOK,
					ResponseBody:   `[invalid`,
				},
			}),
			expected: editExpected{
				displays: []string{stubs.ExpectedDeviceAuthDisplay},
				getCalls: 1,
				err:      expectedInvalidResourceErr,
			},
		},
		{
			name: "invalid edit resource",
			input: editInput{
				resourceType: constants.JobType,
				resourceIDs:  []string{testResourceID1},
				command:      osutil.ExecutableScript("testdata", "edit"),
			},
			mockEndpoints: getEditEndpoints(editConfig{
				getURL: postgrest.Endpoint(constants.WorkbenchEndpointJobs),
				getConfig: stubs.StubHTTPPartialGetConfig{
					ExpectedQuery:  "id=in.(resource-1)",
					ResponseStatus: http.StatusOK,
					ResponseBody:   oneJobJSON,
				},
				postURL: postgrest.Endpoint(constants.WorkbenchEndpointRPCJobBulkUpdate),
				postConfig: stubs.StubHTTPPartialPostConfig{
					ExpectedRequestBody: oneUpdatedJobJSON,
					ResponseStatus:      http.StatusBadRequest,
					ResponseBody:        `error`,
				},
			}),
			expected: editExpected{
				displays:  []string{stubs.ExpectedDeviceAuthDisplay},
				getCalls:  1,
				postCalls: 1,
				err:       expectedEditErr,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer viper.Reset()
			srv, hitCounts := stubs.NewStubHTTPServer(t, tt.mockEndpoints)
			defer srv.Close()

			stubUI := &stubs.StubUI{}
			if tt.input.endpoint == "" {
				tt.input.endpoint = srv.URL
			}
			t.Setenv("EDITOR", tt.input.command)
			fs := afero.NewOsFs()
			if tt.input.fs != nil {
				fs = tt.input.fs
			}
			mockCmd := edit.Command(context.Background(), fs, stubUI)
			cmd.AddRootFlags(mockCmd)
			mockCmd.SetArgs(getEditArgs(tt))
			mockCmd.PersistentPreRun = flags.BindToViper
			err := mockCmd.Execute()
			replacements := map[string]string{
				"${PORT}": stubs.Port(t, srv.URL),
				"${HOST}": stubs.Host(t, srv.URL),
				"${URL}":  srv.URL,
			}
			if tt.expected.err != "" {
				expectedErr, actualErr := stubs.Sanitize(tt.expected.err, err.Error(), replacements)
				require.Equal(t, expectedErr, actualErr)
			} else {
				require.NoError(t, err)
			}

			expectedDisplays, actualDisplays := stubs.SanitizeSlice(
				tt.expected.displays,
				stubUI.Displays,
				replacements,
			)
			require.Equal(t,
				expectedDisplays,
				actualDisplays,
				"unexpected display messages: %v",
				stubUI.Displays,
			)
			require.Equal(t, tt.expected.warns, stubUI.Warns, strings.Join(stubUI.Warns, "\n"))
			require.Empty(t, stubUI.Tables)

			require.Equal(
				t,
				tt.expected.getCalls,
				hitCounts.CountPrefix("GET"+constants.PostgrestAPIEndpoint),
				"get calls",
			)
			require.Equal(
				t,
				tt.expected.postCalls,
				hitCounts.CountPrefix("POST"+constants.PostgrestAPIEndpoint),
				"update calls",
			)
		})
	}
}

func TestSupportedResourceTypes(t *testing.T) {
	actual := edit.SupportedResourceTypes()

	require.Equal(t, []string{
		constants.ImageType,
		constants.JobType,
		constants.SnapshotType,
		constants.VolumeType,
		constants.WorkbenchType,
	}, actual)
}

func getEditArgs(
	tt editTestCase,
) []string {
	args := []string{}
	args = append(args, tt.input.resourceType)
	args = append(args, "--"+constants.AuthModeFlag, constants.AuthModeDevice)
	args = append(args, "--"+constants.OrganizationFlag, "org-1")
	if tt.input.endpoint != "" {
		args = append(args, "--"+constants.EndpointFlag, tt.input.endpoint)
	}
	args = append(args, tt.input.resourceIDs...)
	return args
}

type editConfig struct {
	getURL     string
	getConfig  stubs.StubHTTPPartialGetConfig
	postURL    string
	postConfig stubs.StubHTTPPartialPostConfig
}

func getEditEndpoints(
	config editConfig,
) []stubs.StubHTTPEndpoint {
	endpoints := stubs.GetStubAuthEndpoints()
	endpoints = append(endpoints,
		stubs.StubHTTPEndpoint{
			Request: stubs.StubHTTPRequest{
				Method:          http.MethodGet,
				URL:             config.getURL,
				ExpectedHeaders: stubs.StubAuthHeaders,
				ExpectedQuery:   config.getConfig.ExpectedQuery,
			},
			Response: stubs.StubHTTPResponse{
				Status: config.getConfig.ResponseStatus,
				Body:   config.getConfig.ResponseBody,
			},
		},
		stubs.StubHTTPEndpoint{
			Request: stubs.StubHTTPRequest{
				Method:          http.MethodPost,
				URL:             config.postURL,
				ExpectedHeaders: stubs.StubAuthHeaders,
				ExpectedBody:    config.postConfig.ExpectedRequestBody,
			},
			Response: stubs.StubHTTPResponse{
				Status: config.postConfig.ResponseStatus,
				Body:   config.postConfig.ResponseBody,
			},
		},
	)
	return endpoints
}
