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
	storageClassesEmptyResourcesJSON = `[]`

	storageClassesGetOneResponseJSON = `
[
  {
    "id": "sc-1",
    "display_name": "standard",
    "description": "default class",
    "access_mode": "ReadWriteOnce",
    "dataplane_id": "dp-1",
    "properties": {
      "default": "true"
    }
  }
]
`

	storageClassesGetTwoResponseJSON = `
[
  {
    "id": "id-a",
    "display_name": "fast",
    "description": "ssd",
    "access_mode": "ReadWriteOnce",
    "dataplane_id": "dp-a",
    "properties": {
      "default": "false"
    }
  },
  {
    "id": "id-b",
    "display_name": "bulk",
    "description": "hdd",
    "access_mode": "ReadWriteMany",
    "dataplane_id": "dp-b",
    "properties": {
      "default": "true"
    }
  }
]
`

	storageClassesPrintTwoResourcesJSON = `
[
  {
    "id": "sc-1",
    "display_name": "standard",
    "description": "default class",
    "properties": {
      "default": "true"
    }
  },
  {
    "id": "sc-2",
    "display_name": "fast",
    "description": "ssd",
    "properties": {}
  }
]
`

	expectedStorageClassesEmptyTable   = "ID\tNAME\tDESCRIPTION\tDEFAULT"
	expectedStorageClassesTwoRowsTable = "ID\tNAME\tDESCRIPTION\tDEFAULT\n" +
		"sc-1\tstandard\tdefault class\ttrue\n" +
		"sc-2\tfast\tssd\t"
)

var (
	expectedResponseOneStorageclass = types.StorageClasses{
		{
			ID:          "sc-1",
			DisplayName: "standard",
			Description: "default class",
			AccessMode:  "ReadWriteOnce",
			DataplaneID: "dp-1",
			Properties:  map[string]interface{}{"default": "true"},
		},
	}

	expectedResponseTwoStorageclasses = types.StorageClasses{
		{
			ID:          "id-a",
			DisplayName: "fast",
			Description: "ssd",
			AccessMode:  "ReadWriteOnce",
			DataplaneID: "dp-a",
			Properties:  map[string]interface{}{"default": "false"},
		},
		{
			ID:          "id-b",
			DisplayName: "bulk",
			Description: "hdd",
			AccessMode:  "ReadWriteMany",
			DataplaneID: "dp-b",
			Properties:  map[string]interface{}{"default": "true"},
		},
	}
)

