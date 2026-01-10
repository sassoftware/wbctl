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

	"github.com/sas-institute-rnd-saas/wbctl/internal/constants"
	"github.com/sas-institute-rnd-saas/wbctl/internal/postgrest"
	"github.com/sas-institute-rnd-saas/wbctl/internal/services/workbench"
	"github.com/sas-institute-rnd-saas/wbctl/internal/stubs"
	"github.com/sas-institute-rnd-saas/wbctl/internal/timeutil"
	"github.com/sas-institute-rnd-saas/wbctl/internal/types"
	"github.com/stretchr/testify/require"
)

const (
	jobsEmptyResourcesJSON = `[]`

	jobsGetOneResponseJSON = `
[
  {
    "id": "job-1",
    "name": "daily-job",
    "description": "first",
    "application_id": "app-1"
  }
]
`

	jobsGetTwoResponseJSON = `
[
  {
    "id": "id-a",
    "name": "job-a",
    "application_id": "app-a"
  },
  {
    "id": "id-b",
    "name": "job-b",
    "application_id": "app-b"
  }
]
`

	jobsPrintOneResourcesJSON = `
[
  {
    "id": "job-1",
    "name": "daily-job",
    "username": "alice",
    "application_id": "app-1",
    "services": [
      {"id": "svc-a"},
      {"id": "svc-b"}
    ],
    "created_at": null
  }
]
`

	jobsPrintTwoResourcesJSON = `
[
  {
    "id": "job-1",
    "name": "daily-job",
    "username": "alice",
    "application_id": "app-1",
    "services": [
      {"id": "svc-a"},
      {"id": "svc-b"}
    ],
    "created_at": null
  },
  {
    "id": "job-2",
    "name": "weekly-job",
    "username": "bob",
    "application_id": "app-2",
    "services": [
      {"image": "img-only"}
    ],
    "created_at": null
  }
]
`

	jobsPrintIgnoresEmptyServiceResourcesJSON = `
[
  {
	"id": "job-3",
	"name": "monthly-job",
	"username": "carol",
	"application_id": "app-3",
	"services": [
	  {},
	  {"id": "svc-c"}
	],
	"created_at": null
  }
]
`

	jobsDescribeOneResourcesJSON = `
[
  {
    "id": "job-1",
    "name": "daily-job",
    "description": "first",
    "username": "alice",
    "email": "alice@example.com",
    "created_at": "2026-05-08T00:00:00Z",
    "modified_at": "2026-05-08T00:00:00Z",
    "most_recent_job_submission_at": "2026-05-08T00:00:00Z",
    "most_recent_job_submission_id": "js-1",
    "nodepool": "np-a",
    "application_id": "app-1",
    "context": "ctx",
    "spec": {
      "script": "run.sh"
    },
    "sub_path": "/work",
    "schedule": "0 * * * *"
  }
]
`

	jobsDescribeTwoResourcesJSON = `
[
  {
    "id": "job-1",
    "name": "daily-job",
    "description": "first",
    "username": "alice",
    "email": "alice@example.com",
    "created_at": "2026-05-08T00:00:00Z",
    "modified_at": "2026-05-08T00:00:00Z",
    "most_recent_job_submission_at": "2026-05-08T00:00:00Z",
    "most_recent_job_submission_id": "js-1",
    "nodepool": "np-a",
    "application_id": "app-1",
    "context": "ctx",
    "spec": {
      "script": "run.sh"
    },
    "sub_path": "/work",
    "schedule": "0 * * * *"
  },
  {
    "id": "job-2",
    "name": "daily-job2",
    "description": "second",
    "username": "bob",
    "email": "bob@example.com",
    "created_at": "2026-05-08T00:00:00Z",
    "modified_at": "2026-05-08T00:00:00Z",
    "most_recent_job_submission_at": "2026-05-08T00:00:00Z",
    "most_recent_job_submission_id": "js-1",
    "nodepool": "np-a",
    "application_id": "app-1",
    "context": "ctx",
    "spec": {
      "script": "run.sh"
    },
    "sub_path": "/work",
    "schedule": "0 * * * *"
  }
]
`

	jobsDescribeWithSectionsResourcesJSON = `
[
  {
    "id": "job-2",
    "name": "weekly-job",
    "description": "second",
    "username": "bob",
    "email": "bob@example.com",
    "created_at": "2026-05-08T00:00:00Z",
    "modified_at": "2026-05-08T00:00:00Z",
    "most_recent_job_submission_id": "js-2",
    "nodepool": "np-b",
    "services": [
      {
        "id": "svc-1",
        "image": "img:v1",
        "env": {
          "ENV_A": "value-a"
        }
      }
    ],
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
    "labels": {
      "team": "analytics"
    },
    "properties": {
      "owner": "data"
    }
  }
]
`

	expectedJobsEmptyTable  = "ID\tNAME\tAPPLICATION_ID\tSERVICES\tAGE"
	expectedJobsOneRowTable = "ID\tNAME\tAPPLICATION_ID\tSERVICES\tAGE\n" +
		"job-1\tdaily-job\tapp-1\tsvc-a, svc-b\t"
	expectedJobsTwoRowsTable = "ID\tNAME\tAPPLICATION_ID\tSERVICES\tAGE\n" +
		"job-1\tdaily-job\tapp-1\tsvc-a, svc-b\t\n" +
		"job-2\tweekly-job\tapp-2\timg-only\t"
	expectedJobsAdminEmptyTable  = "ID\tNAME\tUSERNAME\tAPPLICATION_ID\tSERVICES\tAGE"
	expectedJobsAdminOneRowTable = "ID\tNAME\tUSERNAME\tAPPLICATION_ID\tSERVICES\tAGE\n" +
		"job-1\tdaily-job\talice\tapp-1\tsvc-a, svc-b\t"
	expectedJobsIgnoresEmptyServiceRowTable = "ID\tNAME\tAPPLICATION_ID\tSERVICES\tAGE\n" +
		"job-3\tmonthly-job\tapp-3\tsvc-c\t"

	expectedDescribeJobOneDisplay = `ID:                       job-1
Name:                     daily-job
Description:              first
Created By:               alice (alice@example.com)
Created At:               2026-05-08T00:00:00Z (0s ago)
Modified At:              2026-05-08T00:00:00Z (0s ago)
Most Recent Job Run At:   2026-05-08T00:00:00Z
Most Recent Job Run ID:   js-1
Nodepool:                 np-a
Application:              app-1
Context:                  ctx
Spec:                     {
                 "script": "run.sh"
               }
SubPath:    /work
Schedule:   0 * * * *
`

	expectedDescribeJobTwoDisplay = `ID:                       job-1
Name:                     daily-job
Description:              first
Created By:               alice (alice@example.com)
Created At:               2026-05-08T00:00:00Z (0s ago)
Modified At:              2026-05-08T00:00:00Z (0s ago)
Most Recent Job Run At:   2026-05-08T00:00:00Z
Most Recent Job Run ID:   js-1
Nodepool:                 np-a
Application:              app-1
Context:                  ctx
Spec:                     {
                 "script": "run.sh"
               }
SubPath:    /work
Schedule:   0 * * * *


ID:                       job-2
Name:                     daily-job2
Description:              second
Created By:               bob (bob@example.com)
Created At:               2026-05-08T00:00:00Z (0s ago)
Modified At:              2026-05-08T00:00:00Z (0s ago)
Most Recent Job Run At:   2026-05-08T00:00:00Z
Most Recent Job Run ID:   js-1
Nodepool:                 np-a
Application:              app-1
Context:                  ctx
Spec:                     {
                 "script": "run.sh"
               }
SubPath:    /work
Schedule:   0 * * * *
`

	expectedDescribeJobWithSectionsDisplay = `ID:                       job-2
Name:                     weekly-job
Description:              second
Created By:               bob (bob@example.com)
Created At:               2026-05-08T00:00:00Z (0s ago)
Modified At:              2026-05-08T00:00:00Z (0s ago)
Most Recent Job Run ID:   js-2
Nodepool:                 np-b
Labels:
  team:   analytics
Services:
  0:
    ID:      svc-1
    Image:   img:v1
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
  owner:   data
`

	expectedDescribeJobSpecMarshalErrorDisplay = `ID:                       job-1
Name:                     daily-job
Description:              first
Created By:               alice (alice@example.com)
Created At:               2026-05-08T00:00:00Z (0s ago)
Modified At:              2026-05-08T00:00:00Z (0s ago)
Most Recent Job Run At:   2026-05-08T00:00:00Z
Most Recent Job Run ID:   js-1
Nodepool:                 np-a
Application:              app-1
Context:                  ctx
Spec:                     <error: json: unsupported type: func()>
SubPath:                  /work
Schedule:                 0 * * * *
`

	jobsBulkUpdateResponseJSON = `
[
  {
    "id": "job-1",
    "name": "job-one-updated"
  },
  {
    "id": "job-2",
    "name": "job-two-updated"
  }
]
`
	expectedDeleteJobRequestJSON     = `{"ids":["job-1","job-2"]}`
	createJobOneResponseBodyJSON     = `{"id":"job-1","name":"daily-job"}`
	expectedCreateJobRequestJSON     = `{"param":{"name":"daily-job"}}`
	expectedBulkUpdateJobRequestJSON = `{"ids":["job-1","job-2"],"param":{"name":"updated-name"}}`
)

