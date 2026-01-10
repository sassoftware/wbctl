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
	"github.com/sas-institute-rnd-saas/wbctl/internal/i18n"
	"github.com/sas-institute-rnd-saas/wbctl/internal/postgrest"
	"github.com/sas-institute-rnd-saas/wbctl/internal/services/workbench"
	"github.com/sas-institute-rnd-saas/wbctl/internal/stubs"
	"github.com/sas-institute-rnd-saas/wbctl/internal/timeutil"
	"github.com/sas-institute-rnd-saas/wbctl/internal/types"
	"github.com/stretchr/testify/require"
)

const (
	workspacesEmptyResourcesJSON = `[]`

	workspacesGetOneResponseJSON = `
[
  {
    "id": "ws-1",
    "controller_id": "ctrl-1",
    "name": "workspace-1",
    "username": "alice",
    "state": "Running",
    "message": "ok"
  }
]
`

	workspacesGetMixedIDsResponseJSON = `
[
  {
    "id": "123e4567-e89b-12d3-a456-426614174000",
    "controller_id": "ctrl-uuid",
    "name": "workspace-uuid",
    "username": "alice"
  },
  {
    "id": "ws-controller",
    "controller_id": "controller-ws-2",
    "name": "workspace-controller",
    "username": "bob"
  }
]
`

	workspacesPrintOneResourcesJSON = `
[
  {
    "id": "ws-1",
    "controller_id": "",
    "name": "workspace-1",
    "username": "alice",
    "storage_size_g": 10,
    "storage_spec": {
      "storageClassName": "standard"
    },
    "properties": {
      "HOME_DIRECTORY": "true"
    },
    "state": "Running",
    "message": "ok",
    "created_at": null
  }
]
`

	workspacesPrintTwoResourcesJSON = `
[
  {
    "id": "ws-1",
    "controller_id": "",
    "name": "workspace-1",
    "username": "alice",
    "storage_size_g": 10,
    "storage_spec": {
      "storageClassName": "standard"
    },
    "properties": {
      "HOME_DIRECTORY": "true"
    },
    "state": "Running",
    "message": "ok",
    "created_at": null
  },
  {
    "id": "ws-2",
    "controller_id": "ctrl-2",
    "name": "workspace-2",
    "username": "bob",
    "storage_size_g": 20,
    "properties": {},
    "state": "Stopped",
    "message": "idle",
    "created_at": null
  }
]
`

	workspacesPrintEmptyStorageClassResourcesJSON = `
[
  {
	"id": "ws-3",
	"controller_id": "ctrl-3",
	"name": "workspace-3",
	"username": "carol",
	"storage_size_g": 30,
	"storage_spec": {},
	"properties": {},
	"state": "Running",
	"message": "ok",
	"created_at": null
  }
]
`

	workspacesDescribeOneResourcesJSON = `
[
  {
    "id": "ws-1",
    "controller_id": "ctrl-1",
    "name": "workspace-1",
    "description": "first workspace",
    "storage_type": "pvc",
    "username": "alice",
    "email": "alice@example.com",
    "created_at": "2026-05-08T00:00:00Z",
    "modified_at": "2026-05-08T00:00:00Z",
    "state": "Running",
    "message": "ok",
    "pvc_status": {
      "status": {
        "phase": "Bound"
      }
    }
  }
]
`

	workspacesDescribeTwoResourcesJSON = `
[
  {
    "id": "ws-1",
    "controller_id": "ctrl-1",
    "name": "workspace-1",
    "description": "first workspace",
    "storage_type": "pvc",
    "username": "alice",
    "email": "alice@example.com",
    "created_at": "2026-05-08T00:00:00Z",
    "modified_at": "2026-05-08T00:00:00Z",
    "state": "Running",
    "message": "ok",
    "pvc_status": {
      "status": {
        "phase": "Bound"
      }
    }
  },
  {
    "id": "ws-2",
    "controller_id": "ctrl-2",
    "name": "workspace-2",
    "description": "second workspace",
    "storage_type": "pvc",
    "username": "bob",
    "email": "bob@example.com",
    "created_at": "2026-05-08T00:00:00Z",
    "modified_at": "2026-05-08T00:00:00Z",
    "state": "Running",
    "message": "ok",
    "pvc_status": {
      "status": {
        "phase": "Bound"
      }
    }
  }
]
`

	workspacesDescribeWithSectionsResourcesJSON = `
[
  {
    "id": "ws-2",
    "controller_id": "ctrl-2",
    "name": "workspace-2",
    "storage_type": "pvc",
    "username": "bob",
    "email": "bob@example.com",
    "created_at": "2026-05-08T00:00:00Z",
    "modified_at": "2026-05-08T00:00:00Z",
    "state": "Stopped",
    "initializers": [
      {
        "id": "svc-1",
        "image": "img:v1",
        "env": {
          "ENV_A": "value-a"
        }
      }
    ],
    "properties": {
      "owner": "data"
    }
  }
]
`

	expectedWorkspacesEmptyTable  = "ID\tCONTROLLER_ID\tNAME\tSIZE\tHOME_DIR\tSTORAGECLASSNAME\tSTATE\tMESSAGE\tAGE"
	expectedWorkspacesOneRowTable = "ID\tCONTROLLER_ID\tNAME\tSIZE\tHOME_DIR\tSTORAGECLASSNAME\tSTATE\tMESSAGE\tAGE\n" +
		"ws-1\t<unknown>\tworkspace-1\t10Gi\ttrue\tstandard\tRunning\tok\t"
	expectedWorkspacesTwoRowsTable = "ID\tCONTROLLER_ID\tNAME\tSIZE\tHOME_DIR\tSTORAGECLASSNAME\tSTATE\tMESSAGE\tAGE\n" +
		"ws-1\t<unknown>\tworkspace-1\t10Gi\ttrue\tstandard\tRunning\tok\t\n" +
		"ws-2\tctrl-2\tworkspace-2\t20Gi\t\t\tStopped\tidle\t"
	expectedWorkspacesEmptyStorageClassRowTable = "ID\tCONTROLLER_ID\tNAME\tSIZE\tHOME_DIR\tSTORAGECLASSNAME\t" +
		"STATE\tMESSAGE\tAGE\n" +
		"ws-3\tctrl-3\tworkspace-3\t30Gi\t\t\tRunning\tok\t"
	expectedWorkspacesAdminEmptyTable = "ID\tCONTROLLER_ID\tNAME\tUSERNAME\tSIZE\tHOME_DIR\tSTORAGECLASSNAME\t" +
		"STATE\tMESSAGE\tAGE"
	expectedWorkspacesAdminOneRowTable = "ID\tCONTROLLER_ID\tNAME\tUSERNAME\tSIZE\tHOME_DIR\tSTORAGECLASSNAME\t" +
		"STATE\tMESSAGE\tAGE\n" +
		"ws-1\t<unknown>\tworkspace-1\talice\t10Gi\ttrue\tstandard\tRunning\tok\t"

	expectedDescribeWorkspaceOneDisplay = `ID:              ws-1
Controller ID:   ctrl-1
Name:            workspace-1
Description:     first workspace
Storage Type:    pvc
Created By:      alice (alice@example.com)
Created At:      2026-05-08T00:00:00Z (0s ago)
Modified At:     2026-05-08T00:00:00Z (0s ago)
State:           Running
Message:         ok
Phase:           Bound
`
	expectedDescribeWorkspaceTwoDisplay = `ID:              ws-1
Controller ID:   ctrl-1
Name:            workspace-1
Description:     first workspace
Storage Type:    pvc
Created By:      alice (alice@example.com)
Created At:      2026-05-08T00:00:00Z (0s ago)
Modified At:     2026-05-08T00:00:00Z (0s ago)
State:           Running
Message:         ok
Phase:           Bound


ID:              ws-2
Controller ID:   ctrl-2
Name:            workspace-2
Description:     second workspace
Storage Type:    pvc
Created By:      bob (bob@example.com)
Created At:      2026-05-08T00:00:00Z (0s ago)
Modified At:     2026-05-08T00:00:00Z (0s ago)
State:           Running
Message:         ok
Phase:           Bound
`
	expectedDescribeWorkspaceWithSectionsOneDisplay = `ID:              ws-2
Controller ID:   ctrl-2
Name:            workspace-2
Storage Type:    pvc
Created By:      bob (bob@example.com)
Created At:      2026-05-08T00:00:00Z (0s ago)
Modified At:     2026-05-08T00:00:00Z (0s ago)
State:           Stopped
Phase:           
Initializers:
  0:
    ID:      svc-1
    Image:   img:v1
    Environment:
      ENV_A:   value-a
Properties:
  owner:   data
`

	workspacesBulkUpdateResponseJSON = `
[
  {
    "id": "ws-1",
    "name": "ws-one-updated"
  },
  {
    "id": "ws-2",
    "name": "ws-two-updated"
  }
]
`
	expectedDeleteWorkspaceRequestJSON     = `{"ids":["ws-1", "ws-2"]}`
	expectedBulkUpdateWorkspaceRequestJSON = `{"ids":["ws-1","ws-2"],"workspace":` +
		`{"description":"updated-description","pvc_status":{"status":{}}}}`
	expectedUpdateWorkspaceRequestJSON = `{"param":{"id":"ws-1","name":"workspace-one-renamed",` +
		`"pvc_status":{"status":{}}}}`
	createWorkspaceResponseJSON        = `{"id":"ws-1","name":"workspace-1"}`
	createWorkspaceRenamedResponseJSON = `{"id":"ws-1","name":"workspace-one-renamed"}`
)

