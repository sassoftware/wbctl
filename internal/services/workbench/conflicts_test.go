// Copyright © 2026, SAS Institute Inc., Cary, NC, USA.  All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package workbench_test

import (
	"context"
	"encoding/json"
	"fmt"
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

var (
	emptyConflictsResourcesJSON = `[]`

	oneGetConflictResponseJSON = `[{
				"id":"conflict-1",
				"name":"name-1"
			}]`

	twoGetConflictsResponseJSON = `[{
				"id":"id-a",
				"name":"name-a"
			},{
				"id":"id-b",
				"name":"name-b"
			}]`

	expectedEmptyConflictsTable = strings.Join([]string{
		"ID\tNAME\tTYPE\tCONFLICT_ID\tCONFLICT_NAME\tCONFLICT_TYPE\tVOLUME_ID\tVOLUME_NAME",
	}, "\n")

	expectedOneConflictTable = strings.Join([]string{
		"ID\tNAME\tTYPE\tCONFLICT_ID\tCONFLICT_NAME\tCONFLICT_TYPE\tVOLUME_ID\tVOLUME_NAME",
		"c1\tconflict-a\tTYPE-A\tcx1\tother-a\tTYPE-X\tw1\tworkspace-a",
	}, "\n")

	multipleDescribeConflictsResourcesJSON = `[
				{
					"id":"c1",
					"controller_id":"ctrl-1",
					"name":"conflict-a",
					"type":"TYPE-A",
					"conflict_id":"cx1",
					"conflict_controller_id":"cc1",
					"conflict_name":"other-a",
					"conflict_type":"TYPE-X",
					"workspace_id":"w1",
					"workspace_controller_id":"wc1",
					"workspace_name":"workspace-a"
				},
				{
					"id":"c2",
					"controller_id":"ctrl-2",
					"name":"conflict-b",
					"type":"TYPE-B",
					"conflict_id":"cx2",
					"conflict_controller_id":"cc2",
					"conflict_name":"other-b",
					"conflict_type":"TYPE-Y",
					"workspace_id":"w2",
					"workspace_controller_id":"wc2",
					"workspace_name":"workspace-b"
				}
			]`

	oneDescribeConflictResourcesJSON = `[
				{
					"id":"c1",
					"controller_id":"ctrl-1",
					"name":"conflict-a",
					"type":"TYPE-A",
					"conflict_id":"cx1",
					"conflict_controller_id":"cc1",
					"conflict_name":"other-a",
					"conflict_type":"TYPE-X",
					"workspace_id":"w1",
					"workspace_controller_id":"wc1",
					"workspace_name":"workspace-a"
				}
			]`

	onePrintConflictResourcesJSON = `[
				{
					"id":"c1",
					"name":"conflict-a",
					"type":"TYPE-A",
					"conflict_id":"cx1",
					"conflict_name":"other-a",
					"conflict_type":"TYPE-X",
					"workspace_id":"w1",
					"workspace_name":"workspace-a"
				}
			]`

	expectedOneConflictOutput = `ID:                        c1
Controller ID:             ctrl-1
Name:                      conflict-a
Type:                      TYPE-A
Conflict ID:               cx1
Conflict Controller ID:    cc1
Conflict Name:             other-a
Conflict Type:             TYPE-X
Workspace ID:              w1
Workspace Controller ID:   wc1
Workspace Name:            workspace-a
`

	expectedMultipleConflictOutput = `ID:                        c1
Controller ID:             ctrl-1
Name:                      conflict-a
Type:                      TYPE-A
Conflict ID:               cx1
Conflict Controller ID:    cc1
Conflict Name:             other-a
Conflict Type:             TYPE-X
Workspace ID:              w1
Workspace Controller ID:   wc1
Workspace Name:            workspace-a


ID:                        c2
Controller ID:             ctrl-2
Name:                      conflict-b
Type:                      TYPE-B
Conflict ID:               cx2
Conflict Controller ID:    cc2
Conflict Name:             other-b
Conflict Type:             TYPE-Y
Workspace ID:              w2
Workspace Controller ID:   wc2
Workspace Name:            workspace-b
`
)

func TestDescribeConflicts(t *testing.T) {
	tests := []struct {
		name             string
		resourcesJSON    string
		expectedDisplays []string
		expectedWarns    []string
	}{
		{
			name:             "describe conflicts describe empty resources prints nothing",
			resourcesJSON:    emptyConflictsResourcesJSON,
			expectedDisplays: []string{""},
		},
		{
			name:             "describe conflicts describe prints one conflict",
			resourcesJSON:    oneDescribeConflictResourcesJSON,
			expectedDisplays: []string{expectedOneConflictOutput},
		},
		{
			name:             "describe conflicts describe prints multiple conflicts in order",
			resourcesJSON:    multipleDescribeConflictsResourcesJSON,
			expectedDisplays: []string{expectedMultipleConflictOutput},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var resources types.Conflicts
			err := json.Unmarshal([]byte(tt.resourcesJSON), &resources)
			require.NoError(t, err)

			stubUI := &stubs.StubUI{}
			err = workbench.DescribeConflicts(stubUI, resources)
			require.NoError(t, err)
			require.Equal(t, tt.expectedDisplays, stubUI.Displays, strings.Join(stubUI.Displays, "\n"))
			require.Equal(t, tt.expectedWarns, stubUI.Warns, strings.Join(stubUI.Warns, "\n"))
		})
	}
}

