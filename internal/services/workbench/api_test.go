// Copyright © 2026, SAS Institute Inc., Cary, NC, USA.  All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package workbench_test

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/go-openapi/spec"
	"github.com/sas-institute-rnd-saas/wbctl/internal/constants"
	"github.com/sas-institute-rnd-saas/wbctl/internal/postgrest"
	"github.com/sas-institute-rnd-saas/wbctl/internal/services/workbench"
	"github.com/sas-institute-rnd-saas/wbctl/internal/stubs"
	"github.com/stretchr/testify/require"
)

const (
	getAPISuccessResponseJSON = `
{
  "swagger": "2.0",
  "paths": {
    "/b": {
      "get": {
        "summary": "Get B"
      }
    },
    "/a": {
      "post": {
        "summary": "Create A"
      }
    }
  }
}
`

	nilPathsResourcesJSON = `{}`

	sortedPathsResourcesJSON = `
{
  "swagger": "2.0",
  "paths": {
    "/z": { "get": { "summary": "  Get Z  " } },
    "/a": { "get": { "summary": "  Read A  " }, "post": { "summary": " Create A " } },
    "/m": { "post": { "summary": "" } }
  }
}
`

	expectedSortedPathsTable = "PATH\tVERB\tDESCRIPTION\n" +
		"/a\tGET, POST\tRead A, Create A\n" +
		"/m\tPOST\t\n" +
		"/z\tGET\tGet Z"
)

func TestGetAPI(t *testing.T) {
	tests := []getTestCase[[]string]{
		{
			name:          "get api success",
			mockEndpoints: getAPIGetEndpoints(http.StatusOK, getAPISuccessResponseJSON),
			expected: getExpected[[]string]{
				result: []string{"/a", "/b"},
			},
		},
		{
			name:          "get api invalid json response",
			mockEndpoints: getAPIGetEndpoints(http.StatusOK, `{"swagger":"2.0","paths":{`),
			expected: getExpected[[]string]{
				err: "GET ${URL}/postgrest/ failed\n" +
					" Caused by:\n" +
					"\t* unexpected EOF",
			},
		},
		{
			name:          "get api non-200 from server",
			mockEndpoints: getAPIGetEndpoints(http.StatusUnauthorized, "unauthorized"),
			expected: getExpected[[]string]{
				err: "GET ${URL}/postgrest/ failed\n" +
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

			actual, err := workbench.GetAPI(context.Background(), svc)
			if tt.expected.err != "" {
				require.EqualError(t, err, strings.ReplaceAll(tt.expected.err, "${URL}", srv.URL))
				return
			}

			require.NoError(t, err)
			require.NotNil(t, actual)
			for _, path := range tt.expected.result {
				_, ok := actual.Paths.Paths[path]
				require.True(t, ok, "expected path %s in API paths", path)
			}
		})
	}
}

func TestPrintAPI(t *testing.T) {
	tests := []struct {
		name               string
		resourcesJSON      string
		expectedTableCount int
		expectedTable      string
		expectedPrintCalls int
	}{
		{
			name:               "print api nil paths does nothing",
			resourcesJSON:      nilPathsResourcesJSON,
			expectedTableCount: 0,
		},
		{
			name:               "print api prints sorted paths with verbs and descriptions",
			resourcesJSON:      sortedPathsResourcesJSON,
			expectedTableCount: 1,
			expectedTable:      expectedSortedPathsTable,
			expectedPrintCalls: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var swagger spec.Swagger
			err := json.Unmarshal([]byte(tt.resourcesJSON), &swagger)
			require.NoError(t, err)

			u := &stubs.StubUI{}
			err = workbench.PrintAPI(u, &swagger)
			require.NoError(t, err)

			require.Len(t, u.Tables, tt.expectedTableCount)
			if tt.expectedTableCount == 0 {
				return
			}

			tbl := u.Tables[0]
			require.Equal(t, tt.expectedTable, tbl.String())
			require.Equal(t, tt.expectedPrintCalls, tbl.PrintCalls)
		})
	}
}

// Helper endpoints.
func getAPIGetEndpoints(
	responseStatus int,
	responseBody string,
) []stubs.StubHTTPEndpoint {
	return []stubs.StubHTTPEndpoint{
		{
			Request: stubs.StubHTTPRequest{
				Method:          http.MethodGet,
				URL:             postgrest.Endpoint(constants.WorkbenchEndpointRoot),
				ExpectedHeaders: authHeaders,
			},
			Response: stubs.StubHTTPResponse{
				Status: responseStatus,
				Body:   responseBody,
			},
		},
	}
}
