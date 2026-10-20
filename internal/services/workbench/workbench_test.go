// Copyright © 2026, SAS Institute Inc., Cary, NC, USA.  All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package workbench_test

import (
	"context"
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
	workbenchGetSuccessBody  = `{"value":"ok"}`
	workbenchPostSuccessBody = `{"ok":"yes"}`
)

type workbenchGetResponse struct {
	Value string `json:"value"`
}

type workbenchPostResponse struct {
	OK string `json:"ok"`
}

func TestNewWorkbenchService(t *testing.T) {
	tests := []struct {
		name     string
		endpoint string
	}{
		{name: "valid endpoint", endpoint: "http://example.com"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := workbench.NewWorkbenchService(tt.endpoint, "access-token", "csrf-token")
			require.NotNil(t, svc)
		})
	}
}

func TestGetAccessToken(t *testing.T) {
	svc := workbench.NewWorkbenchService("http://example.com", "access-token", "csrf-token")
	require.Equal(t, "access-token", workbench.GetAccessToken(svc))
}

func TestGet(t *testing.T) {
	type expected struct {
		err    string
		result workbenchGetResponse
	}
	type testCase struct {
		name          string
		mockEndpoints []stubs.StubHTTPEndpoint
		expected      expected
	}
	tests := []testCase{
		{
			name: "get get success",
			mockEndpoints: getEndpoints(
				"/resource",
				"id=id-a",
				http.StatusOK,
				workbenchGetSuccessBody,
			),
			expected: expected{
				result: workbenchGetResponse{Value: "ok"},
			},
		},
		{
			name: "get get invalid json",
			mockEndpoints: getEndpoints(
				"/resource",
				"id=id-a",
				http.StatusOK,
				`[`,
			),
			expected: expected{
				err: "GET ${URL}/resource?id=id-a failed\n Caused by:\n\t* unexpected EOF",
			},
		},
		{
			name: "get get non-200",
			mockEndpoints: getEndpoints(
				"/resource",
				"id=id-a",
				http.StatusUnauthorized,
				`unauthorized`,
			),
			expected: expected{
				err: "GET ${URL}/resource?id=id-a failed\n Caused by:\n\t* 401 unauthorized",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv, _ := stubs.NewStubHTTPServer(t, tt.mockEndpoints)
			defer srv.Close()
			svc := workbench.NewWorkbenchService(srv.URL, "access-token", "csrf-token")

			actual := workbench.Get[workbenchGetResponse](
				context.Background(),
				svc,
				"/resource",
				map[string][]string{"id": {"id-a"}},
			)
			if tt.expected.err != "" {
				require.EqualError(t, actual.Error(), strings.ReplaceAll(tt.expected.err, "${URL}", srv.URL))
				return
			}
			require.NoError(t, actual.Error())
			require.Equal(t, tt.expected.result, actual.MustGet())
		})
	}
}

func TestGetRaw(t *testing.T) {
	type request struct {
		rawURL string
	}
	type expected struct {
		err    string
		result string
	}
	type testCase struct {
		name          string
		request       request
		mockEndpoints []stubs.StubHTTPEndpoint
		expected      expected
	}
	tests := []testCase{
		{
			name:    "get raw success",
			request: request{rawURL: "/resource?id=id-a"},
			mockEndpoints: getEndpoints(
				"/resource",
				"id=id-a",
				http.StatusOK,
				`raw-body`,
			),
			expected: expected{
				result: "raw-body",
			},
		},
		{
			name:    "get raw non-200",
			request: request{rawURL: "/resource?id=id-a"},
			mockEndpoints: getEndpoints(
				"/resource",
				"id=id-a",
				http.StatusForbidden,
				`forbidden`,
			),
			expected: expected{
				err: "GET ${URL}/resource?id=id-a failed\n Caused by:\n\t* 403 forbidden",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv, _ := stubs.NewStubHTTPServer(t, tt.mockEndpoints)
			defer srv.Close()
			svc := workbench.NewWorkbenchService(srv.URL, "access-token", "csrf-token")

			rawURL := tt.request.rawURL
			if len(rawURL) > 0 && rawURL[0] == '/' {
				rawURL = srv.URL + rawURL
			}
			actual := workbench.GetRaw(context.Background(), svc, rawURL)
			if tt.expected.err != "" {
				require.EqualError(t, actual.Error(), strings.ReplaceAll(tt.expected.err, "${URL}", srv.URL))
				return
			}
			require.NoError(t, actual.Error())
			require.Equal(t, tt.expected.result, actual.MustGet())
		})
	}
}