var (
	expectedResponseOneWorkspace = types.Workspaces{
		{
			ID:           "ws-1",
			ControllerID: "ctrl-1",
			Name:         "workspace-1",
			Username:     "alice",
			State:        "Running",
			Message:      "ok",
		},
	}

	expectedResponseTwoWorkspaces = types.Workspaces{
		{
			ID:           "123e4567-e89b-12d3-a456-426614174000",
			ControllerID: "ctrl-uuid",
			Name:         "workspace-uuid",
			Username:     "alice",
		},
		{
			ID:           "ws-controller",
			ControllerID: "controller-ws-2",
			Name:         "workspace-controller",
			Username:     "bob",
		},
	}
)

func TestDeleteWorkspace(t *testing.T) {
	tests := []deleteTestCase{
		{
			name: "delete workspace success",
			request: deleteRequest{
				ids: []string{"ws-1", "ws-2"},
			},
			mockEndpoints: getWorkspaceDeleteEndpoints(
				expectedDeleteWorkspaceRequestJSON,
				http.StatusOK,
				`["ws-1", "ws-2"]`,
			),
			expected: deleteExpected{
				result: []string{"ws-1", "ws-2"},
			},
		},
		{
			name: "delete workspace non-200 from server",
			request: deleteRequest{
				ids: []string{"ws-1", "ws-2"},
			},
			mockEndpoints: getWorkspaceDeleteEndpoints(
				expectedDeleteWorkspaceRequestJSON,
				http.StatusUnauthorized,
				`unauthorized`,
			),
			expected: deleteExpected{
				err: "POST ${URL}/postgrest/rpc/workspace_bulk_delete failed\n" +
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

			actual, err := workbench.BulkDeleteWorkspaces(context.Background(), svc, tt.request.ids)
			if tt.expected.err != "" {
				require.EqualError(t, err, strings.ReplaceAll(tt.expected.err, "${URL}", srv.URL))
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.expected.result, actual)
		})
	}
}

