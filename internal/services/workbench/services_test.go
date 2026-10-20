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
	servicesEmptyResourcesJSON = `[]`

	servicesGetOneResponseJSON = `
[
  {
    "id": "svc-1",
    "name": "code-server",
    "type": "workbench",
    "dataplane_id": "dp-1"
  }
]
`

	servicesGetTwoResponseJSON = `
[
  {
    "id": "id-a",
    "name": "svc-a",
    "type": "type-a",
    "dataplane_id": "dp-a"
  },
  {
    "id": "id-b",
    "name": "svc-b",
    "type": "type-b",
    "dataplane_id": "dp-b"
  }
]
`

	servicesPrintTwoResourcesJSON = `
[
  {
    "id": "svc-1",
    "name": "code-server",
    "type": "workbench"
  },
  {
    "id": "svc-2",
    "name": "jupyter",
    "type": "workbench"
  }
]
`

	expectedServicesEmptyTable   = "ID\tNAME\tTYPE"
	expectedServicesTwoRowsTable = "ID\tNAME\tTYPE\n" +
		"svc-1\tcode-server\tworkbench\n" +
		"svc-2\tjupyter\tworkbench"
)

var (
	expectedResponseOneService = types.Services{
		{ID: "svc-1", Name: "code-server", Type: "workbench", DataplaneID: "dp-1"},
	}

	expectedResponseTwoServices = types.Services{
		{ID: "id-a", Name: "svc-a", Type: "type-a", DataplaneID: "dp-a"},
		{ID: "id-b", Name: "svc-b", Type: "type-b", DataplaneID: "dp-b"},
	}
)

func TestGetServices(t *testing.T) {
	tests := []getTestCase[types.Services]{
		{
			name:    "get services success without ids query",
			request: getRequest{},
			mockEndpoints: getServicesGetEndpoints(
				"",
				http.StatusOK,
				servicesGetOneResponseJSON,
			),
			expected: getExpected[types.Services]{
				result: expectedResponseOneService,
			},
		},
		{
			name: "get services success with ids query",
			request: getRequest{
				ids:  []string{"id-a", "id-b"},
				sort: "id.asc",
			},
			mockEndpoints: getServicesGetEndpoints(
				"id=in.(id-a,id-b)&order=id.asc",
				http.StatusOK,
				servicesGetTwoResponseJSON,
			),
			expected: getExpected[types.Services]{
				result: expectedResponseTwoServices,
			},
		},
		{
			name: "get services sort by name",
			request: getRequest{
				sort: "name",
			},
			mockEndpoints: getServicesGetEndpoints(
				"order=name",
				http.StatusOK,
				servicesGetOneResponseJSON,
			),
			expected: getExpected[types.Services]{
				result: expectedResponseOneService,
			},
		},
		{
			name: "get services sort by type",
			request: getRequest{
				sort: "type",
			},
			mockEndpoints: getServicesGetEndpoints(
				"order=type",
				http.StatusOK,
				servicesGetOneResponseJSON,
			),
			expected: getExpected[types.Services]{
				result: expectedResponseOneService,
			},
		},
		{
			name: "get services sort by type and name",
			request: getRequest{
				sort: "type.desc,name.asc",
			},
			mockEndpoints: getServicesGetEndpoints(
				"order=type.desc,name.asc",
				http.StatusOK,
				servicesGetOneResponseJSON,
			),
			expected: getExpected[types.Services]{
				result: expectedResponseOneService,
			},
		},
		{
			name: "get services invalid json response",
			request: getRequest{
				ids: []string{"id-a"},
			},
			mockEndpoints: getServicesGetEndpoints(
				"id=in.(id-a)",
				http.StatusOK,
				`[`,
			),
			expected: getExpected[types.Services]{
				err: "GET ${URL}/postgrest/dataplane_workbench_services?id=in.%28id-a%29 failed\n" +
					" Caused by:\n" +
					"\t* unexpected EOF",
			},
		},
		{
			name: "get services non-200 from server",
			request: getRequest{
				ids: []string{"id-a"},
			},
			mockEndpoints: getServicesGetEndpoints(
				"id=in.(id-a)",
				http.StatusUnauthorized,
				`unauthorized`,
			),
			expected: getExpected[types.Services]{
				err: "GET ${URL}/postgrest/dataplane_workbench_services?id=in.%28id-a%29 failed\n" +
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

			actual, err := workbench.GetServices(context.Background(), svc, tt.request.ids, tt.request.sort)
			if tt.expected.err != "" {
				require.EqualError(t, err, strings.ReplaceAll(tt.expected.err, "${URL}", srv.URL))
				return
			}

			require.NoError(t, err)
			require.Equal(t, tt.expected.result, actual)
		})
	}
}

func TestGetAllServiceIds(t *testing.T) {
	tests := []getTestCase[[]string]{
		{
			name: "get service ids success",
			mockEndpoints: getServicesGetEndpoints(
				"",
				http.StatusOK,
				servicesGetTwoResponseJSON,
			),
			expected: getExpected[[]string]{
				result: []string{"id-a", "id-b"},
			},
		},
		{
			name: "get service ids invalid json response",
			mockEndpoints: getServicesGetEndpoints(
				"",
				http.StatusOK,
				`[`,
			),
			expected: getExpected[[]string]{
				err: "GET ${URL}/postgrest/dataplane_workbench_services failed\n" +
					" Caused by:\n" +
					"\t* unexpected EOF",
			},
		},
		{
			name: "get service ids non-200 from server",
			mockEndpoints: getServicesGetEndpoints(
				"",
				http.StatusUnauthorized,
				`unauthorized`,
			),
			expected: getExpected[[]string]{
				err: "GET ${URL}/postgrest/dataplane_workbench_services failed\n" +
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

			actual, err := workbench.GetServiceIDs(context.Background(), svc)
			if tt.expected.err != "" {
				require.EqualError(t, err, strings.ReplaceAll(tt.expected.err, "${URL}", srv.URL))
				return
			}

			require.NoError(t, err)
			require.Equal(t, tt.expected.result, actual)
		})
	}
}

func TestPrintServices(t *testing.T) {
	tests := []struct {
		name               string
		resourcesJSON      string
		expectedTableCount int
		expectedTable      string
		expectedPrintCalls int
	}{
		{
			name:               "get services empty resources",
			resourcesJSON:      servicesEmptyResourcesJSON,
			expectedTableCount: 1,
			expectedTable:      expectedServicesEmptyTable,
			expectedPrintCalls: 1,
		},
		{
			name:               "get services prints resources",
			resourcesJSON:      servicesPrintTwoResourcesJSON,
			expectedTableCount: 1,
			expectedTable:      expectedServicesTwoRowsTable,
			expectedPrintCalls: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var resources types.Services
			err := json.Unmarshal([]byte(tt.resourcesJSON), &resources)
			require.NoError(t, err)

			u := &stubs.StubUI{}
			err = workbench.PrintServices(u, "ignored-token", resources)
			require.NoError(t, err)

			require.Len(t, u.Tables, tt.expectedTableCount)
			tbl := u.Tables[0]
			require.Equal(t, tt.expectedTable, tbl.String())
			require.Equal(t, tt.expectedPrintCalls, tbl.PrintCalls)
		})
	}
}

func getServicesGetEndpoints(
	expectedQuery string,
	responseStatus int,
	responseBody string,
) []stubs.StubHTTPEndpoint {
	return []stubs.StubHTTPEndpoint{
		{
			Request: stubs.StubHTTPRequest{
				Method:          http.MethodGet,
				URL:             postgrest.Endpoint(constants.WorkbenchEndpointDataplaneWorkbenchServices),
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
