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

	"github.com/sassoftware/wbctl/internal/constants"
	"github.com/sassoftware/wbctl/internal/postgrest"
	"github.com/sassoftware/wbctl/internal/services/workbench"
	"github.com/sassoftware/wbctl/internal/stubs"
	"github.com/sassoftware/wbctl/internal/timeutil"
	"github.com/sassoftware/wbctl/internal/types"
	"github.com/stretchr/testify/require"
)

const (
	dataplanesEmptyResourcesJSON = `[]`

	dataplanesGetOneResponseJSON = `
[
  {
    "id": "dp-1",
    "name": "Dataplane 1",
    "description": "first"
  }
]
`

	dataplanesGetTwoResponseJSON = `
[
  {
    "id": "id-a",
    "name": "Dataplane A"
  },
  {
    "id": "id-b",
    "name": "Dataplane B"
  }
]
`

	dataplanesPrintOneResourcesJSON = `
[
  {
    "id": "dp-1",
    "name": "Dataplane 1",
    "description": "first",
    "message": "ok",
    "created_at": null
  }
]
`

	dataplanesPrintTwoResourcesJSON = `
[
  {
	"id": "dp-1",
	"name": "Dataplane 1",
	"description": "first",
	"message": "ok",
	"created_at": null
  },
  {
	"id": "dp-2",
	"name": "Dataplane 2",
	"description": "second",
	"message": "ready",
	"created_at": null
  }
]
`

	dataplanesDescribeOneResourcesJSON = `
[
  {
    "id": "dp-1",
    "name": "Dataplane 1",
    "description": "first",
    "created_by": "creator",
    "created_at": "2024-01-02T03:04:05Z",
    "modified_at": "2024-01-02T04:05:06Z",
    "state": "READY",
    "spec": {
      "cluster": "c1"
    },
    "message": "healthy",
    "properties": {
      "zone": "us-east"
    }
  }
]
`

	dataplanesDescribeTwoResourcesJSON = `
[
  {
	"id": "dp-1",
	"name": "Dataplane 1",
	"description": "first",
	"created_by": "creator",
	"created_at": "2024-01-02T03:04:05Z",
	"modified_at": "2024-01-02T04:05:06Z",
	"state": "READY",
	"spec": {
	  "cluster": "c1"
	},
	"message": "healthy",
	"properties": {
	  "zone": "us-east"
	}
  },
  {
	"id": "dp-3",
	"name": "Dataplane 3",
	"description": "third",
	"created_by": "creator-3",
	"created_at": "2024-03-04T05:06:07Z",
	"modified_at": "2024-03-04T06:07:08Z",
	"state": "READY",
	"spec": {
	  "cluster": "c3"
	},
	"message": "green",
	"properties": {
	  "zone": "eu-west"
	}
  }
]
`

	expectedDataplanesEmptyTable = "ID\tNAME\tDESCRIPTION\tMESSAGE\tAGE"

	expectedDataplanesOneRowTable = "ID\tNAME\tDESCRIPTION\tMESSAGE\tAGE\n" +
		"dp-1\tDataplane 1\tfirst\tok\t"

	expectedDataplanesTwoRowsTable = "ID\tNAME\tDESCRIPTION\tMESSAGE\tAGE\n" +
		"dp-1\tDataplane 1\tfirst\tok\t\n" +
		"dp-2\tDataplane 2\tsecond\tready\t"

	expectedDescribeDataplaneOneDisplay = `ID:            dp-1
Name:          Dataplane 1
Description:   first
Created By:    creator
Created At:    2024-01-02T03:04:05Z (2y126d ago)
Modified At:   2024-01-02T04:05:06Z (2y126d ago)
State:         READY
Spec:          {
                 "cluster": "c1"
               }
Message:   healthy
Properties:
  zone:   us-east
`

	expectedDescribeDataplaneSpecMarshalErrorDisplay = `ID:            dp-1
Name:          Dataplane 1
Description:   first
Created By:    creator
Created At:    2024-01-02T03:04:05Z (2y126d ago)
Modified At:   2024-01-02T04:05:06Z (2y126d ago)
State:         READY
Spec:          <error: json: unsupported type: func()>
Message:       healthy
Properties:
  zone:   us-east
`

	expectedDescribeDataplanesTwoDisplay = expectedDescribeDataplaneOneDisplay +
		"\n\nID:            dp-3\n" +
		"Name:          Dataplane 3\n" +
		"Description:   third\n" +
		"Created By:    creator-3\n" +
		"Created At:    2024-03-04T05:06:07Z (2y64d ago)\n" +
		"Modified At:   2024-03-04T06:07:08Z (2y64d ago)\n" +
		"State:         READY\n" +
		"Spec:          {\n" +
		"                 \"cluster\": \"c3\"\n" +
		"               }\n" +
		"Message:   green\n" +
		"Properties:\n" +
		"  zone:   eu-west\n"

	dataplanesDescribeEmptySpecResourcesJSON = `
[
  {
    "id": "dp-2",
    "name": "Dataplane 2",
    "created_by": "creator-2",
    "created_at": "2024-02-03T01:02:03Z",
    "modified_at": "2024-02-03T02:03:04Z",
    "state": "IDLE",
    "spec": {}
  }
]
`

	expectedDescribeDataplaneEmptySpecDisplay = `ID:            dp-2
Name:          Dataplane 2
Created By:    creator-2
Created At:    2024-02-03T01:02:03Z (2y94d ago)
Modified At:   2024-02-03T02:03:04Z (2y94d ago)
State:         IDLE
`
)

