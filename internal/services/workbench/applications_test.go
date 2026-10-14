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
	applicationsGetOneResponseJSON = `[{
				"id":"app-1",
				"display_name":"RStudio",
				"description":"R app",
				"services":["svc-a"]
			}]`

	applicationsGetTwoResponseJSON = `[{
				"id":"id-a",
				"display_name":"App A",
				"description":"A",
				"services":["svc-a"]
			},{
				"id":"id-b",
				"display_name":"App B",
				"description":"B",
				"services":["svc-b"]
			}]`

	applicationsEmptyResourcesJSON = `[]`

	applicationsPrintTwoResourcesJSON = `[
				{
					"id":"app-1",
					"display_name":"RStudio",
					"description":"R app",
					"services":["svc-a","svc-b"]
				},
				{
					"id":"app-2",
					"display_name":"Jupyter",
					"description":"Python app",
					"services":[]
				}
			]`

	expectedApplicationsEmptyTable = "ID\tNAME\tDESCRIPTION\tSERVICES"

	expectedApplicationsTwoRowsTable = "ID\tNAME\tDESCRIPTION\tSERVICES\n" +
		"app-1\tRStudio\tR app\tsvc-a, svc-b\n" +
		"app-2\tJupyter\tPython app\t"
)

func TestGetApplications(t *testing.T) {
	tests := []getTestCase[types.Applications]{
		{
			name:          "get applications success without ids query",
			request:       getRequest{},
			mockEndpoints: getApplicationsGetEndpoints("", http.StatusOK, applicationsGetOneResponseJSON),
			expected: getExpected[types.Applications]{
				result: types.Applications{
					{ID: "app-1", DisplayName: "RStudio", Description: "R app", Services: []string{"svc-a"}},
				},
			},
		},
		{
			name: "get applications sort by name asc",
			request: getRequest{
				sort: "name.asc",
			},
			mockEndpoints: getApplicationsGetEndpoints("order=display_name.asc", http.StatusOK, applicationsGetOneResponseJSON),
			expected: getExpected[types.Applications]{
				result: types.Applications{
					{ID: "app-1", DisplayName: "RStudio", Description: "R app", Services: []string{"svc-a"}},
				},
			},
		},
		{
			name: "get applications sort by name desc",
			request: getRequest{
				sort: "name.desc",
			},
			mockEndpoints: getApplicationsGetEndpoints("order=display_name.desc", http.StatusOK, applicationsGetOneResponseJSON),
			expected: getExpected[types.Applications]{
				result: types.Applications{
					{ID: "app-1", DisplayName: "RStudio", Description: "R app", Services: []string{"svc-a"}},
				},
			},
		},
		{
			name: "get applications sort by id",
			request: getRequest{
				sort: "id",
			},
			mockEndpoints: getApplicationsGetEndpoints("order=id", http.StatusOK, applicationsGetOneResponseJSON),
			expected: getExpected[types.Applications]{
				result: types.Applications{
					{ID: "app-1", DisplayName: "RStudio", Description: "R app", Services: []string{"svc-a"}},
				},
			},
		},
		{
			name: "get applications sort by description",
			request: getRequest{
				sort: "description",
			},
			mockEndpoints: getApplicationsGetEndpoints("order=description", http.StatusOK, applicationsGetOneResponseJSON),
			expected: getExpected[types.Applications]{
				result: types.Applications{
					{ID: "app-1", DisplayName: "RStudio", Description: "R app", Services: []string{"svc-a"}},
				},
			},
		},
		{
			name: "get applications sort by services",
			request: getRequest{
				sort: "services",
			},
			mockEndpoints: getApplicationsGetEndpoints("order=services", http.StatusOK, applicationsGetOneResponseJSON),
			expected: getExpected[types.Applications]{
				result: types.Applications{
					{ID: "app-1", DisplayName: "RStudio", Description: "R app", Services: []string{"svc-a"}},
				},
			},
		},
		{
			name: "get applications sort by name and services",
			request: getRequest{
				sort: "name.asc,services.desc",
			},
			mockEndpoints: getApplicationsGetEndpoints(
				"order=display_name.asc,services.desc",
				http.StatusOK, applicationsGetOneResponseJSON,
			),
			expected: getExpected[types.Applications]{
				result: types.Applications{
					{ID: "app-1", DisplayName: "RStudio", Description: "R app", Services: []string{"svc-a"}},
				},
			},
		},
		{
			name: "get applications success with ids query",
			request: getRequest{
				ids: []string{"id-a", "id-b"},
			},
			mockEndpoints: getApplicationsGetEndpoints("id=in.(id-a,id-b)", http.StatusOK, applicationsGetTwoResponseJSON),
			expected: getExpected[types.Applications]{
				result: types.Applications{
					{ID: "id-a", DisplayName: "App A", Description: "A", Services: []string{"svc-a"}},
					{ID: "id-b", DisplayName: "App B", Description: "B", Services: []string{"svc-b"}},
				},
			},
		},
		{
			name: "get applications invalid json response",
			request: getRequest{
				ids: []string{"id-a"},
			},
			mockEndpoints: getApplicationsGetEndpoints("id=in.(id-a)", http.StatusOK, "["),
			expected: getExpected[types.Applications]{
				err: "GET ${URL}/postgrest/dataplane_workbench_applications?id=in.%28id-a%29 failed\n" +
					" Caused by:\n" +
					"\t* unexpected EOF",
			},
		},
		{
			name: "get applications non-200 from server",
			request: getRequest{
				ids: []string{"id-a"},
			},
			mockEndpoints: getApplicationsGetEndpoints("id=in.(id-a)", http.StatusUnauthorized, "unauthorized"),
			expected: getExpected[types.Applications]{
				err: "GET ${URL}/postgrest/dataplane_workbench_applications?id=in.%28id-a%29 failed\n" +
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

			actual, err := workbench.GetApplications(context.Background(), svc, tt.request.ids, tt.request.sort)
			if tt.expected.err != "" {
				require.EqualError(t, err, strings.ReplaceAll(tt.expected.err, "${URL}", srv.URL))
				return
			}

			require.NoError(t, err)
			require.Equal(t, tt.expected.result, actual)
		})
	}
}