var (
	expectedResponseOneJob = types.Jobs{
		{ID: "job-1", Name: "daily-job", Description: "first", ApplicationID: "app-1"},
	}

	expectedResponseTwoJobs = types.Jobs{
		{ID: "id-a", Name: "job-a", ApplicationID: "app-a"},
		{ID: "id-b", Name: "job-b", ApplicationID: "app-b"},
	}
)

func TestDeleteJob(t *testing.T) {
	tests := []deleteTestCase{
		{
			name: "delete jobs success",
			request: deleteRequest{
				ids: []string{"job-1", "job-2"},
			},
			mockEndpoints: getJobsDeleteEndpoints(
				expectedDeleteJobRequestJSON,
				http.StatusOK,
				`["job-1", "job-2"]`,
			),
			expected: deleteExpected{
				result: []string{"job-1", "job-2"},
			},
		},
		{
			name: "delete jobs non-200 from server",
			request: deleteRequest{
				ids: []string{"job-1", "job-2"},
			},
			mockEndpoints: getJobsDeleteEndpoints(
				expectedDeleteJobRequestJSON,
				http.StatusUnauthorized,
				"unauthorized",
			),
			expected: deleteExpected{
				err: "POST ${URL}/postgrest/rpc/job_bulk_delete failed\n" +
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

			actual, err := workbench.BulkDeleteJobs(context.Background(), svc, tt.request.ids)
			if tt.expected.err != "" {
				require.EqualError(t, err, strings.ReplaceAll(tt.expected.err, "${URL}", srv.URL))
				return
			}

			require.NoError(t, err)
			require.Equal(t, tt.expected.result, actual)
		})
	}
}

