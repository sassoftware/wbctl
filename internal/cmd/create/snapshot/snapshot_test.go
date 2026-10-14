// Copyright © 2026, SAS Institute Inc., Cary, NC, USA.  All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package snapshot_test

import (
	"context"
	"net/http"
	"strings"
	"testing"

	"github.com/sassoftware/wbctl/internal/cmd"
	"github.com/sassoftware/wbctl/internal/cmd/create/snapshot"
	"github.com/sassoftware/wbctl/internal/constants"
	"github.com/sassoftware/wbctl/internal/postgrest"
	"github.com/sassoftware/wbctl/internal/stubs"
	"github.com/spf13/afero"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
)

const (
	snapshotFileJSON = `{"id":"","created_by":"","modified_by":"","created_at":null,"modified_at":null,` +
		`"tenant_id":"","dataplane_id":"","name":"snapshot-file",` +
		`"description":"file desc","spec":{"source":{"persistentVolumeClaimName":""}},` +
		`"properties":null,"workspace_source":null,"controller_id":"","state":"",` +
		`"message":"","email":"","username":"",` +
		`"volume_status":null}`
	snapshotFilePrettyJSON = `{
    "dataplane_id": "dp-1",
    "description": "file desc",
    "name": "snapshot-file",
    "spec": {
        "source": {
            "persistentVolumeClaimName": ""
        }
    }
}`
	snapshotInlineCreateBody = `{"param":{"dataplane_id":"dp-1","name":"snapshot-1","description":` +
		`"snapshot desc","spec":{"source":{"persistentVolumeClaimName":"ctrl-1"}}}}`
	snapshotUpdateRequestBodyMapping = `{"ids":["snap-cli-1"],"param":{` +
		`"description":"snapshot desc","name":"snapshot-1",` +
		`"spec":{"source":{"persistentVolumeClaimName":"ctrl-1"}}}}`
	snapshotUpdateRequestBodyIDsOnly = `{"ids":["snap-cli-1"],"param":{}}`
	snapshotUpdateFileContent        = `{"id":"snap-file","created_by":"","modified_by":"",` +
		`"created_at":null,"modified_at":null,` +
		`"tenant_id":"","dataplane_id":"","name":"snapshot-file",` +
		`"description":"file desc","spec":{"source":{"persistentVolumeClaimName":""}},` +
		`"properties":null,"workspace_source":null,"controller_id":"","state":"",` +
		`"message":"","email":"","username":"",` +
		`"volume_status":null}`
	snapshotUpdateFileDryRunOutput = `{
    "ids": [
        "snap-file"
    ],
    "param": {
        "description": "file desc",
        "name": "snapshot-file",
        "spec": {
            "source": {
                "persistentVolumeClaimName": ""
            }
        }
    }
}`
	snapshotUpdateMissingIDErr = "at least one resource ID must be provided for update"
	snapshotUpdateServerErr    = "POST ${URL}/postgrest/rpc/snapshot_bulk_update failed\n" +
		" Caused by:\n" +
		"\t* 401 unauthorized"
	snapshotUpdateLongFlagsDryRunOutput = `{
    "ids": [
        "snap-cli-1"
    ],
    "param": {
        "description": "snapshot desc",
        "name": "snapshot-1",
        "spec": {
            "source": {
                "persistentVolumeClaimName": "ctrl-1"
            }
        }
    }
}`
	expectedCreateSnapshotInvalidEndpointErr = "failed to get username/password token" +
		"\n Caused by:" +
		"\n\t* parse \"a\\nb\": net/url: invalid control character in URL"
	expectedInvlidDataplaneErr = "GET ${URL}/postgrest/dataplanes?id=in.%28invalid%29 failed" +
		"\n Caused by:" +
		"\n\t* 400 error"
	expectedInvalidVolumeErr = "" +
		"failed to get volume invalid: GET ${URL}/postgrest/workspaces?controller_id=in.%28invalid%29 failed" +
		"\n Caused by:" + "\n\t* 400 error"
	expectedCreateImageInvalidFileErr = "unexpected end of JSON input"
)

type createInput struct {
	endpoint    string
	filename    string
	fileContent string
	dataplane   string
	name        string
	description string
	workspace   string
	dryRun      bool
}

type createExpected struct {
	displays       []string
	warns          []string
	err            string
	createCalls    int
	dataplaneCalls int
	workspaceCalls int
}

type createSnapshotTestCase struct {
	name          string
	input         createInput
	mockEndpoints []stubs.StubHTTPEndpoint
	expected      createExpected
}

