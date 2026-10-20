// Copyright © 2026, SAS Institute Inc., Cary, NC, USA.  All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package workbench_test

import (
	"context"
	"net/http"
	"strings"
	"testing"

	"github.com/sassoftware/wbctl/internal/cmd"
	"github.com/sassoftware/wbctl/internal/cmd/create/workbench"
	"github.com/sassoftware/wbctl/internal/constants"
	"github.com/sassoftware/wbctl/internal/postgrest"
	"github.com/sassoftware/wbctl/internal/stubs"
	"github.com/spf13/afero"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
)

const (
	workbenchFileJSON       = `{"name":"wb-file","nodepool":"np-file","state":"RUNNING", "application_id":"vscode"}`
	workbenchFilePrettyJSON = `{
    "application_id": "vscode",
    "dataplane_id": "dp-1",
    "name": "wb-file",
    "nodepool": "np-file",
    "state": "RUNNING"
}`
	workbenchIncompleteFileJSON = `{"name":"wb-file","nodepool":"np-file","state":"RUNNING"}`
	workbenchIncompleteFileErr  = "either --application or at least one --service is required\n" +
		"  Valid Applications: application-1, application-2\n" +
		"  Valid Services: service-1, service-2"
	workbenchNoWorkspaceCreateBody = `{"param":{"application_id":"app-1", "dataplane_id":"dp-1","name":"wb-1",` +
		`"services":[{"id":"svc-a"},{"id":"svc-b"}],"nodepool":"np-1","state":"STOPPED"}}`
	workbenchWithWorkspaceCreateBody = `{"param":{"application_id":"app-1", "dataplane_id":"dp-1","name":"wb-2",` +
		`"services":[{"id":"svc-a","env":{"DEFAULT_MOUNTNAME":"data"}},` +
		`{"id":"svc-b","env":{"DEFAULT_MOUNTNAME":"data"}}],` +
		`"workspaces":[{"id":"ctrl-home","mountName":"sas"},{"id":"ctrl-data",` +
		`"mountName":"data"}],"nodepool":"np-2","state":"STOPPED"}}`
	workbenchWithServicesCreateBody = `{"param":{"dataplane_id":"dp-1","name":"wb-2",` +
		`"services":[{"id":"service-1","env":{"DEFAULT_MOUNTNAME":"data"}},` +
		`{"id":"service-2","env":{"DEFAULT_MOUNTNAME":"data"}}],` +
		`"workspaces":[{"id":"ctrl-home","mountName":"sas"},{"id":"ctrl-data",` +
		`"mountName":"data"}],"nodepool":"np-2","state":"STOPPED"}}`
	workbenchUpdateDryRunOutput = `{
    "ids": [
        "wb-cli-1",
        "wb-cli-2"
    ],
    "workbench": {
        "application_id": "app-1",
        "nodepool": "np-1",
        "services": [
            {
                "id": "svc-a"
            },
            {
                "id": "svc-b"
            }
        ],
        "state": "STOPPED"
    }
}`
	workbenchUpdateRequestBodyMapping = `{"ids":["wb-cli-1","wb-cli-2"],` +
		`"workbench":{"application_id":"app-1","nodepool":"np-1",` +
		`"services":[{"id":"svc-a"},{"id":"svc-b"}],"state":"STOPPED"}}`
	workbenchUpdateRequestBodyWorkspace = `{"ids":["wb-cli-1"],` +
		`"workbench":{"workspaces":[{"id":"ctrl-home","mountName":"sas"}]}}`
	workbenchUpdateRequestBodyIDsOnly = `{"ids":["wb-cli-1"],"workbench":{}}`
	workbenchUpdateFileContent        = `{"id":"wb-file","nodepool":"np-file"}`
	workbenchUpdateFileDryRunOutput   = `{
    "ids": [
        "wb-file"
    ],
    "workbench": {
        "nodepool": "np-file"
    }
}`
	workbenchUpdateMissingIDErr  = "at least one resource ID must be provided for update"
	workbenchBulkUpdateServerErr = "POST ${URL}/postgrest/rpc/workbench_bulk_update failed\n" +
		" Caused by:\n" +
		"\t* 401 unauthorized"
	workbenchUpdateRequestBodyNameLabels  = `{"param":{"id":"wb-cli-1","labels":{"team":"core"},"name":"wb-name"}}`
	workbenchUpdateNameLabelsDryRunOutput = `{
    "ids": [
        "wb-cli-1"
    ],
    "workbench": {
        "labels": {
            "team": "core"
        },
        "name": "wb-name"
    }
}`
	workbenchUpdateWorkspaceDryRunOutput = `{
    "ids": [
        "wb-cli-1"
    ],
    "workbench": {
        "workspaces": [
            {
                "id": "ctrl-home",
                "mountName": "sas"
            }
        ]
    }
}`
	expectedCreateSnapshotInvalidEndpointErr = "failed to get username/password token" +
		"\n Caused by:" +
		"\n\t* parse \"a\\nb\": net/url: invalid control character in URL"
	expectedInvlidDataplaneErr = "GET ${URL}/postgrest/dataplanes?id=in.%28invalid%29 failed" +
		"\n Caused by:" +
		"\n\t* 400 error"
	expectedInvlidWorkspaceErr = "GET ${URL}/postgrest/workspaces?controller_id=in.%28invalid%29 failed" +
		"\n Caused by:" +
		"\n\t* 400 error"
	expectedInvlidApplicationErr = "GET ${URL}/postgrest/dataplane_workbench_applications?id=in.%28invalid%29 failed" +
		"\n Caused by:" +
		"\n\t* 400 error"
	expectedApplicationQuery = "id=in.(app-1)"
	applicationResponseBody  = `[{"id":"app-1","services":["svc-a","svc-b"]}]`
	servicesResponseBody     = `[{"id":"service-1","name":"service 1"},` +
		`{"id":"service-2","name":"service 2"}]`
	expectedCreateImageInvalidFileErr = "unexpected end of JSON input"
)