func TestCreateJob(t *testing.T) {
	tests := []createTestCase[types.Job]{
		{
			name: "create job success",
			request: createRequest[types.Job]{
				body: types.Job{
					Name: "daily-job",
				},
			},
			mockEndpoints: getJobsCreateEndpoints(
				expectedCreateJobRequestJSON,
				http.StatusOK,
				createJobOneResponseBodyJSON,
			),
			expected: createExpected[types.Job]{
				result: types.Job{
					ID:   "job-1",
					Name: "daily-job",
				},
			},
		},
		{
			name: "create job non-200 from server",
			request: createRequest[types.Job]{
				body: types.Job{
					Name: "daily-job",
				},
			},
			mockEndpoints: getJobsCreateEndpoints(
				expectedCreateJobRequestJSON,
				http.StatusUnauthorized,
				"unauthorized",
			),
			expected: createExpected[types.Job]{
				err: "POST ${URL}/postgrest/rpc/job_create failed\n" +
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

			actual, err := workbench.CreateJob(context.Background(), svc, tt.request.body)
			if tt.expected.err != "" {
				require.EqualError(t, err, strings.ReplaceAll(tt.expected.err, "${URL}", srv.URL))
				return
			}

			require.NoError(t, err)
			require.Equal(t, tt.expected.result, actual)
		})
	}
}

