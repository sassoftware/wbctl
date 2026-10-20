// Copyright © 2026, SAS Institute Inc., Cary, NC, USA.  All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package workbench_test

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/sassoftware/wbctl/internal/constants"
	"github.com/sassoftware/wbctl/internal/postgrest"
	"github.com/sassoftware/wbctl/internal/services/workbench"
	"github.com/sassoftware/wbctl/internal/stubs"
	"github.com/sassoftware/wbctl/internal/types"
	"github.com/stretchr/testify/require"
)

const (
	nodepoolsEmptyResourcesJSON = `[]`

	nodepoolsGetOneResponseJSON = `
[
  {
    "id": "np-1",
    "display_name": "Nodepool A",
    "description": "first",
    "dataplane_id": "dp-1"
  }
]
`

	nodepoolsGetTwoResponseJSON = `
[
  {
    "id": "id-a",
    "display_name": "Nodepool A",
    "description": "A",
    "dataplane_id": "dp-a"
  },
  {
    "id": "id-b",
    "display_name": "Nodepool B",
    "description": "B",
    "dataplane_id": "dp-b"
  }
]
`

	nodepoolsPrintTwoResourcesJSON = `
[
  {
    "id": "np-1",
    "display_name": "Nodepool A",
    "description": "first"
  },
  {
    "id": "np-2",
    "display_name": "Nodepool B",
    "description": "second"
  }
]
`

	expectedNodepoolsEmptyTable   = "ID\tNAME\tDESCRIPTION"
	expectedNodepoolsTwoRowsTable = "ID\tNAME\tDESCRIPTION\nnp-1\tNodepool A\tfirst\nnp-2\tNodepool B\tsecond"
)

var (
	expectedResponseOneNodepool = types.Nodepools{
		{ID: "np-1", DisplayName: "Nodepool A", Description: "first", DataplaneID: "dp-1"},
	}

	expectedResponseTwoNodepools = types.Nodepools{
		{ID: "id-a", DisplayName: "Nodepool A", Description: "A", DataplaneID: "dp-a"},
		{ID: "id-b", DisplayName: "Nodepool B", Description: "B", DataplaneID: "dp-b"},
	}
)

func TestGetNodepools(t *testing.T) {
	tests := []getTestCase[types.Nodepools]{
		{
			name:    "get nodepools success without ids query",
			request: getRequest{},
			mockEndpoints: getNodepoolsGetEndpoints(
				"",
				http.StatusOK,
				nodepoolsGetOneResponseJSON,
			),
			expected: getExpected[types.Nodepools]{
				result: expectedResponseOneNodepool,
			},
		},
		{
			name: "get nodepools sort by name asc",
			request: getRequest{
				sort: "name.asc",
			},
			mockEndpoints: getNodepoolsGetEndpoints(
				"order=display_name.asc",
				http.StatusOK,
				nodepoolsGetOneResponseJSON,
			),
			expected: getExpected[types.Nodepools]{
				result: expectedResponseOneNodepool,
			},
		},
		{
			name: "get nodepools sort by name desc",
			request: getRequest{
				sort: "name.desc",
			},
			mockEndpoints: getNodepoolsGetEndpoints(
				"order=display_name.desc",
				http.StatusOK,
				nodepoolsGetOneResponseJSON,
			),
			expected: getExpected[types.Nodepools]{
				result: expectedResponseOneNodepool,
			},
		},
		{
			name: "get nodepools success with ids query",
			request: getRequest{
				ids: []string{"id-a", "id-b"},
			},
			mockEndpoints: getNodepoolsGetEndpoints(
				"id=in.(id-a,id-b)",
				http.StatusOK,
				nodepoolsGetTwoResponseJSON,
			),
			expected: getExpected[types.Nodepools]{
				result: expectedResponseTwoNodepools,
			},
		},
		{
			name: "get nodepools sort by id",
			request: getRequest{
				sort: "id",
			},
			mockEndpoints: getNodepoolsGetEndpoints(
				"order=id",
				http.StatusOK,
				nodepoolsGetOneResponseJSON,
			),
			expected: getExpected[types.Nodepools]{
				result: expectedResponseOneNodepool,
			},
		},
		{
			name: "get nodepools sort by description",
			request: getRequest{
				sort: "description",
			},
			mockEndpoints: getNodepoolsGetEndpoints(
				"order=description",
				http.StatusOK,
				nodepoolsGetOneResponseJSON,
			),
			expected: getExpected[types.Nodepools]{
				result: expectedResponseOneNodepool,
			},
		},
		{
			name: "get nodepools sort by name and description",
			request: getRequest{
				sort: "name.desc,description.asc",
			},
			mockEndpoints: getNodepoolsGetEndpoints(
				"order=display_name.desc,description.asc",
				http.StatusOK,
				nodepoolsGetOneResponseJSON,
			),
			expected: getExpected[types.Nodepools]{
				result: expectedResponseOneNodepool,
			},
		},
		{
			name: "get nodepools invalid json response",
			request: getRequest{
				ids: []string{"id-a"},
			},
			mockEndpoints: getNodepoolsGetEndpoints(
				"id=in.(id-a)",
				http.StatusOK,
				"[",
			),
			expected: getExpected[types.Nodepools]{
				err: "GET ${URL}/postgrest/dataplane_nodepools?id=in.%28id-a%29 failed\n Caused by:\n\t* unexpected EOF",
			},
		},
		{
			name: "get nodepools non-200 from server",
			request: getRequest{
				ids: []string{"id-a"},
			},
			mockEndpoints: getNodepoolsGetEndpoints(
				"id=in.(id-a)",
				http.StatusUnauthorized,
				"unauthorized",
			),
			expected: getExpected[types.Nodepools]{
				err: "GET ${URL}/postgrest/dataplane_nodepools?id=in.%28id-a%29 failed\n Caused by:\n\t* 401 unauthorized",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv, _ := stubs.NewStubHTTPServer(t, tt.mockEndpoints)
			defer srv.Close()
			svc := workbench.NewWorkbenchService(srv.URL, "access-token", "csrf-token")

			actual, err := workbench.GetNodepools(context.Background(), svc, tt.request.ids, tt.request.sort)
			if tt.expected.err != "" {
				require.EqualError(t, err, strings.ReplaceAll(tt.expected.err, "${URL}", srv.URL))
				return
			}

			require.NoError(t, err)
			require.Equal(t, tt.expected.result, actual)
		})
	}
}

