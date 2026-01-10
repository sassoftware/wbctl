// Copyright © 2026, SAS Institute Inc., Cary, NC, USA.  All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package workspace_test

import (
	"context"
	"net/http"
	"strings"
	"testing"

	"github.com/sas-institute-rnd-saas/wbctl/internal/cmd"
	"github.com/sas-institute-rnd-saas/wbctl/internal/cmd/create/workspace"
	"github.com/sas-institute-rnd-saas/wbctl/internal/constants"
	"github.com/sas-institute-rnd-saas/wbctl/internal/postgrest"
	"github.com/sas-institute-rnd-saas/wbctl/internal/stubs"
	"github.com/spf13/afero"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
)

const (
	workspaceFileJSON = `{"dataplane_id":"dp-1","name":"ws-1",` +
		`"storage_type":"PersistentVolumeClaim",` +
		`"storage_spec":{"accessModes":["ReadWriteOnce"],` +
		`"resources":{"requests":{"storage":"10Gi"}},` +
		`"storageClassName":"standard"},"properties":{"HOME_DIRECTORY":"TRUE"}}`
	workspaceFilePrettyJSON = `{
    "dataplane_id": "dp-1",
    "name": "ws-1",
    "properties": {
        "HOME_DIRECTORY": "TRUE"
    },
    "storage_spec": {
        "accessModes": [
            "ReadWriteOnce"
        ],
        "resources": {
            "requests": {
                "storage": "10Gi"
            }
        },
        "storageClassName": "standard"
    },
    "storage_type": "PersistentVolumeClaim"
}`
	workspaceStorageCreateBody = `{"param":{"dataplane_id":"dp-1","name":"ws-1",` +
		`"storage_type":"PersistentVolumeClaim",` +
		`"storage_spec":{"accessModes":["ReadWriteOnce"],` +
		`"resources":{"requests":{"storage":"10Gi"}},` +
		`"storageClassName":"standard"},"properties":{"HOME_DIRECTORY":"TRUE"},` +
		`"pvc_status":{"status":{}}}}`
	workspaceSnapshotCreateBody = `{"param":{"dataplane_id":"dp-1","name":"ws-2","pvc_status":{"status":{}},` +
		`"storage_spec":{"accessModes":["ReadWriteOnce"],"resources":{"requests":{"storage":"1Gi"}},` +
		`"storageClassName":"snap-sc","dataSource":{"apiGroup":"snapshot.storage.k8s.io","kind":"VolumeSnapshot"` +
		`,"name":"ctrl-snap"}},"storage_type":"PersistentVolumeClaim"}}`
	workspaceSnapshotLookupBody = `[{"id":"snap-1","created_by":"","modified_by":"","created_at":null,` +
		`"modified_at":null,"tenant_id":"","dataplane_id":"","name":"",` +
		`"description":"","spec":null,"properties":null,` +
		`"workspace_source":{"storage_spec":{"accessModes":["ReadWriteOnce"],` +
		`"resources":{},"storageClassName":"snap-sc"},"pvc_status":{"status":{}}},` +
		`"controller_id":"ctrl-snap","state":"","message":"","email":"",` +
		`"username":"","volume_status":{"status":{"readyToUse":true,` +
		`"restoreSize":"1Gi"}}}]`
	workspaceUpdateDryRunOutput = `{
    "ids": [
        "ws-cli-1"
    ],
    "workspace": {
        "initializers": [
            {
                "id": "init-1"
            }
        ],
        "properties": {
            "HOME_DIRECTORY": "TRUE"
        },
        "storage_spec": {
            "resources": {
                "requests": {
                    "storage": "10Gi"
                }
            }
        }
    }
}`
	workspaceUpdateRequestBodyIDsOnly = `{"ids":["ws-cli-1"],"workspace":{"pvc_status":{"status":{}}}}`
	workspaceUpdateFileContent        = `{"id":"ws-file","description":"from file"}`
	workspaceUpdateFileDryRunOutput   = `{
    "ids": [
        "ws-file"
    ],
    "workspace": {
        "description": "from file"
    }
}`
	workspaceUpdateMissingIDErr  = "at least one resource ID must be provided for update"
	workspaceBulkUpdateServerErr = "POST ${URL}/postgrest/rpc/workspace_bulk_update failed\n" +
		" Caused by:\n" +
		"\t* 401 unauthorized"
	workspaceUpdateRequestBodyNameDesc = `{"param":{"description":"ws-desc","id":"ws-cli-1",` +
		`"name":"ws-name","pvc_status":{"status":{}}}}`
	workspaceUpdateNameDescDryRunOutput = `{
    "ids": [
        "ws-cli-1"
    ],
    "workspace": {
        "description": "ws-desc",
        "name": "ws-name"
    }
}`
	expectedCreateSnapshotInvalidEndpointErr = "failed to get username/password token" +
		"\n Caused by:" +
		"\n\t* parse \"a\\nb\": net/url: invalid control character in URL"
	expectedInvlidDataplaneErr = "GET ${URL}/postgrest/dataplanes?id=in.%28invalid%29 failed" +
		"\n Caused by:" +
		"\n\t* 400 error"
	expectedInvalidInitalizersErr = "" +
		"GET ${URL}/postgrest/dataplane_workspace_services?id=in.%28invalid%29 failed" +
		"\n Caused by:" +
		"\n\t* 400 error"
	expectedInvalidSizeErr = "" +
		"quantities must match the regular expression '^([+-]?[0-9.]+)([eEinumkKMGTP]*[-+]?[0-9]*)$'"
	expectedInvalidStorageClassErr = "" +
		"GET ${URL}/postgrest/dataplane_storageclasses?id=in.%28invalid%29 failed" +
		"\n Caused by:" +
		"\n\t* 400 error"
	expectedInvalidSnapshotErr = "" +
		"GET ${URL}/postgrest/snapshots?controller_id=in.%28snap-1%29 failed" +
		"\n Caused by:" +
		"\n\t* invalid character 'i' looking for beginning of value"
)