func TestPrintApplications(t *testing.T) {
	tests := []struct {
		name               string
		resourcesJSON      string
		expectedTableCount int
		expectedTable      string
		expectedPrintCalls int
	}{
		{
			name:               "empty resources",
			resourcesJSON:      applicationsEmptyResourcesJSON,
			expectedTableCount: 1,
			expectedTable:      expectedApplicationsEmptyTable,
			expectedPrintCalls: 1,
		},
		{
			name:               "prints resources and joins services",
			resourcesJSON:      applicationsPrintTwoResourcesJSON,
			expectedTableCount: 1,
			expectedTable:      expectedApplicationsTwoRowsTable,
			expectedPrintCalls: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var resources types.Applications
			err := json.Unmarshal([]byte(tt.resourcesJSON), &resources)
			require.NoError(t, err)

			u := &stubs.StubUI{}
			err = workbench.PrintApplications(u, "ignored-token", resources)
			require.NoError(t, err)

			require.Len(t, u.Tables, tt.expectedTableCount)
			tbl := u.Tables[0]
			require.Equal(t, tt.expectedTable, tbl.String())
			require.Equal(t, tt.expectedPrintCalls, tbl.PrintCalls)
		})
	}
}

func getApplicationsGetEndpoints(
	exectedQuery string,
	responseStatus int,
	responseBody string,
) []stubs.StubHTTPEndpoint {
	return []stubs.StubHTTPEndpoint{
		{
			Request: stubs.StubHTTPRequest{
				Method:          http.MethodGet,
				URL:             postgrest.Endpoint(constants.WorkbenchEndpointDataplaneWorkbenchApplications),
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
