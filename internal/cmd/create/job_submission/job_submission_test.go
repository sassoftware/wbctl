// Copyright © 2026, SAS Institute Inc., Cary, NC, USA.  All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package job_submission_test

import (
	"context"
	"net/http"
	"strings"
	"testing"

	"github.com/sassoftware/wbctl/internal/cmd"
	"github.com/sassoftware/wbctl/internal/cmd/create/job_submission"
	"github.com/sassoftware/wbctl/internal/constants"
	"github.com/sassoftware/wbctl/internal/postgrest"
	"github.com/sassoftware/wbctl/internal/stubs"
	"github.com/spf13/afero"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
)

const (
	expectedCreateJobSubmissionDryRunOutput = `{
    "dataplane_id": "dp-1",
    "job_id": "js-file",
    "nodepool": "np-file"
}`
	expectedCreateJobSubmissionFileContent           = `{"dataplane_id":"dp-1","job_id":"js-file","nodepool":"np-file"}`
	expectedCreateJobSubmissionIncompleteFileContent = `{"dataplane_id":"dp-1","nodepool":"np-file"}`
	expectedCreateJobSubmissionIncompleteFileErr     = `--job-id is required`
	expectedCreateJobSubmissionID                    = "js-1"
	expectedCreateJobSubmissionErr                   = "--job-id is required"
	expectedCreateJobSubmissionServerErr             = "POST ${URL}/postgrest/rpc/job_run_create failed\n" +
		" Caused by:\n" +
		"\t* 401 unauthorized"
	expectedCreateJobSubmissionRequestBody = `{"param":{"dataplane_id":"dp-1","job_id":"js-1",` +
		`"env":{"MODE":"nightly"},` +
		`"workspaces":[{"id":"ctrl-home","mountName":"sas"},` +
		`{"id":"ctrl-data","mountName":"data"}],` +
		`"labels":{"team":"analytics"},"nodepool":"np-1"}}`
	expectedCreateJobSubmissionRequestBodyNoExtras = `{"param":{"dataplane_id":"dp-1","job_id":"js-1","nodepool":"np-1"}}`
	expectedCreateJobInvalidEndpointErr            = "failed to get device token" +
		"\n Caused by:" +
		"\n\t* parse \"a\\nb\": net/url: invalid control character in URL"
	expectedInvlidDataplaneErr = "GET ${URL}/postgrest/dataplanes?id=in.%28invalid%29 failed" +
		"\n Caused by:" +
		"\n\t* 400 error"
	expectedInvlidWorkspaceErr = "GET ${URL}/postgrest/workspaces?controller_id=in.%28invalid%29 failed" +
		"\n Caused by:" +
		"\n\t* 400 error"
)

type createInput struct {
	endpoint        string
	filename        string
	fileContent     string
	dataplane       string
	jobID           string
	nodepoolValue   string
	labelValues     []string
	envValues       []string
	workspaceValues []string
	volumeValues    []string
	dryRun          bool
}

type createExpected struct {
	displays       []string
	warns          []string
	err            string
	createCalls    int
	dataplaneCalls int
	workspaceCalls int
}

type createJobSubmissionTestCase struct {
	name          string
	input         createInput
	mockEndpoints []stubs.StubHTTPEndpoint
	expected      createExpected
}