func TestPost(t *testing.T) {
	type request struct {
		requestBody map[string]string
	}
	type expected struct {
		err   string
		value workbenchPostResponse
	}
	type testCase struct {
		name          string
		request       request
		mockEndpoints []stubs.StubHTTPEndpoint
		expected      expected
	}
	tests := []testCase{
		{
			name:    "success",
			request: request{requestBody: map[string]string{"id": "x"}},
			mockEndpoints: postEndpoints(
				"/resource",
				`{"id":"x"}`,
				http.StatusOK,
				workbenchPostSuccessBody,
			),
			expected: expected{value: workbenchPostResponse{OK: "yes"}},
		},
		{
			name:    "invalid json",
			request: request{requestBody: map[string]string{"id": "x"}},
			mockEndpoints: postEndpoints(
				"/resource",
				`{"id":"x"}`,
				http.StatusOK,
				`[`,
			),
			expected: expected{err: "POST ${URL}/resource failed\n Caused by:\n\t* unexpected EOF"},
		},
		{
			name:    "non-200",
			request: request{requestBody: map[string]string{"id": "x"}},
			mockEndpoints: postEndpoints(
				"/resource",
				`{"id":"x"}`,
				http.StatusUnauthorized,
				`unauthorized`,
			),
			expected: expected{err: "POST ${URL}/resource failed\n Caused by:\n\t* 401 unauthorized"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv, _ := stubs.NewStubHTTPServer(t, tt.mockEndpoints)
			defer srv.Close()
			svc := workbench.NewWorkbenchService(srv.URL, "access-token", "csrf-token")

			actual := workbench.Post[workbenchPostResponse](context.Background(), svc, "/resource", tt.request.requestBody)
			if tt.expected.err != "" {
				require.EqualError(t, actual.Error(), strings.ReplaceAll(tt.expected.err, "${URL}", srv.URL))
				return
			}
			require.NoError(t, actual.Error())
			require.Equal(t, tt.expected.value, actual.MustGet())
		})
	}
}

func TestGetValidDataplaneID(t *testing.T) {
	type request struct {
		dataplaneID string
	}
	type expected struct {
		id  string
		err string
	}
	type testCase struct {
		name          string
		request       request
		mockEndpoints []stubs.StubHTTPEndpoint
		expected      expected
	}
	tests := []testCase{
		{
			name:    "single dataplane without explicit id",
			request: request{dataplaneID: ""},
			mockEndpoints: getEndpoints(
				postgrest.Endpoint(constants.WorkbenchEndpointDataplanes),
				"",
				http.StatusOK,
				`[ {"id":"dp-1"} ]`,
			),
			expected: expected{id: "dp-1"},
		},
		{
			name:    "single dataplane with explicit id",
			request: request{dataplaneID: "dp-1"},
			mockEndpoints: getEndpoints(
				postgrest.Endpoint(constants.WorkbenchEndpointDataplanes),
				"id=in.(dp-1)",
				http.StatusOK,
				`[ {"id":"dp-1"} ]`,
			),
			expected: expected{id: "dp-1"},
		},
		{
			name:    "no dataplanes found",
			request: request{dataplaneID: ""},
			mockEndpoints: getEndpoints(
				postgrest.Endpoint(constants.WorkbenchEndpointDataplanes),
				"",
				http.StatusOK,
				`[]`,
			),
			expected: expected{err: "dataplane not found"},
		},
		{
			name:    "single dataplane not found",
			request: request{dataplaneID: "dp-1"},
			mockEndpoints: getEndpoints(
				postgrest.Endpoint(constants.WorkbenchEndpointDataplanes),
				"id=in.(dp-1)",
				http.StatusOK,
				`[]`,
			),
			expected: expected{err: "dataplane not found: dp-1"},
		},
		{
			name:    "multiple dataplanes with explicit id",
			request: request{dataplaneID: "dp-1"},
			mockEndpoints: getEndpoints(
				postgrest.Endpoint(constants.WorkbenchEndpointDataplanes),
				"id=in.(dp-1)",
				http.StatusOK,
				`[ {"id":"dp-1"}, {"id":"dp-2"} ]`,
			),
			expected: expected{err: "multiple dataplanes found, specify one"},
		},
		{
			name:    "invalid dataplane response",
			request: request{dataplaneID: "dp-1"},
			mockEndpoints: getEndpoints(
				postgrest.Endpoint(constants.WorkbenchEndpointDataplanes),
				"id=in.(dp-1)",
				http.StatusOK,
				`[invalid`,
			),
			expected: expected{
				err: "GET ${URL}/postgrest/dataplanes?id=in.%28dp-1%29 failed" +
					"\n Caused by:" +
					"\n\t* invalid character 'i' looking for beginning of value",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv, _ := stubs.NewStubHTTPServer(t, tt.mockEndpoints)
			defer srv.Close()
			svc := workbench.NewWorkbenchService(srv.URL, "access-token", "csrf-token")

			actual, err := workbench.GetValidDataplaneID(context.Background(), svc, tt.request.dataplaneID)
			if tt.expected.err != "" {
				expectedErr := strings.ReplaceAll(tt.expected.err, "${URL}", srv.URL)
				require.EqualError(t, err, expectedErr)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.expected.id, actual)
		})
	}
}