type createInput struct {
	endpoint         string
	filename         string
	fileContent      string
	dataplane        string
	name             string
	description      string
	storageClassName string
	size             string
	initializers     []string
	homeDir          bool
	snapshot         string
	dryRun           bool
}

type createExpected struct {
	displays          []string
	warns             []string
	err               string
	createCalls       int
	dataplaneCalls    int
	storageclassCalls int
	snapshotCalls     int
}

type createWorkspaceTestCase struct {
	name          string
	input         createInput
	mockEndpoints []stubs.StubHTTPEndpoint
	expected      createExpected
}

func TestCreateWorkspaceCommand(t *testing.T) {
	tests := []createWorkspaceTestCase{
		{
			name: "dry run from file",
			input: createInput{
				filename:    "/input.json",
				fileContent: workspaceFileJSON,
				dryRun:      true,
			},
			mockEndpoints: getCreateWorkspaceEndpoints(workspaceEndpointsConfig{
				dataplaneConfig: stubs.StubHTTPPartialGetConfig{
					ResponseBody:   `[{"id":"dp-1"}]`,
					ResponseStatus: http.StatusOK,
				},
			}),
			expected: createExpected{
				displays:       []string{workspaceFilePrettyJSON},
				dataplaneCalls: 1,
			},
		},
		{
			name: "create success with storage class",
			input: createInput{
				dataplane:        "dp-1",
				name:             "ws-1",
				storageClassName: "standard",
				size:             "10Gi",
				homeDir:          true,
			},
			mockEndpoints: getCreateWorkspaceEndpoints(workspaceEndpointsConfig{
				dataplaneConfig: stubs.StubHTTPPartialGetConfig{
					ExpectedQuery:  `id=in.(dp-1)`,
					ResponseBody:   `[{"id":"dp-1"}]`,
					ResponseStatus: http.StatusOK,
				},
				storageClassesConfigs: []stubs.StubHTTPPartialGetConfig{{
					ExpectedQuery:  `id=in.(standard)`,
					ResponseStatus: http.StatusOK,
					ResponseBody:   `[{"id":"standard","access_mode":"ReadWriteOnce"}]`,
				}},
				workspaceConfig: stubs.StubHTTPPartialPostConfig{
					ExpectedRequestBody: workspaceStorageCreateBody,
					ResponseStatus:      http.StatusOK,
					ResponseBody:        `{"id":"ws-1"}`,
				},
			}),
			expected: createExpected{
				displays:          []string{"ws-1"},
				createCalls:       1,
				dataplaneCalls:    1,
				storageclassCalls: 1,
			},
		},
		{
			name: "create success with default storage class",
			input: createInput{
				dataplane: "dp-1",
				name:      "ws-1",
				size:      "10Gi",
				homeDir:   true,
			},
			mockEndpoints: getCreateWorkspaceEndpoints(workspaceEndpointsConfig{
				dataplaneConfig: stubs.StubHTTPPartialGetConfig{
					ExpectedQuery:  `id=in.(dp-1)`,
					ResponseBody:   `[{"id":"dp-1"}]`,
					ResponseStatus: http.StatusOK,
				},
				storageClassesConfigs: []stubs.StubHTTPPartialGetConfig{
					{
						ResponseStatus: http.StatusOK,
						ResponseBody:   `[{"id":"standard","access_mode":"ReadWriteOnce", "properties":{"default": true}}]`,
					},
					{
						ExpectedQuery:  `id=in.(standard)`,
						ResponseStatus: http.StatusOK,
						ResponseBody:   `[{"id":"standard","access_mode":"ReadWriteOnce", "properties":{"default": true}}]`,
					},
				},
				workspaceConfig: stubs.StubHTTPPartialPostConfig{
					ExpectedRequestBody: workspaceStorageCreateBody,
					ResponseStatus:      http.StatusOK,
					ResponseBody:        `{"id":"ws-1"}`,
				},
			}),
			expected: createExpected{
				displays:          []string{"ws-1"},
				warns:             []string{"defaulting to 'standard' storage class"},
				createCalls:       1,
				dataplaneCalls:    1,
				storageclassCalls: 2,
			},
		},
		{
			name: "create success with snapshot",
			input: createInput{
				dataplane: "dp-1",
				name:      "ws-2",
				snapshot:  "snap-1",
			},
			mockEndpoints: getCreateWorkspaceEndpoints(workspaceEndpointsConfig{
				dataplaneConfig: stubs.StubHTTPPartialGetConfig{
					ExpectedQuery:  `id=in.(dp-1)`,
					ResponseBody:   `[{"id":"dp-1"}]`,
					ResponseStatus: http.StatusOK,
				},
				snapshotsConfig: stubs.StubHTTPPartialGetConfig{
					ExpectedQuery:  `controller_id=in.(snap-1)`,
					ResponseStatus: http.StatusOK,
					ResponseBody:   workspaceSnapshotLookupBody,
				},
				workspaceConfig: stubs.StubHTTPPartialPostConfig{
					ExpectedRequestBody: workspaceSnapshotCreateBody,
					ResponseStatus:      http.StatusOK,
					ResponseBody:        `{"id":"ws-2"}`,
				},
			}),
			expected: createExpected{
				displays:       []string{"ws-2"},
				createCalls:    1,
				dataplaneCalls: 1,
				snapshotCalls:  1,
			},
		},
		{
			name: "create error",
			input: createInput{
				dataplane:        "dp-1",
				name:             "ws-1",
				storageClassName: "standard",
				size:             "10Gi",
				homeDir:          true,
			},
			mockEndpoints: getCreateWorkspaceEndpoints(workspaceEndpointsConfig{
				dataplaneConfig: stubs.StubHTTPPartialGetConfig{
					ExpectedQuery:  `id=in.(dp-1)`,
					ResponseBody:   `[{"id":"dp-1"}]`,
					ResponseStatus: http.StatusOK,
				},
				storageClassesConfigs: []stubs.StubHTTPPartialGetConfig{{
					ExpectedQuery:  `id=in.(standard)`,
					ResponseStatus: http.StatusOK,
					ResponseBody:   `[{"id":"standard","access_mode":"ReadWriteOnce"}]`,
				}},
				workspaceConfig: stubs.StubHTTPPartialPostConfig{
					ExpectedRequestBody: workspaceStorageCreateBody,
					ResponseStatus:      http.StatusUnauthorized,
					ResponseBody:        `unauthorized`,
				},
			}),
			expected: createExpected{
				err: "POST ${URL}/postgrest/rpc/workspace_create failed\n" +
					" Caused by:\n" +
					"\t* 401 unauthorized",
				createCalls:       1,
				dataplaneCalls:    1,
				storageclassCalls: 1,
			},
		},
		{
			name: "name required",
			input: createInput{
				storageClassName: "standard",
				size:             "10Gi",
			},
			mockEndpoints: stubs.GetStubAuthEndpoints(),
			expected: createExpected{
				err: "--name is required",
			},
		},
		{
			name: "snapshot or storageclass required",
			input: createInput{
				name: "ws-1",
				size: "10Gi",
			},
			mockEndpoints: stubs.GetStubAuthEndpoints(),
			expected: createExpected{
				err: "one, and only one, of --snapshot or --storageclass is required",
			},
		},
		{
			name: "size required when storageclass",
			input: createInput{
				name:             "ws-1",
				storageClassName: "standard",
			},
			mockEndpoints: stubs.GetStubAuthEndpoints(),
			expected: createExpected{
				err: "--size required when --storageclass specified",
			},
		},
		{
			name: "storageclass required when size, snapshot ignored",
			input: createInput{
				name:     "ws-1",
				size:     "standard",
				snapshot: "snap-1",
			},
			mockEndpoints: stubs.GetStubAuthEndpoints(),
			expected: createExpected{
				err: "--storageclass required when --size specified",
			},
		},
		{
			name: "storage class not found",
			input: createInput{
				dataplane:        "dp-1",
				name:             "ws-3",
				storageClassName: "invalid-sc",
				size:             "10Gi",
			},
			mockEndpoints: getCreateWorkspaceEndpoints(workspaceEndpointsConfig{
				storageClassesConfigs: []stubs.StubHTTPPartialGetConfig{{
					ExpectedQuery:  `id=in.(invalid-sc)`,
					ResponseBody:   `[]`,
					ResponseStatus: http.StatusOK,
				}},
			}),
			expected: createExpected{
				err:               "storage class not found: invalid-sc: no elements found",
				storageclassCalls: 1,
			},
		},
		{
			name: "snapshot not found",
			input: createInput{
				dataplane: "dp-1",
				name:      "ws-4",
				snapshot:  "invalid-snap",
			},
			mockEndpoints: getCreateWorkspaceEndpoints(workspaceEndpointsConfig{
				snapshotsConfig: stubs.StubHTTPPartialGetConfig{
					ExpectedQuery:  `controller_id=in.(invalid-snap)`,
					ResponseBody:   `[]`,
					ResponseStatus: http.StatusOK,
				},
			}),
			expected: createExpected{
				err:           "failed to add snapshot invalid-snap: no elements found",
				snapshotCalls: 1,
			},
		},
		{
			name: "snapshot not ready",
			input: createInput{
				dataplane: "dp-1",
				name:      "ws-5",
				snapshot:  "snap-not-ready",
			},
			mockEndpoints: getCreateWorkspaceEndpoints(workspaceEndpointsConfig{
				snapshotsConfig: stubs.StubHTTPPartialGetConfig{
					ExpectedQuery: `controller_id=in.(snap-not-ready)`,
					ResponseBody: `[{"id":"snap-not-ready","controller_id":"","volume_status":{"status":` +
						`{"readyToUse":false}}}]`,
					ResponseStatus: http.StatusOK,
				},
			}),
			expected: createExpected{
				err: "failed to add snapshot snap-not-ready: snapshot not ready: volume status " +
					"missing or not ready",
				snapshotCalls: 1,
			},
		},
		{
			name: "snapshot no restore size",
			input: createInput{
				dataplane: "dp-1",
				name:      "ws-6",
				snapshot:  "snap-no-size",
			},
			mockEndpoints: getCreateWorkspaceEndpoints(workspaceEndpointsConfig{
				snapshotsConfig: stubs.StubHTTPPartialGetConfig{
					ExpectedQuery: `controller_id=in.(snap-no-size)`,
					ResponseBody: `[{"id":"snap-no-size","controller_id":"ctrl-snap","workspace_source":{` +
						`"storage_spec":{"accessModes":["ReadWriteOnce"],"storageClassName":"snap-sc"}},` +
						`"volume_status":{"status":{"readyToUse":true}}}]`,
					ResponseStatus: http.StatusOK,
				},
			}),
			expected: createExpected{
				err: "failed to add snapshot snap-no-size: snapshot not ready: restore size " +
					"not available",
				snapshotCalls: 1,
			},
		},
		{
			name: "snapshot missing volume source",
			input: createInput{
				dataplane: "dp-1",
				name:      "ws-7",
				snapshot:  "snap-1",
			},
			mockEndpoints: getCreateWorkspaceEndpoints(workspaceEndpointsConfig{
				snapshotsConfig: stubs.StubHTTPPartialGetConfig{
					ExpectedQuery: `controller_id=in.(snap-1)`,
					ResponseBody: `[{"id":"snap-1","controller_id":"ctrl-snap","volume_status":` +
						`{"status":{"readyToUse":true}}}]`,
					ResponseStatus: http.StatusOK,
				},
			}),
			expected: createExpected{
				err: "failed to add snapshot snap-1: snapshot not ready: source or storage " +
					"spec missing",
				snapshotCalls: 1,
			},
		},
		{
			name: "snapshot missing access mode",
			input: createInput{
				dataplane: "dp-1",
				name:      "ws-8",
				snapshot:  "snap-1",
			},
			mockEndpoints: getCreateWorkspaceEndpoints(workspaceEndpointsConfig{
				snapshotsConfig: stubs.StubHTTPPartialGetConfig{
					ExpectedQuery: `controller_id=in.(snap-1)`,
					ResponseBody: `[{"id":"snap-1","controller_id":"ctrl-snap","workspace_source":{"storage_spec":` +
						`{"accessModes":[],"storageClassName":"snap-sc"}},"volume_status":{"status":` +
						`{"readyToUse":true,"restoreSize":"1Gi"}}}]`,
					ResponseStatus: http.StatusOK,
				},
			}),
			expected: createExpected{
				err: "failed to add snapshot snap-1: invalid snapshot: access mode not " +
					"specified",
				snapshotCalls: 1,
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
				name:      "ws-9",
				dataplane: "invalid",
				snapshot:  "snap-1",
			},
			mockEndpoints: getCreateWorkspaceEndpoints(workspaceEndpointsConfig{
				dataplaneConfig: stubs.StubHTTPPartialGetConfig{
					ExpectedQuery:  `id=in.(invalid)`,
					ResponseStatus: http.StatusBadRequest,
					ResponseBody:   `error`,
				},
				snapshotsConfig: stubs.StubHTTPPartialGetConfig{
					ExpectedQuery:  `controller_id=in.(snap-1)`,
					ResponseBody:   workspaceSnapshotLookupBody,
					ResponseStatus: http.StatusOK,
				},
			}),
			expected: createExpected{
				err:            expectedInvlidDataplaneErr,
				dataplaneCalls: 1,
				snapshotCalls:  1,
			},
		},
		{
			name: "invalid storagesize",
			input: createInput{
				dataplane:        "dp-1",
				name:             "ws-1",
				storageClassName: "standard",
				size:             "invalid size",
			},
			mockEndpoints: getCreateWorkspaceEndpoints(workspaceEndpointsConfig{
				storageClassesConfigs: []stubs.StubHTTPPartialGetConfig{{
					ExpectedQuery:  `id=in.(standard)`,
					ResponseStatus: http.StatusOK,
					ResponseBody:   `[{"id":"standard","access_mode":"ReadWriteOnce"}]`,
				}},
			}),
			expected: createExpected{
				err:               expectedInvalidSizeErr,
				storageclassCalls: 1,
			},
		},
		{
			name: "invalid storage class",
			input: createInput{
				dataplane:        "dp-1",
				name:             "ws-1",
				storageClassName: "invalid",
				size:             "1Gi",
			},
			mockEndpoints: getCreateWorkspaceEndpoints(workspaceEndpointsConfig{
				storageClassesConfigs: []stubs.StubHTTPPartialGetConfig{{
					ExpectedQuery:  `id=in.(invalid)`,
					ResponseStatus: http.StatusBadRequest,
					ResponseBody:   `error`,
				}},
			}),
			expected: createExpected{
				err:               expectedInvalidStorageClassErr,
				storageclassCalls: 1,
			},
		},
		{
			name: "invalid snapshot",
			input: createInput{
				dataplane: "dp-1",
				name:      "ws-1",
				snapshot:  "snap-1",
			},
			mockEndpoints: getCreateWorkspaceEndpoints(workspaceEndpointsConfig{
				snapshotsConfig: stubs.StubHTTPPartialGetConfig{
					ExpectedQuery:  `controller_id=in.(snap-1)`,
					ResponseStatus: http.StatusOK,
					ResponseBody:   `[invalid`,
				},
			}),
			expected: createExpected{
				err:           expectedInvalidSnapshotErr,
				snapshotCalls: 1,
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
			mockCmd := workspace.CreateCommand(context.Background(), fsys, stubUI)
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
				hitCounts.Count("POST"+postgrest.Endpoint(constants.WorkbenchEndpointRPCWorkspaceCreate)),
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
				tt.expected.storageclassCalls,
				hitCounts.Count("GET"+postgrest.Endpoint(constants.WorkbenchEndpointDataplaneStorageclasses)),
				"storageclass calls",
			)
			require.Equal(
				t,
				tt.expected.snapshotCalls,
				hitCounts.Count("GET"+postgrest.Endpoint(constants.WorkbenchEndpointSnapshots)),
				"workspace calls",
			)
		})
	}
}

