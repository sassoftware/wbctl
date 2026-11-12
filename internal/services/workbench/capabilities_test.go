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
	capabilitiesGetOneResponseJSON = `[{
				"id":"cap-1",
				"name":"Capability1",
				"description":"The first capability",
                "category":"language",
				"disabled":false,
                "version":"1.0.1"
			}]`

	capabilitiesGetTwoResponseJSON = `[{
				"id":"cap-1",
				"name":"Capability1",
				"description":"The first capability",
                "category":"language",
				"disabled":false,
                "version":"1.0.1"
			},{
				"id":"cap-2",
				"name":"Capability2",
				"description":"The second capability",
                "category":"language",
				"disabled":true,
                "version":"1.0.2"
			}]`

	capabilitiesEmptyResourcesJSON = `[]`

	capabilitiesPrintTwoResourcesJSON = `[
		  {
			"id": "cap-1",
			"name": "Capability1",
			"description": "The first capability",
			"category": "language",
			"disabled": false,
			"version": "1.0.1"
		  },
		  {
			"id": "cap-2",
			"name": "Capability2",
			"description": "The second capability",
			"category": "language",
			"disabled": true,
			"version": "1.0.2"
		  }
		]`

	expectedCapabilitiesEmptyTable = "ID\tNAME\tDESCRIPTION\tVERSION\tCATEGORY\tENABLED"

	expectedCapabilitiesTwoRowsTable = "ID\tNAME\tDESCRIPTION\tVERSION\tCATEGORY\tENABLED\n" +
		"cap-1\tCapability1\tThe first capability\t1.0.1\tlanguage\ttrue\n" +
		"cap-2\tCapability2\tThe second capability\t1.0.2\tlanguage\tfalse"
)

