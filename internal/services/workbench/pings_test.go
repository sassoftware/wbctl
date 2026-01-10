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
	pingsEmptyResourcesJSON = `[]`

	pingsGetOneResponseJSON = `
[
  {
    "dataplane_id": "dp-1",
    "request_at": "2026-05-08T00:00:00Z",
    "status_updated_at": "2026-05-08T00:00:02Z",
    "status_state": "ok",
    "status_message": "pong"
  }
]
`

	pingsGetTwoResponseJSON = `
[
  {
    "dataplane_id": "dp-a",
    "request_at": "2026-05-08T00:00:00Z",
    "status_updated_at": "2026-05-08T00:00:01Z",
    "status_state": "ok",
    "status_message": "fast"
  },
  {
    "dataplane_id": "dp-b",
    "request_at": "2026-05-08T00:00:00Z",
    "status_updated_at": "2026-05-08T00:00:03Z",
    "status_state": "warn",
    "status_message": "slow"
  }
]
`

	pingsPrintTwoResourcesJSON = `
[
  {
    "dataplane_id": "dp-a",
    "request_at": "2026-05-08T00:00:00Z",
    "status_updated_at": "2026-05-08T00:00:01Z",
    "status_state": "ok",
    "status_message": "fast"
  },
  {
    "dataplane_id": "dp-b",
    "request_at": "2026-05-08T00:00:00Z",
    "status_updated_at": "2026-05-08T00:00:03Z",
    "status_state": "warn",
    "status_message": "slow"
  }
]
`

	expectedPingsEmptyTable = "DATAPLANE_ID\tREQUEST_AT\tRESPONSE_AT\tRESPONSE_TIME\tSTATE\tMESSAGE\tAGE"

	expectedPingsTwoRowsTable = "DATAPLANE_ID\tREQUEST_AT\tRESPONSE_AT\tRESPONSE_TIME\tSTATE\tMESSAGE\tAGE\n" +
		"dp-a\t2026-05-08T00:00:00Z\t2026-05-08T00:00:01Z\t1s\tok\tfast\t0s\n" +
		"dp-b\t2026-05-08T00:00:00Z\t2026-05-08T00:00:03Z\t3s\twarn\tslow\t0s"
)

var (
	expectedResponseOnePing = types.Pings{
		{
			DataplaneID: "dp-1",
			RequestAt:   mustTimePtr("2026-05-08T00:00:00Z"),
			ResponseAt:  mustTimePtr("2026-05-08T00:00:02Z"),
			State:       "ok",
			Message:     "pong",
		},
	}

	expectedResponseTwoPings = types.Pings{
		{
			DataplaneID: "dp-a",
			RequestAt:   mustTimePtr("2026-05-08T00:00:00Z"),
			ResponseAt:  mustTimePtr("2026-05-08T00:00:01Z"),
			State:       "ok",
			Message:     "fast",
		},
		{
			DataplaneID: "dp-b",
			RequestAt:   mustTimePtr("2026-05-08T00:00:00Z"),
			ResponseAt:  mustTimePtr("2026-05-08T00:00:03Z"),
			State:       "warn",
			Message:     "slow",
		},
	}
)