func TestCreateWorkspace(t *testing.T) {
	tests := []createTestCase[types.Workspace]{
		{
			name: "create workspace success",
			request: createRequest[types.Workspace]{
				body: types.Workspace{
					Name: "workspace-1",
				},
			},
			mockEndpoints: getWorkspaceCreateEndpoints(
				stubs.AsString(
					t,
					paramPayload[types.Workspace]{Param: types.Workspace{Name: "workspace-1"}},
				),
				http.StatusOK,
				createWorkspaceResponseJSON,
			),
			expected: createExpected[types.Workspace]{
				result: types.Workspace{
					ID:   "ws-1",
					Name: "workspace-1",
				},
			},
		},
		{
			name: "create workspace non-200 from server",
			request: createRequest[types.Workspace]{
				body: types.Workspace{
					Name: "workspace-1",
				},
			},
			mockEndpoints: getWorkspaceCreateEndpoints(
				stubs.AsString(
					t,
					paramPayload[types.Workspace]{Param: types.Workspace{Name: "workspace-1"}},
				),
				http.StatusUnauthorized,
				`unauthorized`,
			),
			expected: createExpected[types.Workspace]{
				err: "POST ${URL}/postgrest/rpc/workspace_create failed\n" +
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

			actual, err := workbench.CreateWorkspace(context.Background(), svc, tt.request.body)
			if tt.expected.err != "" {
				require.EqualError(t, err, strings.ReplaceAll(tt.expected.err, "${URL}", srv.URL))
				return
			}

			require.NoError(t, err)
			require.Equal(t, tt.expected.result, actual)
		})
	}
}