type createInput struct {
	endpoint    string
	filename    string
	fileContent string
	dataplane   string
	name        string
	labels      []string
	application string
	nodepool    string
	workspaces  []string
	services    []string
	dryRun      bool
}

type createExpected struct {
	displays          []string
	warns             []string
	err               string
	createCalls       int
	dataplaneCalls    int
	applicationsCalls int
	servicesCalls     int
	nodepoolCalls     int
	workspaceCalls    int
}

type createWorkbenchTestCase struct {
	name          string
	input         createInput
	mockEndpoints []stubs.StubHTTPEndpoint
	expected      createExpected
}

func TestCreateWorkbenchCommand(t *testing.T) {
	tests := []createWorkbenchTestCase{
		{
			name: "dry run from file",
			input: createInput{
				filename:    "/input.json",
				fileContent: workbenchFileJSON,
				dryRun:      true,
			},
			mockEndpoints: getCreateWorkbenchEndpoints(workbenchEndpointsConfig{
				dataplaneConfig: stubs.StubHTTPPartialGetConfig{
					ResponseStatus: http.StatusOK,
					ResponseBody:   `[{"id":"dp-1"}]`,
				},
			}),
			expected: createExpected{
				displays:       []string{workbenchFilePrettyJSON},
				warns:          []string{"no volume specified"},
				dataplaneCalls: 1,
			},
		},
		{
			name: "dry run from incomplete file",
			input: createInput{
				filename:    "/input.json",
				fileContent: workbenchIncompleteFileJSON,
				dryRun:      true,
			},
			mockEndpoints: getCreateWorkbenchEndpoints(workbenchEndpointsConfig{
				dataplaneConfig: stubs.StubHTTPPartialGetConfig{
					ResponseStatus: http.StatusOK,
					ResponseBody:   `[{"id":"dp-1"}]`,
				},
				applicationsConfig: stubs.StubHTTPPartialGetConfig{
					ResponseBody:   `[{"id": "application-1"}, {"id": "application-2"}]`,
					ResponseStatus: http.StatusOK,
				},
				servicesConfig: stubs.StubHTTPPartialGetConfig{
					ResponseBody:   `[{"id": "service-1"}, {"id": "service-2"}]`,
					ResponseStatus: http.StatusOK,
				},
			}),
			expected: createExpected{
				err:               workbenchIncompleteFileErr,
				dataplaneCalls:    1,
				applicationsCalls: 1,
				servicesCalls:     1,
			},
		},
		{
			name: "create success without workspaces",
			input: createInput{
				dataplane:   "dp-1",
				name:        "wb-1",
				application: "app-1",
				nodepool:    "np-1",
			},
			mockEndpoints: getCreateWorkbenchEndpoints(workbenchEndpointsConfig{
				dataplaneConfig: stubs.StubHTTPPartialGetConfig{
					ExpectedQuery:  `id=in.(dp-1)`,
					ResponseStatus: http.StatusOK,
					ResponseBody:   `[{"id":"dp-1"}]`,
				},
				applicationsConfig: stubs.StubHTTPPartialGetConfig{
					ExpectedQuery:  expectedApplicationQuery,
					ResponseStatus: http.StatusOK,
					ResponseBody:   applicationResponseBody,
				},
				workbenchesConfig: stubs.StubHTTPPartialPostConfig{
					ExpectedRequestBody: workbenchNoWorkspaceCreateBody,
					ResponseStatus:      http.StatusOK,
					ResponseBody:        `{"id":"wb-1"}`,
				},
			}),
			expected: createExpected{
				displays:          []string{"wb-1"},
				warns:             []string{"no volume specified"},
				createCalls:       1,
				dataplaneCalls:    1,
				applicationsCalls: 1,
			},
		},
		{
			name: "create success with workspaces",
			input: createInput{
				dataplane:   "dp-1",
				name:        "wb-2",
				application: "app-1",
				nodepool:    "np-2",
				workspaces:  []string{"ws-home", "ws-data"},
			},
			mockEndpoints: getCreateWorkbenchEndpoints(workbenchEndpointsConfig{
				dataplaneConfig: stubs.StubHTTPPartialGetConfig{
					ExpectedQuery:  `id=in.(dp-1)`,
					ResponseStatus: http.StatusOK,
					ResponseBody:   `[{"id":"dp-1"}]`,
				},
				applicationsConfig: stubs.StubHTTPPartialGetConfig{
					ExpectedQuery:  expectedApplicationQuery,
					ResponseStatus: http.StatusOK,
					ResponseBody:   applicationResponseBody,
				},
				workbenchesConfig: stubs.StubHTTPPartialPostConfig{
					ExpectedRequestBody: workbenchWithWorkspaceCreateBody,
					ResponseStatus:      http.StatusOK,
					ResponseBody:        `{"id":"wb-2"}`,
				},
			}),
			expected: createExpected{
				displays:          []string{"wb-2"},
				createCalls:       1,
				dataplaneCalls:    1,
				applicationsCalls: 1,
				workspaceCalls:    2,
			},
		},
		{
			name: "create success with services",
			input: createInput{
				dataplane:  "dp-1",
				name:       "wb-2",
				services:   []string{"service-1", "service-2"},
				nodepool:   "np-2",
				workspaces: []string{"ws-home", "ws-data"},
			},
			mockEndpoints: getCreateWorkbenchEndpoints(workbenchEndpointsConfig{
				dataplaneConfig: stubs.StubHTTPPartialGetConfig{
					ExpectedQuery:  `id=in.(dp-1)`,
					ResponseStatus: http.StatusOK,
					ResponseBody:   `[{"id":"dp-1"}]`,
				},
				servicesConfig: stubs.StubHTTPPartialGetConfig{
					ResponseStatus: http.StatusOK,
					ResponseBody:   servicesResponseBody,
				},
				workbenchesConfig: stubs.StubHTTPPartialPostConfig{
					ExpectedRequestBody: workbenchWithServicesCreateBody,
					ResponseStatus:      http.StatusOK,
					ResponseBody:        `{"id":"wb-2"}`,
				},
			}),
			expected: createExpected{
				displays:       []string{"wb-2"},
				createCalls:    1,
				dataplaneCalls: 1,
				servicesCalls:  1,
				workspaceCalls: 2,
			},
		},
		{
			name: "create error",
			input: createInput{
				dataplane:   "dp-1",
				name:        "wb-1",
				application: "app-1",
				nodepool:    "np-1",
			},
			mockEndpoints: getCreateWorkbenchEndpoints(workbenchEndpointsConfig{
				dataplaneConfig: stubs.StubHTTPPartialGetConfig{
					ExpectedQuery:  `id=in.(dp-1)`,
					ResponseStatus: http.StatusOK,
					ResponseBody:   `[{"id":"dp-1"}]`,
				},
				applicationsConfig: stubs.StubHTTPPartialGetConfig{
					ExpectedQuery:  expectedApplicationQuery,
					ResponseStatus: http.StatusOK,
					ResponseBody:   applicationResponseBody,
				},
				workbenchesConfig: stubs.StubHTTPPartialPostConfig{
					ExpectedRequestBody: workbenchNoWorkspaceCreateBody,
					ResponseStatus:      http.StatusUnauthorized,
					ResponseBody:        `unauthorized`,
				},
			}),
			expected: createExpected{
				err: "POST ${URL}/postgrest/rpc/workbench_create failed\n" +
					" Caused by:\n" +
					"\t* 401 unauthorized",
				warns:             []string{"no volume specified"},
				createCalls:       1,
				dataplaneCalls:    1,
				applicationsCalls: 1,
			},
		},
		{
			name: "name required",
			input: createInput{
				application: "app-1",
				nodepool:    "np-1",
			},
			mockEndpoints: getCreateWorkbenchEndpoints(workbenchEndpointsConfig{
				dataplaneConfig: stubs.StubHTTPPartialGetConfig{
					ResponseStatus: http.StatusOK,
					ResponseBody:   `[{"id":"dp-1"}]`,
				},
				applicationsConfig: stubs.StubHTTPPartialGetConfig{
					ExpectedQuery:  expectedApplicationQuery,
					ResponseStatus: http.StatusOK,
					ResponseBody:   applicationResponseBody,
				},
			}),
			expected: createExpected{
				err:               "--name is required",
				applicationsCalls: 1,
				dataplaneCalls:    1,
			},
		},
		{
			name: "application or service required",
			input: createInput{
				name:     "wb-1",
				nodepool: "np-1",
			},
			mockEndpoints: getCreateWorkbenchEndpoints(workbenchEndpointsConfig{
				dataplaneConfig: stubs.StubHTTPPartialGetConfig{
					ResponseStatus: http.StatusOK,
					ResponseBody:   `[{"id":"dp-1"}]`,
				},
				applicationsConfig: stubs.StubHTTPPartialGetConfig{
					ResponseBody:   `[{"id": "application-1"}, {"id": "application-2"}]`,
					ResponseStatus: http.StatusOK,
				},
				servicesConfig: stubs.StubHTTPPartialGetConfig{
					ResponseBody:   `[{"id": "service-1"}, {"id": "service-2"}]`,
					ResponseStatus: http.StatusOK,
				},
			}),
			expected: createExpected{
				err: "either --application or at least one --service is required\n" +
					"  Valid Applications: application-1, application-2\n" +
					"  Valid Services: service-1, service-2",
				dataplaneCalls:    1,
				applicationsCalls: 1,
				servicesCalls:     1,
			},
		},
		{
			name: "nodepool required",
			input: createInput{
				name:        "wb-1",
				application: "app-1",
			},
			mockEndpoints: getCreateWorkbenchEndpoints(workbenchEndpointsConfig{
				dataplaneConfig: stubs.StubHTTPPartialGetConfig{
					ResponseStatus: http.StatusOK,
					ResponseBody:   `[{"id":"dp-1"}]`,
				},
				applicationsConfig: stubs.StubHTTPPartialGetConfig{
					ExpectedQuery:  expectedApplicationQuery,
					ResponseStatus: http.StatusOK,
					ResponseBody:   applicationResponseBody,
				},
				nodepoolsConfig: stubs.StubHTTPPartialGetConfig{
					ResponseBody:   `[{"id": "np-1"}, {"id": "np-2"}]`,
					ResponseStatus: http.StatusOK,
				},
			}),
			expected: createExpected{
				err:               "--nodepool is required\n  Valid Nodepools: np-1, np-2",
				applicationsCalls: 1,
				dataplaneCalls:    1,
				nodepoolCalls:     1,
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
			mockEndpoints: getCreateWorkbenchEndpoints(workbenchEndpointsConfig{
				dataplaneConfig: stubs.StubHTTPPartialGetConfig{
					ExpectedQuery:  `id=in.(invalid)`,
					ResponseStatus: http.StatusBadRequest,
					ResponseBody:   `error`,
				},
			}),
			expected: createExpected{
				err:            expectedInvlidDataplaneErr,
				dataplaneCalls: 1,
			},
		},
		{
			name: "invalid workspace id",
			input: createInput{
				dryRun:     true,
				workspaces: []string{"invalid"},
			},
			mockEndpoints: getCreateWorkbenchEndpoints(workbenchEndpointsConfig{
				workspacesConfig: stubs.StubHTTPPartialGetConfig{
					ExpectedQuery:  `controller_id=in.(invalid)`,
					ResponseStatus: http.StatusBadRequest,
					ResponseBody:   `error`,
				},
			}),
			expected: createExpected{
				workspaceCalls: 1,
				err:            expectedInvlidWorkspaceErr,
			},
		},
		{
			name: "invalid application id",
			input: createInput{
				dryRun:      true,
				application: "invalid",
			},
			mockEndpoints: getCreateWorkbenchEndpoints(workbenchEndpointsConfig{
				applicationsConfig: stubs.StubHTTPPartialGetConfig{
					ExpectedQuery:  `id=in.(invalid)`,
					ResponseStatus: http.StatusBadRequest,
					ResponseBody:   `error`,
				},
			}),
			expected: createExpected{
				applicationsCalls: 1,
				err:               expectedInvlidApplicationErr,
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
			mockCmd := workbench.CreateCommand(context.Background(), fsys, stubUI)
			cmd.AddRootFlags(mockCmd)
			mockCmd.SetArgs(getCreateArgs(tt))
			err := mockCmd.Execute()
			if tt.expected.err != "" {
				expectedErr := strings.ReplaceAll(tt.expected.err, "${URL}", srv.URL)
				require.EqualError(t, err, expectedErr)
			} else {
				require.NoError(t, err)
			}

			require.Equal(t, tt.expected.displays, stubUI.Displays, strings.Join(stubUI.Displays, "\n"))
			require.Equal(t, tt.expected.warns, stubUI.Warns, strings.Join(stubUI.Warns, "\n"))
			require.Equal(
				t,
				tt.expected.createCalls,
				hitCounts.Count("POST"+postgrest.Endpoint(constants.WorkbenchEndpointRPCWorkbenchCreate)),
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
				tt.expected.applicationsCalls,
				hitCounts.Count("GET"+postgrest.Endpoint(constants.WorkbenchEndpointDataplaneWorkbenchApplications)),
				"application calls",
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
	endpoint     string
	filename     string
	fileContent  string
	name         string
	labels       []string
	nodepool     string
	desiredState string
	application  string
	workspaces   []string
	longArgs     []string
	dryRun       bool
	args         []string
}

type updateExpected struct {
	displays          []string
	warns             []string
	err               string
	updateCalls       int
	bulkUpdateCalls   int
	applicationsCalls int
	servicesCalls     int
	nodepoolsCalls    int
	workspaceCalls    int
}

type updateWorkbenchTestCase struct {
	name          string
	input         updateInput
	mockEndpoints []stubs.StubHTTPEndpoint
	expected      updateExpected
}

func TestUpdateWorkbenchComand(t *testing.T) {
	tests := []updateWorkbenchTestCase{
		{
			name: "dry run with CLI ids and mapped fields",
			input: updateInput{
				nodepool:     "np-1",
				desiredState: "STOPPED",
				application:  "app-1",
				dryRun:       true,
				args:         []string{"wb-cli-1", "wb-cli-2"},
			},
			mockEndpoints: getBulkUpdateWorkbenchEndpoints(workbenchEndpointsConfig{
				applicationsConfig: stubs.StubHTTPPartialGetConfig{
					ExpectedQuery:  expectedApplicationQuery,
					ResponseStatus: http.StatusOK,
					ResponseBody:   applicationResponseBody,
				},
			}),
			expected: updateExpected{
				displays:          []string{workbenchUpdateDryRunOutput},
				applicationsCalls: 1,
			},
		},
		{
			name: "dry run long flags map application to services",
			input: updateInput{
				longArgs: []string{
					"--" + constants.DryRunFlag,
					"--" + constants.NodepoolFlag, "np-1",
					"--" + constants.DesiredStateFlag, "STOPPED",
					"--" + constants.ApplicationFlag, "app-1",
				},
				args: []string{"wb-cli-1", "wb-cli-2"},
			},
			mockEndpoints: getBulkUpdateWorkbenchEndpoints(workbenchEndpointsConfig{
				applicationsConfig: stubs.StubHTTPPartialGetConfig{
					ExpectedQuery:  expectedApplicationQuery,
					ResponseStatus: http.StatusOK,
					ResponseBody:   applicationResponseBody,
				},
			}),
			expected: updateExpected{
				displays:          []string{workbenchUpdateDryRunOutput},
				applicationsCalls: 1,
			},
		},
		{
			name: "dry run long flags include name and labels",
			input: updateInput{
				longArgs: []string{
					"--" + constants.DryRunFlag,
					"--" + constants.NameFlag, "wb-name",
					"--" + constants.LabelFlag, "team=core",
				},
				args: []string{"wb-cli-1"},
			},
			mockEndpoints: getBulkUpdateWorkbenchEndpoints(workbenchEndpointsConfig{
				applicationsConfig: stubs.StubHTTPPartialGetConfig{
					ExpectedQuery:  expectedApplicationQuery,
					ResponseStatus: http.StatusOK,
					ResponseBody:   applicationResponseBody,
				},
			}),
			expected: updateExpected{
				displays: []string{workbenchUpdateNameLabelsDryRunOutput},
			},
		},
		{
			name: "dry run long flags include workspace lookup",
			input: updateInput{
				longArgs: []string{
					"--" + constants.DryRunFlag,
					"--" + constants.WorkspaceFlag, "ws-home",
				},
				args: []string{"wb-cli-1"},
			},
			mockEndpoints: getBulkUpdateWorkbenchEndpoints(workbenchEndpointsConfig{
				applicationsConfig: stubs.StubHTTPPartialGetConfig{
					ExpectedQuery:  expectedApplicationQuery,
					ResponseStatus: http.StatusOK,
					ResponseBody:   applicationResponseBody,
				},
			}),
			expected: updateExpected{
				displays:       []string{workbenchUpdateWorkspaceDryRunOutput},
				workspaceCalls: 1,
			},
		},
		{
			name: "dry run long flags include volume lookup",
			input: updateInput{
				longArgs: []string{
					"--" + constants.DryRunFlag,
					"--" + constants.VolumeFlag, "ws-home",
				},
				args: []string{"wb-cli-1"},
			},
			mockEndpoints: getBulkUpdateWorkbenchEndpoints(workbenchEndpointsConfig{
				applicationsConfig: stubs.StubHTTPPartialGetConfig{
					ExpectedQuery:  expectedApplicationQuery,
					ResponseStatus: http.StatusOK,
					ResponseBody:   applicationResponseBody,
				},
			}),
			expected: updateExpected{
				displays:       []string{workbenchUpdateWorkspaceDryRunOutput},
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
			name: "server error",
			input: updateInput{
				args: []string{"wb-cli-1"},
			},
			mockEndpoints: getBulkUpdateWorkbenchEndpoints(workbenchEndpointsConfig{
				workbenchesConfig: stubs.StubHTTPPartialPostConfig{
					ExpectedRequestBody: workbenchUpdateRequestBodyIDsOnly,
					ResponseStatus:      http.StatusUnauthorized,
					ResponseBody:        `unauthorized`,
				},
			}),
			expected: updateExpected{
				err:             workbenchBulkUpdateServerErr,
				bulkUpdateCalls: 1,
			},
		},
		{
			name: "ids fallback from file body id",
			input: updateInput{
				filename:    "/update-workbench.json",
				fileContent: workbenchUpdateFileContent,
				dryRun:      true,
			},
			mockEndpoints: stubs.GetStubAuthEndpoints(),
			expected: updateExpected{
				displays: []string{workbenchUpdateFileDryRunOutput},
			},
		},
		{
			name: "error when no ID source",
			input: updateInput{
				nodepool: "np-1",
			},
			mockEndpoints: stubs.GetStubAuthEndpoints(),
			expected: updateExpected{
				err: workbenchUpdateMissingIDErr,
			},
		},
		{
			name: "ids only with empty body accepted",
			input: updateInput{
				args: []string{"wb-cli-1"},
			},
			mockEndpoints: getBulkUpdateWorkbenchEndpoints(workbenchEndpointsConfig{
				workbenchesConfig: stubs.StubHTTPPartialPostConfig{
					ExpectedRequestBody: workbenchUpdateRequestBodyIDsOnly,
					ResponseStatus:      http.StatusOK,
					ResponseBody:        `[{"id":"wb-cli-1"}]`,
				},
			}),
			expected: updateExpected{
				displays:        []string{"wb-cli-1"},
				bulkUpdateCalls: 1,
			},
		},
		{
			name: "mapping path uses dataplane and services",
			input: updateInput{
				nodepool:     "np-1",
				desiredState: "STOPPED",
				application:  "app-1",
				args:         []string{"wb-cli-1", "wb-cli-2"},
			},
			mockEndpoints: getBulkUpdateWorkbenchEndpoints(workbenchEndpointsConfig{
				applicationsConfig: stubs.StubHTTPPartialGetConfig{
					ExpectedQuery:  expectedApplicationQuery,
					ResponseStatus: http.StatusOK,
					ResponseBody:   applicationResponseBody,
				},
				workbenchesConfig: stubs.StubHTTPPartialPostConfig{
					ExpectedRequestBody: workbenchUpdateRequestBodyMapping,
					ResponseStatus:      http.StatusOK,
					ResponseBody:        `[{"id":"wb-cli-1"},{"id":"wb-cli-2"}]`,
				},
			}),
			expected: updateExpected{
				displays:          []string{"wb-cli-1", "wb-cli-2"},
				bulkUpdateCalls:   1,
				applicationsCalls: 1,
			},
		},
		{
			name: "mapping path uses workspace",
			input: updateInput{
				workspaces: []string{"ws-home"},
				args:       []string{"wb-cli-1"},
			},
			mockEndpoints: getBulkUpdateWorkbenchEndpoints(workbenchEndpointsConfig{
				workbenchesConfig: stubs.StubHTTPPartialPostConfig{
					ExpectedRequestBody: workbenchUpdateRequestBodyWorkspace,
					ResponseStatus:      http.StatusOK,
					ResponseBody:        `[{"id":"wb-cli-1"}]`,
				},
			}),
			expected: updateExpected{
				displays:        []string{"wb-cli-1"},
				bulkUpdateCalls: 1,
				workspaceCalls:  1,
			},
		},
		{
			name: "mapping path uses name and labels",
			input: updateInput{
				name:   "wb-name",
				labels: []string{"team=core"},
				args:   []string{"wb-cli-1"},
			},
			mockEndpoints: getUpdateWorkbenchEndpoints(workbenchEndpointsConfig{
				workbenchesConfig: stubs.StubHTTPPartialPostConfig{
					ExpectedRequestBody: workbenchUpdateRequestBodyNameLabels,
					ResponseStatus:      http.StatusOK,
					ResponseBody:        `{"id":"wb-cli-1"}`,
				},
			}),
			expected: updateExpected{
				displays:    []string{"wb-cli-1"},
				updateCalls: 1,
			},
		},
		{
			name: "invalid endpoint",
			input: updateInput{
				endpoint: "a\nb",
			},
			expected: updateExpected{
				err: expectedCreateSnapshotInvalidEndpointErr,
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
			mockCmd := workbench.UpdateCommand(context.Background(), fsys, stubUI)
			cmd.AddRootFlags(mockCmd)
			mockCmd.SetArgs(getUpdateArgs(tt))
			err := mockCmd.Execute()
			if tt.expected.err != "" {
				expectedErr := strings.ReplaceAll(tt.expected.err, "${URL}", srv.URL)
				require.EqualError(t, err, expectedErr)
			} else {
				require.NoError(t, err)
			}

			require.Equal(t, tt.expected.displays, stubUI.Displays, strings.Join(stubUI.Displays, "\n"))
			require.Equal(t, tt.expected.warns, stubUI.Warns, strings.Join(stubUI.Warns, "\n"))
			require.Equal(
				t,
				tt.expected.updateCalls,
				hitCounts.Count("POST"+postgrest.Endpoint(constants.WorkbenchEndpointRPCWorkbenchUpdate)),
				"update calls",
			)
			require.Equal(
				t,
				tt.expected.bulkUpdateCalls,
				hitCounts.Count("POST"+postgrest.Endpoint(constants.WorkbenchEndpointRPCWorkbenchBulkUpdate)),
				"bulk update calls",
			)
			require.Equal(
				t,
				tt.expected.applicationsCalls,
				hitCounts.Count("GET"+postgrest.Endpoint(constants.WorkbenchEndpointDataplaneWorkbenchApplications)),
				"applications calls",
			)
			require.Equal(
				t,
				tt.expected.servicesCalls,
				hitCounts.Count("GET"+postgrest.Endpoint(constants.WorkbenchEndpointDataplaneWorkbenchServices)),
				"services calls",
			)
			require.Equal(
				t,
				tt.expected.nodepoolsCalls,
				hitCounts.Count("GET"+postgrest.Endpoint(constants.WorkbenchEndpointDataplaneNodepools)),
				"nodepools calls",
			)
			require.Equal(
				t,
				tt.expected.workspaceCalls,
				hitCounts.Count("GET"+postgrest.Endpoint(constants.WorkbenchEndpointWorkspaces)),
				"workspaces calls",
			)
			require.Equal(
				t,
				0,
				hitCounts.Count("GET"+postgrest.Endpoint(constants.WorkbenchEndpointDataplanes)),
				"dataplane calls",
			)
		})
	}
}

func getCreateArgs(
	tt createWorkbenchTestCase,
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
	if tt.input.application != "" {
		args = append(args, "--"+constants.ApplicationFlag, tt.input.application)
	}
	if tt.input.nodepool != "" {
		args = append(args, "--"+constants.NodepoolFlag, tt.input.nodepool)
	}
	for _, label := range tt.input.labels {
		args = append(args, "--"+constants.LabelFlag, label)
	}
	for _, workspaceID := range tt.input.workspaces {
		args = append(args, "--"+constants.WorkspaceFlag, workspaceID)
	}
	for _, serviceID := range tt.input.services {
		args = append(args, "--"+constants.ServiceFlag, serviceID)
	}
	return args
}

func getUpdateArgs(
	tt updateWorkbenchTestCase,
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
	if tt.input.nodepool != "" {
		args = append(args, "--"+constants.NodepoolFlag, tt.input.nodepool)
	}
	if tt.input.desiredState != "" {
		args = append(args, "--"+constants.DesiredStateFlag, tt.input.desiredState)
	}
	if tt.input.application != "" {
		args = append(args, "--"+constants.ApplicationFlag, tt.input.application)
	}
	for _, label := range tt.input.labels {
		args = append(args, "--"+constants.LabelFlag, label)
	}
	for _, workspaceID := range tt.input.workspaces {
		args = append(args, "--"+constants.WorkspaceFlag, workspaceID)
	}
	args = append(args, tt.input.longArgs...)
	args = append(args, tt.input.args...)
	return args
}

type workbenchEndpointsConfig struct {
	dataplaneConfig    stubs.StubHTTPPartialGetConfig
	applicationsConfig stubs.StubHTTPPartialGetConfig
	servicesConfig     stubs.StubHTTPPartialGetConfig
	nodepoolsConfig    stubs.StubHTTPPartialGetConfig
	workspacesConfig   stubs.StubHTTPPartialGetConfig
	workbenchesConfig  stubs.StubHTTPPartialPostConfig
}

func getWorkbenchEndpoints(
	urlUnderTest string,
	config workbenchEndpointsConfig,
) []stubs.StubHTTPEndpoint {
	endpoints := append(stubs.GetStubAuthEndpoints(), stubs.GetStubWorkspaceEndpoints(config.workspacesConfig)...)
	endpoints = append(endpoints, stubs.GetStubDataplaneEndpoints(config.dataplaneConfig)...)
	endpoints = append(endpoints, stubs.GetStubApplicationsEndpoints(config.applicationsConfig)...)
	endpoints = append(endpoints, stubs.GetStubServicesEndpoints(config.servicesConfig)...)
	endpoints = append(endpoints, stubs.GetStubNodepoolsEndpoints(config.nodepoolsConfig)...)
	return append(endpoints,
		[]stubs.StubHTTPEndpoint{
			{
				Request: stubs.StubHTTPRequest{
					Method:          http.MethodPost,
					URL:             urlUnderTest,
					ExpectedHeaders: stubs.StubAuthHeaders,
					ExpectedBody:    config.workbenchesConfig.ExpectedRequestBody,
				},
				Response: stubs.StubHTTPResponse{
					Status: config.workbenchesConfig.ResponseStatus,
					Body:   config.workbenchesConfig.ResponseBody,
				},
			},
		}...)
}

func getCreateWorkbenchEndpoints(config workbenchEndpointsConfig) []stubs.StubHTTPEndpoint {
	return getWorkbenchEndpoints(postgrest.Endpoint(constants.WorkbenchEndpointRPCWorkbenchCreate), config)
}

func getUpdateWorkbenchEndpoints(config workbenchEndpointsConfig) []stubs.StubHTTPEndpoint {
	return getWorkbenchEndpoints(postgrest.Endpoint(constants.WorkbenchEndpointRPCWorkbenchUpdate), config)
}
func getBulkUpdateWorkbenchEndpoints(config workbenchEndpointsConfig) []stubs.StubHTTPEndpoint {
	return getWorkbenchEndpoints(postgrest.Endpoint(constants.WorkbenchEndpointRPCWorkbenchBulkUpdate), config)
}
