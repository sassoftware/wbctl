// Copyright © 2026, SAS Institute Inc., Cary, NC, USA.  All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package workbench_test

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/sassoftware/wbctl/internal/constants"
	"github.com/sassoftware/wbctl/internal/postgrest"
	"github.com/sassoftware/wbctl/internal/services/workbench"
	"github.com/sassoftware/wbctl/internal/stubs"
	"github.com/sassoftware/wbctl/internal/timeutil"
	"github.com/sassoftware/wbctl/internal/types"
	"github.com/stretchr/testify/require"
)

const (
	jobSubmissionsEmptyResourcesJSON = `[]`

	jobSubmissionsGetOneResponseJSON = `
[
  {
    "id": "js-1",
    "controller_id": "ctrl-1",
    "job_name": "daily-job",
    "username": "alice",
    "current_state": "Succeeded",
    "message": "done",
    "runtime": 12.34,
    "log_url": "https://logs/1"
  }
]
`

	jobSubmissionsGetMixedIDsResponseJSON = `
[
  {
    "id": "123e4567-e89b-12d3-a456-426614174000",
    "controller_id": "ctrl-uuid",
    "job_name": "uuid-job",
    "username": "alice",
    "current_state": "Running",
    "runtime": 1.25,
    "log_url": "https://logs/uuid"
  },
  {
    "id": "js-controller",
    "controller_id": "controller-job-2",
    "job_name": "controller-job",
    "username": "bob",
    "current_state": "Queued",
    "runtime": 0.00,
    "log_url": "https://logs/controller"
  }
]
`

	jobSubmissionsPrintOneResourcesJSON = `
[
  {
    "id": "js-1",
    "controller_id": "",
    "job_name": "daily-job",
    "username": "alice",
    "current_state": "Succeeded",
    "message": "done",
    "runtime": 12.34,
    "log_url": "https://logs/1",
    "created_at": null
  }
]
`

	jobSubmissionsPrintTwoResourcesJSON = `
[
  {
    "id": "js-1",
    "controller_id": "",
    "job_name": "daily-job",
    "username": "alice",
    "current_state": "Succeeded",
    "message": "done",
    "runtime": 12.34,
    "log_url": "https://logs/1",
    "created_at": null
  },
  {
    "id": "js-2",
    "controller_id": "ctrl-2",
    "job_name": "weekly-job",
    "username": "bob",
    "current_state": "Running",
    "message": "working",
    "runtime": 1.20,
    "log_url": "https://logs/2",
    "created_at": null
  }
]
`

	jobSubmissionsDescribeOneResourcesJSON = `
[
  {
    "id": "js-1",
    "controller_id": "ctrl-1",
    "job_name": "daily-job",
    "job_id": "job-1",
    "url": "https://job/1",
    "log_url": "https://logs/1",
    "username": "alice",
    "email": "alice@example.com",
    "created_at": "2024-01-02T03:04:05Z",
    "modified_at": "2024-01-02T04:05:06Z",
    "runtime": 12.34,
    "nodepool": "np-a",
    "scheduled": true,
    "current_state": "Succeeded",
    "message": "done"
  }
]
`

	jobSubmissionsDescribeTwoResourcesJSON = `
[
  {
    "id": "js-1",
    "controller_id": "ctrl-1",
    "job_name": "daily-job",
    "job_id": "job-1",
    "url": "https://job/1",
    "log_url": "https://logs/1",
    "username": "alice",
    "email": "alice@example.com",
    "created_at": "2024-01-02T03:04:05Z",
    "modified_at": "2024-01-02T04:05:06Z",
    "runtime": 12.34,
    "nodepool": "np-a",
    "scheduled": true,
    "current_state": "Succeeded",
    "message": "done"
  },
  {
    "id": "js-2",
    "controller_id": "ctrl-2",
    "job_name": "daily-job2",
    "job_id": "job-2",
    "url": "https://job/2",
    "log_url": "https://logs/2",
    "username": "bob",
    "email": "bob@example.com",
    "created_at": "2024-01-02T03:04:05Z",
    "modified_at": "2024-01-02T04:05:06Z",
    "runtime": 12.34,
    "nodepool": "np-a",
    "scheduled": true,
    "current_state": "Succeeded",
    "message": "done"
  }
]
`

	jobSubmissionsDescribeWithSectionsResourcesJSON = `
[
  {
    "id": "js-2",
    "controller_id": "ctrl-2",
    "job_name": "weekly-job",
    "job_id": "job-2",
    "url": "https://job/2",
    "log_url": "https://logs/2",
    "username": "bob",
    "email": "bob@example.com",
    "created_at": "2024-01-03T03:04:05Z",
    "modified_at": "2024-01-03T04:05:06Z",
    "canceled_at": "2024-01-04T05:06:07Z",
    "completed_at": "2024-01-04T06:07:08Z",
    "runtime": 1.2,
    "scheduled": false,
    "current_state": "Failed",
    "message": "error",
    "labels": {
      "run": "nightly"
    },
    "env": {
      "ENV_A": "value-a"
    },
    "workspaces": [
      {
        "id": "ws-1",
        "mountName": "data"
      },
      {
        "id": "ws-2",
        "mountName": "data",
        "ephemeralPVC": {
          "accessModes": [
            "ReadWriteOnce",
            "ReadWriteMany"
          ],
          "storageClassName": "ebs",
          "resources": {
            "requests": {
              "storage": "1G"
            }
          }
        }
      }
    ],
    "properties": {
      "reason": "quota"
    },
    "pod_status": [
      {
        "name": "pod-1",
        "status": {
          "phase": "Running",
          "message": "pod-ok",
          "reason": "scheduled",
          "nominatedNodeName": "node-a",
          "podIP": "10.0.0.1"
        }
      }
    ]
  }
]
`

	expectedJobSubmissionsEmptyTable = "ID\tCONTROLLER_ID\tJOB_NAME\tLOG_URL\tSTATE\tMESSAGE\tRUNTIME\tAGE"

	expectedJobSubmissionsOneRowTable = "ID\tCONTROLLER_ID\tJOB_NAME\tLOG_URL\tSTATE\tMESSAGE\tRUNTIME\tAGE\n" +
		"js-1\t<unknown>\tdaily-job\thttps://logs/1\tSucceeded\tdone\t12.34s\t"

	expectedJobSubmissionsTwoRowsTable = "ID\tCONTROLLER_ID\tJOB_NAME\tLOG_URL\tSTATE\tMESSAGE\tRUNTIME\tAGE\n" +
		"js-1\t<unknown>\tdaily-job\thttps://logs/1\tSucceeded\tdone\t12.34s\t\n" +
		"js-2\tctrl-2\tweekly-job\thttps://logs/2\tRunning\tworking\t1.20s\t"

	expectedJobSubmissionsAdminEmptyTable = "ID\tCONTROLLER_ID\tJOB_NAME\tUSERNAME\tLOG_URL\tSTATE\tMESSAGE\t" +
		"RUNTIME\tAGE"

	expectedJobSubmissionsAdminOneRowTable = "ID\tCONTROLLER_ID\tJOB_NAME\tUSERNAME\tLOG_URL\tSTATE\tMESSAGE\t" +
		"RUNTIME\tAGE\n" +
		"js-1\t<unknown>\tdaily-job\talice\thttps://logs/1\tSucceeded\tdone\t12.34s\t"

	expectedDescribeJobSubmissionOneDisplay = `ID:              js-1
Controller ID:   ctrl-1
Job Name:        daily-job
Job ID:          job-1
URL:             https://job/1
Log URL:         https://logs/1
Created By:      alice (alice@example.com)
Created At:      2024-01-02T03:04:05Z (2y126d ago)
Modified At:     2024-01-02T04:05:06Z (2y126d ago)
Runtime:         12.340000
Nodepool:        np-a
Scheduled:       true
State:           Succeeded
Message:         done
`

	expectedDescribeJobSubmissionTwoDisplay = `ID:              js-1
Controller ID:   ctrl-1
Job Name:        daily-job
Job ID:          job-1
URL:             https://job/1
Log URL:         https://logs/1
Created By:      alice (alice@example.com)
Created At:      2024-01-02T03:04:05Z (2y126d ago)
Modified At:     2024-01-02T04:05:06Z (2y126d ago)
Runtime:         12.340000
Nodepool:        np-a
Scheduled:       true
State:           Succeeded
Message:         done


ID:              js-2
Controller ID:   ctrl-2
Job Name:        daily-job2
Job ID:          job-2
URL:             https://job/2
Log URL:         https://logs/2
Created By:      bob (bob@example.com)
Created At:      2024-01-02T03:04:05Z (2y126d ago)
Modified At:     2024-01-02T04:05:06Z (2y126d ago)
Runtime:         12.340000
Nodepool:        np-a
Scheduled:       true
State:           Succeeded
Message:         done
`

	expectedDescribeJobSubmissionWithSectionsDisplay = `ID:              js-2
Controller ID:   ctrl-2
Job Name:        weekly-job
Job ID:          job-2
URL:             https://job/2
Log URL:         https://logs/2
Created By:      bob (bob@example.com)
Created At:      2024-01-03T03:04:05Z (2y125d ago)
Modified At:     2024-01-03T04:05:06Z (2y125d ago)
Canceled At:     2024-01-04T05:06:07Z (2y124d ago)
Completed At:    2024-01-04T06:07:08Z (2y124d ago)
Runtime:         1.200000
Scheduled:       false
State:           Failed
Message:         error
Labels:
  run:   nightly
Environment:
  ENV_A:   value-a
Volumes:
  0:
    ID:           ws-1
    Mount name:   data
  1:
    ID:           ws-2
    Mount name:   data
    Ephemeral PVC:
      Access Modes:    ReadWriteOnce, ReadWriteMany
      Storage Class:   ebs
      Size:            1G
Properties:
  reason:   quota
Pod pod-1:
  Phase:            Running
  Message:          pod-ok
  Reason:           scheduled
  Nominated node:   node-a
  IPs:              10.0.0.1
`
	expectedDeleteJobRunRequestJSON = `{"ids":["js-1", "js-2"]}`
	expectedCreateJobRunRequestJSON = `{"param":{"job_name":"daily-job"}}`
	createJobRunOneResponseBodyJSON = `{"id":"js-1","job_name":"daily-job"}`
)