func TestGetWorkspaceRequests(t *testing.T) {
	type request struct {
		workspaceIDs []string
	}
	type expected struct {
		value []types.WorkspaceRequestWithMetadata
		err   string
	}
	type testCase struct {
		name          string
		request       request
		mockEndpoints []stubs.StubHTTPEndpoint
		expected      expected
	}
	tests := []testCase{
		{
			name:    "uses controller id and home dir mount",
			request: request{workspaceIDs: []string{"ws-1"}},
			mockEndpoints: getEndpoints(
				postgrest.Endpoint(constants.WorkbenchEndpointWorkspaces),
				"controller_id=in.(ws-1)",
				http.StatusOK,
				`[ {"id":"ws-1","controller_id":"ctrl-1","name":"myws","properties":{"HOME_DIRECTORY":"true"}} ]`,
			),
			expected: expected{value: []types.WorkspaceRequestWithMetadata{{
				WorkspaceRequest: types.WorkspaceRequest{ID: "ctrl-1", MountName: "sas"},
				HomeDir:          true,
				DbID:             "ws-1",
			}}},
		},
		{
			name:    "falls back to workspace id",
			request: request{workspaceIDs: []string{"00000000-0000-0000-0000-000000000001"}},
			mockEndpoints: getEndpoints(
				postgrest.Endpoint(constants.WorkbenchEndpointWorkspaces),
				"id=in.(00000000-0000-0000-0000-000000000001)",
				http.StatusOK,
				`[ {"id":"00000000-0000-0000-0000-000000000001","controller_id":"","name":"wsname","properties":{}} ]`,
			),
			expected: expected{value: []types.WorkspaceRequestWithMetadata{{
				WorkspaceRequest: types.WorkspaceRequest{ID: "00000000-0000-0000-0000-000000000001", MountName: "wsname"},
				HomeDir:          false,
				DbID:             "00000000-0000-0000-0000-000000000001",
			}}},
		},
		{
			name:    "workspace not found",
			request: request{workspaceIDs: []string{"missing"}},
			mockEndpoints: getEndpoints(
				postgrest.Endpoint(constants.WorkbenchEndpointWorkspaces),
				"controller_id=in.(missing)",
				http.StatusOK,
				`[]`,
			),
			expected: expected{err: "volume not found: missing"},
		},
		{
			name:    "invalid workspace response",
			request: request{workspaceIDs: []string{"ws-1"}},
			mockEndpoints: getEndpoints(
				postgrest.Endpoint(constants.WorkbenchEndpointWorkspaces),
				"controller_id=in.(ws-1)",
				http.StatusOK,
				`[invalid`,
			),
			expected: expected{
				err: "GET ${URL}/postgrest/workspaces?controller_id=in.%28ws-1%29 failed" +
					"\n Caused by:" +
					"\n\t* invalid character 'i' looking for beginning of value",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv, _ := stubs.NewStubHTTPServer(t, tt.mockEndpoints)
			defer srv.Close()
			svc := workbench.NewWorkbenchService(srv.URL, "access-token", "csrf-token")

			actual, err := workbench.GetValidWorkspaceRequests(context.Background(), svc, tt.request.workspaceIDs)
			if tt.expected.err != "" {
				expectedErr := strings.ReplaceAll(tt.expected.err, "${URL}", srv.URL)
				require.EqualError(t, err, expectedErr)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.expected.value, actual)
		})
	}
}