func TestCreateJobSubmissionCommand(t *testing.T) {
	tests := []createJobSubmissionTestCase{
		{
			name: "dry run from file",
			input: createInput{
				filename:    "/input.json",
				fileContent: expectedCreateJobSubmissionFileContent,
				dryRun:      true,
			},
			mockEndpoints: stubs.GetStubAuthEndpoints(),
			expected: createExpected{
				displays: []string{stubs.ExpectedDeviceAuthDisplay, expectedCreateJobSubmissionDryRunOutput},
			},
		},
		{
			name: "dry run from incomplete file",
			input: createInput{
				filename:    "/input.json",
				fileContent: expectedCreateJobSubmissionIncompleteFileContent,
				dryRun:      true,
			},
			mockEndpoints: stubs.GetStubAuthEndpoints(),
			expected: createExpected{
				displays: []string{stubs.ExpectedDeviceAuthDisplay},
				err:      expectedCreateJobSubmissionIncompleteFileErr,
			},
		},
		{
			name: "create success with inline fields",
			input: createInput{
				dataplane:     "dp-1",
				jobID:         "js-1",
				nodepoolValue: "np-1",
				labelValues:   []string{"team=analytics"},
				envValues:     []string{"MODE=nightly"},
				volumeValues:  []string{"ws-home", "ws-data"},
			},
			mockEndpoints: getJobRunEndpoints(jobSubmissionEndpointsConfig{
				dataplaneConfig: stubs.StubHTTPPartialGetConfig{
					ExpectedQuery:  "id=in.(dp-1)",
					ResponseBody:   `[{"id":"dp-1"}]`,
					ResponseStatus: http.StatusOK,
				},
				jobConfig: stubs.StubHTTPPartialPostConfig{
					ExpectedRequestBody: expectedCreateJobSubmissionRequestBody,
					ResponseStatus:      http.StatusOK,
					ResponseBody:        `{"id":"js-1"}`,
				},
			}),
			expected: createExpected{
				displays:       []string{stubs.ExpectedDeviceAuthDisplay, expectedCreateJobSubmissionID},
				createCalls:    1,
				dataplaneCalls: 1,
				workspaceCalls: 2,
			},
		},
		{
			name: "create success with inline fields (workspace flag)",
			input: createInput{
				dataplane:       "dp-1",
				jobID:           "js-1",
				nodepoolValue:   "np-1",
				labelValues:     []string{"team=analytics"},
				envValues:       []string{"MODE=nightly"},
				workspaceValues: []string{"ws-home", "ws-data"},
			},
			mockEndpoints: getJobRunEndpoints(jobSubmissionEndpointsConfig{
				dataplaneConfig: stubs.StubHTTPPartialGetConfig{
					ExpectedQuery:  "id=in.(dp-1)",
					ResponseBody:   `[{"id":"dp-1"}]`,
					ResponseStatus: http.StatusOK,
				},
				jobConfig: stubs.StubHTTPPartialPostConfig{
					ExpectedRequestBody: expectedCreateJobSubmissionRequestBody,
					ResponseStatus:      http.StatusOK,
					ResponseBody:        `{"id":"js-1"}`,
				},
			}),
			expected: createExpected{
				displays:       []string{stubs.ExpectedDeviceAuthDisplay, expectedCreateJobSubmissionID},
				createCalls:    1,
				dataplaneCalls: 1,
				workspaceCalls: 2,
			},
		},
		{
			name: "create error",
			input: createInput{
				dataplane:     "dp-1",
				jobID:         "js-1",
				nodepoolValue: "np-1",
			},
			mockEndpoints: getJobRunEndpoints(jobSubmissionEndpointsConfig{
				dataplaneConfig: stubs.StubHTTPPartialGetConfig{
					ExpectedQuery:  "id=in.(dp-1)",
					ResponseBody:   `[{"id":"dp-1"}]`,
					ResponseStatus: http.StatusOK,
				},
				jobConfig: stubs.StubHTTPPartialPostConfig{
					ExpectedRequestBody: expectedCreateJobSubmissionRequestBodyNoExtras,
					ResponseStatus:      http.StatusUnauthorized,
					ResponseBody:        `unauthorized`,
				},
			}),
			expected: createExpected{
				displays:       []string{stubs.ExpectedDeviceAuthDisplay},
				err:            expectedCreateJobSubmissionServerErr,
				createCalls:    1,
				dataplaneCalls: 1,
			},
		},
		{
			name: "job id required",
			input: createInput{
				dataplane:     "dp-1",
				nodepoolValue: "np-1",
			},
			mockEndpoints: getJobRunEndpoints(jobSubmissionEndpointsConfig{
				dataplaneConfig: stubs.StubHTTPPartialGetConfig{
					ExpectedQuery:  "id=in.(dp-1)",
					ResponseBody:   `[{"id":"dp-1"}]`,
					ResponseStatus: http.StatusOK,
				},
			}),
			expected: createExpected{
				displays:       []string{stubs.ExpectedDeviceAuthDisplay},
				err:            expectedCreateJobSubmissionErr,
				dataplaneCalls: 1,
			},
		},
		{
			name: "invalid endpoint",
			input: createInput{
				endpoint: "a\nb",
			},
			expected: createExpected{
				err: expectedCreateJobInvalidEndpointErr,
			},
		},
		{
			name: "invalid dataplane id",
			input: createInput{
				dryRun:    true,
				dataplane: "invalid",
			},
			mockEndpoints: getJobRunEndpoints(jobSubmissionEndpointsConfig{
				dataplaneConfig: stubs.StubHTTPPartialGetConfig{
					ExpectedQuery:  "id=in.(invalid)",
					ResponseBody:   `error`,
					ResponseStatus: http.StatusBadRequest,
				},
			}),
			expected: createExpected{
				displays:       []string{stubs.ExpectedDeviceAuthDisplay},
				err:            expectedInvlidDataplaneErr,
				dataplaneCalls: 1,
			},
		},
		{
			name: "invalid workspace id",
			input: createInput{
				dryRun:          true,
				dataplane:       "dp-1",
				workspaceValues: []string{"invalid"},
			},
			mockEndpoints: getJobRunEndpoints(jobSubmissionEndpointsConfig{
				dataplaneConfig: stubs.StubHTTPPartialGetConfig{
					ExpectedQuery:  "id=in.(dp-1)",
					ResponseBody:   `[{"id":"dp-1"}]`,
					ResponseStatus: http.StatusOK,
				},
				workspaceConfig: stubs.StubHTTPPartialGetConfig{
					ExpectedQuery:  "controller_id=in.(invalid)",
					ResponseStatus: http.StatusBadRequest,
					ResponseBody:   `error`,
				},
			}),
			expected: createExpected{
				displays:       []string{stubs.ExpectedDeviceAuthDisplay},
				dataplaneCalls: 1,
				workspaceCalls: 1,
				err:            expectedInvlidWorkspaceErr,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer viper.Reset()
			t.Setenv("HOME", t.TempDir())

			srv, hitCounts := stubs.NewStubHTTPServer(t, tt.mockEndpoints)
			defer srv.Close()

			fsys := afero.NewMemMapFs()
			if tt.input.filename != "" {
				require.NoError(t, afero.WriteFile(fsys, tt.input.filename, []byte(tt.input.fileContent), 0o600))
			}

			stubUI := &stubs.StubUI{}
			if tt.input.endpoint == "" {
				tt.input.endpoint = srv.URL
			}
			mockCmd := job_submission.CreateCommand(context.Background(), fsys, stubUI)
			cmd.AddRootFlags(mockCmd)
			mockCmd.SetArgs(getCreateArgs(tt))
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
			require.Equal(
				t,
				tt.expected.createCalls,
				hitCounts.Count("POST"+postgrest.Endpoint(constants.WorkbenchEndpointRPCJobRunCreate)),
				"create calls",
			)
			require.Equal(
				t,
				tt.expected.dataplaneCalls,
				hitCounts.Count("GET"+postgrest.Endpoint(constants.WorkbenchEndpointDataplanes)),
				"dataplane calls",
			)
			require.Equal(
				t,
				tt.expected.workspaceCalls,
				hitCounts.Count("GET"+postgrest.Endpoint(constants.WorkbenchEndpointWorkspaces)),
				"workspace calls",
			)
		})
	}
}