func TestGetAllNodepoolIDs(t *testing.T) {
	tests := []getTestCase[[]string]{
		{
			name: "get nodepool ids success",
			mockEndpoints: getNodepoolsGetEndpoints(
				"",
				http.StatusOK,
				nodepoolsGetTwoResponseJSON,
			),
			expected: getExpected[[]string]{
				result: []string{"id-a", "id-b"},
			},
		},
		{
			name: "get nodepool ids invalid json response",
			mockEndpoints: getNodepoolsGetEndpoints(
				"",
				http.StatusOK,
				"[",
			),
			expected: getExpected[[]string]{
				err: "GET ${URL}/postgrest/dataplane_nodepools failed\n Caused by:\n\t* unexpected EOF",
			},
		},
		{
			name: "get nodepool ids non-200 from server",
			mockEndpoints: getNodepoolsGetEndpoints(
				"",
				http.StatusUnauthorized,
				"unauthorized",
			),
			expected: getExpected[[]string]{
				err: "GET ${URL}/postgrest/dataplane_nodepools failed\n Caused by:\n\t* 401 unauthorized",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv, _ := stubs.NewStubHTTPServer(t, tt.mockEndpoints)
			defer srv.Close()
			svc := workbench.NewWorkbenchService(srv.URL, "access-token", "csrf-token")

			actual, err := workbench.GetNodepoolIDs(context.Background(), svc)
			if tt.expected.err != "" {
				require.EqualError(t, err, strings.ReplaceAll(tt.expected.err, "${URL}", srv.URL))
				return
			}

			require.NoError(t, err)
			require.Equal(t, tt.expected.result, actual)
		})
	}
}

func TestPrintNodepools(t *testing.T) {
	tests := []struct {
		name               string
		resourcesJSON      string
		expectedTableCount int
		expectedTable      string
		expectedPrintCalls int
	}{
		{
			name:               "print nodepools empty resources",
			resourcesJSON:      nodepoolsEmptyResourcesJSON,
			expectedTableCount: 1,
			expectedTable:      expectedNodepoolsEmptyTable,
			expectedPrintCalls: 1,
		},
		{
			name:               "print nodepools prints resources",
			resourcesJSON:      nodepoolsPrintTwoResourcesJSON,
			expectedTableCount: 1,
			expectedTable:      expectedNodepoolsTwoRowsTable,
			expectedPrintCalls: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var resources types.Nodepools
			err := json.Unmarshal([]byte(tt.resourcesJSON), &resources)
			require.NoError(t, err)

			u := &stubs.StubUI{}
			err = workbench.PrintNodepools(u, "ignored-token", resources)
			require.NoError(t, err)

			require.Len(t, u.Tables, tt.expectedTableCount)
			tbl := u.Tables[0]
			require.Equal(t, tt.expectedTable, tbl.String())
			require.Equal(t, tt.expectedPrintCalls, tbl.PrintCalls)
		})
	}
}

func getNodepoolsGetEndpoints(expectedQuery string, responseStatus int, responseBody string) []stubs.StubHTTPEndpoint {
	return []stubs.StubHTTPEndpoint{
		{
			Request: stubs.StubHTTPRequest{
				Method:          http.MethodGet,
				URL:             postgrest.Endpoint(constants.WorkbenchEndpointDataplaneNodepools),
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