func TestBulkUpdateJobs(t *testing.T) {
	tests := []bulkupdateTestCase[types.Job, types.Jobs]{
		{
			name: "bulk update jobs success",
			request: bulkupdateRequest[types.Job]{
				ids:  []string{"job-1", "job-2"},
				body: types.Job{Name: "updated-name"},
			},
			mockEndpoints: getJobsBulkUpdateEndpoints(
				expectedBulkUpdateJobRequestJSON,
				http.StatusOK,
				jobsBulkUpdateResponseJSON,
			),
			expected: bulkupdateExpected[types.Jobs]{
				result: types.Jobs{
					{ID: "job-1", Name: "job-one-updated"},
					{ID: "job-2", Name: "job-two-updated"},
				},
			},
		},
		{
			name: "bulk update jobs non-200 from server",
			request: bulkupdateRequest[types.Job]{
				ids:  []string{"job-1", "job-2"},
				body: types.Job{Name: "updated-name"},
			},
			mockEndpoints: getJobsBulkUpdateEndpoints(
				expectedBulkUpdateJobRequestJSON,
				http.StatusUnauthorized,
				"unauthorized",
			),
			expected: bulkupdateExpected[types.Jobs]{
				err: "POST ${URL}/postgrest/rpc/job_bulk_update failed\n" +
					" Caused by:\n" +
					"\t* 401 unauthorized",
			},
		},
		{
			name: "bulk update jobs invalid json response",
			request: bulkupdateRequest[types.Job]{
				ids:  []string{"job-1", "job-2"},
				body: types.Job{Name: "updated-name"},
			},
			mockEndpoints: getJobsBulkUpdateEndpoints(
				expectedBulkUpdateJobRequestJSON,
				http.StatusOK,
				"[",
			),
			expected: bulkupdateExpected[types.Jobs]{
				err: "POST ${URL}/postgrest/rpc/job_bulk_update failed\n" +
					" Caused by:\n" +
					"\t* unexpected EOF",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv, _ := stubs.NewStubHTTPServer(t, tt.mockEndpoints)
			defer srv.Close()
			svc := workbench.NewWorkbenchService(srv.URL, "access-token", "csrf-token")

			actual, err := workbench.BulkUpdateJobs(context.Background(), svc, tt.request.ids, tt.request.body)
			if tt.expected.err != "" {
				require.EqualError(t, err, strings.ReplaceAll(tt.expected.err, "${URL}", srv.URL))
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.expected.result, actual)
		})
	}
}

func TestGetJobs(t *testing.T) {
	tests := []getTestCase[types.Jobs]{
		{
			name:    "get jobs success without ids query",
			request: getRequest{},
			mockEndpoints: getJobsGetEndpoints(
				"",
				http.StatusOK,
				jobsGetOneResponseJSON,
			),
			expected: getExpected[types.Jobs]{
				result: expectedResponseOneJob,
			},
		},
		{
			name:    "get jobs success with ids query",
			request: getRequest{ids: []string{"id-a"}},
			mockEndpoints: getJobsGetEndpoints(
				"id=in.(id-a)",
				http.StatusOK,
				jobsGetTwoResponseJSON,
			),
			expected: getExpected[types.Jobs]{
				result: expectedResponseTwoJobs,
			},
		},
		{
			name: "get jobs sort by name",
			request: getRequest{
				sort: "name",
			},
			mockEndpoints: getJobsGetEndpoints(
				"order=name",
				http.StatusOK,
				jobsGetOneResponseJSON,
			),
			expected: getExpected[types.Jobs]{
				result: expectedResponseOneJob,
			},
		},
		{
			name: "get jobs sort by username",
			request: getRequest{
				sort: "username",
			},
			mockEndpoints: getJobsGetEndpoints(
				"order=username",
				http.StatusOK,
				jobsGetOneResponseJSON,
			),
			expected: getExpected[types.Jobs]{
				result: expectedResponseOneJob,
			},
		},
		{
			name: "get jobs sort by application_id",
			request: getRequest{
				sort: "application_id",
			},
			mockEndpoints: getJobsGetEndpoints(
				"order=application_id",
				http.StatusOK,
				jobsGetOneResponseJSON,
			),
			expected: getExpected[types.Jobs]{
				result: expectedResponseOneJob,
			},
		},
		{
			name: "get jobs sort by services",
			request: getRequest{
				sort: "services",
			},
			mockEndpoints: getJobsGetEndpoints(
				"order=services",
				http.StatusOK,
				jobsGetOneResponseJSON,
			),
			expected: getExpected[types.Jobs]{
				result: expectedResponseOneJob,
			},
		},
		{
			name: "get jobs sort by age",
			request: getRequest{
				sort: "age",
			},
			mockEndpoints: getJobsGetEndpoints(
				"order=created_at",
				http.StatusOK,
				jobsGetOneResponseJSON,
			),
			expected: getExpected[types.Jobs]{
				result: expectedResponseOneJob,
			},
		},
		{
			name: "get jobs sort by age and state",
			request: getRequest{
				sort: "age.desc,state.asc",
			},
			mockEndpoints: getJobsGetEndpoints(
				"order=created_at.desc,current_state.asc",
				http.StatusOK,
				jobsGetOneResponseJSON,
			),
			expected: getExpected[types.Jobs]{
				result: expectedResponseOneJob,
			},
		},
		{
			name:    "get jobs invalid json response",
			request: getRequest{ids: []string{"id-a"}},
			mockEndpoints: getJobsGetEndpoints(
				"id=in.(id-a)",
				http.StatusOK,
				"[",
			),
			expected: getExpected[types.Jobs]{
				// jobs do not have controller ids so this query is for `id`
				// even though it is not a uuid.
				err: "GET ${URL}/postgrest/jobs?id=in.%28id-a%29 failed\n" +
					" Caused by:\n" +
					"\t* unexpected EOF",
			},
		},
		{
			name:    "get jobs non-200 from server",
			request: getRequest{ids: []string{"id-a"}},
			mockEndpoints: getJobsGetEndpoints(
				"id=in.(id-a)",
				http.StatusUnauthorized,
				"unauthorized",
			),
			expected: getExpected[types.Jobs]{
				err: "GET ${URL}/postgrest/jobs?id=in.%28id-a%29 failed\n" +
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

			actual, err := workbench.GetJobs(context.Background(), svc, tt.request.ids, tt.request.sort)
			if tt.expected.err != "" {
				require.EqualError(t, err, strings.ReplaceAll(tt.expected.err, "${URL}", srv.URL))
				return
			}

			require.NoError(t, err)
			require.Equal(t, tt.expected.result, actual)
		})
	}
}