func TestStartJobCommand(t *testing.T) {
	tests := []createJobSubmissionTestCase{
		{
			name: "success with inline fields",
			input: createInput{
				dataplane:     "dp-1",
				jobID:         "js-1",
				nodepoolValue: "np-1",
				labelValues:   []string{"team=analytics"},
				envValues:     []string{"MODE=nightly"},
				volumeValues:  []string{"ws-home", "ws-data"},
			},
			mockEndpoints: getJobRunEndpoints(jobSubmissionEndpointsConfig{
				dataplaneConfig: stubs.StubHTTPPartialGetConfig{
					ExpectedQuery:  "id=in.(dp-1)",
					ResponseBody:   `[{"id":"dp-1"}]`,
					ResponseStatus: http.StatusOK,
				},
				jobConfig: stubs.StubHTTPPartialPostConfig{
					ExpectedRequestBody: expectedCreateJobSubmissionRequestBody,
					ResponseStatus:      http.StatusOK,
					ResponseBody:        `{"id":"js-1"}`,
				},
			}),
			expected: createExpected{
				displays:       []string{stubs.ExpectedDeviceAuthDisplay, expectedCreateJobSubmissionID},
				createCalls:    1,
				dataplaneCalls: 1,
				workspaceCalls: 2,
			},
		},
		{
			name: "success with inline fields (workspace flag)",
			input: createInput{
				dataplane:       "dp-1",
				jobID:           "js-1",
				nodepoolValue:   "np-1",
				labelValues:     []string{"team=analytics"},
				envValues:       []string{"MODE=nightly"},
				workspaceValues: []string{"ws-home", "ws-data"},
			},
			mockEndpoints: getJobRunEndpoints(jobSubmissionEndpointsConfig{
				dataplaneConfig: stubs.StubHTTPPartialGetConfig{
					ExpectedQuery:  "id=in.(dp-1)",
					ResponseBody:   `[{"id":"dp-1"}]`,
					ResponseStatus: http.StatusOK,
				},
				jobConfig: stubs.StubHTTPPartialPostConfig{
					ExpectedRequestBody: expectedCreateJobSubmissionRequestBody,
					ResponseStatus:      http.StatusOK,
					ResponseBody:        `{"id":"js-1"}`,
				},
			}),
			expected: createExpected{
				displays:       []string{stubs.ExpectedDeviceAuthDisplay, expectedCreateJobSubmissionID},
				createCalls:    1,
				dataplaneCalls: 1,
				workspaceCalls: 2,
			},
		},
		{
			name: "error",
			input: createInput{
				dataplane:     "dp-1",
				jobID:         "js-1",
				nodepoolValue: "np-1",
			},
			mockEndpoints: getJobRunEndpoints(jobSubmissionEndpointsConfig{
				dataplaneConfig: stubs.StubHTTPPartialGetConfig{
					ExpectedQuery:  "id=in.(dp-1)",
					ResponseBody:   `[{"id":"dp-1"}]`,
					ResponseStatus: http.StatusOK,
				},
				jobConfig: stubs.StubHTTPPartialPostConfig{
					ExpectedRequestBody: expectedCreateJobSubmissionRequestBodyNoExtras,
					ResponseStatus:      http.StatusUnauthorized,
					ResponseBody:        `unauthorized`,
				},
			}),
			expected: createExpected{
				displays:       []string{stubs.ExpectedDeviceAuthDisplay},
				err:            expectedCreateJobSubmissionServerErr,
				createCalls:    1,
				dataplaneCalls: 1,
			},
		},
		{
			name: "job id required",
			input: createInput{
				dataplane:     "dp-1",
				nodepoolValue: "np-1",
			},
			mockEndpoints: getJobRunEndpoints(jobSubmissionEndpointsConfig{
				dataplaneConfig: stubs.StubHTTPPartialGetConfig{
					ExpectedQuery:  "id=in.(dp-1)",
					ResponseBody:   `[{"id":"dp-1"}]`,
					ResponseStatus: http.StatusOK,
				},
			}),
			expected: createExpected{
				err: "requires at least 1 arg(s), only received 0",
			},
		},
		{
			name: "invalid endpoint",
			input: createInput{
				jobID:    "js-1",
				endpoint: "a\nb",
			},
			expected: createExpected{
				err: expectedCreateJobInvalidEndpointErr,
			},
		},
		{
			name: "invalid dataplane id",
			input: createInput{
				jobID:     "js-1",
				dryRun:    true,
				dataplane: "invalid",
			},
			mockEndpoints: getJobRunEndpoints(jobSubmissionEndpointsConfig{
				dataplaneConfig: stubs.StubHTTPPartialGetConfig{
					ExpectedQuery:  "id=in.(invalid)",
					ResponseBody:   `error`,
					ResponseStatus: http.StatusBadRequest,
				},
			}),
			expected: createExpected{
				displays:       []string{stubs.ExpectedDeviceAuthDisplay},
				err:            expectedInvlidDataplaneErr,
				dataplaneCalls: 1,
			},
		},
		{
			name: "invalid workspace id",
			input: createInput{
				jobID:           "js-1",
				dryRun:          true,
				dataplane:       "dp-1",
				workspaceValues: []string{"invalid"},
			},
			mockEndpoints: getJobRunEndpoints(jobSubmissionEndpointsConfig{
				dataplaneConfig: stubs.StubHTTPPartialGetConfig{
					ExpectedQuery:  "id=in.(dp-1)",
					ResponseBody:   `[{"id":"dp-1"}]`,
					ResponseStatus: http.StatusOK,
				},
				workspaceConfig: stubs.StubHTTPPartialGetConfig{
					ExpectedQuery:  "controller_id=in.(invalid)",
					ResponseStatus: http.StatusBadRequest,
					ResponseBody:   `error`,
				},
			}),
			expected: createExpected{
				displays:       []string{stubs.ExpectedDeviceAuthDisplay},
				dataplaneCalls: 1,
				workspaceCalls: 1,
				err:            expectedInvlidWorkspaceErr,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer viper.Reset()
			t.Setenv("HOME", t.TempDir())

			srv, hitCounts := stubs.NewStubHTTPServer(t, tt.mockEndpoints)
			defer srv.Close()

			fsys := afero.NewMemMapFs()
			if tt.input.filename != "" {
				require.NoError(t, afero.WriteFile(fsys, tt.input.filename, []byte(tt.input.fileContent), 0o600))
			}

			stubUI := &stubs.StubUI{}
			if tt.input.endpoint == "" {
				tt.input.endpoint = srv.URL
			}
			mockCmd := job_submission.StartCommand(context.Background(), fsys, stubUI)
			cmd.AddRootFlags(mockCmd)
			mockCmd.SetArgs(getStartArgs(tt))
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
			require.Equal(
				t,
				tt.expected.createCalls,
				hitCounts.Count("POST"+postgrest.Endpoint(constants.WorkbenchEndpointRPCJobRunCreate)),
				"create calls",
			)
			require.Equal(
				t,
				tt.expected.dataplaneCalls,
				hitCounts.Count("GET"+postgrest.Endpoint(constants.WorkbenchEndpointDataplanes)),
				"dataplane calls",
			)
			require.Equal(
				t,
				tt.expected.workspaceCalls,
				hitCounts.Count("GET"+postgrest.Endpoint(constants.WorkbenchEndpointWorkspaces)),
				"workspace calls",
			)
		})
	}
}

