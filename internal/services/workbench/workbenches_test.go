// Copyright © 2026, SAS Institute Inc., Cary, NC, USA.  All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package workbench_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/sassoftware/wbctl/internal/constants"
	"github.com/sassoftware/wbctl/internal/i18n"
	"github.com/sassoftware/wbctl/internal/postgrest"
	"github.com/sassoftware/wbctl/internal/services/workbench"
	"github.com/sassoftware/wbctl/internal/stubs"
	"github.com/sassoftware/wbctl/internal/timeutil"
	"github.com/sassoftware/wbctl/internal/types"
	"github.com/stretchr/testify/require"
)

const (
	workbenchesEmptyResourcesJSON = `[]`

	workbenchesGetOneResponseJSON = `
[
  {
    "id": "wb-1",
    "controller_id": "ctrl-1",
    "name": "wb-one",
    "username": "alice",
    "application_id": "app-1",
    "url": "https://wb/1",
    "current_state": "Running",
    "message": "ok"
  }
]
`

	workbenchesGetMixedIDsResponseJSON = `
[
  {
    "id": "123e4567-e89b-12d3-a456-426614174000",
    "controller_id": "ctrl-uuid",
    "name": "wb-uuid",
    "username": "alice",
    "application_id": "app-1"
  },
  {
    "id": "wb-controller",
    "controller_id": "controller-wb-2",
    "name": "wb-controller",
    "username": "bob",
    "application_id": "app-2"
  }
]
`

	workbenchesPrintOneResourcesJSON = `
[
  {
    "id": "wb-1",
    "controller_id": "",
    "name": "wb-one",
    "username": "alice",
    "application_id": "app-1",
    "services": [{"id":"svc-a"},{"id":"svc-b"}],
    "url": "https://wb/1",
    "current_state": "Running",
    "message": "ok",
    "created_at": null
  }
]
`

	workbenchesPrintTwoResourcesJSON = `
[
  {
    "id": "wb-1",
    "controller_id": "",
    "name": "wb-one",
    "username": "alice",
    "application_id": "app-1",
    "services": [{"id":"svc-a"},{"id":"svc-b"}],
    "url": "https://wb/1",
    "current_state": "Running",
    "message": "ok",
    "created_at": null
  },
  {
    "id": "wb-2",
    "controller_id": "ctrl-2",
    "name": "wb-two",
    "username": "bob",
    "application_id": "app-2",
    "services": [{"image":"img-only"}],
    "url": "https://wb/2",
    "current_state": "Stopped",
    "message": "idle",
    "created_at": null
  }
]
`

	workbenchesDescribeOneResourcesJSON = `
[
  {
    "id": "wb-1",
    "controller_id": "ctrl-1",
    "name": "wb-one",
    "url": "https://wb/1",
    "username": "alice",
    "email": "alice@example.com",
    "created_at": "2026-05-08T00:00:00Z",
    "modified_at": "2026-05-08T00:00:00Z",
    "nodepool": "np-a",
    "desired_state": "Running",
    "current_state": "Running",
    "message": "ok",
    "application_id": "app-1"
  }
]
`

	workbenchesDescribeTwoResourcesJSON = `
[
  {
    "id": "wb-1",
    "controller_id": "ctrl-1",
    "name": "wb-one",
    "url": "https://wb/1",
    "username": "alice",
    "email": "alice@example.com",
    "created_at": "2026-05-08T00:00:00Z",
    "modified_at": "2026-05-08T00:00:00Z",
    "nodepool": "np-a",
    "desired_state": "Running",
    "current_state": "Running",
    "message": "ok",
    "application_id": "app-1"
  },
  {
    "id": "wb-2",
    "controller_id": "ctrl-2",
    "name": "wb-two",
    "url": "https://wb/2",
    "username": "bob",
    "email": "bob@example.com",
    "created_at": "2026-05-08T00:00:00Z",
    "modified_at": "2026-05-08T00:00:00Z",
    "nodepool": "np-a",
    "desired_state": "Running",
    "current_state": "Running",
    "message": "ok",
    "application_id": "app-1"
  }
]
`

	workbenchesDescribeWithSectionsResourcesJSON = `
[
  {
    "id": "wb-2",
    "controller_id": "ctrl-2",
    "name": "wb-two",
    "url": "https://wb/2",
    "username": "bob",
    "email": "bob@example.com",
    "created_at": "2026-05-08T00:00:00Z",
    "modified_at": "2026-05-08T00:00:00Z",
    "nodepool": "np-b",
    "desired_state": "Stopped",
    "current_state": "Stopped",
    "labels": {
      "team": "analytics"
    },
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
    "properties": {
      "owner": "data"
    }
  }
]
`

	workbenchesBulkUpdateResponseJSON = `
[
  {
    "id": "wb-1",
    "name": "wb-one-updated"
  },
  {
    "id": "wb-2",
    "name": "wb-two-updated"
  }
]
`

	expectedWorkbenchesEmptyTable = "" +
		"ID\tCONTROLLER_ID\tNAME\tAPPLICATION_ID\tSERVICES\tURL\tSTATE\tMESSAGE\tAGE"

	expectedWorkbenchesOneRowTable = "" +
		"ID\tCONTROLLER_ID\tNAME\tAPPLICATION_ID\tSERVICES\tURL\tSTATE\tMESSAGE\tAGE\n" +
		"wb-1\t<unknown>\twb-one\tapp-1\tsvc-a, svc-b\thttps://wb/1\tRunning\tok\t"

	expectedWorkbenchesTwoRowsTable = "" +
		"ID\tCONTROLLER_ID\tNAME\tAPPLICATION_ID\tSERVICES\tURL\tSTATE\tMESSAGE\tAGE\n" +
		"wb-1\t<unknown>\twb-one\tapp-1\tsvc-a, svc-b\thttps://wb/1\tRunning\tok\t\n" +
		"wb-2\tctrl-2\twb-two\tapp-2\timg-only\thttps://wb/2\tStopped\tidle\t"

	expectedWorkbenchesAdminEmptyTable = "" +
		"ID\tCONTROLLER_ID\tNAME\tUSERNAME\tAPPLICATION_ID\tSERVICES\tURL\tSTATE\tMESSAGE\tAGE"

	expectedWorkbenchesAdminOneRowTable = "" +
		"ID\tCONTROLLER_ID\tNAME\tUSERNAME\tAPPLICATION_ID\tSERVICES\tURL\tSTATE\tMESSAGE\tAGE\n" +
		"wb-1\t<unknown>\twb-one\talice\tapp-1\tsvc-a, svc-b\thttps://wb/1\tRunning\tok\t"

	expectedDescribeWorkbenchOneDisplay = `ID:              wb-1
Controller ID:   ctrl-1
Name:            wb-one
URL:             https://wb/1
Created By:      alice (alice@example.com)
Created At:      2026-05-08T00:00:00Z (0s ago)
Modified At:     2026-05-08T00:00:00Z (0s ago)
Nodepool:        np-a
Desired State:   Running
Current State:   Running
Message:         ok
Application:     app-1
`
	expectedDescribeWorkbenchTwoDisplay = `ID:              wb-1
Controller ID:   ctrl-1
Name:            wb-one
URL:             https://wb/1
Created By:      alice (alice@example.com)
Created At:      2026-05-08T00:00:00Z (0s ago)
Modified At:     2026-05-08T00:00:00Z (0s ago)
Nodepool:        np-a
Desired State:   Running
Current State:   Running
Message:         ok
Application:     app-1


ID:              wb-2
Controller ID:   ctrl-2
Name:            wb-two
URL:             https://wb/2
Created By:      bob (bob@example.com)
Created At:      2026-05-08T00:00:00Z (0s ago)
Modified At:     2026-05-08T00:00:00Z (0s ago)
Nodepool:        np-a
Desired State:   Running
Current State:   Running
Message:         ok
Application:     app-1
`
	expectedDescribeWorkbenchWithSectionsOneDisplay = `ID:              wb-2
Controller ID:   ctrl-2
Name:            wb-two
URL:             https://wb/2
Created By:      bob (bob@example.com)
Created At:      2026-05-08T00:00:00Z (0s ago)
Modified At:     2026-05-08T00:00:00Z (0s ago)
Nodepool:        np-b
Desired State:   Stopped
Current State:   Stopped
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
	expectedInvalidSingleUpdateResponseErr = "" +
		"POST ${URL}/postgrest/rpc/workbench_update failed" +
		"\n Caused by:" +
		"\n\t* invalid character 'i' looking for beginning of value"
	expectedDeleteWorkbenchRequestJSON     = `{"ids":["wb-1","wb-2"]}`
	createWorkbenchResponseJSON            = `{"id":"wb-1","name":"wb-one"}`
	expectedBulkUpdateWorkbenchRequestJSON = `{"ids":["wb-1","wb-2"],"workbench":{"desired_state":"STOPPED"}}`
	expectedUpdateWorkbenchRequestJSON     = `{"param":{"id":"wb-1","name":"UPDATED"}}`
	expectedUpdateWorkbenchRequest2JSON    = `{"param":{"id":"wb-1","name":"wb-one-renamed"}}`
)

var (
	expectedResponseOneWorkbench = types.Workbenches{
		{
			ID:            "wb-1",
			ControllerID:  "ctrl-1",
			Name:          "wb-one",
			Username:      "alice",
			ApplicationID: "app-1",
			URL:           "https://wb/1",
			CurrentState:  "Running",
			Message:       "ok",
		},
	}

	expectedResponseTwoWorkbenches = types.Workbenches{
		{
			ID:            "123e4567-e89b-12d3-a456-426614174000",
			ControllerID:  "ctrl-uuid",
			Name:          "wb-uuid",
			Username:      "alice",
			ApplicationID: "app-1",
		},
		{
			ID:            "wb-controller",
			ControllerID:  "controller-wb-2",
			Name:          "wb-controller",
			Username:      "bob",
			ApplicationID: "app-2",
		},
	}
)

func TestDeleteWorkbench(t *testing.T) {
	tests := []deleteTestCase{
		{
			name: "delete workbench success",
			request: deleteRequest{
				ids: []string{"wb-1", "wb-2"},
			},
			mockEndpoints: getWorkbenchesDeleteEndpoints(
				expectedDeleteWorkbenchRequestJSON,
				http.StatusOK,
				`["wb-1", "wb-2"]`,
			),
			expected: deleteExpected{
				result: []string{"wb-1", "wb-2"},
			},
		},
		{
			name: "delete workbench non-200 from server",
			request: deleteRequest{
				ids: []string{"wb-1", "wb-2"},
			},
			mockEndpoints: getWorkbenchesDeleteEndpoints(
				expectedDeleteWorkbenchRequestJSON,
				http.StatusUnauthorized,
				"unauthorized",
			),
			expected: deleteExpected{
				err: "POST %s/postgrest/rpc/workbench_bulk_delete failed\n" +
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
			actual, err := workbench.BulkDeleteWorkbenches(context.Background(), svc, tt.request.ids)
			if tt.expected.err != "" {
				require.EqualError(t, err, fmt.Sprintf(tt.expected.err, srv.URL))
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.expected.result, actual)
		})
	}
}

func TestCreateWorkbench(t *testing.T) {
	tests := []createTestCase[types.Workbench]{
		{
			name: "create workbench success",
			request: createRequest[types.Workbench]{
				body: types.Workbench{
					Name: "wb-one",
				},
			},
			mockEndpoints: getWorkbenchesCreateEndpoints(
				stubs.AsString(
					paramPayload[types.Workbench]{Param: types.Workbench{Name: "wb-one"}},
				),
				http.StatusOK,
				createWorkbenchResponseJSON,
			),
			expected: createExpected[types.Workbench]{
				result: types.Workbench{
					ID:   "wb-1",
					Name: "wb-one",
				},
			},
		},
		{
			name: "create workbench non-200 from server",
			request: createRequest[types.Workbench]{
				body: types.Workbench{
					Name: "wb-one",
				},
			},
			mockEndpoints: getWorkbenchesCreateEndpoints(
				stubs.AsString(
					paramPayload[types.Workbench]{Param: types.Workbench{Name: "wb-one"}},
				),
				http.StatusUnauthorized,
				"unauthorized",
			),
			expected: createExpected[types.Workbench]{
				err: "POST %s/postgrest/rpc/workbench_create failed\n" +
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

			actual, err := workbench.CreateWorkbench(context.Background(), svc, tt.request.body)
			if tt.expected.err != "" {
				require.EqualError(t, err, fmt.Sprintf(tt.expected.err, srv.URL))
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.expected.result, actual)
		})
	}
}

func TestBulkUpdateWorkbenches(t *testing.T) {
	tests := []bulkupdateTestCase[types.Workbench, types.Workbenches]{
		{
			name: "bulk update workbenches success",
			request: bulkupdateRequest[types.Workbench]{
				ids:  []string{"wb-1", "wb-2"},
				body: types.Workbench{DesiredState: "STOPPED"},
			},
			mockEndpoints: getWorkbenchesBulkUpdateEndpoints(
				expectedBulkUpdateWorkbenchRequestJSON,
				http.StatusOK,
				workbenchesBulkUpdateResponseJSON,
			),
			expected: bulkupdateExpected[types.Workbenches]{
				result: types.Workbenches{
					{ID: "wb-1", Name: "wb-one-updated"},
					{ID: "wb-2", Name: "wb-two-updated"},
				},
			},
		},
		{
			name: "bulk update workbenches non-200 from server",
			request: bulkupdateRequest[types.Workbench]{
				ids:  []string{"wb-1", "wb-2"},
				body: types.Workbench{DesiredState: "STOPPED"},
			},
			mockEndpoints: getWorkbenchesBulkUpdateEndpoints(
				expectedBulkUpdateWorkbenchRequestJSON,
				http.StatusUnauthorized,
				"unauthorized",
			),
			expected: bulkupdateExpected[types.Workbenches]{
				err: "POST ${URL}/postgrest/rpc/workbench_bulk_update failed\n" +
					" Caused by:\n" +
					"\t* 401 unauthorized",
			},
		},
		{
			name: "bulk update workbenches invalid json response",
			request: bulkupdateRequest[types.Workbench]{
				ids:  []string{"wb-1", "wb-2"},
				body: types.Workbench{DesiredState: "STOPPED"},
			},
			mockEndpoints: getWorkbenchesBulkUpdateEndpoints(
				expectedBulkUpdateWorkbenchRequestJSON,
				http.StatusOK,
				"[",
			),
			expected: bulkupdateExpected[types.Workbenches]{
				err: "POST ${URL}/postgrest/rpc/workbench_bulk_update failed\n" +
					" Caused by:\n" +
					"\t* unexpected EOF",
			},
		},
		{
			name: "bulk update workbench name with single id uses single update endpoint",
			request: bulkupdateRequest[types.Workbench]{
				ids:  []string{"wb-1"},
				body: types.Workbench{Name: "wb-one-renamed"},
			},
			mockEndpoints: getWorkbenchesUpdateEndpoints(
				expectedUpdateWorkbenchRequest2JSON,
				http.StatusOK,
				`{"id":"wb-1","name":"wb-one-renamed"}`,
			),
			expected: bulkupdateExpected[types.Workbenches]{
				result: types.Workbenches{{ID: "wb-1", Name: "wb-one-renamed"}},
			},
		},
		{
			name: "bulk update workbench name with multiple ids returns validation error without rpc",
			request: bulkupdateRequest[types.Workbench]{
				ids:  []string{"wb-1", "wb-2"},
				body: types.Workbench{Name: "wb-one-renamed"},
			},
			expected: bulkupdateExpected[types.Workbenches]{
				err: i18n.Format(i18n.BulkUpdateNameError,
					i18n.A{constants.I18nArgKeyType: constants.WorkbenchesType}),
			},
		},
		{
			name: "update single workbench name invalid response",
			request: bulkupdateRequest[types.Workbench]{
				ids:  []string{"wb-1"},
				body: types.Workbench{Name: "UPDATED"},
			},
			mockEndpoints: getWorkbenchesUpdateEndpoints(
				expectedUpdateWorkbenchRequestJSON,
				http.StatusOK,
				"[invalid",
			),
			expected: bulkupdateExpected[types.Workbenches]{
				err: expectedInvalidSingleUpdateResponseErr,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv, _ := stubs.NewStubHTTPServer(t, tt.mockEndpoints)
			defer srv.Close()
			svc := workbench.NewWorkbenchService(srv.URL, "access-token", "csrf-token")

			actual, err := workbench.BulkUpdateWorkbenches(context.Background(), svc, tt.request.ids, tt.request.body)
			if tt.expected.err != "" {
				require.EqualError(t, err, strings.ReplaceAll(tt.expected.err, "${URL}", srv.URL))
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.expected.result, actual)
		})
	}
}

func TestGetWorkbenches(t *testing.T) {
	tests := []getTestCase[types.Workbenches]{
		{
			name:    "get workbench success without ids query",
			request: getRequest{},
			mockEndpoints: getWorkbenchesGetEndpoints(
				"",
				http.StatusOK,
				workbenchesGetOneResponseJSON,
			),
			expected: getExpected[types.Workbenches]{
				result: expectedResponseOneWorkbench,
			},
		},
		{
			name: "get workbench success with mixed uuid and controller ids query",
			request: getRequest{
				ids:  []string{"123e4567-e89b-12d3-a456-426614174000", "controller-wb-2"},
				sort: "id.asc",
			},
			mockEndpoints: getWorkbenchesGetEndpoints(
				"or=(id.in.(123e4567-e89b-12d3-a456-426614174000),"+
					"controller_id.in.(controller-wb-2))&order=id.asc",
				http.StatusOK,
				workbenchesGetMixedIDsResponseJSON,
			),
			expected: getExpected[types.Workbenches]{
				result: expectedResponseTwoWorkbenches,
			},
		},
		{
			name: "get workbench sort by controller_id",
			request: getRequest{
				sort: "controller_id",
			},
			mockEndpoints: getWorkbenchesGetEndpoints(
				"order=controller_id",
				http.StatusOK,
				workbenchesGetOneResponseJSON,
			),
			expected: getExpected[types.Workbenches]{
				result: expectedResponseOneWorkbench,
			},
		},
		{
			name: "get workbench sort by name",
			request: getRequest{
				sort: "name",
			},
			mockEndpoints: getWorkbenchesGetEndpoints(
				"order=name",
				http.StatusOK,
				workbenchesGetOneResponseJSON,
			),
			expected: getExpected[types.Workbenches]{
				result: expectedResponseOneWorkbench,
			},
		},
		{
			name: "get workbench sort by username",
			request: getRequest{
				sort: "username",
			},
			mockEndpoints: getWorkbenchesGetEndpoints(
				"order=username",
				http.StatusOK,
				workbenchesGetOneResponseJSON,
			),
			expected: getExpected[types.Workbenches]{
				result: expectedResponseOneWorkbench,
			},
		},
		{
			name: "get workbench sort by application_id",
			request: getRequest{
				sort: "application_id",
			},
			mockEndpoints: getWorkbenchesGetEndpoints(
				"order=application_id",
				http.StatusOK,
				workbenchesGetOneResponseJSON,
			),
			expected: getExpected[types.Workbenches]{
				result: expectedResponseOneWorkbench,
			},
		},
		{
			name: "get workbench sort by services",
			request: getRequest{
				sort: "services",
			},
			mockEndpoints: getWorkbenchesGetEndpoints(
				"order=services",
				http.StatusOK,
				workbenchesGetOneResponseJSON,
			),
			expected: getExpected[types.Workbenches]{
				result: expectedResponseOneWorkbench,
			},
		},
		{
			name: "get workbench sort by url",
			request: getRequest{
				sort: "url",
			},
			mockEndpoints: getWorkbenchesGetEndpoints(
				"order=url",
				http.StatusOK,
				workbenchesGetOneResponseJSON,
			),
			expected: getExpected[types.Workbenches]{
				result: expectedResponseOneWorkbench,
			},
		},
		{
			name: "get workbench sort by state",
			request: getRequest{
				sort: "state",
			},
			mockEndpoints: getWorkbenchesGetEndpoints(
				"order=current_state",
				http.StatusOK,
				workbenchesGetOneResponseJSON,
			),
			expected: getExpected[types.Workbenches]{
				result: expectedResponseOneWorkbench,
			},
		},
		{
			name: "get workbench sort by message",
			request: getRequest{
				sort: "message",
			},
			mockEndpoints: getWorkbenchesGetEndpoints(
				"order=message",
				http.StatusOK,
				workbenchesGetOneResponseJSON,
			),
			expected: getExpected[types.Workbenches]{
				result: expectedResponseOneWorkbench,
			},
		},
		{
			name: "get workbench sort by age",
			request: getRequest{
				sort: "age",
			},
			mockEndpoints: getWorkbenchesGetEndpoints(
				"order=created_at",
				http.StatusOK,
				workbenchesGetOneResponseJSON,
			),
			expected: getExpected[types.Workbenches]{
				result: expectedResponseOneWorkbench,
			},
		},
		{
			name: "get workbench sort by age and state",
			request: getRequest{
				sort: "age.desc,state.asc",
			},
			mockEndpoints: getWorkbenchesGetEndpoints(
				"order=created_at.desc,current_state.asc",
				http.StatusOK,
				workbenchesGetOneResponseJSON,
			),
			expected: getExpected[types.Workbenches]{
				result: expectedResponseOneWorkbench,
			},
		},
		{
			name: "get workbench invalid json response",
			request: getRequest{
				ids: []string{"id-a"},
			},
			mockEndpoints: getWorkbenchesGetEndpoints(
				"controller_id=in.(id-a)",
				http.StatusOK,
				"[",
			),
			expected: getExpected[types.Workbenches]{
				err: "GET ${URL}/postgrest/workbenches?controller_id=in.%28id-a%29 failed\n" +
					" Caused by:\n" +
					"\t* unexpected EOF",
			},
		},
		{
			name: "get workbench non-200 from server",
			request: getRequest{
				ids: []string{"id-a"},
			},
			mockEndpoints: getWorkbenchesGetEndpoints(
				"controller_id=in.(id-a)",
				http.StatusUnauthorized,
				"unauthorized",
			),
			expected: getExpected[types.Workbenches]{
				err: "GET ${URL}/postgrest/workbenches?controller_id=in.%28id-a%29 failed\n" +
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

			actual, err := workbench.GetWorkbenches(context.Background(), svc, tt.request.ids, tt.request.sort)
			if tt.expected.err != "" {
				require.EqualError(t, err, strings.ReplaceAll(tt.expected.err, "${URL}", srv.URL))
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.expected.result, actual)
		})
	}
}

func TestPrintWorkbenches(t *testing.T) {
	tests := []struct {
		name               string
		resourcesJSON      string
		token              string
		expectedTableCount int
		expectedTable      string
		expectedPrintCalls int
	}{
		{
			name:               "print workbenches empty resources with non-admin token",
			resourcesJSON:      workbenchesEmptyResourcesJSON,
			token:              "token-without-admin-role",
			expectedTableCount: 1,
			expectedTable:      expectedWorkbenchesEmptyTable,
			expectedPrintCalls: 1,
		},
		{
			name:               "print workbenches empty resources with admin token",
			resourcesJSON:      workbenchesEmptyResourcesJSON,
			token:              createTokenWithRole(constants.WorkbenchAdminRole),
			expectedTableCount: 1,
			expectedTable:      expectedWorkbenchesAdminEmptyTable,
			expectedPrintCalls: 1,
		},
		{
			name:               "print workbenches prints one resource with non-admin token",
			resourcesJSON:      workbenchesPrintOneResourcesJSON,
			token:              "token-without-admin-role",
			expectedTableCount: 1,
			expectedTable:      expectedWorkbenchesOneRowTable,
			expectedPrintCalls: 1,
		},
		{
			name:               "print workbenches prints one resource with admin token",
			resourcesJSON:      workbenchesPrintOneResourcesJSON,
			token:              createTokenWithRole(constants.WorkbenchAdminRole),
			expectedTableCount: 1,
			expectedTable:      expectedWorkbenchesAdminOneRowTable,
			expectedPrintCalls: 1,
		},
		{
			name:               "print workbenches prints two resources with non-admin token",
			resourcesJSON:      workbenchesPrintTwoResourcesJSON,
			token:              "token-without-admin-role",
			expectedTableCount: 1,
			expectedTable:      expectedWorkbenchesTwoRowsTable,
			expectedPrintCalls: 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var resources types.Workbenches
			err := json.Unmarshal([]byte(tt.resourcesJSON), &resources)
			require.NoError(t, err)
			u := &stubs.StubUI{}
			err = workbench.PrintWorkbenches(u, tt.token, resources)
			require.NoError(t, err)
			require.Len(t, u.Tables, tt.expectedTableCount)
			tbl := u.Tables[0]
			require.Equal(t, tt.expectedTable, tbl.String())
			require.Equal(t, tt.expectedPrintCalls, tbl.PrintCalls)
		})
	}
}

func TestDescribeWorkbenches(t *testing.T) {
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
			name:             "describe workbenches describe empty resources prints nothing",
			resourcesJSON:    workbenchesEmptyResourcesJSON,
			expectedDisplays: []string{""},
		},
		{
			name:             "describe workbenches describe prints one workbench",
			resourcesJSON:    workbenchesDescribeOneResourcesJSON,
			expectedDisplays: []string{expectedDescribeWorkbenchOneDisplay},
		},
		{
			name:             "describe workbenches describe prints two workbenches",
			resourcesJSON:    workbenchesDescribeTwoResourcesJSON,
			expectedDisplays: []string{expectedDescribeWorkbenchTwoDisplay},
		},
		{
			name:             "describe workbenches describe prints workbench with sections",
			resourcesJSON:    workbenchesDescribeWithSectionsResourcesJSON,
			expectedDisplays: []string{expectedDescribeWorkbenchWithSectionsOneDisplay},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var resources types.Workbenches
			err := json.Unmarshal([]byte(tt.resourcesJSON), &resources)
			require.NoError(t, err)
			stubUI := &stubs.StubUI{}
			err = workbench.DescribeWorkbenches(stubUI, resources)
			require.NoError(t, err)
			require.Equal(t, tt.expectedDisplays, stubUI.Displays, strings.Join(stubUI.Displays, ""))
			require.Equal(t, tt.expectedWarns, stubUI.Warns, strings.Join(stubUI.Warns, ""))
		})
	}
}

func getWorkbenchesDeleteEndpoints(
	expectedBody string,
	responseStatus int,
	responseBody string,
) []stubs.StubHTTPEndpoint {
	return []stubs.StubHTTPEndpoint{
		{
			Request: stubs.StubHTTPRequest{
				Method:          http.MethodPost,
				URL:             postgrest.Endpoint(constants.WorkbenchEndpointRPCWorkbenchBulkDelete),
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

func getWorkbenchesCreateEndpoints(
	expectedBody string,
	responseStatus int,
	responseBody string,
) []stubs.StubHTTPEndpoint {
	return []stubs.StubHTTPEndpoint{
		{
			Request: stubs.StubHTTPRequest{
				Method:          http.MethodPost,
				URL:             postgrest.Endpoint(constants.WorkbenchEndpointRPCWorkbenchCreate),
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

func getWorkbenchesBulkUpdateEndpoints(
	expectedBody string,
	responseStatus int,
	responseBody string,
) []stubs.StubHTTPEndpoint {
	return []stubs.StubHTTPEndpoint{
		{
			Request: stubs.StubHTTPRequest{
				Method:          http.MethodPost,
				URL:             postgrest.Endpoint(constants.WorkbenchEndpointRPCWorkbenchBulkUpdate),
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

func getWorkbenchesUpdateEndpoints(
	expectedBody string,
	responseStatus int,
	responseBody string,
) []stubs.StubHTTPEndpoint {
	return []stubs.StubHTTPEndpoint{
		{
			Request: stubs.StubHTTPRequest{
				Method:          http.MethodPost,
				URL:             postgrest.Endpoint(constants.WorkbenchEndpointRPCWorkbenchUpdate),
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

func getWorkbenchesGetEndpoints(
	expectedQuery string,
	responseStatus int,
	responseBody string,
) []stubs.StubHTTPEndpoint {
	return []stubs.StubHTTPEndpoint{
		{
			Request: stubs.StubHTTPRequest{
				Method:          http.MethodGet,
				URL:             postgrest.Endpoint(constants.WorkbenchEndpointWorkbenches),
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