func TestBulkUpdateWorkspaces(t *testing.T) {
	tests := []bulkupdateTestCase[types.Workspace, types.Workspaces]{
		{
			name: "bulk update workspaces success",
			request: bulkupdateRequest[types.Workspace]{
				ids:  []string{"ws-1", "ws-2"},
				body: types.Workspace{Description: "updated-description"},
			},
			mockEndpoints: getWorkspaceBulkUpdateEndpoints(
				expectedBulkUpdateWorkspaceRequestJSON,
				http.StatusOK,
				workspacesBulkUpdateResponseJSON,
			),
			expected: bulkupdateExpected[types.Workspaces]{
				result: types.Workspaces{
					{ID: "ws-1", Name: "ws-one-updated"},
					{ID: "ws-2", Name: "ws-two-updated"},
				},
			},
		},
		{
			name: "bulk update workspaces non-200 from server",
			request: bulkupdateRequest[types.Workspace]{
				ids:  []string{"ws-1", "ws-2"},
				body: types.Workspace{Description: "updated-description"},
			},
			mockEndpoints: getWorkspaceBulkUpdateEndpoints(
				expectedBulkUpdateWorkspaceRequestJSON,
				http.StatusUnauthorized,
				`unauthorized`,
			),
			expected: bulkupdateExpected[types.Workspaces]{
				err: "POST ${URL}/postgrest/rpc/workspace_bulk_update failed\n" +
					" Caused by:\n" +
					"\t* 401 unauthorized",
			},
		},
		{
			name: "bulk update workspaces invalid json response",
			request: bulkupdateRequest[types.Workspace]{
				ids:  []string{"ws-1", "ws-2"},
				body: types.Workspace{Description: "updated-description"},
			},
			mockEndpoints: getWorkspaceBulkUpdateEndpoints(
				expectedBulkUpdateWorkspaceRequestJSON,
				http.StatusOK,
				`[`,
			),
			expected: bulkupdateExpected[types.Workspaces]{
				err: "POST ${URL}/postgrest/rpc/workspace_bulk_update failed\n" +
					" Caused by:\n" +
					"\t* unexpected EOF",
			},
		},
		{
			name: "bulk update workspace name with single id uses single update endpoint",
			request: bulkupdateRequest[types.Workspace]{
				ids:  []string{"ws-1"},
				body: types.Workspace{Name: "workspace-one-renamed"},
			},
			mockEndpoints: getWorkspaceUpdateEndpoints(
				expectedUpdateWorkspaceRequestJSON,
				http.StatusOK,
				createWorkspaceRenamedResponseJSON,
			),
			expected: bulkupdateExpected[types.Workspaces]{
				result: types.Workspaces{{ID: "ws-1", Name: "workspace-one-renamed"}},
			},
		},
		{
			name: "bulk update workspace name with multiple ids returns validation error without rpc",
			request: bulkupdateRequest[types.Workspace]{
				ids:  []string{"ws-1", "ws-2"},
				body: types.Workspace{Name: "workspace-one-renamed"},
			},
			expected: bulkupdateExpected[types.Workspaces]{
				err: i18n.Format(i18n.BulkUpdateNameError,
					i18n.A{constants.I18nArgKeyType: constants.VolumesType}),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv, counts := stubs.NewStubHTTPServer(t, tt.mockEndpoints)
			defer srv.Close()
			svc := workbench.NewWorkbenchService(srv.URL, "access-token", "csrf-token")

			actual, err := workbench.BulkUpdateWorkspaces(context.Background(), svc, tt.request.ids, tt.request.body)
			if tt.expected.err != "" {
				require.EqualError(t, err, strings.ReplaceAll(tt.expected.err, "${URL}", srv.URL))
				return
			}
			_ = counts
			require.NoError(t, err)
			require.Equal(t, tt.expected.result, actual)
		})
	}
}