func TestAddEnvsToServiceRequests(t *testing.T) {
	tests := []struct {
		name             string
		envs             []string
		serviceRequests  []types.ServiceRequest
		expectedRequests []types.ServiceRequest
	}{
		{
			name:            "adds envs to nil map",
			envs:            []string{"A=1", "B=2"},
			serviceRequests: []types.ServiceRequest{{ID: "svc-1"}},
			expectedRequests: []types.ServiceRequest{{
				ID:  "svc-1",
				Env: map[string]string{"A": "1", "B": "2"},
			}},
		},
		{
			name:            "overrides existing keys",
			envs:            []string{"A=new"},
			serviceRequests: []types.ServiceRequest{{ID: "svc-1", Env: map[string]string{"A": "old", "X": "x"}}},
			expectedRequests: []types.ServiceRequest{{
				ID:  "svc-1",
				Env: map[string]string{"A": "new", "X": "x"},
			}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := workbench.AddEnvsToServiceRequests(tt.envs, tt.serviceRequests)
			require.Equal(t, tt.expectedRequests, actual)
		})
	}
}

func TestGetServiceRequests(t *testing.T) {
	type request struct {
		application string
		serviceIDs  []string
	}
	type expected struct {
		value []types.ServiceRequest
		err   string
	}
	type testCase struct {
		name          string
		request       request
		mockEndpoints []stubs.StubHTTPEndpoint
		expected      expected
	}
	tests := []testCase{
		{
			name:     "empty request",
			request:  request{},
			expected: expected{value: nil},
		},
		{
			name:    "from application",
			request: request{application: "app-1"},
			mockEndpoints: getEndpoints(
				postgrest.Endpoint(constants.WorkbenchEndpointDataplaneWorkbenchApplications),
				"id=in.(app-1)",
				http.StatusOK,
				`[ {"id":"app-1","services":["svc-a","svc-b"]} ]`,
			),
			expected: expected{value: []types.ServiceRequest{{ID: "svc-a"}, {ID: "svc-b"}}},
		},
		{
			name:    "application not found",
			request: request{application: "missing"},
			mockEndpoints: getEndpoints(
				postgrest.Endpoint(constants.WorkbenchEndpointDataplaneWorkbenchApplications),
				"id=in.(missing)",
				http.StatusOK,
				`[]`,
			),
			expected: expected{err: "application not found: missing"},
		},
		{
			name:    "from service ids",
			request: request{serviceIDs: []string{"svc-a"}},
			mockEndpoints: getEndpoints(
				postgrest.Endpoint(constants.WorkbenchEndpointDataplaneWorkbenchServices),
				"",
				http.StatusOK,
				`[ {"id":"svc-a"} ]`,
			),
			expected: expected{value: []types.ServiceRequest{{ID: "svc-a"}}},
		},
		{
			name:    "service not found",
			request: request{serviceIDs: []string{"missing"}},
			mockEndpoints: getEndpoints(
				postgrest.Endpoint(constants.WorkbenchEndpointDataplaneWorkbenchServices),
				"",
				http.StatusOK,
				`[]`,
			),
			expected: expected{err: "service not found: missing"},
		},
		{
			name:    "invalid application response",
			request: request{application: "app-1"},
			mockEndpoints: getEndpoints(
				postgrest.Endpoint(constants.WorkbenchEndpointDataplaneWorkbenchApplications),
				"id=in.(app-1)",
				http.StatusOK,
				`[invalid`,
			),
			expected: expected{
				err: "GET ${URL}/postgrest/dataplane_workbench_applications?id=in.%28app-1%29 failed" +
					"\n Caused by:" +
					"\n\t* invalid character 'i' looking for beginning of value",
			},
		},
		{
			name:    "invalid service response",
			request: request{serviceIDs: []string{"service-1"}},
			mockEndpoints: getEndpoints(
				postgrest.Endpoint(constants.WorkbenchEndpointDataplaneWorkbenchServices),
				"",
				http.StatusOK,
				`[invalid`,
			),
			expected: expected{
				err: "GET ${URL}/postgrest/dataplane_workbench_services failed" +
					"\n Caused by:" +
					"\n\t* invalid character 'i' looking for beginning of value",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv, _ := stubs.NewStubHTTPServer(t, tt.mockEndpoints)
			defer srv.Close()
			svc := workbench.NewWorkbenchService(srv.URL, "access-token", "csrf-token")

			actual, err := workbench.GetServiceRequests(context.Background(), svc, tt.request.application, tt.request.serviceIDs)
			if tt.expected.err != "" {
				expectedErr := strings.ReplaceAll(tt.expected.err, "${URL}", srv.URL)
				require.EqualError(t, err, expectedErr)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.expected.value, actual)
		})
	}
}