var (
	expectedResponseOneSubmission = types.JobSubmissions{
		{
			ID:           "js-1",
			ControllerID: "ctrl-1",
			JobName:      "daily-job",
			Username:     "alice",
			CurrentState: "Succeeded",
			Message:      "done",
			Runtime:      12.34,
			LogURL:       "https://logs/1",
		},
	}

	expectedResponseTwoSubmissions = types.JobSubmissions{
		{
			ID:           "123e4567-e89b-12d3-a456-426614174000",
			ControllerID: "ctrl-uuid",
			JobName:      "uuid-job",
			Username:     "alice",
			CurrentState: "Running",
			Runtime:      1.25,
			LogURL:       "https://logs/uuid",
		},
		{
			ID:           "js-controller",
			ControllerID: "controller-job-2",
			JobName:      "controller-job",
			Username:     "bob",
			CurrentState: "Queued",
			Runtime:      0,
			LogURL:       "https://logs/controller",
		},
	}
)

func TestDeleteJobSubmission(t *testing.T) {
	tests := []deleteTestCase{
		{
			name: "delete job-submission success",
			request: deleteRequest{
				ids: []string{"js-1", "js-2"},
			},
			mockEndpoints: getJobSubmissionDeleteEndpoints(
				expectedDeleteJobRunRequestJSON,
				http.StatusOK,
				`["js-1", "js-2"]`,
			),
			expected: deleteExpected{
				result: []string{"js-1", "js-2"},
			},
		},
		{
			name: "delete job-submission non-200 from server",
			request: deleteRequest{
				ids: []string{"js-1", "js-2"},
			},
			mockEndpoints: getJobSubmissionDeleteEndpoints(
				expectedDeleteJobRunRequestJSON,
				http.StatusUnauthorized,
				"unauthorized",
			),
			expected: deleteExpected{
				err: "POST ${URL}/postgrest/rpc/job_run_bulk_delete failed\n" +
					" Caused by:\n" +
					"\t* 401 unauthorized",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv, _ := stubs.NewStubHTTPServer(t, tt.mockEndpoints)
			defer srv.Close()
			svc := workbench.NewWorkbenchService(srv.URL, "access-token", "csrf-token")

			actual, err := workbench.BulkDeleteJobSubmissions(context.Background(), svc, tt.request.ids)
			if tt.expected.err != "" {
				require.EqualError(t, err, strings.ReplaceAll(tt.expected.err, "${URL}", srv.URL))
				return
			}

			require.NoError(t, err)
			require.Equal(t, tt.expected.result, actual)
		})
	}
}