func TestGetWorkspaces(t *testing.T) {
	tests := []getTestCase[types.Workspaces]{
		{
			name:    "get workspaces success without ids query",
			request: getRequest{},
			mockEndpoints: getWorkspaceGetEndpoints(
				"",
				http.StatusOK,
				workspacesGetOneResponseJSON,
			),
			expected: getExpected[types.Workspaces]{
				result: expectedResponseOneWorkspace,
			},
		},
		{
			name: "get workspaces success with mixed uuid and controller ids query",
			request: getRequest{
				ids:  []string{"123e4567-e89b-12d3-a456-426614174000", "controller-ws-2"},
				sort: "id.asc",
			},
			mockEndpoints: getWorkspaceGetEndpoints(
				"or=(id.in.(123e4567-e89b-12d3-a456-426614174000),"+
					"controller_id.in.(controller-ws-2))&order=id.asc",
				http.StatusOK,
				workspacesGetMixedIDsResponseJSON,
			),
			expected: getExpected[types.Workspaces]{
				result: expectedResponseTwoWorkspaces,
			},
		},
		{
			name: "get workspaces sort by controller_id",
			request: getRequest{
				sort: "controller_id",
			},
			mockEndpoints: getWorkspaceGetEndpoints(
				"order=controller_id",
				http.StatusOK,
				workspacesGetOneResponseJSON,
			),
			expected: getExpected[types.Workspaces]{
				result: expectedResponseOneWorkspace,
			},
		},
		{
			name: "get workspaces sort by name",
			request: getRequest{
				sort: "name",
			},
			mockEndpoints: getWorkspaceGetEndpoints(
				"order=name",
				http.StatusOK,
				workspacesGetOneResponseJSON,
			),
			expected: getExpected[types.Workspaces]{
				result: expectedResponseOneWorkspace,
			},
		},
		{
			name: "get workspaces sort by username",
			request: getRequest{
				sort: "username",
			},
			mockEndpoints: getWorkspaceGetEndpoints(
				"order=username",
				http.StatusOK,
				workspacesGetOneResponseJSON,
			),
			expected: getExpected[types.Workspaces]{
				result: expectedResponseOneWorkspace,
			},
		},
		{
			name: "get workspaces sort by size",
			request: getRequest{
				sort: "size",
			},
			mockEndpoints: getWorkspaceGetEndpoints(
				"order=storage_size_g",
				http.StatusOK,
				workspacesGetOneResponseJSON,
			),
			expected: getExpected[types.Workspaces]{
				result: expectedResponseOneWorkspace,
			},
		},
		{
			name: "get workspaces sort by home_dir",
			request: getRequest{
				sort: "home_dir",
			},
			mockEndpoints: getWorkspaceGetEndpoints(
				"order=properties->HOME_DIRECTORY",
				http.StatusOK,
				workspacesGetOneResponseJSON,
			),
			expected: getExpected[types.Workspaces]{
				result: expectedResponseOneWorkspace,
			},
		},
		{
			name: "get workspaces sort by storageclassname",
			request: getRequest{
				sort: "storageclassname",
			},
			mockEndpoints: getWorkspaceGetEndpoints(
				"order=storage_spec->storageClassName",
				http.StatusOK,
				workspacesGetOneResponseJSON,
			),
			expected: getExpected[types.Workspaces]{
				result: expectedResponseOneWorkspace,
			},
		},
		{
			name: "get workspaces sort by state",
			request: getRequest{
				sort: "state",
			},
			mockEndpoints: getWorkspaceGetEndpoints(
				"order=state",
				http.StatusOK,
				workspacesGetOneResponseJSON,
			),
			expected: getExpected[types.Workspaces]{
				result: expectedResponseOneWorkspace,
			},
		},
		{
			name: "get workspaces sort by message",
			request: getRequest{
				sort: "message",
			},
			mockEndpoints: getWorkspaceGetEndpoints(
				"order=message",
				http.StatusOK,
				workspacesGetOneResponseJSON,
			),
			expected: getExpected[types.Workspaces]{
				result: expectedResponseOneWorkspace,
			},
		},
		{
			name: "get workspaces sort by age",
			request: getRequest{
				sort: "age",
			},
			mockEndpoints: getWorkspaceGetEndpoints(
				"order=created_at",
				http.StatusOK,
				workspacesGetOneResponseJSON,
			),
			expected: getExpected[types.Workspaces]{
				result: expectedResponseOneWorkspace,
			},
		},
		{
			name: "get workspaces sort by age, size, homedir, and storageclassname",
			request: getRequest{
				sort: "age.desc,size.asc,home_dir,storageclassname.desc",
			},
			mockEndpoints: getWorkspaceGetEndpoints(
				"order=created_at.desc,storage_size_g.asc,properties->HOME_DIRECTORY,storage_spec->storageClassName.desc",
				http.StatusOK,
				workspacesGetOneResponseJSON,
			),
			expected: getExpected[types.Workspaces]{
				result: expectedResponseOneWorkspace,
			},
		},
		{
			name: "get workspaces invalid json response",
			request: getRequest{
				ids: []string{"id-a"},
			},
			mockEndpoints: getWorkspaceGetEndpoints(
				"controller_id=in.(id-a)",
				http.StatusOK,
				`[`,
			),
			expected: getExpected[types.Workspaces]{
				err: "GET ${URL}/postgrest/workspaces?controller_id=in.%28id-a%29 failed\n" +
					" Caused by:\n" +
					"\t* unexpected EOF",
			},
		},
		{
			name: "get workspaces non-200 from server",
			request: getRequest{
				ids: []string{"id-a"},
			},
			mockEndpoints: getWorkspaceGetEndpoints(
				"controller_id=in.(id-a)",
				http.StatusUnauthorized,
				`unauthorized`,
			),
			expected: getExpected[types.Workspaces]{
				err: "GET ${URL}/postgrest/workspaces?controller_id=in.%28id-a%29 failed\n" +
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
			actual, err := workbench.GetWorkspaces(context.Background(), svc, tt.request.ids, tt.request.sort)
			if tt.expected.err != "" {
				require.EqualError(t, err, strings.ReplaceAll(tt.expected.err, "${URL}", srv.URL))
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.expected.result, actual)
		})
	}
}