func TestGetPings(t *testing.T) {
	tests := []getTestCase[types.Pings]{
		{
			name:    "get pings success without ids query",
			request: getRequest{},
			mockEndpoints: getPingsGetEndpoints(
				"",
				http.StatusOK,
				pingsGetOneResponseJSON,
			),
			expected: getExpected[types.Pings]{
				result: expectedResponseOnePing,
			},
		},
		{
			name: "get pings success with dataplane_id query",
			request: getRequest{
				ids:  []string{"id-a", "id-b"},
				sort: "dataplane_id.asc",
			},
			mockEndpoints: getPingsGetEndpoints(
				"id=in.(id-a,id-b)&order=dataplane_id.asc",
				http.StatusOK,
				pingsGetTwoResponseJSON,
			),
			expected: getExpected[types.Pings]{
				result: expectedResponseTwoPings,
			},
		},
		{
			name: "get pings sort by request_at",
			request: getRequest{
				sort: "request_at",
			},
			mockEndpoints: getPingsGetEndpoints(
				"order=request_at",
				http.StatusOK,
				pingsGetOneResponseJSON,
			),
			expected: getExpected[types.Pings]{
				result: expectedResponseOnePing,
			},
		},
		{
			name: "get pings sort by response_at",
			request: getRequest{
				sort: "response_at",
			},
			mockEndpoints: getPingsGetEndpoints(
				"order=status_updated_at",
				http.StatusOK,
				pingsGetOneResponseJSON,
			),
			expected: getExpected[types.Pings]{
				result: expectedResponseOnePing,
			},
		},
		{
			name: "get pings sort by response_time",
			request: getRequest{
				sort: "response_time",
			},
			mockEndpoints: getPingsGetEndpoints(
				"order=response_time",
				http.StatusOK,
				pingsGetOneResponseJSON,
			),
			expected: getExpected[types.Pings]{
				result: expectedResponseOnePing,
			},
		},
		{
			name: "get pings sort by state",
			request: getRequest{
				sort: "state",
			},
			mockEndpoints: getPingsGetEndpoints(
				"order=status_state",
				http.StatusOK,
				pingsGetOneResponseJSON,
			),
			expected: getExpected[types.Pings]{
				result: expectedResponseOnePing,
			},
		},
		{
			name: "get pings sort by message",
			request: getRequest{
				sort: "message",
			},
			mockEndpoints: getPingsGetEndpoints(
				"order=status_message",
				http.StatusOK,
				pingsGetOneResponseJSON,
			),
			expected: getExpected[types.Pings]{
				result: expectedResponseOnePing,
			},
		},
		{
			name: "get pings sort by age",
			request: getRequest{
				sort: "age",
			},
			mockEndpoints: getPingsGetEndpoints(
				"order=request_at",
				http.StatusOK,
				pingsGetOneResponseJSON,
			),
			expected: getExpected[types.Pings]{
				result: expectedResponseOnePing,
			},
		},
		{
			name: "get pings sort by age, response, state, and message",
			request: getRequest{
				sort: "age.desc,response_at.asc,state,message.desc",
			},
			mockEndpoints: getPingsGetEndpoints(
				"order=request_at.desc,status_updated_at.asc,status_state,status_message.desc",
				http.StatusOK,
				pingsGetOneResponseJSON,
			),
			expected: getExpected[types.Pings]{
				result: expectedResponseOnePing,
			},
		},
		{
			name: "get pings invalid json response",
			request: getRequest{
				ids: []string{"id-a"},
			},
			mockEndpoints: getPingsGetEndpoints(
				"id=in.(id-a)",
				http.StatusOK,
				`[`,
			),
			expected: getExpected[types.Pings]{
				err: "GET ${URL}/postgrest/pings?id=in.%28id-a%29 failed\n" +
					" Caused by:\n" +
					"\t* unexpected EOF",
			},
		},
		{
			name: "get pings non-200 from server",
			request: getRequest{
				ids: []string{"id-a"},
			},
			mockEndpoints: getPingsGetEndpoints(
				"id=in.(id-a)",
				http.StatusUnauthorized,
				`unauthorized`,
			),
			expected: getExpected[types.Pings]{
				err: "GET ${URL}/postgrest/pings?id=in.%28id-a%29 failed\n" +
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

			actual, err := workbench.GetPings(context.Background(), svc, tt.request.ids, tt.request.sort)
			if tt.expected.err != "" {
				require.EqualError(t, err, strings.ReplaceAll(tt.expected.err, "${URL}", srv.URL))
				return
			}

			require.NoError(t, err)
			require.Equal(t, tt.expected.result, actual)
		})
	}
}

func TestPrintPings(t *testing.T) {
	fixedNow := time.Date(2026, time.May, 8, 0, 0, 0, 0, time.UTC)
	timeutil.SetNow(func() time.Time { return fixedNow })
	t.Cleanup(func() { timeutil.SetNow(time.Now) })

	tests := []struct {
		name               string
		resourcesJSON      string
		expectedTableCount int
		expectedTable      string
		expectedPrintCalls int
	}{
		{
			name:               "print pings empty resources",
			resourcesJSON:      pingsEmptyResourcesJSON,
			expectedTableCount: 1,
			expectedTable:      expectedPingsEmptyTable,
			expectedPrintCalls: 1,
		},
		{
			name:               "print pings prints resources",
			resourcesJSON:      pingsPrintTwoResourcesJSON,
			expectedTableCount: 1,
			expectedTable:      expectedPingsTwoRowsTable,
			expectedPrintCalls: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var resources types.Pings
			err := json.Unmarshal([]byte(tt.resourcesJSON), &resources)
			require.NoError(t, err)

			u := &stubs.StubUI{}
			err = workbench.PrintPings(u, "ignored-token", resources)
			require.NoError(t, err)

			require.Len(t, u.Tables, tt.expectedTableCount)
			tbl := u.Tables[0]
			require.Equal(t, tt.expectedTable, tbl.String())
			require.Equal(t, tt.expectedPrintCalls, tbl.PrintCalls)
		})
	}
}

// Helper endpoints.
func getPingsGetEndpoints(
	expectedQuery string,
	responseStatus int,
	responseBody string,
) []stubs.StubHTTPEndpoint {
	return []stubs.StubHTTPEndpoint{
		{
			Request: stubs.StubHTTPRequest{
				Method:          http.MethodGet,
				URL:             postgrest.Endpoint(constants.WorkbenchEndpointPings),
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

func mustTimePtr(value string) *time.Time {
	t, err := time.Parse(time.RFC3339, value)
	if err != nil {
		panic(err)
	}
	return &t
}