func TestCreateJobSubmission(t *testing.T) {
	tests := []createTestCase[types.JobSubmission]{
		{
			name: "create job-submission success",
			request: createRequest[types.JobSubmission]{
				body: types.JobSubmission{
					JobName: "daily-job",
				},
			},
			mockEndpoints: getJobSubmissionCreateEndpoints(
				expectedCreateJobRunRequestJSON,
				http.StatusOK,
				createJobRunOneResponseBodyJSON,
			),
			expected: createExpected[types.JobSubmission]{
				result: types.JobSubmission{
					ID:      "js-1",
					JobName: "daily-job",
				},
			},
		},
		{
			name: "create job-submission non-200 from server",
			request: createRequest[types.JobSubmission]{
				body: types.JobSubmission{
					JobName: "daily-job",
				},
			},
			mockEndpoints: getJobSubmissionCreateEndpoints(
				expectedCreateJobRunRequestJSON,
				http.StatusUnauthorized,
				"unauthorized",
			),
			expected: createExpected[types.JobSubmission]{
				err: "POST ${URL}/postgrest/rpc/job_run_create failed\n" +
					" Caused by:\n" +
					"\t* 401 unauthorized",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv, _ := stubs.NewStubHTTPServer(t, tt.mockEndpoints)
			defer srv.Close()
			svc := workbench.NewWorkbenchService(srv.URL, "access-token", "csrf-token")

			actual, err := workbench.CreateJobSubmission(context.Background(), svc, tt.request.body)
			if tt.expected.err != "" {
				require.EqualError(t, err, strings.ReplaceAll(tt.expected.err, "${URL}", srv.URL))
				return
			}

			require.NoError(t, err)
			require.Equal(t, tt.expected.result, actual)
		})
	}
}

