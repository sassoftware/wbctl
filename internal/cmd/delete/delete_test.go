// Copyright © 2026, SAS Institute Inc., Cary, NC, USA.  All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package delete_test

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/sassoftware/wbctl/internal/cmd"
	cmddelete "github.com/sassoftware/wbctl/internal/cmd/delete"
	"github.com/sassoftware/wbctl/internal/constants"
	"github.com/sassoftware/wbctl/internal/postgrest"
	"github.com/sassoftware/wbctl/internal/stubs"
	"github.com/spf13/afero"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
)

const (
	image1ID                   = "image-1"
	image2ID                   = "image-2"
	job1ID                     = "job-1"
	job2ID                     = "job-2"
	snapshot1ID                = "snapshot-1"
	snapshot2ID                = "snapshot-2"
	submission1ID              = "submission-1"
	submission2ID              = "submission-2"
	workbench1ID               = "workbench-1"
	workbench2ID               = "workbench-2"
	workspace1ID               = "workspace-1"
	workspace2ID               = "workspace-2"
	expectedInvalidEndpointErr = "failed to get username/password token" +
		"\n Caused by:" +
		"\n\t* parse \"a\\nb\": net/url: invalid control character in URL"
)

var (
	expectedImageDisplay          = strings.Join([]string{image1ID, image2ID}, "\n")
	expectedJobDisplay            = strings.Join([]string{job1ID, job2ID}, "\n")
	expectedSnapshotDisplay       = strings.Join([]string{snapshot1ID, snapshot2ID}, "\n")
	expectedSubmissionDisplay     = strings.Join([]string{submission1ID, submission2ID}, "\n")
	expectedWorkbenchDisplay      = strings.Join([]string{workbench1ID, workbench2ID}, "\n")
	expectedWorkspaceDisplay      = strings.Join([]string{workspace1ID, workspace2ID}, "\n")
	expectedImageRequestBody      = `{"ids":["` + image1ID + `", "` + image2ID + `"]}`
	expectedJobRequestBody        = `{"ids":["` + job1ID + `", "` + job2ID + `"]}`
	expectedSnapshotRequestBody   = `{"ids":["` + snapshot1ID + `", "` + snapshot2ID + `"]}`
	expectedSubmissionRequestBody = `{"ids":["` + submission1ID + `", "` + submission2ID + `"]}`
	expectedWorkbenchRequestBody  = `{"ids":["` + workbench1ID + `", "` + workbench2ID + `"]}`
	expectedWorkspaceRequestBody  = `{"ids":["` + workspace1ID + `", "` + workspace2ID + `"]}`
	imageResponseBody             = `["` + image1ID + `", "` + image2ID + `"]`
	jobResponseBody               = `["` + job1ID + `", "` + job2ID + `"]`
	snapshotResponseBody          = `["` + snapshot1ID + `", "` + snapshot2ID + `"]`
	submissionResponseBody        = `["` + submission1ID + `", "` + submission2ID + `"]`
	workbenchResponseBody         = `["` + workbench1ID + `", "` + workbench2ID + `"]`
	workspaceResponseBody         = `["` + workspace1ID + `", "` + workspace2ID + `"]`
)

type deleteInput struct {
	endpoint     string
	resourceIDs  []string
	resourceType string
}

type deleteExpected struct {
	err         string
	displays    []string
	warns       []string
	deleteCalls int
}

type deleteTestCase struct {
	name          string
	input         deleteInput
	mockEndpoints []stubs.StubHTTPEndpoint
	expected      deleteExpected
}