func TestDescribeDataplanes(t *testing.T) {
	fixedNow := time.Date(2026, time.May, 8, 0, 0, 0, 0, time.UTC)
	timeutil.SetNow(func() time.Time { return fixedNow })
	t.Cleanup(func() { timeutil.SetNow(time.Now) })

	tests := []struct {
		name             string
		resourcesJSON    string
		mutate           func(types.Dataplanes)
		expectedDisplays []string
		expectedWarns    []string
	}{
		{
			name:             "describe dataplanes describe empty resources prints nothing",
			resourcesJSON:    dataplanesEmptyResourcesJSON,
			expectedDisplays: []string{""},
		},
		{
			name:             "describe dataplanes describe prints one dataplane and properties",
			resourcesJSON:    dataplanesDescribeOneResourcesJSON,
			expectedDisplays: []string{expectedDescribeDataplaneOneDisplay},
		},
		{
			name:             "describe dataplanes describe prints two dataplanes and properties",
			resourcesJSON:    dataplanesDescribeTwoResourcesJSON,
			expectedDisplays: []string{expectedDescribeDataplanesTwoDisplay},
		},
		{
			name:          "describe dataplanes describe spec marshal error path",
			resourcesJSON: dataplanesDescribeOneResourcesJSON,
			mutate: func(resources types.Dataplanes) {
				resources[0].Spec["bad"] = func() {}
			},
			expectedDisplays: []string{expectedDescribeDataplaneSpecMarshalErrorDisplay},
		},
		{
			name:             "describe dataplanes describe empty spec and no properties omits both sections",
			resourcesJSON:    dataplanesDescribeEmptySpecResourcesJSON,
			expectedDisplays: []string{expectedDescribeDataplaneEmptySpecDisplay},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var resources types.Dataplanes
			err := json.Unmarshal([]byte(tt.resourcesJSON), &resources)
			require.NoError(t, err)
			if tt.mutate != nil {
				tt.mutate(resources)
			}

			stubUI := &stubs.StubUI{}
			err = workbench.DescribeDataplanes(stubUI, resources)
			require.NoError(t, err)

			require.Equal(t, tt.expectedDisplays, stubUI.Displays, strings.Join(stubUI.Displays, "\n"))
			require.Equal(t, tt.expectedWarns, stubUI.Warns, strings.Join(stubUI.Warns, "\n"))
		})
	}
}

