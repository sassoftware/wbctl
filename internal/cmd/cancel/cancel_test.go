// Copyright © 2026, SAS Institute Inc., Cary, NC, USA.  All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package cancel_test

import (
	"context"
	"net/http"
	"strings"
	"testing"

	"github.com/sassoftware/wbctl/internal/cmd"
	"github.com/sassoftware/wbctl/internal/cmd/cancel"
	"github.com/sassoftware/wbctl/internal/constants"
	"github.com/sassoftware/wbctl/internal/postgrest"
	"github.com/sassoftware/wbctl/internal/stubs"
	"github.com/spf13/afero"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
)

const cancelJobRunsResponse = `[
  {
    "id": "submission-1",
    "controller_id": "controller-1"
  },
  {
    "id": "submission-2",
    "controller_id": "controller-2"
  }
]`

const (
	cancelTestControllerID1        = "controller-1"
	cancelTestControllerID2        = "controller-2"
	cancelTestGetJobSubmissionsErr = "" +
		"GET ${URL}/postgrest/job_runs?controller_id=in.%28controller-1%2Ccontroller-2%29 failed" +
		"\n Caused by:" +
		"\n\t* unexpected EOF"
	cancelTestBulkCancelErr = "" +
		"POST ${URL}/postgrest/rpc/job_run_bulk_cancel failed" +
		"\n Caused by:" +
		"\n\t* unexpected EOF"
	cancelTestOneCall                       = 1
	cancelExpectedDisplaySubmission1        = "submission-1"
	cancelExpectedDisplaySubmission2        = "submission-2"
	cancelExpectedBulkCancelResponseBody    = `["submission-1","submission-2"]`
	cancelExpectedBulkCancelInvalidJSONBody = `[`
	cancelExpectedGetInvalidJSONBody        = `[`
	expectedInvalidEndpointErr              = "failed to get username/password token" +
		"\n Caused by:" +
		"\n\t* parse \"a\\nb\": net/url: invalid control character in URL"
	cancelExpectedGetQuery = "controller_id=in.(controller-1,controller-2)"
)

var cancelExpectedDisplaysStoppedIDs = []string{
	cancelExpectedDisplaySubmission1,
	cancelExpectedDisplaySubmission2,
}

var cancelExpectedNoWarns []string

func TestSupportedResourceTypes(t *testing.T) {
	require.Equal(t, []string{constants.JobSubmissionsType}, cancel.SupportedResourceTypes())
}

type cancelInput struct {
	endpoint     string
	resourceType string
	resourceIDs  []string
}

type cancelExpected struct {
	err       string
	displays  []string
	warns     []string
	getCalls  int
	postCalls int
}

type cancelTestCase struct {
	name          string
	input         cancelInput
	mockEndpoints []stubs.StubHTTPEndpoint
	expected      cancelExpected
}

func TestCancelCommand(t *testing.T) {
	tests := []cancelTestCase{
		{
			name: "invalid endpoint",
			input: cancelInput{
				endpoint:     "a\nb",
				resourceType: constants.JobSubmissionType,
				resourceIDs:  []string{cancelTestControllerID1},
			},
			expected: cancelExpected{err: expectedInvalidEndpointErr},
		},
		{
			name: "job submission get error",
			input: cancelInput{
				resourceType: constants.JobSubmissionType,
				resourceIDs:  []string{cancelTestControllerID1, cancelTestControllerID2},
			},
			mockEndpoints: getJobRunCancelEndpoints(
				cancelExpectedGetQuery,
				cancelExpectedGetInvalidJSONBody,
				cancelExpectedBulkCancelResponseBody,
			),
			expected: cancelExpected{
				err:      cancelTestGetJobSubmissionsErr,
				getCalls: cancelTestOneCall,
			},
		},
		{
			name: "job submission bulk cancel error",
			input: cancelInput{
				resourceType: constants.JobSubmissionType,
				resourceIDs:  []string{cancelTestControllerID1, cancelTestControllerID2},
			},
			mockEndpoints: getJobRunCancelEndpoints(
				cancelExpectedGetQuery,
				cancelJobRunsResponse,
				cancelExpectedBulkCancelInvalidJSONBody,
			),
			expected: cancelExpected{
				err:       cancelTestBulkCancelErr,
				getCalls:  cancelTestOneCall,
				postCalls: cancelTestOneCall,
			},
		},
		{
			name: "job submission cancel prints stopped ids",
			input: cancelInput{
				resourceType: constants.JobSubmissionType,
				resourceIDs:  []string{cancelTestControllerID1, cancelTestControllerID2},
			},
			mockEndpoints: getJobRunCancelEndpoints(
				cancelExpectedGetQuery,
				cancelJobRunsResponse,
				cancelExpectedBulkCancelResponseBody,
			),
			expected: cancelExpected{
				displays:  cancelExpectedDisplaysStoppedIDs,
				warns:     cancelExpectedNoWarns,
				getCalls:  cancelTestOneCall,
				postCalls: cancelTestOneCall,
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
			mockCmd := cancel.Command(context.Background(), afero.NewMemMapFs(), stubUI)
			cmd.AddRootFlags(mockCmd)
			mockCmd.SetArgs(getCancelArgs(tt))
			err := mockCmd.Execute()
			if tt.expected.err != "" {
				expectedErr := strings.ReplaceAll(tt.expected.err, "${URL}", srv.URL)
				require.EqualError(t, err, expectedErr)
			} else {
				require.NoError(t, err)
			}

			require.Equal(t, tt.expected.getCalls, hitCounts.Count("GET"+
				postgrest.Endpoint(constants.WorkbenchEndpointJobRuns)))
			require.Equal(t, tt.expected.postCalls, hitCounts.Count("POST"+
				postgrest.Endpoint(constants.WorkbenchEndpointRPCJobRunBulkCancel)))

			require.Equal(t, tt.expected.displays, stubUI.Displays, strings.Join(stubUI.Displays, "\n"))
			require.Equal(t, tt.expected.warns, stubUI.Warns, strings.Join(stubUI.Warns, "\n"))
		})
	}
}

func getCancelArgs(
	tt cancelTestCase,
) []string {
	args := []string{}
	args = append(args, tt.input.resourceType)
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

func getJobRunCancelEndpoints(
	expectedJobRunsQuery string,
	jobRunsResponseBody string,
	cancelResponseBody string,
) []stubs.StubHTTPEndpoint {
	return append(stubs.GetStubAuthEndpoints(),
		[]stubs.StubHTTPEndpoint{
			{
				Request: stubs.StubHTTPRequest{
					Method:          http.MethodGet,
					URL:             postgrest.Endpoint(constants.WorkbenchEndpointJobRuns),
					ExpectedHeaders: stubs.StubAuthHeaders,
					ExpectedQuery:   expectedJobRunsQuery,
				},
				Response: stubs.StubHTTPResponse{
					Body:   jobRunsResponseBody,
					Status: http.StatusOK,
				},
			},
			{
				Request: stubs.StubHTTPRequest{
					Method:          http.MethodPost,
					URL:             postgrest.Endpoint(constants.WorkbenchEndpointRPCJobRunBulkCancel),
					ExpectedHeaders: stubs.StubAuthHeaders,
					ExpectedBody:    `{"ids":["submission-1","submission-2"]}`,
				},
				Response: stubs.StubHTTPResponse{
					Body:   cancelResponseBody,
					Status: http.StatusOK,
				},
			},
		}...)
}