func TestGetConflicts(t *testing.T) {
	tests := []getTestCase[types.Conflicts]{
		{
			name:          "get conflicts success without ids query",
			request:       getRequest{},
			mockEndpoints: getConflictsGetEndpoints("", http.StatusOK, oneGetConflictResponseJSON),
			expected: getExpected[types.Conflicts]{
				result: types.Conflicts{
					{ID: "conflict-1", Name: "name-1"},
				},
			},
		},
		{
			name: "get conflicts sort by id",
			request: getRequest{
				sort: "id",
			},
			mockEndpoints: getConflictsGetEndpoints("order=id", http.StatusOK, oneGetConflictResponseJSON),
			expected: getExpected[types.Conflicts]{
				result: types.Conflicts{
					{ID: "conflict-1", Name: "name-1"},
				},
			},
		},
		{
			name: "get conflicts sort by name",
			request: getRequest{
				sort: "name",
			},
			mockEndpoints: getConflictsGetEndpoints("order=name", http.StatusOK, oneGetConflictResponseJSON),
			expected: getExpected[types.Conflicts]{
				result: types.Conflicts{
					{ID: "conflict-1", Name: "name-1"},
				},
			},
		},
		{
			name: "get conflicts sort by type",
			request: getRequest{
				sort: "type",
			},
			mockEndpoints: getConflictsGetEndpoints("order=type", http.StatusOK, oneGetConflictResponseJSON),
			expected: getExpected[types.Conflicts]{
				result: types.Conflicts{
					{ID: "conflict-1", Name: "name-1"},
				},
			},
		},
		{
			name: "get conflicts sort by conflict_id",
			request: getRequest{
				sort: "conflict_id",
			},
			mockEndpoints: getConflictsGetEndpoints("order=conflict_id", http.StatusOK, oneGetConflictResponseJSON),
			expected: getExpected[types.Conflicts]{
				result: types.Conflicts{
					{ID: "conflict-1", Name: "name-1"},
				},
			},
		},
		{
			name: "get conflicts sort by conflict_name",
			request: getRequest{
				sort: "conflict_name",
			},
			mockEndpoints: getConflictsGetEndpoints("order=conflict_name", http.StatusOK, oneGetConflictResponseJSON),
			expected: getExpected[types.Conflicts]{
				result: types.Conflicts{
					{ID: "conflict-1", Name: "name-1"},
				},
			},
		},
		{
			name: "get conflicts sort by conflict_type",
			request: getRequest{
				sort: "conflict_type",
			},
			mockEndpoints: getConflictsGetEndpoints("order=conflict_type", http.StatusOK, oneGetConflictResponseJSON),
			expected: getExpected[types.Conflicts]{
				result: types.Conflicts{
					{ID: "conflict-1", Name: "name-1"},
				},
			},
		},
		{
			name: "get conflicts sort by workspace_id",
			request: getRequest{
				sort: "workspace_id",
			},
			mockEndpoints: getConflictsGetEndpoints("order=workspace_id", http.StatusOK, oneGetConflictResponseJSON),
			expected: getExpected[types.Conflicts]{
				result: types.Conflicts{
					{ID: "conflict-1", Name: "name-1"},
				},
			},
		},
		{
			name: "get conflicts sort by workspace_name",
			request: getRequest{
				sort: "workspace_name",
			},
			mockEndpoints: getConflictsGetEndpoints("order=workspace_name", http.StatusOK, oneGetConflictResponseJSON),
			expected: getExpected[types.Conflicts]{
				result: types.Conflicts{
					{ID: "conflict-1", Name: "name-1"},
				},
			},
		},
		{
			name: "get conflicts sort by volume_name and volume_id",
			request: getRequest{
				sort: "volume_name.desc,volume_id.asc",
			},
			mockEndpoints: getConflictsGetEndpoints(
				"order=workspace_name.desc,workspace_id.asc",
				http.StatusOK,
				oneGetConflictResponseJSON,
			),
			expected: getExpected[types.Conflicts]{
				result: types.Conflicts{
					{ID: "conflict-1", Name: "name-1"},
				},
			},
		},
		{
			name: "get conflicts success with ids query",
			request: getRequest{
				ids: []string{"id-a", "id-b"},
			},
			mockEndpoints: getConflictsGetEndpoints("id=in.(id-a,id-b)", http.StatusOK, twoGetConflictsResponseJSON),
			expected: getExpected[types.Conflicts]{
				result: types.Conflicts{
					{ID: "id-a", Name: "name-a"},
					{ID: "id-b", Name: "name-b"},
				},
			},
		},
		{
			name: "get conflicts invalid json response",
			request: getRequest{
				ids: []string{"id-a"},
			},
			mockEndpoints: getConflictsGetEndpoints("id=in.(id-a)", http.StatusOK, "["),
			expected: getExpected[types.Conflicts]{
				err: "GET %s/postgrest/workspace_conflicts?id=in.%%28id-a%%29 failed\n" +
					" Caused by:\n" +
					"\t* unexpected EOF",
			},
		},
		{
			name: "get conflicts non-200 from server",
			request: getRequest{
				ids: []string{"id-a"},
			},
			mockEndpoints: getConflictsGetEndpoints("id=in.(id-a)", http.StatusUnauthorized, "unauthorized"),
			expected: getExpected[types.Conflicts]{
				err: "GET %s/postgrest/workspace_conflicts?id=in.%%28id-a%%29 failed\n" +
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

			actual, err := workbench.GetConflicts(context.Background(), svc, tt.request.ids, tt.request.sort)
			if tt.expected.err != "" {
				require.EqualError(t, err, fmt.Sprintf(tt.expected.err, srv.URL))
				return
			}

			require.NoError(t, err)
			require.Equal(t, tt.expected.result, actual)
		})
	}
}