func TestGetCapabilities(t *testing.T) {
	tests := []getTestCase[types.Capabilities]{
		{
			name:          "get capabilities success without ids query",
			request:       getRequest{},
			mockEndpoints: getCapabilitiesGetEndpoints("", http.StatusOK, capabilitiesGetOneResponseJSON),
			expected: getExpected[types.Capabilities]{
				result: types.Capabilities{
					{
						ID:          "cap-1",
						Name:        "Capability1",
						Description: "The first capability",
						Version:     "1.0.1",
						Category:    "language",
						Disabled:    false,
					},
				},
			},
		},
		{
			name: "get capabilities sort by name asc",
			request: getRequest{
				sort: "name.asc",
			},
			mockEndpoints: getCapabilitiesGetEndpoints("order=name.asc", http.StatusOK, capabilitiesGetOneResponseJSON),
			expected: getExpected[types.Capabilities]{
				result: types.Capabilities{
					{
						ID:          "cap-1",
						Name:        "Capability1",
						Description: "The first capability",
						Version:     "1.0.1",
						Category:    "language",
						Disabled:    false,
					},
				},
			},
		},
		{
			name: "get capabilities sort by name desc",
			request: getRequest{
				sort: "name.desc",
			},
			mockEndpoints: getCapabilitiesGetEndpoints("order=name.desc", http.StatusOK, capabilitiesGetOneResponseJSON),
			expected: getExpected[types.Capabilities]{
				result: types.Capabilities{
					{
						ID:          "cap-1",
						Name:        "Capability1",
						Description: "The first capability",
						Version:     "1.0.1",
						Category:    "language",
						Disabled:    false,
					},
				},
			},
		},
		{
			name: "get capabilities sort by id",
			request: getRequest{
				sort: "id",
			},
			mockEndpoints: getCapabilitiesGetEndpoints("order=id", http.StatusOK, capabilitiesGetOneResponseJSON),
			expected: getExpected[types.Capabilities]{
				result: types.Capabilities{
					{
						ID:          "cap-1",
						Name:        "Capability1",
						Description: "The first capability",
						Version:     "1.0.1",
						Category:    "language",
						Disabled:    false,
					},
				},
			},
		},
		{
			name: "get capabilities sort by description",
			request: getRequest{
				sort: "description",
			},
			mockEndpoints: getCapabilitiesGetEndpoints("order=description", http.StatusOK, capabilitiesGetOneResponseJSON),
			expected: getExpected[types.Capabilities]{
				result: types.Capabilities{
					{
						ID:          "cap-1",
						Name:        "Capability1",
						Description: "The first capability",
						Version:     "1.0.1",
						Category:    "language",
						Disabled:    false,
					},
				},
			},
		},
		{
			name: "get capabilities sort by category",
			request: getRequest{
				sort: "category",
			},
			mockEndpoints: getCapabilitiesGetEndpoints("order=category", http.StatusOK, capabilitiesGetOneResponseJSON),
			expected: getExpected[types.Capabilities]{
				result: types.Capabilities{
					{
						ID:          "cap-1",
						Name:        "Capability1",
						Description: "The first capability",
						Version:     "1.0.1",
						Category:    "language",
						Disabled:    false,
					},
				},
			},
		},
		{
			name: "get capabilities sort by version",
			request: getRequest{
				sort: "version",
			},
			mockEndpoints: getCapabilitiesGetEndpoints("order=version", http.StatusOK, capabilitiesGetOneResponseJSON),
			expected: getExpected[types.Capabilities]{
				result: types.Capabilities{
					{
						ID:          "cap-1",
						Name:        "Capability1",
						Description: "The first capability",
						Version:     "1.0.1",
						Category:    "language",
						Disabled:    false,
					},
				},
			},
		},
		{
			name: "get capabilities sort by enabled",
			request: getRequest{
				sort: "enabled",
			},
			mockEndpoints: getCapabilitiesGetEndpoints("order=disabled", http.StatusOK, capabilitiesGetOneResponseJSON),
			expected: getExpected[types.Capabilities]{
				result: types.Capabilities{
					{
						ID:          "cap-1",
						Name:        "Capability1",
						Description: "The first capability",
						Version:     "1.0.1",
						Category:    "language",
						Disabled:    false,
					},
				},
			},
		},
		{
			name: "get capabilities sort by name and category",
			request: getRequest{
				sort: "name.asc,category.desc",
			},
			mockEndpoints: getCapabilitiesGetEndpoints(
				"order=name.asc,category.desc",
				http.StatusOK, capabilitiesGetOneResponseJSON,
			),
			expected: getExpected[types.Capabilities]{
				result: types.Capabilities{
					{
						ID:          "cap-1",
						Name:        "Capability1",
						Description: "The first capability",
						Version:     "1.0.1",
						Category:    "language",
						Disabled:    false,
					},
				},
			},
		},
		{
			name: "get capabilities success with ids query",
			request: getRequest{
				ids: []string{"cap-1", "cap-2"},
			},
			mockEndpoints: getCapabilitiesGetEndpoints("id=in.(cap-1,cap-2)", http.StatusOK, capabilitiesGetTwoResponseJSON),
			expected: getExpected[types.Capabilities]{
				result: types.Capabilities{
					{
						ID:          "cap-1",
						Name:        "Capability1",
						Description: "The first capability",
						Version:     "1.0.1",
						Category:    "language",
						Disabled:    false,
					},
					{
						ID:          "cap-2",
						Name:        "Capability2",
						Description: "The second capability",
						Version:     "1.0.2",
						Category:    "language",
						Disabled:    true,
					},
				},
			},
		},
		{
			name: "get capabilities invalid json response",
			request: getRequest{
				ids: []string{"cap-1"},
			},
			mockEndpoints: getCapabilitiesGetEndpoints("id=in.(cap-1)", http.StatusOK, "["),
			expected: getExpected[types.Capabilities]{
				err: "GET ${URL}/postgrest/dataplane_workbench_capabilities?id=in.%28cap-1%29 failed\n" +
					" Caused by:\n" +
					"\t* unexpected EOF",
			},
		},
		{
			name: "get capabilities non-200 from server",
			request: getRequest{
				ids: []string{"cap-1"},
			},
			mockEndpoints: getCapabilitiesGetEndpoints("id=in.(cap-1)", http.StatusUnauthorized, "unauthorized"),
			expected: getExpected[types.Capabilities]{
				err: "GET ${URL}/postgrest/dataplane_workbench_capabilities?id=in.%28cap-1%29 failed\n" +
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

			actual, err := workbench.GetCapabilities(context.Background(), svc, tt.request.ids, tt.request.sort)
			if tt.expected.err != "" {
				require.EqualError(t, err, strings.ReplaceAll(tt.expected.err, "${URL}", srv.URL))
				return
			}

			require.NoError(t, err)
			require.Equal(t, tt.expected.result, actual)
		})
	}
}

func TestPrintCapabilities(t *testing.T) {
	tests := []struct {
		name               string
		resourcesJSON      string
		expectedTableCount int
		expectedTable      string
		expectedPrintCalls int
	}{
		{
			name:               "empty resources",
			resourcesJSON:      capabilitiesEmptyResourcesJSON,
			expectedTableCount: 1,
			expectedTable:      expectedCapabilitiesEmptyTable,
			expectedPrintCalls: 1,
		},
		{
			name:               "prints resources",
			resourcesJSON:      capabilitiesPrintTwoResourcesJSON,
			expectedTableCount: 1,
			expectedTable:      expectedCapabilitiesTwoRowsTable,
			expectedPrintCalls: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var resources types.Capabilities
			err := json.Unmarshal([]byte(tt.resourcesJSON), &resources)
			require.NoError(t, err)

			u := &stubs.StubUI{}
			err = workbench.PrintCapabilities(u, "ignored-token", resources)
			require.NoError(t, err)

			require.Len(t, u.Tables, tt.expectedTableCount)
			tbl := u.Tables[0]
			require.Equal(t, tt.expectedTable, tbl.String())
			require.Equal(t, tt.expectedPrintCalls, tbl.PrintCalls)
		})
	}
}

func getCapabilitiesGetEndpoints(
	exectedQuery string,
	responseStatus int,
	responseBody string,
) []stubs.StubHTTPEndpoint {
	return []stubs.StubHTTPEndpoint{
		{
			Request: stubs.StubHTTPRequest{
				Method:          http.MethodGet,
				URL:             postgrest.Endpoint(constants.WorkbenchEndpointDataplaneWorkbenchCapabilities),
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
