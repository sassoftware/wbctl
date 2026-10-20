// Copyright © 2026, SAS Institute Inc., Cary, NC, USA.  All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package job_test

import (
	"context"
	"net/http"
	"strings"
	"testing"

	"github.com/sassoftware/wbctl/internal/cmd"
	"github.com/sassoftware/wbctl/internal/cmd/create/job"
	"github.com/sassoftware/wbctl/internal/constants"
	"github.com/sassoftware/wbctl/internal/postgrest"
	"github.com/sassoftware/wbctl/internal/stubs"
	"github.com/spf13/afero"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
)

const (
	expectedCreateJobDryRunOutput = `{
    "application_id": "vscode",
    "context": "ws-1",
    "dataplane_id": "123e4567-e89b-12d3-a456-426614174000",
    "name": "job-from-file",
    "nodepool": "np-file"
}`
	expectedCreateJobFileContent = `{"name":"job-from-file",` +
		`"nodepool":"np-file","context":"ws-1","application_id":"vscode"}`
	expectedCreateJobIncompleteFileContent = `{"name":"job-from-incomplete-file",` +
		`"nodepool":"np-file","context":"ws-1"}`
	expectedCrateJobIncompleteFileErr = "either --application or at least one --service is required\n" +
		"  Valid Applications: application-1, application-2\n" +
		"  Valid Services: service-1, service-2"
	expectedCreateJobID         = "job-1"
	expectedCreateJobNameError  = "--name is required"
	expectedCreateJobContextErr = "either --context or --command is required"
	expectedCreateJobAppErr     = "either --application or at least one --service is required\n" +
		"  Valid Applications: application-1, application-2\n" +
		"  Valid Services: service-1, service-2"
	expectedCreateJobNodepoolErr = "--nodepool is required\n" +
		"  Valid Nodepools: np-1, np-2"
	expectedCreateJobServerErr = "POST ${URL}/postgrest/rpc/job_create failed\n" +
		" Caused by:\n" +
		"\t* 401 unauthorized"
	expectedCreateJobWarn           = "no volume specified"
	expectedDataplaneQuery          = "id=in.(123e4567-e89b-12d3-a456-426614174000)"
	expectedApplicationQuery        = "id=in.(app-1)"
	expectedApplicationResponseBody = `[{"id":"app-1","display_name":"RStudio",` +
		`"description":"R app","services":["svc-a","svc-b"]}]`
	expectedServicesResponseBody = `[{"id":"service-1","name":"service 1"},` +
		`{"id":"service-2","name":"service 2"}]`
	expectedCreateJobRequestBodyWithContext = `
{
  "param": {
    "application_id": "app-1",
    "context": "build/src",
    "dataplane_id": "123e4567-e89b-12d3-a456-426614174000",
    "description": "job desc",
    "name": "job-1",
    "nodepool": "np-1",
    "services": [ { "id": "svc-a" }, { "id": "svc-b" } ],
    "properties": {
      "isWorkbenchAppCreated": true
    }
  }
}
`
	expectedCreateJobRequestBodyNoWorkspaces = `
{
  "param": {
    "application_id": "app-1",
    "dataplane_id": "123e4567-e89b-12d3-a456-426614174000",
    "description": "job desc",
    "name": "job-1",
    "nodepool": "np-1",
    "services": [ { "id": "svc-a" }, { "id": "svc-b" } ],
    "spec": {
      "tasks": {
        "default": {
          "cmds": [
            "run.sh"
          ]
        }
      },
      "version": "3"
    },
    "properties": {
      "isWorkbenchAppCreated": true
    }
  }
}
`
	expectedCreateJobRequestBodyWithWorkspaces = `
{
  "param": {
    "application_id": "app-1",
    "dataplane_id": "123e4567-e89b-12d3-a456-426614174000",
    "description": "job desc",
    "name": "job-2",
    "nodepool": "np-2",
    "services": [
      {
        "id": "svc-a",
        "env": {
          "DEFAULT_MOUNTNAME": "data"
        }
      },
      {
        "id": "svc-b",
        "env": {
          "DEFAULT_MOUNTNAME": "data"
        }
      }
    ],
    "spec": {
      "tasks": {
        "default": {
          "cmds": [
            "run.sh"
          ]
        }
      },
      "version": "3"
    },
    "sub_path": "src",
    "workspaces": [
      {
        "id": "ctrl-home",
        "mountName": "sas"
      },
      {
        "id": "ctrl-data",
        "mountName": "data"
      }
    ],
    "properties": {
      "isWorkbenchAppCreated": true,
      "primaryVolume": {
        "controller_id": "ctrl-home",
        "folderName": "sas",
        "id": "ws-home"
      }
    }
  }
}
`
	expectedCreateJobRequestBodyWithServices = `
{
  "param": {
    "dataplane_id": "123e4567-e89b-12d3-a456-426614174000",
    "description": "job desc",
    "name": "job-2",
    "nodepool": "np-2",
    "services": [
      {
        "id": "service-1",
        "env": {
          "DEFAULT_MOUNTNAME": "data"
        }
      },
      {
        "id": "service-2",
        "env": {
          "DEFAULT_MOUNTNAME": "data"
        }
      }
    ],
    "spec": {
      "tasks": {
        "default": {
          "cmds": [
            "run.sh"
          ]
        }
      },
      "version": "3"
    },
    "sub_path": "src",
    "workspaces": [
      {
        "id": "ctrl-home",
        "mountName": "sas"
      },
      {
        "id": "ctrl-data",
        "mountName": "data"
      }
    ],
    "properties": {
      "isWorkbenchAppCreated": true,
      "primaryVolume": {
        "controller_id": "ctrl-home",
        "folderName": "sas",
        "id": "ws-home"
      }
    }
  }
}
`
	expectedUpdateJobDryRunOutput = `{
    "ids": [
        "job-cli-1",
        "job-cli-2"
    ],
    "param": {
        "description": "job update desc",
        "properties": {
            "isWorkbenchAppCreated": true
        },
        "spec": {
            "tasks": {
                "default": {
                    "cmds": [
                        "run.sh"
                    ]
                }
            },
            "version": "3"
        },
        "sub_path": "src",
        "workspaces": [
            {
                "id": "ctrl-home",
                "mountName": "sas"
            }
        ]
    }
}`
	expectedUpdateJobRequestBodyMapping = `
{
  "ids": [
    "job-cli-1",
    "job-cli-2"
  ],
  "param": {
    "description": "job update desc",
    "properties": {
      "isWorkbenchAppCreated": true
    },
    "spec": {
      "tasks": {
        "default": {
          "cmds": [
            "run.sh"
          ]
        }
      },
      "version": "3"
    },
    "sub_path": "src",
    "workspaces": [
      {
        "id": "ctrl-home",
        "mountName": "sas"
      }
    ]
  }
}
`
	expectedUpdateJobRequestBodyIDsOnly = `
{
  "ids": [
    "job-cli-1"
  ],
  "param": {
    "properties": {
      "isWorkbenchAppCreated": true
    }
  }
}
`
	expectedUpdateJobFileContent      = `{"id":"job-file","description":"from file"}`
	expectedUpdateJobFileDryRunOutput = `{
    "ids": [
        "job-file"
    ],
    "param": {
        "description": "from file"
    }
}`
	expectedUpdateJobMissingIDErr = "at least one resource ID must be provided for update"
	expectedUpdateJobServerErr    = "POST ${URL}/postgrest/rpc/job_bulk_update failed\n" +
		" Caused by:\n" +
		"\t* 401 unauthorized"
	expectedUpdateJobRequestBodyAllFlags = `
{
  "ids": [
    "job-cli-1"
  ],
  "param": {
    "application_id": "app-1",
    "context": "build/src",
    "description": "job update desc",
    "labels": {
      "team": "core"
    },
    "name": "job-name",
    "nodepool": "np-1",
    "properties": {
      "isWorkbenchAppCreated": true,
      "schedule": "simple"
    },
    "schedule": "0 0 * * *",
    "services": [
      {
        "id": "svc-a",
        "env": {
          "A": "B"
        }
      },
      {
        "id": "svc-b",
        "env": {
          "A": "B"
        }
      }
    ],
    "spec": {
      "tasks": {
        "default": {
          "cmds": [
            "run.sh"
          ]
        }
      },
      "version": "3"
    },
    "sub_path": "src"
  }
}
	`
	expectedUpdateJobLongFlagsDryRunOutput = `{
    "ids": [
        "job-cli-1"
    ],
    "param": {
        "application_id": "app-1",
        "context": "build/src",
        "description": "job update desc",
        "labels": {
            "team": "core"
        },
        "name": "job-name",
        "nodepool": "np-1",
        "properties": {
            "isWorkbenchAppCreated": true,
            "schedule": "simple"
        },
        "schedule": "0 0 * * *",
        "services": [
            {
                "id": "svc-a",
                "env": {
                    "A": "B"
                }
            },
            {
                "id": "svc-b",
                "env": {
                    "A": "B"
                }
            }
        ],
        "spec": {
            "tasks": {
                "default": {
                    "cmds": [
                        "run.sh"
                    ]
                }
            },
            "version": "3"
        },
        "sub_path": "src",
        "workspaces": [
            {
                "id": "ctrl-home",
                "mountName": "sas"
            }
        ]
    }
}`
	expectedInvalidEndpointErr = "failed to get username/password token" +
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
	expectedCreateImageInvalidFileErr = "unexpected end of JSON input"
)