func TestPrintWorkspaces(t *testing.T) {
	tests := []struct {
		name               string
		resourcesJSON      string
		token              string
		expectedTableCount int
		expectedTable      string
		expectedPrintCalls int
	}{
		{
			name:               "print workspaces empty resources with non-admin token",
			resourcesJSON:      workspacesEmptyResourcesJSON,
			token:              "token-without-admin-role",
			expectedTableCount: 1,
			expectedTable:      expectedWorkspacesEmptyTable,
			expectedPrintCalls: 1,
		},
		{
			name:               "print workspaces empty resources with admin token",
			resourcesJSON:      workspacesEmptyResourcesJSON,
			token:              createTokenWithRole(constants.WorkbenchAdminRole),
			expectedTableCount: 1,
			expectedTable:      expectedWorkspacesAdminEmptyTable,
			expectedPrintCalls: 1,
		},
		{
			name:               "print workspaces prints one resource with non-admin token",
			resourcesJSON:      workspacesPrintOneResourcesJSON,
			token:              "token-without-admin-role",
			expectedTableCount: 1,
			expectedTable:      expectedWorkspacesOneRowTable,
			expectedPrintCalls: 1,
		},
		{
			name:               "print workspaces prints one resource with admin token",
			resourcesJSON:      workspacesPrintOneResourcesJSON,
			token:              createTokenWithRole(constants.WorkbenchAdminRole),
			expectedTableCount: 1,
			expectedTable:      expectedWorkspacesAdminOneRowTable,
			expectedPrintCalls: 1,
		},
		{
			name:               "print workspaces prints two resources with non-admin token",
			resourcesJSON:      workspacesPrintTwoResourcesJSON,
			token:              "token-without-admin-role",
			expectedTableCount: 1,
			expectedTable:      expectedWorkspacesTwoRowsTable,
			expectedPrintCalls: 1,
		},
		{
			name:               "print workspaces prints resource with non-nil storage spec and empty class",
			resourcesJSON:      workspacesPrintEmptyStorageClassResourcesJSON,
			token:              "token-without-admin-role",
			expectedTableCount: 1,
			expectedTable:      expectedWorkspacesEmptyStorageClassRowTable,
			expectedPrintCalls: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var resources types.Workspaces
			err := json.Unmarshal([]byte(tt.resourcesJSON), &resources)
			require.NoError(t, err)
			u := &stubs.StubUI{}
			err = workbench.PrintWorkspaces(u, tt.token, resources)
			require.NoError(t, err)
			require.Len(t, u.Tables, tt.expectedTableCount)
			tbl := u.Tables[0]
			require.Equal(t, tt.expectedTable, tbl.String())
			require.Equal(t, tt.expectedPrintCalls, tbl.PrintCalls)
		})
	}
}