func TestDeleteCommand(t *testing.T) {
	tests := []deleteTestCase{
		{
			name: "missing resource ids",
			input: deleteInput{
				resourceType: constants.ImageType,
			},
			mockEndpoints: stubs.GetStubAuthEndpoints(),
			expected: deleteExpected{
				err: "at least one resource ID must be provided for delete",
			},
		},
		{
			name: "delete non-200 from server",
			input: deleteInput{
				resourceType: constants.VolumesType,
				resourceIDs:  []string{workspace1ID},
			},
			mockEndpoints: getDeleteEndpoints(
				postgrest.Endpoint(constants.WorkbenchEndpointRPCWorkspaceBulkDelete),
				stubs.StubHTTPPartialPostConfig{
					ExpectedRequestBody: `{"ids":["workspace-1"]}`,
					ResponseStatus:      http.StatusBadRequest,
					ResponseBody:        `error`,
				},
			),
			expected: deleteExpected{
				err: "POST %s/postgrest/rpc/workspace_bulk_delete failed\n" +
					" Caused by:\n" +
					"\t* 400 error",
				deleteCalls: 1,
			},
		},
		{
			name: "deletes images by singular type",
			input: deleteInput{
				resourceType: constants.ImageType,
				resourceIDs:  []string{image1ID, image2ID},
			},
			mockEndpoints: getDeleteEndpoints(
				postgrest.Endpoint(constants.WorkbenchEndpointRPCImageBulkDelete),
				stubs.StubHTTPPartialPostConfig{
					ExpectedRequestBody: expectedImageRequestBody,
					ResponseStatus:      http.StatusOK,
					ResponseBody:        imageResponseBody,
				},
			),
			expected: deleteExpected{
				displays:    []string{expectedImageDisplay},
				deleteCalls: 1,
			},
		},
		{
			name: "deletes images by plural type",
			input: deleteInput{
				resourceType: constants.ImagesType,
				resourceIDs:  []string{image1ID, image2ID},
			},
			mockEndpoints: getDeleteEndpoints(
				postgrest.Endpoint(constants.WorkbenchEndpointRPCImageBulkDelete),
				stubs.StubHTTPPartialPostConfig{
					ExpectedRequestBody: expectedImageRequestBody,
					ResponseStatus:      http.StatusOK,
					ResponseBody:        imageResponseBody,
				},
			),
			expected: deleteExpected{
				displays:    []string{expectedImageDisplay},
				deleteCalls: 1,
			},
		},
		{
			name: "deletes job submissions by singular type",
			input: deleteInput{
				resourceType: constants.JobSubmissionType,
				resourceIDs:  []string{submission1ID, submission2ID},
			},
			mockEndpoints: getDeleteEndpoints(
				postgrest.Endpoint(constants.WorkbenchEndpointRPCJobRunBulkDelete),
				stubs.StubHTTPPartialPostConfig{
					ExpectedRequestBody: expectedSubmissionRequestBody,
					ResponseStatus:      http.StatusOK,
					ResponseBody:        submissionResponseBody,
				},
			),
			expected: deleteExpected{
				displays:    []string{expectedSubmissionDisplay},
				deleteCalls: 1,
			},
		},
		{
			name: "deletes job submissions by plural type",
			input: deleteInput{
				resourceType: constants.JobSubmissionsType,
				resourceIDs:  []string{submission1ID, submission2ID},
			},
			mockEndpoints: getDeleteEndpoints(
				postgrest.Endpoint(constants.WorkbenchEndpointRPCJobRunBulkDelete),
				stubs.StubHTTPPartialPostConfig{
					ExpectedRequestBody: expectedSubmissionRequestBody,
					ResponseStatus:      http.StatusOK,
					ResponseBody:        submissionResponseBody,
				},
			),
			expected: deleteExpected{
				displays:    []string{expectedSubmissionDisplay},
				deleteCalls: 1,
			},
		},
		{
			name: "deletes jobs by singular type",
			input: deleteInput{
				resourceType: constants.JobType,
				resourceIDs:  []string{job1ID, job2ID},
			},
			mockEndpoints: getDeleteEndpoints(
				postgrest.Endpoint(constants.WorkbenchEndpointRPCJobBulkDelete),
				stubs.StubHTTPPartialPostConfig{
					ExpectedRequestBody: expectedJobRequestBody,
					ResponseStatus:      http.StatusOK,
					ResponseBody:        jobResponseBody,
				},
			),
			expected: deleteExpected{
				displays:    []string{expectedJobDisplay},
				deleteCalls: 1,
			},
		},
		{
			name: "deletes jobs by plural type",
			input: deleteInput{
				resourceType: constants.JobsType,
				resourceIDs:  []string{job1ID, job2ID},
			},
			mockEndpoints: getDeleteEndpoints(
				postgrest.Endpoint(constants.WorkbenchEndpointRPCJobBulkDelete),
				stubs.StubHTTPPartialPostConfig{
					ExpectedRequestBody: expectedJobRequestBody,
					ResponseStatus:      http.StatusOK,
					ResponseBody:        jobResponseBody,
				},
			),
			expected: deleteExpected{
				displays:    []string{expectedJobDisplay},
				deleteCalls: 1,
			},
		},
		{
			name: "deletes snapshots by singular type",
			input: deleteInput{
				resourceType: constants.SnapshotType,
				resourceIDs:  []string{snapshot1ID, snapshot2ID},
			},
			mockEndpoints: getDeleteEndpoints(
				postgrest.Endpoint(constants.WorkbenchEndpointRPCSnapshotBulkDelete),
				stubs.StubHTTPPartialPostConfig{
					ExpectedRequestBody: expectedSnapshotRequestBody,
					ResponseStatus:      http.StatusOK,
					ResponseBody:        snapshotResponseBody,
				},
			),
			expected: deleteExpected{
				displays:    []string{expectedSnapshotDisplay},
				deleteCalls: 1,
			},
		},
		{
			name: "deletes snapshots by plural type",
			input: deleteInput{
				resourceType: constants.SnapshotsType,
				resourceIDs:  []string{snapshot1ID, snapshot2ID},
			},
			mockEndpoints: getDeleteEndpoints(
				postgrest.Endpoint(constants.WorkbenchEndpointRPCSnapshotBulkDelete),
				stubs.StubHTTPPartialPostConfig{
					ExpectedRequestBody: expectedSnapshotRequestBody,
					ResponseStatus:      http.StatusOK,
					ResponseBody:        snapshotResponseBody,
				},
			),
			expected: deleteExpected{
				displays:    []string{expectedSnapshotDisplay},
				deleteCalls: 1,
			},
		},
		{
			name: "deletes workbenches by singular type",
			input: deleteInput{
				resourceType: constants.WorkbenchType,
				resourceIDs:  []string{workbench1ID, workbench2ID},
			},
			mockEndpoints: getDeleteEndpoints(
				postgrest.Endpoint(constants.WorkbenchEndpointRPCWorkbenchBulkDelete),
				stubs.StubHTTPPartialPostConfig{
					ExpectedRequestBody: expectedWorkbenchRequestBody,
					ResponseStatus:      http.StatusOK,
					ResponseBody:        workbenchResponseBody,
				},
			),
			expected: deleteExpected{
				displays:    []string{expectedWorkbenchDisplay},
				deleteCalls: 1,
			},
		},
		{
			name: "deletes workbenches by plural type",
			input: deleteInput{
				resourceType: constants.WorkbenchesType,
				resourceIDs:  []string{workbench1ID, workbench2ID},
			},
			mockEndpoints: getDeleteEndpoints(
				postgrest.Endpoint(constants.WorkbenchEndpointRPCWorkbenchBulkDelete),
				stubs.StubHTTPPartialPostConfig{
					ExpectedRequestBody: expectedWorkbenchRequestBody,
					ResponseStatus:      http.StatusOK,
					ResponseBody:        workbenchResponseBody,
				},
			),
			expected: deleteExpected{
				displays:    []string{expectedWorkbenchDisplay},
				deleteCalls: 1,
			},
		},
		{
			name: "deletes volumes by singular type",
			input: deleteInput{
				resourceType: constants.VolumeType,
				resourceIDs:  []string{workspace1ID, workspace2ID},
			},
			mockEndpoints: getDeleteEndpoints(
				postgrest.Endpoint(constants.WorkbenchEndpointRPCWorkspaceBulkDelete),
				stubs.StubHTTPPartialPostConfig{
					ExpectedRequestBody: expectedWorkspaceRequestBody,
					ResponseStatus:      http.StatusOK,
					ResponseBody:        workspaceResponseBody,
				},
			),
			expected: deleteExpected{
				displays:    []string{expectedWorkspaceDisplay},
				deleteCalls: 1,
			},
		},
		{
			name: "deletes volumes by plural type",
			input: deleteInput{
				resourceType: constants.VolumesType,
				resourceIDs:  []string{workspace1ID, workspace2ID},
			},
			mockEndpoints: getDeleteEndpoints(
				postgrest.Endpoint(constants.WorkbenchEndpointRPCWorkspaceBulkDelete),
				stubs.StubHTTPPartialPostConfig{
					ExpectedRequestBody: expectedWorkspaceRequestBody,
					ResponseStatus:      http.StatusOK,
					ResponseBody:        workspaceResponseBody,
				},
			),
			expected: deleteExpected{
				displays:    []string{expectedWorkspaceDisplay},
				deleteCalls: 1,
			},
		},
		{
			name: "invalid endpoint",
			input: deleteInput{
				resourceType: constants.VolumeType,
				resourceIDs:  []string{workbench1ID, workbench2ID},
				endpoint:     "a\nb",
			},
			expected: deleteExpected{
				err: expectedInvalidEndpointErr,
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
			mockCmd := cmddelete.Command(context.Background(), afero.NewMemMapFs(), stubUI)
			cmd.AddRootFlags(mockCmd)
			mockCmd.SetArgs(getDeleteArgs(tt))
			err := mockCmd.Execute()
			if tt.expected.err != "" {
				if strings.Contains(tt.expected.err, "%s") {
					expectedErr := fmt.Sprintf(tt.expected.err, srv.URL)
					require.EqualError(t, err, expectedErr)
				} else {
					require.EqualError(t, err, tt.expected.err)
				}
			} else {
				require.NoError(t, err)
			}

			require.Equal(t, tt.expected.displays, stubUI.Displays, strings.Join(stubUI.Displays, "\n"))
			require.Equal(t, tt.expected.warns, stubUI.Warns, strings.Join(stubUI.Warns, "\n"))
			require.Empty(t, stubUI.Tables)
			require.Equal(
				t,
				tt.expected.deleteCalls,
				hitCounts.CountPrefix("POST"+constants.PostgrestAPIEndpoint),
				"delete calls",
			)
		})
	}
}