func TestPrintJobs(t *testing.T) {
	tests := []struct {
		name               string
		resourcesJSON      string
		token              string
		expectedTableCount int
		expectedTable      string
		expectedPrintCalls int
	}{
		{
			name:               "print jobs empty resources with non-admin token",
			resourcesJSON:      jobsEmptyResourcesJSON,
			token:              "token-without-admin-role",
			expectedTableCount: 1,
			expectedTable:      expectedJobsEmptyTable,
			expectedPrintCalls: 1,
		},
		{
			name:               "print jobs empty resources with admin token",
			resourcesJSON:      jobsEmptyResourcesJSON,
			token:              createTokenWithRole(constants.WorkbenchAdminRole),
			expectedTableCount: 1,
			expectedTable:      expectedJobsAdminEmptyTable,
			expectedPrintCalls: 1,
		},
		{
			name:               "print jobs prints one resource with non-admin token",
			resourcesJSON:      jobsPrintOneResourcesJSON,
			token:              "token-without-admin-role",
			expectedTableCount: 1,
			expectedTable:      expectedJobsOneRowTable,
			expectedPrintCalls: 1,
		},
		{
			name:               "print jobs prints two resources with non-admin token",
			resourcesJSON:      jobsPrintTwoResourcesJSON,
			token:              "token-without-admin-role",
			expectedTableCount: 1,
			expectedTable:      expectedJobsTwoRowsTable,
			expectedPrintCalls: 1,
		},
		{
			name:               "print jobs prints one resource with admin token",
			resourcesJSON:      jobsPrintOneResourcesJSON,
			token:              createTokenWithRole(constants.WorkbenchAdminRole),
			expectedTableCount: 1,
			expectedTable:      expectedJobsAdminOneRowTable,
			expectedPrintCalls: 1,
		},
		{
			name:               "print jobs ignores empty service entries",
			resourcesJSON:      jobsPrintIgnoresEmptyServiceResourcesJSON,
			token:              "token-without-admin-role",
			expectedTableCount: 1,
			expectedTable:      expectedJobsIgnoresEmptyServiceRowTable,
			expectedPrintCalls: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var resources types.Jobs
			err := json.Unmarshal([]byte(tt.resourcesJSON), &resources)
			require.NoError(t, err)

			u := &stubs.StubUI{}
			err = workbench.PrintJobs(u, tt.token, resources)
			require.NoError(t, err)

			require.Len(t, u.Tables, tt.expectedTableCount)
			tbl := u.Tables[0]
			require.Equal(t, tt.expectedTable, tbl.String())
			require.Equal(t, tt.expectedPrintCalls, tbl.PrintCalls)
		})
	}
}