func TestGetStorageClasses(t *testing.T) {
	tests := []getTestCase[types.StorageClasses]{
		{
			name:    "get storageclasses success without ids query",
			request: getRequest{},
			mockEndpoints: getStorageClassesGetEndpoints(
				"",
				http.StatusOK,
				storageClassesGetOneResponseJSON,
			),
			expected: getExpected[types.StorageClasses]{
				result: expectedResponseOneStorageclass,
			},
		},
		{
			name: "get storageclasses sort by name asc",
			request: getRequest{
				ids:  nil,
				sort: "display_name.asc",
			},
			mockEndpoints: getStorageClassesGetEndpoints(
				"order=display_name.asc",
				http.StatusOK,
				storageClassesGetOneResponseJSON,
			),
			expected: getExpected[types.StorageClasses]{
				result: expectedResponseOneStorageclass,
			},
		},
		{
			name: "get storageclasses sort by name desc",
			request: getRequest{
				ids:  nil,
				sort: "display_name.desc",
			},
			mockEndpoints: getStorageClassesGetEndpoints(
				"order=display_name.desc",
				http.StatusOK,
				storageClassesGetOneResponseJSON,
			),
			expected: getExpected[types.StorageClasses]{
				result: expectedResponseOneStorageclass,
			},
		},
		{
			name: "get storageclasses success with ids query",
			request: getRequest{
				ids: []string{"id-a", "id-b"},
			},
			mockEndpoints: getStorageClassesGetEndpoints(
				"id=in.(id-a,id-b)",
				http.StatusOK,
				storageClassesGetTwoResponseJSON,
			),
			expected: getExpected[types.StorageClasses]{
				result: expectedResponseTwoStorageclasses,
			},
		},
		{
			name: "get storageclasses sort by name",
			request: getRequest{
				sort: "name",
			},
			mockEndpoints: getStorageClassesGetEndpoints(
				"order=display_name",
				http.StatusOK,
				storageClassesGetOneResponseJSON,
			),
			expected: getExpected[types.StorageClasses]{
				result: expectedResponseOneStorageclass,
			},
		},
		{
			name: "get storageclasses sort by description",
			request: getRequest{
				sort: "description",
			},
			mockEndpoints: getStorageClassesGetEndpoints(
				"order=description",
				http.StatusOK,
				storageClassesGetOneResponseJSON,
			),
			expected: getExpected[types.StorageClasses]{
				result: expectedResponseOneStorageclass,
			},
		},
		{
			name: "get storageclasses sort by default",
			request: getRequest{
				sort: "default",
			},
			mockEndpoints: getStorageClassesGetEndpoints(
				"order=properties->default",
				http.StatusOK,
				storageClassesGetOneResponseJSON,
			),
			expected: getExpected[types.StorageClasses]{
				result: expectedResponseOneStorageclass,
			},
		},
		{
			name: "get storageclasses sort by name and default",
			request: getRequest{
				sort: "name.asc,default.desc",
			},
			mockEndpoints: getStorageClassesGetEndpoints(
				"order=display_name.asc,properties->default.desc",
				http.StatusOK,
				storageClassesGetOneResponseJSON,
			),
			expected: getExpected[types.StorageClasses]{
				result: expectedResponseOneStorageclass,
			},
		},
		{
			name: "get storageclasses invalid json response",
			request: getRequest{
				ids: []string{"id-a"},
			},
			mockEndpoints: getStorageClassesGetEndpoints(
				"id=in.(id-a)",
				http.StatusOK,
				`[`,
			),
			expected: getExpected[types.StorageClasses]{
				err: "GET ${URL}/postgrest/dataplane_storageclasses?id=in.%28id-a%29 failed\n" +
					" Caused by:\n" +
					"\t* unexpected EOF",
			},
		},
		{
			name: "get storageclasses non-200 from server",
			request: getRequest{
				ids: []string{"id-a"},
			},
			mockEndpoints: getStorageClassesGetEndpoints(
				"id=in.(id-a)",
				http.StatusUnauthorized,
				`unauthorized`,
			),
			expected: getExpected[types.StorageClasses]{
				err: "GET ${URL}/postgrest/dataplane_storageclasses?id=in.%28id-a%29 failed\n" +
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

			actual, err := workbench.GetStorageClasses(context.Background(), svc, tt.request.ids, tt.request.sort)
			if tt.expected.err != "" {
				require.EqualError(t, err, strings.ReplaceAll(tt.expected.err, "${URL}", srv.URL))
				return
			}

			require.NoError(t, err)
			require.Equal(t, tt.expected.result, actual)
		})
	}
}

func TestPrintStorageClasses(t *testing.T) {
	tests := []struct {
		name               string
		resourcesJSON      string
		expectedTableCount int
		expectedTable      string
		expectedPrintCalls int
	}{
		{
			name:               "print storageclasses empty resources",
			resourcesJSON:      storageClassesEmptyResourcesJSON,
			expectedTableCount: 1,
			expectedTable:      expectedStorageClassesEmptyTable,
			expectedPrintCalls: 1,
		},
		{
			name:               "print storageclasses prints resources",
			resourcesJSON:      storageClassesPrintTwoResourcesJSON,
			expectedTableCount: 1,
			expectedTable:      expectedStorageClassesTwoRowsTable,
			expectedPrintCalls: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var resources types.StorageClasses
			err := json.Unmarshal([]byte(tt.resourcesJSON), &resources)
			require.NoError(t, err)

			u := &stubs.StubUI{}
			err = workbench.PrintStorageClasses(u, "ignored-token", resources)
			require.NoError(t, err)

			require.Len(t, u.Tables, tt.expectedTableCount)
			tbl := u.Tables[0]
			require.Equal(t, tt.expectedTable, tbl.String())
			require.Equal(t, tt.expectedPrintCalls, tbl.PrintCalls)
		})
	}
}

func getStorageClassesGetEndpoints(
	expectedQuery string,
	responseStatus int,
	responseBody string,
) []stubs.StubHTTPEndpoint {
	return []stubs.StubHTTPEndpoint{
		{
			Request: stubs.StubHTTPRequest{
				Method:          http.MethodGet,
				URL:             postgrest.Endpoint(constants.WorkbenchEndpointDataplaneStorageclasses),
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