func TestSupportedResourceTypes(t *testing.T) {
	actual := cmddelete.SupportedResourceTypes()

	require.Equal(t, []string{
		constants.ImagesType,
		constants.JobsType,
		constants.JobSubmissionsType,
		constants.SnapshotsType,
		constants.VolumesType,
		constants.WorkbenchesType,
	}, actual)
}

func getDeleteArgs(
	tt deleteTestCase,
) []string {
	args := []string{tt.input.resourceType}
	args = append(args, "--"+constants.AuthModeFlag, constants.AuthModePassword)
	args = append(args, "--"+constants.ClientIDFlag, constants.OAuthClientID)
	args = append(args, "--"+constants.PasswordFlag, "pass")
	args = append(args, "--"+constants.UsernameFlag, "user")
	if tt.input.endpoint != "" {
		args = append(args, "--"+constants.EndpointFlag, tt.input.endpoint)
	}
	args = append(args, tt.input.resourceIDs...)
	return args
}

func getDeleteEndpoints(
	urlUnderTest string,
	config ...stubs.StubHTTPPartialPostConfig,
) []stubs.StubHTTPEndpoint {
	endpoints := stubs.GetStubAuthEndpoints()
	for _, cfg := range config {
		endpoints = append(endpoints,
			stubs.StubHTTPEndpoint{
				Request: stubs.StubHTTPRequest{
					Method:          http.MethodPost,
					URL:             urlUnderTest,
					ExpectedHeaders: stubs.StubAuthHeaders,
					ExpectedBody:    cfg.ExpectedRequestBody,
				},
				Response: stubs.StubHTTPResponse{
					Status: cfg.ResponseStatus,
					Body:   cfg.ResponseBody,
				},
			})
	}
	return endpoints
}