func TestGetJobSubmissions(t *testing.T) {
	tests := []getTestCase[types.JobSubmissions]{
		{
			name:    "success without ids query",
			request: getRequest{},
			mockEndpoints: getJobSubmissionsGetEndpoints(
				"",
				http.StatusOK,
				jobSubmissionsGetOneResponseJSON,
			),
			expected: getExpected[types.JobSubmissions]{
				result: expectedResponseOneSubmission,
			},
		},
		{
			name: "sort by job_name asc",
			request: getRequest{
				sort: "job_name.asc",
			},
			mockEndpoints: getJobSubmissionsGetEndpoints(
				"order=job_name.asc",
				http.StatusOK,
				jobSubmissionsGetOneResponseJSON,
			),
			expected: getExpected[types.JobSubmissions]{
				result: expectedResponseOneSubmission,
			},
		},
		{
			name: "sort by job_name desc",
			request: getRequest{
				sort: "job_name.desc",
			},
			mockEndpoints: getJobSubmissionsGetEndpoints(
				"order=job_name.desc",
				http.StatusOK,
				jobSubmissionsGetOneResponseJSON,
			),
			expected: getExpected[types.JobSubmissions]{
				result: expectedResponseOneSubmission,
			},
		},
		{
			name: "sort by id",
			request: getRequest{
				sort: "id",
			},
			mockEndpoints: getJobSubmissionsGetEndpoints(
				"order=id",
				http.StatusOK,
				jobSubmissionsGetOneResponseJSON,
			),
			expected: getExpected[types.JobSubmissions]{
				result: expectedResponseOneSubmission,
			},
		},
		{
			name: "sort by controller_id",
			request: getRequest{
				sort: "controller_id",
			},
			mockEndpoints: getJobSubmissionsGetEndpoints(
				"order=controller_id",
				http.StatusOK,
				jobSubmissionsGetOneResponseJSON,
			),
			expected: getExpected[types.JobSubmissions]{
				result: expectedResponseOneSubmission,
			},
		},
		{
			name: "sort by job_name",
			request: getRequest{
				sort: "job_name",
			},
			mockEndpoints: getJobSubmissionsGetEndpoints(
				"order=job_name",
				http.StatusOK,
				jobSubmissionsGetOneResponseJSON,
			),
			expected: getExpected[types.JobSubmissions]{
				result: expectedResponseOneSubmission,
			},
		},
		{
			name: "sort by username",
			request: getRequest{
				sort: "username",
			},
			mockEndpoints: getJobSubmissionsGetEndpoints(
				"order=username",
				http.StatusOK,
				jobSubmissionsGetOneResponseJSON,
			),
			expected: getExpected[types.JobSubmissions]{
				result: expectedResponseOneSubmission,
			},
		},
		{
			name: "sort by log_url",
			request: getRequest{
				sort: "log_url",
			},
			mockEndpoints: getJobSubmissionsGetEndpoints(
				"order=log_url",
				http.StatusOK,
				jobSubmissionsGetOneResponseJSON,
			),
			expected: getExpected[types.JobSubmissions]{
				result: expectedResponseOneSubmission,
			},
		},
		{
			name: "sort by state",
			request: getRequest{
				sort: "state",
			},
			mockEndpoints: getJobSubmissionsGetEndpoints(
				"order=current_state",
				http.StatusOK,
				jobSubmissionsGetOneResponseJSON,
			),
			expected: getExpected[types.JobSubmissions]{
				result: expectedResponseOneSubmission,
			},
		},
		{
			name: "sort by message",
			request: getRequest{
				sort: "message",
			},
			mockEndpoints: getJobSubmissionsGetEndpoints(
				"order=message",
				http.StatusOK,
				jobSubmissionsGetOneResponseJSON,
			),
			expected: getExpected[types.JobSubmissions]{
				result: expectedResponseOneSubmission,
			},
		},
		{
			name: "sort by runtime",
			request: getRequest{
				sort: "runtime",
			},
			mockEndpoints: getJobSubmissionsGetEndpoints(
				"order=runtime",
				http.StatusOK,
				jobSubmissionsGetOneResponseJSON,
			),
			expected: getExpected[types.JobSubmissions]{
				result: expectedResponseOneSubmission,
			},
		},
		{
			name: "sort by age",
			request: getRequest{
				sort: "created_at",
			},
			mockEndpoints: getJobSubmissionsGetEndpoints(
				"order=created_at",
				http.StatusOK,
				jobSubmissionsGetOneResponseJSON,
			),
			expected: getExpected[types.JobSubmissions]{
				result: expectedResponseOneSubmission,
			},
		},
		{
			name: "sort by age and state",
			request: getRequest{
				sort: "age.desc,state.asc",
			},
			mockEndpoints: getJobSubmissionsGetEndpoints(
				"order=created_at.desc,current_state.asc",
				http.StatusOK,
				jobSubmissionsGetOneResponseJSON,
			),
			expected: getExpected[types.JobSubmissions]{
				result: expectedResponseOneSubmission,
			},
		},
		{
			name: "success with mixed uuid and controller ids query",
			request: getRequest{
				ids: []string{"123e4567-e89b-12d3-a456-426614174000", "controller-job-2"},
			},
			mockEndpoints: getJobSubmissionsGetEndpoints(
				"or=(id.in.(123e4567-e89b-12d3-a456-426614174000),controller_id.in.(controller-job-2))",
				http.StatusOK,
				jobSubmissionsGetMixedIDsResponseJSON,
			),
			expected: getExpected[types.JobSubmissions]{
				result: expectedResponseTwoSubmissions,
			},
		},
		{
			name: "invalid json response",
			request: getRequest{
				ids: []string{"id-a"},
			},
			mockEndpoints: getJobSubmissionsGetEndpoints(
				"controller_id=in.(id-a)",
				http.StatusOK,
				"[",
			),
			expected: getExpected[types.JobSubmissions]{
				err: "GET ${URL}/postgrest/job_runs?controller_id=in.%28id-a%29 failed\n" +
					" Caused by:\n" +
					"\t* unexpected EOF",
			},
		},
		{
			name: "non-200 from server",
			request: getRequest{
				ids: []string{"id-a"},
			},
			mockEndpoints: getJobSubmissionsGetEndpoints(
				"controller_id=in.(id-a)",
				http.StatusUnauthorized,
				"unauthorized",
			),
			expected: getExpected[types.JobSubmissions]{
				err: "GET ${URL}/postgrest/job_runs?controller_id=in.%28id-a%29 failed\n" +
					" Caused by:\n" +
					"\t* 401 unauthorized",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv, _ := stubs.NewStubHTTPServer(t, tt.mockEndpoints)
			defer srv.Close()
			svc := workbench.NewWorkbenchService(srv.URL, "access-token", "csrf-token")

			actual, err := workbench.GetJobSubmissions(context.Background(), svc, tt.request.ids, tt.request.sort)
			if tt.expected.err != "" {
				require.EqualError(t, err, strings.ReplaceAll(tt.expected.err, "${URL}", srv.URL))
				return
			}

			require.NoError(t, err)
			require.Equal(t, tt.expected.result, actual)
		})
	}
}