func TestCreateSnapshotCommand(t *testing.T) {
	tests := []createSnapshotTestCase{
		{
			name: "dry run from file",
			input: createInput{
				filename:    "/input.json",
				fileContent: snapshotFileJSON,
				dryRun:      true,
			},
			mockEndpoints: getCreateSnapshotEndpoints(snapshotEndpointsConfig{
				dataplaneConfig: stubs.StubHTTPPartialGetConfig{
					ResponseBody:   `[{"id":"dp-1"}]`,
					ResponseStatus: http.StatusOK,
				},
			}),
			expected: createExpected{
				displays:       []string{snapshotFilePrettyJSON},
				dataplaneCalls: 1,
			},
		},
		{
			name: "create success",
			input: createInput{
				dataplane:   "dp-1",
				name:        "snapshot-1",
				description: "snapshot desc",
				workspace:   "ws-1",
			},
			mockEndpoints: getCreateSnapshotEndpoints(snapshotEndpointsConfig{
				dataplaneConfig: stubs.StubHTTPPartialGetConfig{
					ExpectedQuery:  `id=in.(dp-1)`,
					ResponseBody:   `[{"id":"dp-1"}]`,
					ResponseStatus: http.StatusOK,
				},
				workspaceConfig: stubs.StubHTTPPartialGetConfig{
					ExpectedQuery:  `controller_id=in.(ws-1)`,
					ResponseBody:   `[{"id":"ws-1","controller_id":"ctrl-1"}]`,
					ResponseStatus: http.StatusOK,
				},
				snapshotConfig: stubs.StubHTTPPartialPostConfig{
					ExpectedRequestBody: snapshotInlineCreateBody,
					ResponseStatus:      http.StatusOK,
					ResponseBody:        `{"id":"snapshot-1"}`,
				},
			}),
			expected: createExpected{
				displays:       []string{`snapshot-1`},
				createCalls:    1,
				dataplaneCalls: 1,
				workspaceCalls: 1,
			},
		},
		{
			name: "create error",
			input: createInput{
				dataplane:   "dp-1",
				name:        "snapshot-1",
				description: "snapshot desc",
				workspace:   "ws-1",
			},
			mockEndpoints: getCreateSnapshotEndpoints(snapshotEndpointsConfig{
				dataplaneConfig: stubs.StubHTTPPartialGetConfig{
					ExpectedQuery:  `id=in.(dp-1)`,
					ResponseBody:   `[{"id":"dp-1"}]`,
					ResponseStatus: http.StatusOK,
				},
				workspaceConfig: stubs.StubHTTPPartialGetConfig{
					ExpectedQuery:  `controller_id=in.(ws-1)`,
					ResponseBody:   `[{"id":"ws-1","controller_id":"ctrl-1"}]`,
					ResponseStatus: http.StatusOK,
				},
				snapshotConfig: stubs.StubHTTPPartialPostConfig{
					ExpectedRequestBody: snapshotInlineCreateBody,
					ResponseStatus:      http.StatusUnauthorized,
					ResponseBody:        `unauthorized`,
				},
			}),
			expected: createExpected{
				err: "POST ${URL}/postgrest/rpc/snapshot_create failed\n" +
					" Caused by:\n" +
					"\t* 401 unauthorized",
				createCalls:    1,
				dataplaneCalls: 1,
				workspaceCalls: 1,
			},
		},
		{
			name: "name required",
			input: createInput{
				workspace: "ws-1",
			},
			mockEndpoints: getCreateSnapshotEndpoints(snapshotEndpointsConfig{
				dataplaneConfig: stubs.StubHTTPPartialGetConfig{
					ResponseBody:   `[{"id":"dp-1"}]`,
					ResponseStatus: http.StatusOK,
				},
				workspaceConfig: stubs.StubHTTPPartialGetConfig{
					ExpectedQuery:  `controller_id=in.(ws-1)`,
					ResponseBody:   `[{"id":"ws-1","controller_id":"ctrl-1"}]`,
					ResponseStatus: http.StatusOK,
				},
			}),
			expected: createExpected{
				err:            "--name is required",
				workspaceCalls: 1,
				dataplaneCalls: 1,
			},
		},
		{
			name: "workspace required",
			input: createInput{
				name: "snapshot-1",
			},
			mockEndpoints: getCreateSnapshotEndpoints(snapshotEndpointsConfig{
				dataplaneConfig: stubs.StubHTTPPartialGetConfig{
					ResponseBody:   `[{"id":"dp-1"}]`,
					ResponseStatus: http.StatusOK,
				},
			}),
			expected: createExpected{
				err:            "--volume is required",
				dataplaneCalls: 1,
			},
		},
		{
			name: "volume not found",
			input: createInput{
				dataplane:   "dp-1",
				name:        "snapshot-1",
				description: "snapshot desc",
				workspace:   "missing",
			},
			mockEndpoints: getCreateSnapshotEndpoints(snapshotEndpointsConfig{
				workspaceConfig: stubs.StubHTTPPartialGetConfig{
					ExpectedQuery:  `controller_id=in.(missing)`,
					ResponseBody:   `[]`,
					ResponseStatus: http.StatusOK,
				},
			}),
			expected: createExpected{
				err:            "failed to get volume missing: no elements found",
				workspaceCalls: 1,
			},
		},
		{
			name: "volume error",
			input: createInput{
				dataplane:   "dp-1",
				name:        "snapshot-1",
				description: "snapshot desc",
				workspace:   "invalid",
			},
			mockEndpoints: getCreateSnapshotEndpoints(snapshotEndpointsConfig{
				workspaceConfig: stubs.StubHTTPPartialGetConfig{
					ExpectedQuery:  `controller_id=in.(invalid)`,
					ResponseBody:   `error`,
					ResponseStatus: http.StatusBadRequest,
				},
			}),
			expected: createExpected{
				err:            expectedInvalidVolumeErr,
				workspaceCalls: 1,
			},
		},
		{
			name: "invalid endpoint",
			input: createInput{
				endpoint: "a\nb",
			},
			expected: createExpected{
				err: expectedCreateSnapshotInvalidEndpointErr,
			},
		},
		{
			name: "invalid dataplane id",
			input: createInput{
				dryRun:    true,
				dataplane: "invalid",
			},
			mockEndpoints: getCreateSnapshotEndpoints(snapshotEndpointsConfig{
				dataplaneConfig: stubs.StubHTTPPartialGetConfig{
					ExpectedQuery:  `id=in.(invalid)`,
					ResponseBody:   `error`,
					ResponseStatus: http.StatusBadRequest,
				},
			}),
			expected: createExpected{
				err:            expectedInvlidDataplaneErr,
				dataplaneCalls: 1,
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
			mockCmd := snapshot.CreateCommand(context.Background(), fsys, stubUI)
			cmd.AddRootFlags(mockCmd)
			mockCmd.SetArgs(getCreateArgs(tt))
			err := mockCmd.Execute()
			if tt.expected.err != "" {
				expectedErr := strings.ReplaceAll(tt.expected.err, "${URL}", srv.URL)
				require.EqualError(t, err, expectedErr, expectedErr)
			} else {
				require.NoError(t, err)
			}

			require.Equal(t, tt.expected.displays, stubUI.Displays, strings.Join(stubUI.Displays, "\n"))
			require.Equal(t, tt.expected.warns, stubUI.Warns, strings.Join(stubUI.Warns, "\n"))
			require.Equal(
				t,
				tt.expected.createCalls,
				hitCounts.Count("POST"+postgrest.Endpoint(constants.WorkbenchEndpointRPCSnapshotCreate)),
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

type updateInput struct {
	endpoint    string
	filename    string
	fileContent string
	name        string
	description string
	workspace   string
	dryRun      bool
	longArgs    []string
	args        []string
}

type updateExpected struct {
	displays       []string
	warns          []string
	err            string
	updateCalls    int
	workspaceCalls int
}

type updateSnapshotTestCase struct {
	name          string
	input         updateInput
	mockEndpoints []stubs.StubHTTPEndpoint
	expected      updateExpected
}

func TestUpdateSnapshotCommand(t *testing.T) {
	tests := []updateSnapshotTestCase{
		{
			name: "dry run from file",
			input: updateInput{
				filename:    "/update-snapshot.json",
				fileContent: snapshotUpdateFileContent,
				dryRun:      true,
			},
			mockEndpoints: stubs.GetStubAuthEndpoints(),
			expected: updateExpected{
				displays: []string{snapshotUpdateFileDryRunOutput},
			},
		},
		{
			name: "dry run long flags include mapped fields",
			input: updateInput{
				longArgs: []string{
					"--" + constants.DryRunFlag,
					"--" + constants.NameFlag, "snapshot-1",
					"--" + constants.DescriptionFlag, "snapshot desc",
					"--" + constants.WorkspaceFlag, "ws-1",
				},
				args: []string{"snap-cli-1"},
			},
			mockEndpoints: getUpdateSnapshotEndpoints(snapshotEndpointsConfig{
				workspaceConfig: stubs.StubHTTPPartialGetConfig{
					ExpectedQuery:  `controller_id=in.(ws-1)`,
					ResponseBody:   `[{"id":"ws-1","controller_id":"ctrl-1"}]`,
					ResponseStatus: http.StatusOK,
				},
			}),
			expected: updateExpected{
				displays:       []string{snapshotUpdateLongFlagsDryRunOutput},
				workspaceCalls: 1,
			},
		},
		{
			name: "update from invalid file",
			input: updateInput{
				filename:    "/input.json",
				fileContent: `[`,
			},
			mockEndpoints: stubs.GetStubAuthEndpoints(),
			expected: updateExpected{
				err: expectedCreateImageInvalidFileErr,
			},
		},
		{
			name: "update success prints IDs only",
			input: updateInput{
				args: []string{"snap-cli-1"},
			},
			mockEndpoints: getUpdateSnapshotEndpoints(snapshotEndpointsConfig{
				snapshotConfig: stubs.StubHTTPPartialPostConfig{
					ExpectedRequestBody: snapshotUpdateRequestBodyIDsOnly,
					ResponseStatus:      http.StatusOK,
					ResponseBody:        `[{"id":"snap-a"},{"id":"snap-b"}]`,
				},
			}),
			expected: updateExpected{
				displays:    []string{"snap-a", "snap-b"},
				updateCalls: 1,
			},
		},
		{
			name: "server error",
			input: updateInput{
				args: []string{"snap-cli-1"},
			},
			mockEndpoints: getUpdateSnapshotEndpoints(snapshotEndpointsConfig{
				snapshotConfig: stubs.StubHTTPPartialPostConfig{
					ExpectedRequestBody: snapshotUpdateRequestBodyIDsOnly,
					ResponseStatus:      http.StatusUnauthorized,
					ResponseBody:        `unauthorized`,
				},
			}),
			expected: updateExpected{
				err:         snapshotUpdateServerErr,
				updateCalls: 1,
			},
		},
		{
			name: "ids fallback from file body id",
			input: updateInput{
				filename:    "/update-snapshot.json",
				fileContent: snapshotUpdateFileContent,
				dryRun:      true,
			},
			mockEndpoints: stubs.GetStubAuthEndpoints(),
			expected: updateExpected{
				displays: []string{snapshotUpdateFileDryRunOutput},
			},
		},
		{
			name: "error when no ID source",
			input: updateInput{
				name: "snapshot-1",
			},
			mockEndpoints: stubs.GetStubAuthEndpoints(),
			expected: updateExpected{
				err: snapshotUpdateMissingIDErr,
			},
		},
		{
			name: "ids only with empty body accepted",
			input: updateInput{
				args: []string{"snap-cli-1"},
			},
			mockEndpoints: getUpdateSnapshotEndpoints(snapshotEndpointsConfig{
				snapshotConfig: stubs.StubHTTPPartialPostConfig{
					ExpectedRequestBody: snapshotUpdateRequestBodyIDsOnly,
					ResponseStatus:      http.StatusOK,
					ResponseBody:        `[{"id":"snap-cli-1"}]`,
				},
			}),
			expected: updateExpected{
				displays:    []string{"snap-cli-1"},
				updateCalls: 1,
			},
		},
		{
			name: "mapping path uses dataplane and workspace spec",
			input: updateInput{
				name:        "snapshot-1",
				description: "snapshot desc",
				workspace:   "ws-1",
				args:        []string{"snap-cli-1"},
			},
			mockEndpoints: getUpdateSnapshotEndpoints(snapshotEndpointsConfig{
				workspaceConfig: stubs.StubHTTPPartialGetConfig{
					ExpectedQuery:  `controller_id=in.(ws-1)`,
					ResponseBody:   `[{"id":"ws-1","controller_id":"ctrl-1"}]`,
					ResponseStatus: http.StatusOK,
				},
				snapshotConfig: stubs.StubHTTPPartialPostConfig{
					ExpectedRequestBody: snapshotUpdateRequestBodyMapping,
					ResponseStatus:      http.StatusOK,
					ResponseBody:        `[{"id":"snap-cli-1"}]`,
				},
			}),
			expected: updateExpected{
				displays:       []string{"snap-cli-1"},
				updateCalls:    1,
				workspaceCalls: 1,
			},
		},
		{
			name: "invalid endpoint",
			input: updateInput{
				endpoint: "a\nb",
			},
			expected: updateExpected{
				err: "failed to get username/password token" +
					"\n Caused by:" +
					"\n\t* parse \"a\\nb\": net/url: invalid control character in URL",
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
			mockCmd := snapshot.UpdateCommand(context.Background(), fsys, stubUI)
			cmd.AddRootFlags(mockCmd)
			mockCmd.SetArgs(getUpdateArgs(tt))
			err := mockCmd.Execute()
			if tt.expected.err != "" {
				expectedErr := strings.ReplaceAll(tt.expected.err, "${URL}", srv.URL)
				require.EqualError(t, err, expectedErr, expectedErr)
			} else {
				require.NoError(t, err)
			}

			require.Equal(t, tt.expected.displays, stubUI.Displays, strings.Join(stubUI.Displays, "\n"))
			require.Equal(t, tt.expected.warns, stubUI.Warns, strings.Join(stubUI.Warns, "\n"))
			require.Equal(
				t,
				tt.expected.updateCalls,
				hitCounts.Count("POST"+postgrest.Endpoint(constants.WorkbenchEndpointRPCSnapshotBulkUpdate)),
				"update calls",
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
	tt createSnapshotTestCase,
) []string {
	args := []string{}
	args = append(args, "--"+constants.AuthModeFlag, constants.AuthModePassword)
	args = append(args, "--"+constants.ClientIDFlag, constants.OAuthClientID)
	args = append(args, "--"+constants.PasswordFlag, "pass")
	args = append(args, "--"+constants.UsernameFlag, "user")
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
	if tt.input.name != "" {
		args = append(args, "--"+constants.NameFlag, tt.input.name)
	}
	if tt.input.description != "" {
		args = append(args, "--"+constants.DescriptionFlag, tt.input.description)
	}
	if tt.input.workspace != "" {
		args = append(args, "--"+constants.WorkspaceFlag, tt.input.workspace)
	}
	return args
}

func getUpdateArgs(
	tt updateSnapshotTestCase,
) []string {
	args := []string{}
	args = append(args, "--"+constants.AuthModeFlag, constants.AuthModePassword)
	args = append(args, "--"+constants.ClientIDFlag, constants.OAuthClientID)
	args = append(args, "--"+constants.PasswordFlag, "pass")
	args = append(args, "--"+constants.UsernameFlag, "user")
	if tt.input.endpoint != "" {
		args = append(args, "--"+constants.EndpointFlag, tt.input.endpoint)
	}
	if tt.input.dryRun {
		args = append(args, "--"+constants.DryRunFlag)
	}
	if tt.input.filename != "" {
		args = append(args, "--"+constants.FilenameFlag, tt.input.filename)
	}
	if tt.input.name != "" {
		args = append(args, "--"+constants.NameFlag, tt.input.name)
	}
	if tt.input.description != "" {
		args = append(args, "--"+constants.DescriptionFlag, tt.input.description)
	}
	if tt.input.workspace != "" {
		args = append(args, "--"+constants.WorkspaceFlag, tt.input.workspace)
	}
	args = append(args, tt.input.longArgs...)
	args = append(args, tt.input.args...)
	return args
}

type snapshotEndpointsConfig struct {
	dataplaneConfig stubs.StubHTTPPartialGetConfig
	workspaceConfig stubs.StubHTTPPartialGetConfig
	snapshotConfig  stubs.StubHTTPPartialPostConfig
}

func getSnapshotEndpoints(
	urlUnderTest string,
	config snapshotEndpointsConfig,
) []stubs.StubHTTPEndpoint {
	endpoints := append(stubs.GetStubAuthEndpoints(), stubs.GetStubWorkspaceEndpoints(config.workspaceConfig)...)
	endpoints = append(endpoints, stubs.GetStubDataplaneEndpoints(config.dataplaneConfig)...)
	return append(endpoints,
		[]stubs.StubHTTPEndpoint{
			{
				Request: stubs.StubHTTPRequest{
					Method:          http.MethodPost,
					URL:             urlUnderTest,
					ExpectedHeaders: stubs.StubAuthHeaders,
					ExpectedBody:    config.snapshotConfig.ExpectedRequestBody,
				},
				Response: stubs.StubHTTPResponse{
					Status: config.snapshotConfig.ResponseStatus,
					Body:   config.snapshotConfig.ResponseBody,
				},
			},
		}...)
}

func getCreateSnapshotEndpoints(config snapshotEndpointsConfig) []stubs.StubHTTPEndpoint {
	return getSnapshotEndpoints(
		postgrest.Endpoint(constants.WorkbenchEndpointRPCSnapshotCreate),
		config,
	)
}

func getUpdateSnapshotEndpoints(config snapshotEndpointsConfig) []stubs.StubHTTPEndpoint {
	return getSnapshotEndpoints(
		postgrest.Endpoint(constants.WorkbenchEndpointRPCSnapshotBulkUpdate),
		config,
	)
}
