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
	eventsEmptyResourcesJSON = `[]`

	eventsGetOneResponseJSON = `
[
  {
    "id": "evt-1",
    "type": "INFO",
    "resource_type": "workbench",
    "kind": "audit",
    "action": "create",
    "note": "created",
    "reason": "manual"
  }
]
`

	eventsGetTwoResponseJSON = `
[
  {
    "id": "id-a",
    "type": "INFO",
    "resource_type": "workspace",
    "kind": "audit",
    "action": "update",
    "note": "one",
    "reason": "r1"
  },
  {
    "id": "id-b",
    "type": "WARN",
    "resource_type": "workspace",
    "kind": "audit",
    "action": "delete",
    "note": "two",
    "reason": "r2"
  }
]
`

	eventsPrintOneResourcesJSON = `
[
  {
    "id": "evt-1",
    "type": "INFO",
    "resource_type": "workbench",
    "kind": "audit",
    "action": "create",
    "note": "line1\nline2",
    "reason": "manual",
    "created_at": null
  }
]
`

	expectedEventsEmptyTable = "ID\tTYPE\tRESOURCE_TYPE\tKIND\tACTION\tNOTE\tREASON\tAGE"

	expectedEventsOneRowTable = "ID\tTYPE\tRESOURCE_TYPE\tKIND\tACTION\tNOTE\tREASON\tAGE\n" +
		"evt-1\tINFO\tworkbench\taudit\tcreate\tline1 line2\tmanual\t"

	expectedEventsTwoRowTable = "ID\tTYPE\tRESOURCE_TYPE\tKIND\tACTION\tNOTE\tREASON\tAGE\n" +
		"id-a\tINFO\tworkspace\taudit\tupdate\tone\tr1\t\n" +
		"id-b\tWARN\tworkspace\taudit\tdelete\ttwo\tr2\t"
)

var (
	expectedResponseOneEvent = types.Events{
		{
			ID:           "evt-1",
			Type:         "INFO",
			ResourceType: "workbench",
			Kind:         "audit",
			Action:       "create",
			Note:         "created",
			Reason:       "manual",
		},
	}

	expectedResponseTwoEvents = types.Events{
		{
			ID:           "id-a",
			Type:         "INFO",
			ResourceType: "workspace",
			Kind:         "audit",
			Action:       "update",
			Note:         "one",
			Reason:       "r1",
		},
		{
			ID:           "id-b",
			Type:         "WARN",
			ResourceType: "workspace",
			Kind:         "audit",
			Action:       "delete",
			Note:         "two",
			Reason:       "r2",
		},
	}
)