func TestPrintJobSubmissions(t *testing.T) {
	tests := []struct {
		name               string
		resourcesJSON      string
		token              string
		expectedTableCount int
		expectedTable      string
		expectedPrintCalls int
	}{
		{
			name:               "empty resources with non-admin token",
			resourcesJSON:      jobSubmissionsEmptyResourcesJSON,
			token:              "token-without-admin-role",
			expectedTableCount: 1,
			expectedTable:      expectedJobSubmissionsEmptyTable,
			expectedPrintCalls: 1,
		},
		{
			name:               "empty resources with admin token",
			resourcesJSON:      jobSubmissionsEmptyResourcesJSON,
			token:              createTokenWithRole(constants.WorkbenchAdminRole),
			expectedTableCount: 1,
			expectedTable:      expectedJobSubmissionsAdminEmptyTable,
			expectedPrintCalls: 1,
		},
		{
			name:               "prints one resource with non-admin token",
			resourcesJSON:      jobSubmissionsPrintOneResourcesJSON,
			token:              "token-without-admin-role",
			expectedTableCount: 1,
			expectedTable:      expectedJobSubmissionsOneRowTable,
			expectedPrintCalls: 1,
		},
		{
			name:               "prints two resources with non-admin token",
			resourcesJSON:      jobSubmissionsPrintTwoResourcesJSON,
			token:              "token-without-admin-role",
			expectedTableCount: 1,
			expectedTable:      expectedJobSubmissionsTwoRowsTable,
			expectedPrintCalls: 1,
		},
		{
			name:               "prints one resource with admin token",
			resourcesJSON:      jobSubmissionsPrintOneResourcesJSON,
			token:              createTokenWithRole(constants.WorkbenchAdminRole),
			expectedTableCount: 1,
			expectedTable:      expectedJobSubmissionsAdminOneRowTable,
			expectedPrintCalls: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var resources types.JobSubmissions
			err := json.Unmarshal([]byte(tt.resourcesJSON), &resources)
			require.NoError(t, err)

			u := &stubs.StubUI{}
			err = workbench.PrintJobSubmissions(u, tt.token, resources)
			require.NoError(t, err)

			require.Len(t, u.Tables, tt.expectedTableCount)
			tbl := u.Tables[0]
			require.Equal(t, tt.expectedTable, tbl.String())
			require.Equal(t, tt.expectedPrintCalls, tbl.PrintCalls)
		})
	}
}