type updateInput struct {
	endpoint     string
	filename     string
	fileContent  string
	nameValue    string
	description  string
	sizeValue    string
	initializers []string
	homeDir      bool
	dryRun       bool
	longArgs     []string
	args         []string
}

type updateExpected struct {
	displays          []string
	warns             []string
	err               string
	singleUpdateCalls int
	bulkUpdateCalls   int
	initializerCalls  int
}

type updateWorkspaceTestCase struct {
	name          string
	input         updateInput
	mockEndpoints []stubs.StubHTTPEndpoint
	expected      updateExpected
}

func TestUpdateWorkspaceCommand(t *testing.T) {
	tests := []updateWorkspaceTestCase{
		{
			name: "dry run with CLI ids and mapped fields",
			input: updateInput{
				dryRun:       true,
				sizeValue:    "10Gi",
				initializers: []string{"init-1"},
				homeDir:      true,
				args:         []string{"ws-cli-1"},
			},
			mockEndpoints: append(stubs.GetStubAuthEndpoints(), getInitializerEndpoints(stubs.StubHTTPPartialGetConfig{
				ExpectedQuery:  `id=in.(init-1)`,
				ResponseStatus: http.StatusOK,
				ResponseBody:   `[{"id":"init-1","name":"init"}]`,
			})...),
			expected: updateExpected{
				displays:         []string{workspaceUpdateDryRunOutput},
				initializerCalls: 1,
			},
		},
		{
			name: "dry run long flags include storage branch mapping",
			input: updateInput{
				longArgs: []string{
					"--" + constants.DryRunFlag,
					"--" + constants.SizeFlag, "10Gi",
					"--" + constants.InitializerFlag, "init-1",
					"--" + constants.HomeDirFlag,
				},
				args: []string{"ws-cli-1"},
			},
			mockEndpoints: append(stubs.GetStubAuthEndpoints(), getInitializerEndpoints(stubs.StubHTTPPartialGetConfig{
				ExpectedQuery:  `id=in.(init-1)`,
				ResponseStatus: http.StatusOK,
				ResponseBody:   `[{"id":"init-1","name":"init"}]`,
			})...),
			expected: updateExpected{
				displays:         []string{workspaceUpdateDryRunOutput},
				initializerCalls: 1,
			},
		},
		{
			name: "dry run long flags include name and description",
			input: updateInput{
				longArgs: []string{
					"--" + constants.DryRunFlag,
					"--" + constants.NameFlag, "ws-name",
					"--" + constants.DescriptionFlag, "ws-desc",
				},
				args: []string{"ws-cli-1"},
			},
			mockEndpoints: stubs.GetStubAuthEndpoints(),
			expected: updateExpected{
				displays: []string{workspaceUpdateNameDescDryRunOutput},
			},
		},
		{
			name: "update success prints IDs only",
			input: updateInput{
				args: []string{"ws-cli-1"},
			},
			mockEndpoints: getBulkUpdateWorkspaceEndpoints(workspaceUpdateEndpointsConfig{
				workspaceConfig: stubs.StubHTTPPartialPostConfig{
					ExpectedRequestBody: workspaceUpdateRequestBodyIDsOnly,
					ResponseStatus:      http.StatusOK,
					ResponseBody:        `[{"id":"ws-a"},{"id":"ws-b"}]`,
				},
			}),
			expected: updateExpected{
				displays:        []string{"ws-a", "ws-b"},
				bulkUpdateCalls: 1,
			},
		},
		{
			name: "server error",
			input: updateInput{
				args: []string{"ws-cli-1"},
			},
			mockEndpoints: getBulkUpdateWorkspaceEndpoints(workspaceUpdateEndpointsConfig{
				workspaceConfig: stubs.StubHTTPPartialPostConfig{
					ExpectedRequestBody: workspaceUpdateRequestBodyIDsOnly,
					ResponseStatus:      http.StatusUnauthorized,
					ResponseBody:        `unauthorized`,
				},
			}),
			expected: updateExpected{
				err:             workspaceBulkUpdateServerErr,
				bulkUpdateCalls: 1,
			},
		},
		{
			name: "ids fallback from file body id",
			input: updateInput{
				filename:    "/update-workspace.json",
				fileContent: workspaceUpdateFileContent,
				dryRun:      true,
			},
			mockEndpoints: stubs.GetStubAuthEndpoints(),
			expected: updateExpected{
				displays: []string{workspaceUpdateFileDryRunOutput},
			},
		},
		{
			name: "error when no ID source",
			input: updateInput{
				dryRun: true,
			},
			mockEndpoints: stubs.GetStubAuthEndpoints(),
			expected: updateExpected{
				err: workspaceUpdateMissingIDErr,
			},
		},
		{
			name: "ids only with empty body accepted",
			input: updateInput{
				args: []string{"ws-cli-1"},
			},
			mockEndpoints: getBulkUpdateWorkspaceEndpoints(workspaceUpdateEndpointsConfig{
				workspaceConfig: stubs.StubHTTPPartialPostConfig{
					ExpectedRequestBody: workspaceUpdateRequestBodyIDsOnly,
					ResponseStatus:      http.StatusOK,
					ResponseBody:        `[{"id":"ws-cli-1"}]`,
				},
			}),
			expected: updateExpected{
				displays:        []string{"ws-cli-1"},
				bulkUpdateCalls: 1,
			},
		},
		{
			name: "mapping path uses name and description",
			input: updateInput{
				nameValue:   "ws-name",
				description: "ws-desc",
				args:        []string{"ws-cli-1"},
			},
			mockEndpoints: getSingleUpdateWorkspaceEndpoints(workspaceUpdateEndpointsConfig{
				workspaceConfig: stubs.StubHTTPPartialPostConfig{
					ExpectedRequestBody: workspaceUpdateRequestBodyNameDesc,
					ResponseStatus:      http.StatusOK,
					ResponseBody:        `{"id":"ws-cli-1"}`,
				},
			}),
			expected: updateExpected{
				displays:          []string{"ws-cli-1"},
				singleUpdateCalls: 1,
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
		{
			name: "invalid initializers",
			input: updateInput{
				dryRun:       true,
				sizeValue:    "10Gi",
				initializers: []string{"invalid"},
				homeDir:      true,
				args:         []string{"ws-cli-1"},
			},
			mockEndpoints: append(stubs.GetStubAuthEndpoints(), getInitializerEndpoints(stubs.StubHTTPPartialGetConfig{
				ExpectedQuery:  `id=in.(invalid)`,
				ResponseStatus: http.StatusBadRequest,
				ResponseBody:   `error`,
			})...),
			expected: updateExpected{
				err:              expectedInvalidInitalizersErr,
				initializerCalls: 1,
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
			mockCmd := workspace.UpdateCommand(context.Background(), fsys, stubUI)
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
				tt.expected.singleUpdateCalls,
				hitCounts.Count("POST"+postgrest.Endpoint(constants.WorkbenchEndpointRPCWorkspaceUpdate)),
				"update calls",
			)
			require.Equal(
				t,
				tt.expected.bulkUpdateCalls,
				hitCounts.Count("POST"+postgrest.Endpoint(constants.WorkbenchEndpointRPCWorkspaceBulkUpdate)),
				"bulk update calls",
			)
			require.Equal(
				t,
				tt.expected.initializerCalls,
				hitCounts.Count("GET"+postgrest.Endpoint(constants.WorkbenchEndpointDataplaneWorkspaceServices)),
				"initializer calls",
			)
		})
	}
}

func getCreateArgs(
	tt createWorkspaceTestCase,
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
	if tt.name != "" {
		args = append(args, "--"+constants.NameFlag, tt.input.name)
	}
	if tt.input.description != "" {
		args = append(args, "--"+constants.DescriptionFlag, tt.input.description)
	}
	if tt.input.storageClassName != "" {
		args = append(args, "--"+constants.StorageClassFlag, tt.input.storageClassName)
	}
	if tt.input.size != "" {
		args = append(args, "--"+constants.SizeFlag, tt.input.size)
	}
	if tt.input.homeDir {
		args = append(args, "--"+constants.HomeDirFlag)
	}
	if tt.input.snapshot != "" {
		args = append(args, "--"+constants.SnapshotFlag, tt.input.snapshot)
	}
	for _, initializer := range tt.input.initializers {
		args = append(args, "--"+constants.InitializerFlag, initializer)
	}
	return args
}

func getUpdateArgs(
	tt updateWorkspaceTestCase,
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
	if tt.input.nameValue != "" {
		args = append(args, "--"+constants.NameFlag, tt.input.nameValue)
	}
	if tt.input.description != "" {
		args = append(args, "--"+constants.DescriptionFlag, tt.input.description)
	}
	if tt.input.sizeValue != "" {
		args = append(args, "--"+constants.SizeFlag, tt.input.sizeValue)
	}
	for _, initializer := range tt.input.initializers {
		args = append(args, "--"+constants.InitializerFlag, initializer)
	}
	if tt.input.homeDir {
		args = append(args, "--"+constants.HomeDirFlag)
	}
	args = append(args, tt.input.longArgs...)
	args = append(args, tt.input.args...)
	return args
}

type workspaceEndpointsConfig struct {
	dataplaneConfig       stubs.StubHTTPPartialGetConfig
	storageClassesConfigs []stubs.StubHTTPPartialGetConfig
	snapshotsConfig       stubs.StubHTTPPartialGetConfig
	applicationsConfig    stubs.StubHTTPPartialGetConfig
	workspaceConfig       stubs.StubHTTPPartialPostConfig
}

type workspaceUpdateEndpointsConfig struct {
	initializerConfig stubs.StubHTTPPartialGetConfig
	workspaceConfig   stubs.StubHTTPPartialPostConfig
}

func getWorkspaceEndpoints(
	urlUnderTest string,
	config workspaceEndpointsConfig,
) []stubs.StubHTTPEndpoint {
	endpoints := append(stubs.GetStubAuthEndpoints(), stubs.GetStubDataplaneEndpoints(config.dataplaneConfig)...)
	endpoints = append(endpoints, stubs.GetStubApplicationEndpoints(config.applicationsConfig)...)
	endpoints = append(endpoints,
		[]stubs.StubHTTPEndpoint{
			{
				Request: stubs.StubHTTPRequest{
					Method:          http.MethodGet,
					URL:             postgrest.Endpoint(constants.WorkbenchEndpointSnapshots),
					ExpectedHeaders: stubs.StubAuthHeaders,
					ExpectedQuery:   config.snapshotsConfig.ExpectedQuery,
				},
				Response: stubs.StubHTTPResponse{
					Status: config.snapshotsConfig.ResponseStatus,
					Body:   config.snapshotsConfig.ResponseBody,
				},
			},
			{
				Request: stubs.StubHTTPRequest{
					Method:          http.MethodPost,
					URL:             urlUnderTest,
					ExpectedHeaders: stubs.StubAuthHeaders,
					ExpectedBody:    config.workspaceConfig.ExpectedRequestBody,
				},
				Response: stubs.StubHTTPResponse{
					Status: config.workspaceConfig.ResponseStatus,
					Body:   config.workspaceConfig.ResponseBody,
				},
			},
		}...)
	for _, storageClassesConfig := range config.storageClassesConfigs {
		endpoints = append(endpoints, stubs.StubHTTPEndpoint{
			Request: stubs.StubHTTPRequest{
				Method:          http.MethodGet,
				URL:             postgrest.Endpoint(constants.WorkbenchEndpointDataplaneStorageclasses),
				ExpectedHeaders: stubs.StubAuthHeaders,
				ExpectedQuery:   storageClassesConfig.ExpectedQuery,
			},
			Response: stubs.StubHTTPResponse{
				Status: storageClassesConfig.ResponseStatus,
				Body:   storageClassesConfig.ResponseBody,
			},
		},
		)
	}
	return endpoints
}

func getCreateWorkspaceEndpoints(config workspaceEndpointsConfig) []stubs.StubHTTPEndpoint {
	return getWorkspaceEndpoints(postgrest.Endpoint(constants.WorkbenchEndpointRPCWorkspaceCreate), config)
}

func getInitializerEndpoints(config stubs.StubHTTPPartialGetConfig) []stubs.StubHTTPEndpoint {
	return []stubs.StubHTTPEndpoint{
		{
			Request: stubs.StubHTTPRequest{
				Method:          http.MethodGet,
				URL:             postgrest.Endpoint(constants.WorkbenchEndpointDataplaneWorkspaceServices),
				ExpectedHeaders: stubs.StubAuthHeaders,
				ExpectedQuery:   config.ExpectedQuery,
			},
			Response: stubs.StubHTTPResponse{
				Status: config.ResponseStatus,
				Body:   config.ResponseBody,
			},
		},
	}
}

func getUpdateWorkspaceEndpoints(
	urlUnderTest string,
	config workspaceUpdateEndpointsConfig,
) []stubs.StubHTTPEndpoint {
	endpoints := append(stubs.GetStubAuthEndpoints(),
		getInitializerEndpoints(config.initializerConfig)...)
	return append(endpoints, stubs.StubHTTPEndpoint{
		Request: stubs.StubHTTPRequest{
			Method:          http.MethodPost,
			URL:             urlUnderTest,
			ExpectedHeaders: stubs.StubAuthHeaders,
			ExpectedBody:    config.workspaceConfig.ExpectedRequestBody,
		},
		Response: stubs.StubHTTPResponse{
			Status: config.workspaceConfig.ResponseStatus,
			Body:   config.workspaceConfig.ResponseBody,
		},
	},
	)
}

func getSingleUpdateWorkspaceEndpoints(config workspaceUpdateEndpointsConfig) []stubs.StubHTTPEndpoint {
	return getUpdateWorkspaceEndpoints(
		postgrest.Endpoint(constants.WorkbenchEndpointRPCWorkspaceUpdate),
		config,
	)
}
func getBulkUpdateWorkspaceEndpoints(config workspaceUpdateEndpointsConfig) []stubs.StubHTTPEndpoint {
	return getUpdateWorkspaceEndpoints(
		postgrest.Endpoint(constants.WorkbenchEndpointRPCWorkspaceBulkUpdate),
		config,
	)
}