func TestGetInitializerRequests(t *testing.T) {
	type request struct {
		initializerIDs []string
	}
	type expected struct {
		value []types.ServiceRequest
		err   string
	}
	type testCase struct {
		name          string
		request       request
		mockEndpoints []stubs.StubHTTPEndpoint
		expected      expected
	}
	tests := []testCase{
		{
			name:    "success",
			request: request{initializerIDs: []string{"init-1"}},
			mockEndpoints: getEndpoints(
				postgrest.Endpoint(constants.WorkbenchEndpointDataplaneWorkspaceServices),
				"id=in.(init-1)",
				http.StatusOK,
				`[ {"id":"init-1"} ]`,
			),
			expected: expected{value: []types.ServiceRequest{{ID: "init-1"}}},
		},
		{
			name:    "initializer not found",
			request: request{initializerIDs: []string{"missing"}},
			mockEndpoints: getEndpoints(
				postgrest.Endpoint(constants.WorkbenchEndpointDataplaneWorkspaceServices),
				"id=in.(missing)",
				http.StatusOK,
				`[]`,
			),
			expected: expected{err: "initializer not found: missing"},
		},
		{
			name:    "invalid initializer response",
			request: request{initializerIDs: []string{"init-1"}},
			mockEndpoints: getEndpoints(
				postgrest.Endpoint(constants.WorkbenchEndpointDataplaneWorkspaceServices),
				"id=in.(init-1)",
				http.StatusOK,
				`[invalid`,
			),
			expected: expected{
				err: "GET ${URL}/postgrest/dataplane_workspace_services?id=in.%28init-1%29 failed" +
					"\n Caused by:" +
					"\n\t* invalid character 'i' looking for beginning of value",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv, _ := stubs.NewStubHTTPServer(t, tt.mockEndpoints)
			defer srv.Close()
			svc := workbench.NewWorkbenchService(srv.URL, "access-token", "csrf-token")

			actual, err := workbench.GetInitializerRequests(context.Background(), svc, tt.request.initializerIDs)
			if tt.expected.err != "" {
				expectedErr := strings.ReplaceAll(tt.expected.err, "${URL}", srv.URL)
				require.EqualError(t, err, expectedErr)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.expected.value, actual)
		})
	}
}

func getEndpoints(
	url string,
	expectedQuery string,
	status int,
	body string,
) []stubs.StubHTTPEndpoint {
	return []stubs.StubHTTPEndpoint{{
		Request: stubs.StubHTTPRequest{
			Method:          http.MethodGet,
			URL:             url,
			ExpectedHeaders: authHeaders,
			ExpectedQuery:   expectedQuery,
		},
		Response: stubs.StubHTTPResponse{
			Status: status,
			Body:   body,
		},
	}}
}

func postEndpoints(
	url string,
	expectedBody string,
	status int,
	body string,
) []stubs.StubHTTPEndpoint {
	return []stubs.StubHTTPEndpoint{{
		Request: stubs.StubHTTPRequest{
			Method:          http.MethodPost,
			URL:             url,
			ExpectedHeaders: authHeaders,
			ExpectedBody:    expectedBody,
		},
		Response: stubs.StubHTTPResponse{
			Status: status,
			Body:   body,
		},
	}}
}