func TestDescribeJobs(t *testing.T) {
	fixedNow := time.Date(2026, time.May, 8, 0, 0, 0, 0, time.UTC)
	timeutil.SetNow(func() time.Time { return fixedNow })
	t.Cleanup(func() { timeutil.SetNow(time.Now) })

	tests := []struct {
		name             string
		resourcesJSON    string
		mutate           func(types.Jobs)
		expectedDisplays []string
		expectedWarns    []string
	}{
		{
			name:             "describe jobs describe empty resources prints nothing",
			resourcesJSON:    jobsEmptyResourcesJSON,
			expectedDisplays: []string{""},
		},
		{
			name:             "describe jobs describe prints one job",
			resourcesJSON:    jobsDescribeOneResourcesJSON,
			expectedDisplays: []string{expectedDescribeJobOneDisplay},
		},
		{
			name:             "describe jobs describe prints two jobs",
			resourcesJSON:    jobsDescribeTwoResourcesJSON,
			expectedDisplays: []string{expectedDescribeJobTwoDisplay},
		},
		{
			name:             "describe jobs describe prints job with optional sections",
			resourcesJSON:    jobsDescribeWithSectionsResourcesJSON,
			expectedDisplays: []string{expectedDescribeJobWithSectionsDisplay},
		},
		{
			name:          "describe jobs describe spec marshal error path",
			resourcesJSON: jobsDescribeOneResourcesJSON,
			mutate: func(resources types.Jobs) {
				resources[0].Spec["bad"] = func() {}
			},
			expectedDisplays: []string{
				expectedDescribeJobSpecMarshalErrorDisplay,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var resources types.Jobs
			err := json.Unmarshal([]byte(tt.resourcesJSON), &resources)
			require.NoError(t, err)
			if tt.mutate != nil {
				tt.mutate(resources)
			}

			stubUI := &stubs.StubUI{}
			err = workbench.DescribeJobs(stubUI, resources)
			require.NoError(t, err)
			require.Equal(t, tt.expectedDisplays, stubUI.Displays, strings.Join(stubUI.Displays, ""))
			require.Equal(t, tt.expectedWarns, stubUI.Warns, strings.Join(stubUI.Warns, ""))
		})
	}
}

func getJobsDeleteEndpoints(
	expectedBody string,
	responseStatus int,
	responseBody string,
) []stubs.StubHTTPEndpoint {
	return []stubs.StubHTTPEndpoint{
		{
			Request: stubs.StubHTTPRequest{
				Method:          http.MethodPost,
				URL:             postgrest.Endpoint(constants.WorkbenchEndpointRPCJobBulkDelete),
				ExpectedHeaders: authHeaders,
				ExpectedBody:    expectedBody,
			},
			Response: stubs.StubHTTPResponse{
				Status: responseStatus,
				Body:   responseBody,
			},
		},
	}
}

func getJobsCreateEndpoints(
	expectedBody string,
	responseStatus int,
	responseBody string,
) []stubs.StubHTTPEndpoint {
	return []stubs.StubHTTPEndpoint{
		{
			Request: stubs.StubHTTPRequest{
				Method:          http.MethodPost,
				URL:             postgrest.Endpoint(constants.WorkbenchEndpointRPCJobCreate),
				ExpectedHeaders: authHeaders,
				ExpectedBody:    expectedBody,
			},
			Response: stubs.StubHTTPResponse{
				Status: responseStatus,
				Body:   responseBody,
			},
		},
	}
}

func getJobsBulkUpdateEndpoints(
	expectedBody string,
	responseStatus int,
	responseBody string,
) []stubs.StubHTTPEndpoint {
	return []stubs.StubHTTPEndpoint{
		{
			Request: stubs.StubHTTPRequest{
				Method:          http.MethodPost,
				URL:             postgrest.Endpoint(constants.WorkbenchEndpointRPCJobBulkUpdate),
				ExpectedHeaders: authHeaders,
				ExpectedBody:    expectedBody,
			},
			Response: stubs.StubHTTPResponse{
				Status: responseStatus,
				Body:   responseBody,
			},
		},
	}
}

func getJobsGetEndpoints(
	expectedQuery string,
	responseStatus int,
	responseBody string,
) []stubs.StubHTTPEndpoint {
	return []stubs.StubHTTPEndpoint{
		{
			Request: stubs.StubHTTPRequest{
				Method:          http.MethodGet,
				URL:             postgrest.Endpoint(constants.WorkbenchEndpointJobs),
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