func TestDescribeWorkspaces(t *testing.T) {
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
			name:             "describe workspaces describe empty resources prints nothing",
			resourcesJSON:    workspacesEmptyResourcesJSON,
			expectedDisplays: []string{""},
		},
		{
			name:             "describe workspaces describe prints one workspace",
			resourcesJSON:    workspacesDescribeOneResourcesJSON,
			expectedDisplays: []string{expectedDescribeWorkspaceOneDisplay},
		},
		{
			name:             "describe workspaces describe prints two workspaces",
			resourcesJSON:    workspacesDescribeTwoResourcesJSON,
			expectedDisplays: []string{expectedDescribeWorkspaceTwoDisplay},
		},
		{
			name:             "describe workspaces describe prints workspace with sections",
			resourcesJSON:    workspacesDescribeWithSectionsResourcesJSON,
			expectedDisplays: []string{expectedDescribeWorkspaceWithSectionsOneDisplay},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var resources types.Workspaces
			err := json.Unmarshal([]byte(tt.resourcesJSON), &resources)
			require.NoError(t, err)
			stubUI := &stubs.StubUI{}
			err = workbench.DescribeWorkspaces(stubUI, resources)
			require.NoError(t, err)
			require.Equal(t, tt.expectedDisplays, stubUI.Displays, strings.Join(stubUI.Displays, ""))
			require.Equal(t, tt.expectedWarns, stubUI.Warns, strings.Join(stubUI.Warns, ""))
		})
	}
}

func getWorkspaceUpdateEndpoints(
	expectedBody string,
	responseStatus int,
	responseBody string,
) []stubs.StubHTTPEndpoint {
	return []stubs.StubHTTPEndpoint{
		{
			Request: stubs.StubHTTPRequest{
				Method:          http.MethodPost,
				URL:             postgrest.Endpoint(constants.WorkbenchEndpointRPCWorkspaceUpdate),
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

func getWorkspaceBulkUpdateEndpoints(
	expectedBody string,
	responseStatus int,
	responseBody string,
) []stubs.StubHTTPEndpoint {
	return []stubs.StubHTTPEndpoint{
		{
			Request: stubs.StubHTTPRequest{
				Method:          http.MethodPost,
				URL:             postgrest.Endpoint(constants.WorkbenchEndpointRPCWorkspaceBulkUpdate),
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

func getWorkspaceDeleteEndpoints(
	expectedBody string,
	responseStatus int,
	responseBody string,
) []stubs.StubHTTPEndpoint {
	return []stubs.StubHTTPEndpoint{
		{
			Request: stubs.StubHTTPRequest{
				Method:          http.MethodPost,
				URL:             postgrest.Endpoint(constants.WorkbenchEndpointRPCWorkspaceBulkDelete),
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

func getWorkspaceCreateEndpoints(
	expectedBody string,
	responseStatus int,
	responseBody string,
) []stubs.StubHTTPEndpoint {
	return []stubs.StubHTTPEndpoint{
		{
			Request: stubs.StubHTTPRequest{
				Method:          http.MethodPost,
				URL:             postgrest.Endpoint(constants.WorkbenchEndpointRPCWorkspaceCreate),
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

func getWorkspaceGetEndpoints(
	exectedQuery string,
	responseStatus int,
	responseBody string,
) []stubs.StubHTTPEndpoint {
	return []stubs.StubHTTPEndpoint{
		{
			Request: stubs.StubHTTPRequest{
				Method:          http.MethodGet,
				URL:             postgrest.Endpoint(constants.WorkbenchEndpointWorkspaces),
				ExpectedHeaders: authHeaders,
				ExpectedQuery:   exectedQuery,
			},
			Response: stubs.StubHTTPResponse{
				Status: responseStatus,
				Body:   responseBody,
			},
		},
	}
}
