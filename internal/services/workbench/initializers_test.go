// Copyright © 2026, SAS Institute Inc., Cary, NC, USA.  All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package workbench_test

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/sas-institute-rnd-saas/wbctl/internal/constants"
	"github.com/sas-institute-rnd-saas/wbctl/internal/postgrest"
	"github.com/sas-institute-rnd-saas/wbctl/internal/services/workbench"
	"github.com/sas-institute-rnd-saas/wbctl/internal/stubs"
	"github.com/sas-institute-rnd-saas/wbctl/internal/types"
	"github.com/stretchr/testify/require"
)

const (
	initializersEmptyResourcesJSON = `[]`

	initializersGetOneResponseJSON = `
[
  {
    "id": "init-1",
    "name": "git-clone",
    "dataplane_id": "dp-1"
  }
]
`

	initializersGetTwoResponseJSON = `
[
  {
    "id": "id-a",
    "name": "initializer-a",
    "dataplane_id": "dp-a"
  },
  {
    "id": "id-b",
    "name": "initializer-b",
    "dataplane_id": "dp-b"
  }
]
`

	initializersPrintTwoResourcesJSON = `
[
  {
    "id": "init-1",
    "name": "git-clone",
    "dataplane_id": "dp-1"
  },
  {
    "id": "init-2",
    "name": "sync-secrets",
    "dataplane_id": "dp-1"
  }
]
`

	expectedInitializersEmptyTable = "ID\tNAME\tTYPE"

	expectedInitializersTwoRowsTable = "ID\tNAME\tTYPE\n" +
		"init-1\tgit-clone\n" +
		"init-2\tsync-secrets"
)

var (
	expectedResponseOneInitializer = types.Initializers{
		{ID: "init-1", Name: "git-clone", DataplaneID: "dp-1"},
	}

	expectedResponseTwoInitializers = types.Initializers{
		{ID: "id-a", Name: "initializer-a", DataplaneID: "dp-a"},
		{ID: "id-b", Name: "initializer-b", DataplaneID: "dp-b"},
	}
)

func TestGetInitializers(t *testing.T) {
	tests := []getTestCase[types.Initializers]{
		{
			name:          "get initializers success without ids query",
			request:       getRequest{},
			mockEndpoints: getInitializersGetEndpoints("", http.StatusOK, initializersGetOneResponseJSON),
			expected: getExpected[types.Initializers]{
				result: expectedResponseOneInitializer,
			},
		},
		{
			name: "get initializers success with ids query",
			request: getRequest{
				ids:  []string{"id-a", "id-b"},
				sort: "id.asc",
			},
			mockEndpoints: getInitializersGetEndpoints(
				"id=in.(id-a,id-b)&order=id.asc",
				http.StatusOK,
				initializersGetTwoResponseJSON,
			),
			expected: getExpected[types.Initializers]{
				result: expectedResponseTwoInitializers,
			},
		},
		{
			name: "get initializers sort by name",
			request: getRequest{
				sort: "name",
			},
			mockEndpoints: getInitializersGetEndpoints(
				"order=name",
				http.StatusOK,
				initializersGetOneResponseJSON,
			),
			expected: getExpected[types.Initializers]{
				result: expectedResponseOneInitializer,
			},
		},
		{
			name: "get initializers sort by type",
			request: getRequest{
				sort: "type",
			},
			mockEndpoints: getInitializersGetEndpoints(
				"order=type",
				http.StatusOK,
				initializersGetOneResponseJSON,
			),
			expected: getExpected[types.Initializers]{
				result: expectedResponseOneInitializer,
			},
		},
		{
			name: "get initializers sort by id and name",
			request: getRequest{
				sort: "id.desc,name.asc",
			},
			mockEndpoints: getInitializersGetEndpoints(
				"order=id.desc,name.asc",
				http.StatusOK,
				initializersGetOneResponseJSON,
			),
			expected: getExpected[types.Initializers]{
				result: expectedResponseOneInitializer,
			},
		},
		{
			name: "get initializers invalid json response",
			request: getRequest{
				ids: []string{"id-a"},
			},
			mockEndpoints: getInitializersGetEndpoints(
				"id=in.(id-a)",
				http.StatusOK,
				`[`,
			),
			expected: getExpected[types.Initializers]{
				err: "GET ${URL}/postgrest/dataplane_workspace_services?id=in.%28id-a%29 failed\n" +
					" Caused by:\n" +
					"\t* unexpected EOF",
			},
		},
		{
			name: "get initializers non-200 from server",
			request: getRequest{
				ids: []string{"id-a"},
			},
			mockEndpoints: getInitializersGetEndpoints(
				"id=in.(id-a)",
				http.StatusUnauthorized,
				`unauthorized`,
			),
			expected: getExpected[types.Initializers]{
				err: "GET ${URL}/postgrest/dataplane_workspace_services?id=in.%28id-a%29 failed\n" +
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

			actual, err := workbench.GetInitializers(context.Background(), svc, tt.request.ids, tt.request.sort)
			if tt.expected.err != "" {
				require.EqualError(t, err, strings.ReplaceAll(tt.expected.err, "${URL}", srv.URL))
				return
			}

			require.NoError(t, err)
			require.Equal(t, tt.expected.result, actual)
		})
	}
}

func TestPrintInitializers(t *testing.T) {
	tests := []struct {
		name               string
		resourcesJSON      string
		expectedTableCount int
		expectedTable      string
		expectedPrintCalls int
	}{
		{
			name:               "print initializers empty resources",
			resourcesJSON:      initializersEmptyResourcesJSON,
			expectedTableCount: 1,
			expectedTable:      expectedInitializersEmptyTable,
			expectedPrintCalls: 1,
		},
		{
			name:               "print initializers prints two resources",
			resourcesJSON:      initializersPrintTwoResourcesJSON,
			expectedTableCount: 1,
			expectedTable:      expectedInitializersTwoRowsTable,
			expectedPrintCalls: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var resources types.Initializers
			err := json.Unmarshal([]byte(tt.resourcesJSON), &resources)
			require.NoError(t, err)

			u := &stubs.StubUI{}
			err = workbench.PrintInitializers(u, "ignored-token", resources)
			require.NoError(t, err)

			require.Len(t, u.Tables, tt.expectedTableCount)
			tbl := u.Tables[0]
			require.Equal(t, tt.expectedTable, tbl.String())
			require.Equal(t, tt.expectedPrintCalls, tbl.PrintCalls)
		})
	}
}

func getInitializersGetEndpoints(
	exectedQuery string,
	responseStatus int,
	responseBody string,
) []stubs.StubHTTPEndpoint {
	return []stubs.StubHTTPEndpoint{
		{
			Request: stubs.StubHTTPRequest{
				Method:          http.MethodGet,
				URL:             postgrest.Endpoint(constants.WorkbenchEndpointDataplaneWorkspaceServices),
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