func TestPrintConflicts(t *testing.T) {
	tests := []struct {
		name               string
		resourcesJSON      string
		expectedTableCount int
		expectedTable      string
		expectedPrintCalls int
	}{
		{
			name:               "empty resources",
			resourcesJSON:      emptyConflictsResourcesJSON,
			expectedTableCount: 1,
			expectedTable:      expectedEmptyConflictsTable,
			expectedPrintCalls: 1,
		},
		{
			name:               "prints resources",
			resourcesJSON:      onePrintConflictResourcesJSON,
			expectedTableCount: 1,
			expectedTable:      expectedOneConflictTable,
			expectedPrintCalls: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var resources types.Conflicts
			err := json.Unmarshal([]byte(tt.resourcesJSON), &resources)
			require.NoError(t, err)

			u := &stubs.StubUI{}
			err = workbench.PrintConflicts(u, "ignored-token", resources)
			require.NoError(t, err)

			require.Len(t, u.Tables, tt.expectedTableCount)
			tbl := u.Tables[0]
			require.Equal(t, tt.expectedTable, tbl.String())
			require.Equal(t, tt.expectedPrintCalls, tbl.PrintCalls)
		})
	}
}

func getConflictsGetEndpoints(exectedQuery string, responseStatus int, responseBody string) []stubs.StubHTTPEndpoint {
	return []stubs.StubHTTPEndpoint{
		{
			Request: stubs.StubHTTPRequest{
				Method:          http.MethodGet,
				URL:             postgrest.Endpoint(constants.WorkbenchEndpointWorkspaceConflicts),
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