func getCreateArgs(
	tt createJobSubmissionTestCase,
) []string {
	args := getCommonArgs(tt)
	if tt.input.jobID != "" {
		args = append(args, "--"+constants.JobIDFlag, tt.input.jobID)
	}
	return args
}

func getStartArgs(
	tt createJobSubmissionTestCase,
) []string {
	args := getCommonArgs(tt)
	if tt.input.jobID != "" {
		args = append(args, tt.input.jobID)
	}
	return args
}

func getCommonArgs(
	tt createJobSubmissionTestCase,
) []string {
	args := []string{}
	args = append(args, "--"+constants.AuthModeFlag, constants.AuthModeDevice)
	args = append(args, "--"+constants.OrganizationFlag, "org-1")
	if tt.input.endpoint != "" {
		args = append(args, "--"+constants.EndpointFlag, tt.input.endpoint)
	}
	if tt.input.dryRun {
		args = append(args, "--"+constants.DryRunFlag)
	}
	if tt.input.filename != "" {
		args = append(args, "--"+constants.FilenameFlag, tt.input.filename)
	}
	if tt.input.dataplane != "" {
		args = append(args, "--"+constants.DataplaneFlag, tt.input.dataplane)
	}
	if tt.input.nodepoolValue != "" {
		args = append(args, "--"+constants.NodepoolFlag, tt.input.nodepoolValue)
	}
	for _, label := range tt.input.labelValues {
		args = append(args, "--"+constants.LabelFlag, label)
	}
	for _, env := range tt.input.envValues {
		args = append(args, "--"+constants.EnvironmentFlag, env)
	}
	for _, workspaceID := range tt.input.workspaceValues {
		args = append(args, "--"+constants.WorkspaceFlag, workspaceID)
	}
	for _, workspaceID := range tt.input.volumeValues {
		args = append(args, "--"+constants.VolumeFlag, workspaceID)
	}
	return args
}

type jobSubmissionEndpointsConfig struct {
	dataplaneConfig stubs.StubHTTPPartialGetConfig
	workspaceConfig stubs.StubHTTPPartialGetConfig
	jobConfig       stubs.StubHTTPPartialPostConfig
}

func getJobRunEndpoints(
	config jobSubmissionEndpointsConfig,
) []stubs.StubHTTPEndpoint {
	endpoints := append(stubs.GetStubAuthEndpoints(), stubs.GetStubWorkspaceEndpoints(config.workspaceConfig)...)
	endpoints = append(endpoints, stubs.GetStubDataplaneEndpoints(config.dataplaneConfig)...)
	return append(endpoints,
		[]stubs.StubHTTPEndpoint{
			{
				Request: stubs.StubHTTPRequest{
					Method:          http.MethodPost,
					URL:             postgrest.Endpoint(constants.WorkbenchEndpointRPCJobRunCreate),
					ExpectedHeaders: stubs.StubAuthHeaders,
					ExpectedBody:    config.jobConfig.ExpectedRequestBody,
				},
				Response: stubs.StubHTTPResponse{
					Status: config.jobConfig.ResponseStatus,
					Body:   config.jobConfig.ResponseBody,
				},
			},
		}...)
}