type createInput struct {
	endpoint    string
	filename    string
	fileContent string
	dryRun      bool
	dataplane   string
	name        string
	description string
	context     string
	command     string
	application string
	services    []string
	nodepool    string
	subPath     string
	workspaces  []string
}

type createExpected struct {
	displays          []string
	warns             []string
	err               string
	createCalls       int
	dataplanesCalls   int
	applicationsCalls int
	servicesCalls     int
	nodepoolsCalls    int
	workspacesCalls   int
}

type createJobTestCase struct {
	name          string
	input         createInput
	mockEndpoints []stubs.StubHTTPEndpoint
	expected      createExpected
}

func TestCreateJobCommand(t *testing.T) {
	tests := []createJobTestCase{
		{
			name: "dry run from file",
			input: createInput{
				filename:    "/input.json",
				fileContent: expectedCreateJobFileContent,
				dryRun:      true,
			},
			mockEndpoints: getCreateJobEndpoints(jobEndpointsConfig{
				dataplaneConfig: stubs.StubHTTPPartialGetConfig{
					ResponseBody:   `[{"id":"123e4567-e89b-12d3-a456-426614174000"}]`,
					ResponseStatus: http.StatusOK,
				},
			}),
			expected: createExpected{
				displays:        []string{expectedCreateJobDryRunOutput},
				warns:           []string{expectedCreateJobWarn},
				dataplanesCalls: 1,
			},
		},
		{
			name: "dry run from incomplete file",
			input: createInput{
				filename:    "/input.json",
				fileContent: expectedCreateJobIncompleteFileContent,
				dryRun:      true,
			},
			mockEndpoints: getCreateJobEndpoints(jobEndpointsConfig{
				dataplaneConfig: stubs.StubHTTPPartialGetConfig{
					ResponseBody:   `[{"id":"123e4567-e89b-12d3-a456-426614174000"}]`,
					ResponseStatus: http.StatusOK,
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
				err:               expectedCrateJobIncompleteFileErr,
				dataplanesCalls:   1,
				applicationsCalls: 1,
				servicesCalls:     1,
			},
		},
		{
			name: "create success with context",
			input: createInput{
				dataplane:   "123e4567-e89b-12d3-a456-426614174000",
				name:        "job-1",
				description: "job desc",
				context:     "build/src",
				application: "app-1",
				nodepool:    "np-1",
			},
			mockEndpoints: getCreateJobEndpoints(jobEndpointsConfig{
				dataplaneConfig: stubs.StubHTTPPartialGetConfig{
					ExpectedQuery:  expectedDataplaneQuery,
					ResponseBody:   `[{"id":"123e4567-e89b-12d3-a456-426614174000"}]`,
					ResponseStatus: http.StatusOK,
				},
				applicationsConfig: stubs.StubHTTPPartialGetConfig{
					ExpectedQuery:  expectedApplicationQuery,
					ResponseBody:   expectedApplicationResponseBody,
					ResponseStatus: http.StatusOK,
				},
				jobConfig: stubs.StubHTTPPartialPostConfig{
					ExpectedRequestBody: expectedCreateJobRequestBodyWithContext,
					ResponseBody:        `{"id":"job-1"}`,
					ResponseStatus:      http.StatusOK,
				},
			}),
			expected: createExpected{
				displays:          []string{expectedCreateJobID},
				warns:             []string{expectedCreateJobWarn},
				createCalls:       1,
				dataplanesCalls:   1,
				applicationsCalls: 1,
			},
		},
		{
			name: "create success with workspaces",
			input: createInput{
				dataplane:   "123e4567-e89b-12d3-a456-426614174000",
				name:        "job-2",
				description: "job desc",
				command:     "run.sh",
				application: "app-1",
				nodepool:    "np-2",
				subPath:     "src",
				workspaces:  []string{"ws-home", "ws-data"},
			},
			mockEndpoints: getCreateJobEndpoints(jobEndpointsConfig{
				dataplaneConfig: stubs.StubHTTPPartialGetConfig{
					ExpectedQuery:  expectedDataplaneQuery,
					ResponseBody:   `[{"id":"123e4567-e89b-12d3-a456-426614174000"}]`,
					ResponseStatus: http.StatusOK,
				},
				applicationsConfig: stubs.StubHTTPPartialGetConfig{
					ExpectedQuery:  expectedApplicationQuery,
					ResponseBody:   expectedApplicationResponseBody,
					ResponseStatus: http.StatusOK,
				},
				jobConfig: stubs.StubHTTPPartialPostConfig{
					ExpectedRequestBody: expectedCreateJobRequestBodyWithWorkspaces,
					ResponseBody:        `{"id":"job-1"}`,
					ResponseStatus:      http.StatusOK,
				},
			}),
			expected: createExpected{
				displays:          []string{expectedCreateJobID},
				createCalls:       1,
				dataplanesCalls:   1,
				applicationsCalls: 1,
				workspacesCalls:   2,
			},
		},
		{
			name: "create success with services",
			input: createInput{
				dataplane:   "123e4567-e89b-12d3-a456-426614174000",
				name:        "job-2",
				description: "job desc",
				command:     "run.sh",
				services:    []string{"service-1", "service-2"},
				nodepool:    "np-2",
				subPath:     "src",
				workspaces:  []string{"ws-home", "ws-data"},
			},
			mockEndpoints: getCreateJobEndpoints(jobEndpointsConfig{
				dataplaneConfig: stubs.StubHTTPPartialGetConfig{
					ExpectedQuery:  expectedDataplaneQuery,
					ResponseBody:   `[{"id":"123e4567-e89b-12d3-a456-426614174000"}]`,
					ResponseStatus: http.StatusOK,
				},
				servicesConfig: stubs.StubHTTPPartialGetConfig{
					ResponseBody:   expectedServicesResponseBody,
					ResponseStatus: http.StatusOK,
				},
				jobConfig: stubs.StubHTTPPartialPostConfig{
					ExpectedRequestBody: expectedCreateJobRequestBodyWithServices,
					ResponseBody:        `{"id":"job-1"}`,
					ResponseStatus:      http.StatusOK,
				},
			}),
			expected: createExpected{
				displays:        []string{expectedCreateJobID},
				createCalls:     1,
				dataplanesCalls: 1,
				servicesCalls:   1,
				workspacesCalls: 2,
			},
		},
		{
			name: "create error",
			input: createInput{
				dataplane:   "123e4567-e89b-12d3-a456-426614174000",
				name:        "job-1",
				description: "job desc",
				command:     "run.sh",
				application: "app-1",
				nodepool:    "np-1",
			},
			mockEndpoints: getCreateJobEndpoints(jobEndpointsConfig{
				dataplaneConfig: stubs.StubHTTPPartialGetConfig{
					ExpectedQuery:  expectedDataplaneQuery,
					ResponseBody:   `[{"id":"123e4567-e89b-12d3-a456-426614174000"}]`,
					ResponseStatus: http.StatusOK,
				},
				applicationsConfig: stubs.StubHTTPPartialGetConfig{
					ExpectedQuery:  expectedApplicationQuery,
					ResponseBody:   expectedApplicationResponseBody,
					ResponseStatus: http.StatusOK,
				},
				jobConfig: stubs.StubHTTPPartialPostConfig{
					ExpectedRequestBody: expectedCreateJobRequestBodyNoWorkspaces,
					ResponseBody:        `unauthorized`,
					ResponseStatus:      http.StatusUnauthorized,
				},
			}),
			expected: createExpected{
				err:               expectedCreateJobServerErr,
				warns:             []string{expectedCreateJobWarn},
				createCalls:       1,
				dataplanesCalls:   1,
				applicationsCalls: 1,
			},
		},
		{
			name: "name required",
			input: createInput{
				context:     "build/src",
				application: "app-1",
				nodepool:    "np-1",
			},
			mockEndpoints: getCreateJobEndpoints(jobEndpointsConfig{
				dataplaneConfig: stubs.StubHTTPPartialGetConfig{
					ResponseBody:   `[{"id":"123e4567-e89b-12d3-a456-426614174000"}]`,
					ResponseStatus: http.StatusOK,
				},
				applicationsConfig: stubs.StubHTTPPartialGetConfig{
					ExpectedQuery:  expectedApplicationQuery,
					ResponseBody:   expectedApplicationResponseBody,
					ResponseStatus: http.StatusOK,
				},
			}),
			expected: createExpected{
				err:               expectedCreateJobNameError,
				applicationsCalls: 1,
				dataplanesCalls:   1,
			},
		},
		{
			name: "context or command required",
			input: createInput{
				name:        "job-1",
				application: "app-1",
				nodepool:    "np-1",
			},
			mockEndpoints: getCreateJobEndpoints(jobEndpointsConfig{
				dataplaneConfig: stubs.StubHTTPPartialGetConfig{
					ResponseBody:   `[{"id":"123e4567-e89b-12d3-a456-426614174000"}]`,
					ResponseStatus: http.StatusOK,
				},
				applicationsConfig: stubs.StubHTTPPartialGetConfig{
					ExpectedQuery:  expectedApplicationQuery,
					ResponseBody:   expectedApplicationResponseBody,
					ResponseStatus: http.StatusOK,
				},
			}),
			expected: createExpected{
				err:               expectedCreateJobContextErr,
				applicationsCalls: 1,
				dataplanesCalls:   1,
			},
		},
		{
			name: "application or service required",
			input: createInput{
				name:     "job-1",
				command:  "run.sh",
				nodepool: "np-1",
			},
			mockEndpoints: getCreateJobEndpoints(jobEndpointsConfig{
				dataplaneConfig: stubs.StubHTTPPartialGetConfig{
					ResponseBody:   `[{"id":"123e4567-e89b-12d3-a456-426614174000"}]`,
					ResponseStatus: http.StatusOK,
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
				dataplanesCalls:   1,
				err:               expectedCreateJobAppErr,
				applicationsCalls: 1,
				servicesCalls:     1,
			},
		},
		{
			name: "nodepool required",
			input: createInput{
				name:        "job-1",
				command:     "run.sh",
				application: "app-1",
			},
			mockEndpoints: getCreateJobEndpoints(jobEndpointsConfig{
				dataplaneConfig: stubs.StubHTTPPartialGetConfig{
					ResponseBody:   `[{"id":"123e4567-e89b-12d3-a456-426614174000"}]`,
					ResponseStatus: http.StatusOK,
				},
				applicationsConfig: stubs.StubHTTPPartialGetConfig{
					ExpectedQuery:  expectedApplicationQuery,
					ResponseBody:   expectedApplicationResponseBody,
					ResponseStatus: http.StatusOK,
				},
				nodepoolsConfig: stubs.StubHTTPPartialGetConfig{
					ResponseBody:   `[{"id": "np-1"}, {"id": "np-2"}]`,
					ResponseStatus: http.StatusOK,
				},
			}),
			expected: createExpected{
				err:               expectedCreateJobNodepoolErr,
				applicationsCalls: 1,
				dataplanesCalls:   1,
				nodepoolsCalls:    1,
			},
		},
		{
			name: "invalid endpoint",
			input: createInput{
				endpoint: "a\nb",
			},
			expected: createExpected{
				err: expectedInvalidEndpointErr,
			},
		},
		{
			name: "invalid dataplane id",
			input: createInput{
				dryRun:    true,
				dataplane: "invalid",
			},
			mockEndpoints: getCreateJobEndpoints(jobEndpointsConfig{
				dataplaneConfig: stubs.StubHTTPPartialGetConfig{
					ExpectedQuery:  "id=in.(invalid)",
					ResponseBody:   `error`,
					ResponseStatus: http.StatusBadRequest,
				},
			}),
			expected: createExpected{
				err:             expectedInvlidDataplaneErr,
				dataplanesCalls: 1,
			},
		},
		{
			name: "invalid workspace id",
			input: createInput{
				dryRun:     true,
				workspaces: []string{"invalid"},
			},
			mockEndpoints: getCreateJobEndpoints(jobEndpointsConfig{
				workspaceConfig: stubs.StubHTTPPartialGetConfig{
					ExpectedQuery:  "controller_id=in.(invalid)",
					ResponseBody:   `error`,
					ResponseStatus: http.StatusBadRequest,
				},
			}),
			expected: createExpected{
				workspacesCalls: 1,
				err:             expectedInvlidWorkspaceErr,
			},
		},
		{
			name: "invalid application id",
			input: createInput{
				dryRun:      true,
				application: "invalid",
			},
			mockEndpoints: getCreateJobEndpoints(jobEndpointsConfig{
				applicationsConfig: stubs.StubHTTPPartialGetConfig{
					ExpectedQuery:  "id=in.(invalid)",
					ResponseBody:   `error`,
					ResponseStatus: http.StatusBadRequest,
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
			mockCmd := job.CreateCommand(context.Background(), fsys, stubUI)
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
				hitCounts.Count("POST"+postgrest.Endpoint(constants.WorkbenchEndpointRPCJobCreate)),
				"create calls",
			)
			require.Equal(
				t,
				tt.expected.dataplanesCalls,
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
				tt.expected.servicesCalls,
				hitCounts.Count("GET"+postgrest.Endpoint(constants.WorkbenchEndpointDataplaneWorkbenchServices)),
				"service calls",
			)
			require.Equal(
				t,
				tt.expected.nodepoolsCalls,
				hitCounts.Count("GET"+postgrest.Endpoint(constants.WorkbenchEndpointDataplaneNodepools)),
				"nodepool calls",
			)
			require.Equal(
				t,
				tt.expected.workspacesCalls,
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
	dryRun      bool
	dataplane   string
	name        string
	description string
	context     string
	command     string
	subPath     string
	schedule    string
	nodepool    string
	application string
	labels      []string
	envs        []string
	workspaces  []string
	longArgs    []string
	args        []string
}

type updateExpected struct {
	displays          []string
	warns             []string
	err               string
	updateCalls       int
	applicationsCalls int
	servicesCalls     int
	nodepoolsCalls    int
	workspaceCalls    int
}

type updateJobTestCase struct {
	name          string
	input         updateInput
	mockEndpoints []stubs.StubHTTPEndpoint
	expected      updateExpected
}

func TestUpdateJobCommand(t *testing.T) {
	tests := []updateJobTestCase{
		{
			name: "dry run with CLI ids and mapped fields",
			input: updateInput{
				dryRun:      true,
				dataplane:   "123e4567-e89b-12d3-a456-426614174000",
				description: "job update desc",
				command:     "run.sh",
				subPath:     "src",
				workspaces:  []string{"ws-home"},
				args:        []string{"job-cli-1", "job-cli-2"},
			},
			mockEndpoints: getUpdateJobEndpoints(jobEndpointsConfig{}),
			expected: updateExpected{
				displays:       []string{expectedUpdateJobDryRunOutput},
				workspaceCalls: 1,
			},
		},
		{
			name: "dry run long flags include all mapped fields (workspace flag)",
			input: updateInput{
				longArgs: []string{
					"--" + constants.DryRunFlag,
					"--" + constants.NameFlag, "job-name",
					"--" + constants.DescriptionFlag, "job update desc",
					"--" + constants.ContextFlag, "build/src",
					"--" + constants.CommandFlag, "run.sh",
					"--" + constants.SubPathFlag, "src",
					"--" + constants.ScheduleFlag, "0 0 * * *",
					"--" + constants.NodepoolFlag, "np-1",
					"--" + constants.ApplicationFlag, "app-1",
					"--" + constants.LabelFlag, "team=core",
					"--" + constants.EnvironmentFlag, "A=B",
					"--" + constants.WorkspaceFlag, "ws-home",
				},
				args: []string{"job-cli-1"},
			},
			mockEndpoints: getUpdateJobEndpoints(jobEndpointsConfig{
				applicationsConfig: stubs.StubHTTPPartialGetConfig{
					ExpectedQuery:  expectedApplicationQuery,
					ResponseBody:   expectedApplicationResponseBody,
					ResponseStatus: http.StatusOK,
				},
			}),
			expected: updateExpected{
				displays:          []string{expectedUpdateJobLongFlagsDryRunOutput},
				applicationsCalls: 1,
				workspaceCalls:    1,
			},
		},
		{
			name: "dry run long flags include all mapped fields",
			input: updateInput{
				longArgs: []string{
					"--" + constants.DryRunFlag,
					"--" + constants.NameFlag, "job-name",
					"--" + constants.DescriptionFlag, "job update desc",
					"--" + constants.ContextFlag, "build/src",
					"--" + constants.CommandFlag, "run.sh",
					"--" + constants.SubPathFlag, "src",
					"--" + constants.ScheduleFlag, "0 0 * * *",
					"--" + constants.NodepoolFlag, "np-1",
					"--" + constants.ApplicationFlag, "app-1",
					"--" + constants.LabelFlag, "team=core",
					"--" + constants.EnvironmentFlag, "A=B",
					"--" + constants.VolumeFlag, "ws-home",
				},
				args: []string{"job-cli-1"},
			},
			mockEndpoints: getUpdateJobEndpoints(jobEndpointsConfig{
				applicationsConfig: stubs.StubHTTPPartialGetConfig{
					ExpectedQuery:  expectedApplicationQuery,
					ResponseBody:   expectedApplicationResponseBody,
					ResponseStatus: http.StatusOK,
				},
			}),
			expected: updateExpected{
				displays:          []string{expectedUpdateJobLongFlagsDryRunOutput},
				applicationsCalls: 1,
				workspaceCalls:    1,
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
				args: []string{"job-cli-1"},
			},
			mockEndpoints: getUpdateJobEndpoints(jobEndpointsConfig{
				jobConfig: stubs.StubHTTPPartialPostConfig{
					ExpectedRequestBody: expectedUpdateJobRequestBodyIDsOnly,
					ResponseStatus:      http.StatusOK,
					ResponseBody:        `[{"id":"job-a"},{"id":"job-b"}]`,
				},
			}),
			expected: updateExpected{
				displays:    []string{"job-a", "job-b"},
				updateCalls: 1,
			},
		},
		{
			name: "server error",
			input: updateInput{
				args: []string{"job-cli-1"},
			},
			mockEndpoints: getUpdateJobEndpoints(jobEndpointsConfig{
				jobConfig: stubs.StubHTTPPartialPostConfig{
					ExpectedRequestBody: expectedUpdateJobRequestBodyIDsOnly,
					ResponseStatus:      http.StatusUnauthorized,
					ResponseBody:        `unauthorized`,
				},
			}),
			expected: updateExpected{
				err:         expectedUpdateJobServerErr,
				updateCalls: 1,
			},
		},
		{
			name: "ids fallback from file body id",
			input: updateInput{
				filename:    "/update-job.json",
				fileContent: expectedUpdateJobFileContent,
				dryRun:      true,
			},
			mockEndpoints: stubs.GetStubAuthEndpoints(),
			expected: updateExpected{
				displays: []string{expectedUpdateJobFileDryRunOutput},
			},
		},
		{
			name: "error when no ID source",
			input: updateInput{
				description: "job update desc",
			},
			mockEndpoints: stubs.GetStubAuthEndpoints(),
			expected: updateExpected{
				err: expectedUpdateJobMissingIDErr,
			},
		},
		{
			name: "mapping path uses dataplane workspace and spec",
			input: updateInput{
				dataplane:   "123e4567-e89b-12d3-a456-426614174000",
				description: "job update desc",
				command:     "run.sh",
				subPath:     "src",
				workspaces:  []string{"ws-home"},
				args:        []string{"job-cli-1", "job-cli-2"},
			},
			mockEndpoints: getUpdateJobEndpoints(jobEndpointsConfig{
				jobConfig: stubs.StubHTTPPartialPostConfig{
					ExpectedRequestBody: expectedUpdateJobRequestBodyMapping,
					ResponseStatus:      http.StatusOK,
					ResponseBody:        `[{"id":"job-cli-1"},{"id":"job-cli-2"}]`,
				},
			}),
			expected: updateExpected{
				displays:       []string{"job-cli-1", "job-cli-2"},
				updateCalls:    1,
				workspaceCalls: 1,
			},
		},
		{
			name: "mapping path uses remaining update flags",
			input: updateInput{
				dataplane:   "123e4567-e89b-12d3-a456-426614174000",
				name:        "job-name",
				description: "job update desc",
				context:     "build/src",
				command:     "run.sh",
				subPath:     "src",
				schedule:    "0 0 * * *",
				nodepool:    "np-1",
				application: "app-1",
				labels:      []string{"team=core"},
				envs:        []string{"A=B"},
				args:        []string{"job-cli-1"},
			},
			mockEndpoints: getUpdateJobEndpoints(jobEndpointsConfig{
				jobConfig: stubs.StubHTTPPartialPostConfig{
					ExpectedRequestBody: expectedUpdateJobRequestBodyAllFlags,
					ResponseStatus:      http.StatusOK,
					ResponseBody:        `[{"id":"job-cli-1"}]`,
				},
				applicationsConfig: stubs.StubHTTPPartialGetConfig{
					ExpectedQuery:  expectedApplicationQuery,
					ResponseStatus: http.StatusOK,
					ResponseBody:   expectedApplicationResponseBody,
				},
			}),
			expected: updateExpected{
				displays:          []string{"job-cli-1"},
				updateCalls:       1,
				applicationsCalls: 1,
			},
		},
		{
			name: "invalid endpoint",
			input: updateInput{
				endpoint: "a\nb",
			},
			expected: updateExpected{
				err: expectedInvalidEndpointErr,
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
			mockCmd := job.UpdateCommand(context.Background(), fsys, stubUI)
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
				hitCounts.Count("POST"+postgrest.Endpoint(constants.WorkbenchEndpointRPCJobBulkUpdate)),
				"update calls",
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
		})
	}
}

func getCreateArgs(
	tt createJobTestCase,
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
	if tt.input.context != "" {
		args = append(args, "--"+constants.ContextFlag, tt.input.context)
	}
	if tt.input.command != "" {
		args = append(args, "--"+constants.CommandFlag, tt.input.command)
	}
	if tt.input.nodepool != "" {
		args = append(args, "--"+constants.NodepoolFlag, tt.input.nodepool)
	}
	if tt.input.application != "" {
		args = append(args, "--"+constants.ApplicationFlag, tt.input.application)
	}
	if len(tt.input.services) != 0 {
		for _, service := range tt.input.services {
			args = append(args, "--"+constants.ServiceFlag, service)
		}
	}
	if tt.input.subPath != "" {
		args = append(args, "--"+constants.SubPathFlag, tt.input.subPath)
	}
	for _, workspaceID := range tt.input.workspaces {
		args = append(args, "--"+constants.WorkspaceFlag, workspaceID)
	}
	return args
}

func getUpdateArgs(
	tt updateJobTestCase,
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
	if tt.input.description != "" {
		args = append(args, "--"+constants.DescriptionFlag, tt.input.description)
	}
	if tt.input.name != "" {
		args = append(args, "--"+constants.NameFlag, tt.input.name)
	}
	if tt.input.context != "" {
		args = append(args, "--"+constants.ContextFlag, tt.input.context)
	}
	if tt.input.command != "" {
		args = append(args, "--"+constants.CommandFlag, tt.input.command)
	}
	if tt.input.subPath != "" {
		args = append(args, "--"+constants.SubPathFlag, tt.input.subPath)
	}
	if tt.input.schedule != "" {
		args = append(args, "--"+constants.ScheduleFlag, tt.input.schedule)
	}
	if tt.input.nodepool != "" {
		args = append(args, "--"+constants.NodepoolFlag, tt.input.nodepool)
	}
	if tt.input.application != "" {
		args = append(args, "--"+constants.ApplicationFlag, tt.input.application)
	}
	for _, label := range tt.input.labels {
		args = append(args, "--"+constants.LabelFlag, label)
	}
	for _, env := range tt.input.envs {
		args = append(args, "--"+constants.EnvironmentFlag, env)
	}
	for _, workspaceID := range tt.input.workspaces {
		args = append(args, "--"+constants.WorkspaceFlag, workspaceID)
	}
	args = append(args, tt.input.longArgs...)
	args = append(args, tt.input.args...)
	return args
}

type jobEndpointsConfig struct {
	dataplaneConfig    stubs.StubHTTPPartialGetConfig
	applicationsConfig stubs.StubHTTPPartialGetConfig
	servicesConfig     stubs.StubHTTPPartialGetConfig
	nodepoolsConfig    stubs.StubHTTPPartialGetConfig
	workspaceConfig    stubs.StubHTTPPartialGetConfig
	jobConfig          stubs.StubHTTPPartialPostConfig
}

func getJobEndpoints(
	urlUnderTest string,
	config jobEndpointsConfig,
) []stubs.StubHTTPEndpoint {
	endpoints := append(stubs.GetStubAuthEndpoints(), stubs.GetStubWorkspaceEndpoints(config.workspaceConfig)...)
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
					ExpectedBody:    config.jobConfig.ExpectedRequestBody,
				},
				Response: stubs.StubHTTPResponse{
					Status: config.jobConfig.ResponseStatus,
					Body:   config.jobConfig.ResponseBody,
				},
			},
		}...)
}

func getCreateJobEndpoints(config jobEndpointsConfig) []stubs.StubHTTPEndpoint {
	return getJobEndpoints(postgrest.Endpoint(constants.WorkbenchEndpointRPCJobCreate), config)
}

func getUpdateJobEndpoints(config jobEndpointsConfig) []stubs.StubHTTPEndpoint {
	return getJobEndpoints(postgrest.Endpoint(constants.WorkbenchEndpointRPCJobBulkUpdate), config)
}