func TestDescribeJobSubmissions(t *testing.T) {
	fixedNow := time.Date(2026, time.May, 8, 0, 0, 0, 0, time.UTC)
	timeutil.SetNow(func() time.Time { return fixedNow })
	t.Cleanup(func() { timeutil.SetNow(time.Now) })

	tests := []struct {
		name             string
		resourcesJSON    string
		expectedDisplays []string
		expectedWarns    []string
	}{
		{
			name:             "describe empty resources prints nothing",
			resourcesJSON:    jobSubmissionsEmptyResourcesJSON,
			expectedDisplays: []string{""},
		},
		{
			name:             "describe prints one job submission",
			resourcesJSON:    jobSubmissionsDescribeOneResourcesJSON,
			expectedDisplays: []string{expectedDescribeJobSubmissionOneDisplay},
		},
		{
			name:             "describe prints two job submissions",
			resourcesJSON:    jobSubmissionsDescribeTwoResourcesJSON,
			expectedDisplays: []string{expectedDescribeJobSubmissionTwoDisplay},
		},
		{
			name:             "describe prints job submission with optional sections",
			resourcesJSON:    jobSubmissionsDescribeWithSectionsResourcesJSON,
			expectedDisplays: []string{expectedDescribeJobSubmissionWithSectionsDisplay},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var resources types.JobSubmissions
			err := json.Unmarshal([]byte(tt.resourcesJSON), &resources)
			require.NoError(t, err)

			stubUI := &stubs.StubUI{}
			err = workbench.DescribeJobSubmissions(stubUI, resources)
			require.NoError(t, err)
			require.Equal(t, tt.expectedDisplays, stubUI.Displays, strings.Join(stubUI.Displays, ""))
			require.Equal(t, tt.expectedWarns, stubUI.Warns, strings.Join(stubUI.Warns, ""))
		})
	}
}