func TestGetDataplanes(t *testing.T) {
	tests := []getTestCase[types.Dataplanes]{
		{
			name:          "get dataplanes success without ids query",
			request:       getRequest{},
			mockEndpoints: getDataplanesGetEndpoints("", http.StatusOK, dataplanesGetOneResponseJSON),
			expected: getExpected[types.Dataplanes]{
				result: types.Dataplanes{
					{ID: "dp-1", Name: "Dataplane 1", Description: "first"},
				},
			},
		},
		{
			name: "get dataplanes success with ids query",
			request: getRequest{
				ids:  []string{"id-a", "id-b"},
				sort: "id.asc",
			},
			mockEndpoints: getDataplanesGetEndpoints(
				"id=in.(id-a,id-b)&order=id.asc",
				http.StatusOK,
				dataplanesGetTwoResponseJSON,
			),
			expected: getExpected[types.Dataplanes]{
				result: types.Dataplanes{
					{ID: "id-a", Name: "Dataplane A"},
					{ID: "id-b", Name: "Dataplane B"},
				},
			},
		},
		{
			name: "get dataplanes sort by name",
			request: getRequest{
				sort: "name",
			},
			mockEndpoints: getDataplanesGetEndpoints("order=name", http.StatusOK, dataplanesGetTwoResponseJSON),
			expected: getExpected[types.Dataplanes]{
				result: types.Dataplanes{
					{ID: "id-a", Name: "Dataplane A"},
					{ID: "id-b", Name: "Dataplane B"},
				},
			},
		},
		{
			name: "get dataplanes sort by description",
			request: getRequest{
				sort: "description",
			},
			mockEndpoints: getDataplanesGetEndpoints("order=description", http.StatusOK, dataplanesGetTwoResponseJSON),
			expected: getExpected[types.Dataplanes]{
				result: types.Dataplanes{
					{ID: "id-a", Name: "Dataplane A"},
					{ID: "id-b", Name: "Dataplane B"},
				},
			},
		},
		{
			name: "get dataplanes sort by message",
			request: getRequest{
				sort: "message",
			},
			mockEndpoints: getDataplanesGetEndpoints("order=message", http.StatusOK, dataplanesGetTwoResponseJSON),
			expected: getExpected[types.Dataplanes]{
				result: types.Dataplanes{
					{ID: "id-a", Name: "Dataplane A"},
					{ID: "id-b", Name: "Dataplane B"},
				},
			},
		},
		{
			name: "get dataplanes sort by age",
			request: getRequest{
				sort: "age",
			},
			mockEndpoints: getDataplanesGetEndpoints("order=created_at", http.StatusOK, dataplanesGetTwoResponseJSON),
			expected: getExpected[types.Dataplanes]{
				result: types.Dataplanes{
					{ID: "id-a", Name: "Dataplane A"},
					{ID: "id-b", Name: "Dataplane B"},
				},
			},
		},
		{
			name: "get dataplanes sort by age and id",
			request: getRequest{
				sort: "age.desc,id.asc",
			},
			mockEndpoints: getDataplanesGetEndpoints(
				"order=created_at.desc,id.asc",
				http.StatusOK,
				dataplanesGetTwoResponseJSON,
			),
			expected: getExpected[types.Dataplanes]{
				result: types.Dataplanes{
					{ID: "id-a", Name: "Dataplane A"},
					{ID: "id-b", Name: "Dataplane B"},
				},
			},
		},
		{
			name: "get dataplanes invalid json response",
			request: getRequest{
				ids: []string{"id-a"},
			},
			mockEndpoints: getDataplanesGetEndpoints("id=in.(id-a)", http.StatusOK, `[`),
			expected: getExpected[types.Dataplanes]{
				err: "GET ${URL}/postgrest/dataplanes?id=in.%28id-a%29 failed\n" +
					" Caused by:\n" +
					"\t* unexpected EOF",
			},
		},
		{
			name: "get dataplanes non-200 from server",
			request: getRequest{
				ids: []string{"id-a"},
			},
			mockEndpoints: getDataplanesGetEndpoints(
				"id=in.(id-a)",
				http.StatusUnauthorized,
				`get dataplane unauthorized`,
			),
			expected: getExpected[types.Dataplanes]{
				err: "GET ${URL}/postgrest/dataplanes?id=in.%28id-a%29 failed\n" +
					" Caused by:\n" +
					"\t* 401 get dataplane unauthorized",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv, _ := stubs.NewStubHTTPServer(t, tt.mockEndpoints)
			defer srv.Close()
			svc := workbench.NewWorkbenchService(srv.URL, "access-token", "csrf-token")

			actual, err := workbench.GetDataplanes(context.Background(), svc, tt.request.ids, tt.request.sort)
			if tt.expected.err != "" {
				require.EqualError(t, err, strings.ReplaceAll(tt.expected.err, "${URL}", srv.URL))
				return
			}

			require.NoError(t, err)
			require.Equal(t, tt.expected.result, actual)
		})
	}
}

func TestPrintDataplanes(t *testing.T) {
	tests := []struct {
		name               string
		resourcesJSON      string
		expectedTableCount int
		expectedTable      string
		expectedPrintCalls int
	}{
		{
			name:               "print dataplane empty resources",
			resourcesJSON:      dataplanesEmptyResourcesJSON,
			expectedTableCount: 1,
			expectedTable:      expectedDataplanesEmptyTable,
			expectedPrintCalls: 1,
		},
		{
			name:               "print dataplane prints resources",
			resourcesJSON:      dataplanesPrintOneResourcesJSON,
			expectedTableCount: 1,
			expectedTable:      expectedDataplanesOneRowTable,
			expectedPrintCalls: 1,
		},
		{
			name:               "print dataplane prints two resources",
			resourcesJSON:      dataplanesPrintTwoResourcesJSON,
			expectedTableCount: 1,
			expectedTable:      expectedDataplanesTwoRowsTable,
			expectedPrintCalls: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var resources types.Dataplanes
			err := json.Unmarshal([]byte(tt.resourcesJSON), &resources)
			require.NoError(t, err)

			u := &stubs.StubUI{}
			err = workbench.PrintDataplanes(u, "ignored-token", resources)
			require.NoError(t, err)

			require.Len(t, u.Tables, tt.expectedTableCount)
			tbl := u.Tables[0]
			require.Equal(t, tt.expectedTable, tbl.String())
			require.Equal(t, tt.expectedPrintCalls, tbl.PrintCalls)
		})
	}
}

func getDataplanesGetEndpoints(
	expectedQuery string,
	responseStatus int,
	responseBody string,
) []stubs.StubHTTPEndpoint {
	return []stubs.StubHTTPEndpoint{
		{
			Request: stubs.StubHTTPRequest{
				Method:          http.MethodGet,
				URL:             postgrest.Endpoint(constants.WorkbenchEndpointDataplanes),
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