func TestGetEvents(t *testing.T) {
	tests := []getTestCase[types.Events]{
		{
			name: "get events success without ids query",
			request: getRequest{
				ids: nil,
			},
			mockEndpoints: getEventsGetEndpoints("", http.StatusOK, eventsGetOneResponseJSON),
			expected: getExpected[types.Events]{
				result: expectedResponseOneEvent,
			},
		},
		{
			name: "get events success with ids query",
			request: getRequest{
				ids:  []string{"id-a", "id-b"},
				sort: "id.asc",
			},
			mockEndpoints: getEventsGetEndpoints("id=in.(id-a,id-b)&order=id.asc", http.StatusOK, eventsGetTwoResponseJSON),
			expected: getExpected[types.Events]{
				result: expectedResponseTwoEvents,
			},
		},
		{
			name: "get events sort by type",
			request: getRequest{
				sort: "type",
			},
			mockEndpoints: getEventsGetEndpoints("order=type", http.StatusOK, eventsGetOneResponseJSON),
			expected: getExpected[types.Events]{
				result: expectedResponseOneEvent,
			},
		},
		{
			name: "get events sort by type",
			request: getRequest{
				sort: "resource_type",
			},
			mockEndpoints: getEventsGetEndpoints("order=resource_type", http.StatusOK, eventsGetOneResponseJSON),
			expected: getExpected[types.Events]{
				result: expectedResponseOneEvent,
			},
		},
		{
			name: "get events sort by kind",
			request: getRequest{
				sort: "kind",
			},
			mockEndpoints: getEventsGetEndpoints("order=kind", http.StatusOK, eventsGetOneResponseJSON),
			expected: getExpected[types.Events]{
				result: expectedResponseOneEvent,
			},
		},
		{
			name: "get events sort by action",
			request: getRequest{
				sort: "action",
			},
			mockEndpoints: getEventsGetEndpoints("order=action", http.StatusOK, eventsGetOneResponseJSON),
			expected: getExpected[types.Events]{
				result: expectedResponseOneEvent,
			},
		},
		{
			name: "get events sort by note",
			request: getRequest{
				sort: "note",
			},
			mockEndpoints: getEventsGetEndpoints("order=note", http.StatusOK, eventsGetOneResponseJSON),
			expected: getExpected[types.Events]{
				result: expectedResponseOneEvent,
			},
		},
		{
			name: "get events sort by reason",
			request: getRequest{
				sort: "reason",
			},
			mockEndpoints: getEventsGetEndpoints("order=reason", http.StatusOK, eventsGetOneResponseJSON),
			expected: getExpected[types.Events]{
				result: expectedResponseOneEvent,
			},
		},
		{
			name: "get events sort by age",
			request: getRequest{
				sort: "age",
			},
			mockEndpoints: getEventsGetEndpoints("order=created_at", http.StatusOK, eventsGetOneResponseJSON),
			expected: getExpected[types.Events]{
				result: expectedResponseOneEvent,
			},
		},
		{
			name: "get events sort by age and kind",
			request: getRequest{
				sort: "age.asc,kind.desc",
			},
			mockEndpoints: getEventsGetEndpoints("order=created_at.asc,kind.desc", http.StatusOK, eventsGetOneResponseJSON),
			expected: getExpected[types.Events]{
				result: expectedResponseOneEvent,
			},
		},
		{
			name: "get events invalid json response",
			request: getRequest{
				ids: []string{"id-a"},
			},
			mockEndpoints: getEventsGetEndpoints("id=in.(id-a)", http.StatusOK, `[`),
			expected: getExpected[types.Events]{
				err: "GET ${URL}/postgrest/events?id=in.%28id-a%29 failed\n" +
					" Caused by:\n" +
					"\t* unexpected EOF",
			},
		},
		{
			name: "get events non-200 from server",
			request: getRequest{
				ids: []string{"id-a"},
			},
			mockEndpoints: getEventsGetEndpoints("id=in.(id-a)", http.StatusUnauthorized, `unauthorized`),
			expected: getExpected[types.Events]{
				err: "GET ${URL}/postgrest/events?id=in.%28id-a%29 failed\n" +
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

			actual, err := workbench.GetEvents(context.Background(), svc, tt.request.ids, tt.request.sort)
			if tt.expected.err != "" {
				require.EqualError(t, err, strings.ReplaceAll(tt.expected.err, "${URL}", srv.URL))
				return
			}

			require.NoError(t, err)
			require.Equal(t, tt.expected.result, actual)
		})
	}
}

func TestPrintEvents(t *testing.T) {
	tests := []struct {
		name               string
		resourcesJSON      string
		expectedTableCount int
		expectedTable      string
		expectedPrintCalls int
	}{
		{
			name:               "print events empty resources",
			resourcesJSON:      eventsEmptyResourcesJSON,
			expectedTableCount: 1,
			expectedTable:      expectedEventsEmptyTable,
			expectedPrintCalls: 1,
		},
		{
			name:               "print events prints resources and normalizes note newlines",
			resourcesJSON:      eventsGetTwoResponseJSON,
			expectedTableCount: 1,
			expectedTable:      expectedEventsTwoRowTable,
			expectedPrintCalls: 1,
		},
		{
			name:               "print events prints two resources and normalizes note newlines",
			resourcesJSON:      eventsPrintOneResourcesJSON,
			expectedTableCount: 1,
			expectedTable:      expectedEventsOneRowTable,
			expectedPrintCalls: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var resources types.Events
			err := json.Unmarshal([]byte(tt.resourcesJSON), &resources)
			require.NoError(t, err)

			u := &stubs.StubUI{}
			err = workbench.PrintEvents(u, "ignored-token", resources)
			require.NoError(t, err)

			require.Len(t, u.Tables, tt.expectedTableCount)
			tbl := u.Tables[0]
			require.Equal(t, tt.expectedTable, tbl.String())
			require.Equal(t, tt.expectedPrintCalls, tbl.PrintCalls)
		})
	}
}

func getEventsGetEndpoints(
	expectedQuery string,
	responseStatus int,
	responseBody string,
) []stubs.StubHTTPEndpoint {
	return []stubs.StubHTTPEndpoint{
		{
			Request: stubs.StubHTTPRequest{
				Method:          http.MethodGet,
				URL:             postgrest.Endpoint(constants.WorkbenchEndpointEvents),
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