func getJobSubmissionDeleteEndpoints(
	expectedBody string,
	responseStatus int,
	responseBody string,
) []stubs.StubHTTPEndpoint {
	return []stubs.StubHTTPEndpoint{
		{
			Request: stubs.StubHTTPRequest{
				Method:          http.MethodPost,
				URL:             postgrest.Endpoint(constants.WorkbenchEndpointRPCJobRunBulkDelete),
				ExpectedHeaders: authHeaders,
				ExpectedBody:    expectedBody,
			},
			Response: stubs.StubHTTPResponse{
				Body:   responseBody,
				Status: responseStatus,
			},
		},
	}
}

func getJobSubmissionCreateEndpoints(
	expectedBody string,
	responseStatus int,
	responseBody string,
) []stubs.StubHTTPEndpoint {
	return []stubs.StubHTTPEndpoint{
		{
			Request: stubs.StubHTTPRequest{
				Method:          http.MethodPost,
				URL:             postgrest.Endpoint(constants.WorkbenchEndpointRPCJobRunCreate),
				ExpectedHeaders: authHeaders,
				ExpectedBody:    expectedBody,
			},
			Response: stubs.StubHTTPResponse{
				Body:   responseBody,
				Status: responseStatus,
			},
		},
	}
}

func getJobSubmissionsGetEndpoints(
	expectedQuery string,
	responseStatus int,
	responseBody string,
) []stubs.StubHTTPEndpoint {
	return []stubs.StubHTTPEndpoint{
		{
			Request: stubs.StubHTTPRequest{
				Method:          http.MethodGet,
				URL:             postgrest.Endpoint(constants.WorkbenchEndpointJobRuns),
				ExpectedHeaders: authHeaders,
				ExpectedQuery:   expectedQuery,
			},
			Response: stubs.StubHTTPResponse{
				Status: responseStatus,
				Body:   responseBody,
			},
		},
	}
}
