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
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	settingsEmptyResourcesJSON      = `{}`
	settingsGetOneResponseJSON      = `{"foo": "bar"}`
	settingsGetInvalidJSON          = `{"foo": `
	expectedSettingsRequestJSON     = `{"name":"foo","value":"bar"}`
	expectedUserSettingsRequestJSON = `{"user_id":"user1","name":"foo","value":"bar"}`
)

type getSettingsRequest struct {
	userID string
	tenant bool
}

func TestGetSettings(t *testing.T) {
	tests := []struct {
		name          string
		request       getSettingsRequest
		mockEndpoints []stubs.StubHTTPEndpoint
		expected      getExpected[types.Settings]
	}{
		{
			name: "get settings success",
			mockEndpoints: getSettingsGetEndpoints(
				"",
				false,
				http.StatusOK,
				settingsGetOneResponseJSON,
			),
			expected: getExpected[types.Settings]{
				result: types.Settings{"foo": "bar"},
			},
		},
		{
			name: "get user settings success",
			request: getSettingsRequest{
				userID: "user1",
			},
			mockEndpoints: getSettingsGetEndpoints(
				"user1",
				false,
				http.StatusOK,
				settingsGetOneResponseJSON,
			),
			expected: getExpected[types.Settings]{
				result: types.Settings{"foo": "bar"},
			},
		},
		{
			name:    "get tenant settings success",
			request: getSettingsRequest{tenant: true},
			mockEndpoints: getSettingsGetEndpoints(
				"",
				true,
				http.StatusOK,
				settingsGetOneResponseJSON,
			),
			expected: getExpected[types.Settings]{
				result: types.Settings{"foo": "bar"},
			},
		},
		{
			name: "get settings invalid json",
			mockEndpoints: getSettingsGetEndpoints(
				"",
				false,
				http.StatusOK,
				settingsGetInvalidJSON,
			),
			expected: getExpected[types.Settings]{
				result: types.Settings{"foo": "bar"},
				err:    "GET ${URL}/postgrest/rpc/get_settings failed\n Caused by:\n\t* unexpected EOF",
			},
		},
		{
			name: "get settings non-200",
			mockEndpoints: getSettingsGetEndpoints(
				"",
				false,
				http.StatusUnauthorized,
				`unauthorized`,
			),
			expected: getExpected[types.Settings]{
				result: types.Settings{"foo": "bar"},
				err:    "GET ${URL}/postgrest/rpc/get_settings failed\n Caused by:\n\t* 401 unauthorized",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv, _ := stubs.NewStubHTTPServer(t, tt.mockEndpoints)
			defer srv.Close()
			svc := workbench.NewWorkbenchService(srv.URL, "access-token", "csrf-token")

			actual, err := workbench.GetSettings(
				context.Background(),
				svc,
				tt.request.userID,
				tt.request.tenant,
			)
			if tt.expected.err != "" {
				require.EqualError(t, err, strings.ReplaceAll(tt.expected.err, "${URL}", srv.URL))
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.expected.result, actual)
		})
	}
}

func TestPrintSettings(t *testing.T) {
	t.Run("error from jsonutils.ToJSON", func(t *testing.T) {
		// Simulate a value that causes jsonutils.ToJSON to error
		resources := types.Settings{"bad": make(chan int)}
		u := &stubs.StubUI{}
		err := workbench.PrintSettings(u, resources)
		require.Error(t, err)
		require.Contains(t, err.Error(), "json: unsupported type")
	})

	t.Run("multiple keys sorted", func(t *testing.T) {
		resources := types.Settings{"z": 1, "a": 2, "m": 3}
		u := &stubs.StubUI{}
		err := workbench.PrintSettings(u, resources)
		require.NoError(t, err)
		require.Len(t, u.Tables, 1)
		tbl := u.Tables[0]
		// Should be sorted: a, m, z
		expected := "NAME\tVALUE\na\t2\nm\t3\nz\t1"
		require.Equal(t, expected, tbl.String())
	})

	tests := []struct {
		name               string
		resourcesJSON      string
		expectedTableCount int
		expectedTable      string
		expectedPrintCalls int
	}{
		{
			name:               "print settings empty",
			resourcesJSON:      settingsEmptyResourcesJSON,
			expectedTableCount: 1,
			expectedTable:      "NAME\tVALUE",
			expectedPrintCalls: 1,
		},
		{
			name:               "print settings one entry",
			resourcesJSON:      settingsGetOneResponseJSON,
			expectedTableCount: 1,
			expectedTable:      "NAME\tVALUE\nfoo\t\"bar\"",
			expectedPrintCalls: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var resources types.Settings
			err := json.Unmarshal([]byte(tt.resourcesJSON), &resources)
			require.NoError(t, err)

			u := &stubs.StubUI{}
			err = workbench.PrintSettings(u, resources)
			require.NoError(t, err)

			require.Len(t, u.Tables, tt.expectedTableCount)
			tbl := u.Tables[0]
			require.Equal(t, tt.expectedTable, tbl.String())
			require.Equal(t, tt.expectedPrintCalls, tbl.PrintCalls)
		})
	}
}

type setSettingsRequest struct {
	userID string
	tenant bool
	name   string
	value  string
}

func TestSetSetting(t *testing.T) {
	tests := []struct {
		name          string
		request       setSettingsRequest
		mockEndpoints []stubs.StubHTTPEndpoint
		expected      getExpected[bool]
	}{
		{
			name:    "set setting success",
			request: setSettingsRequest{name: "foo", value: "bar"},
			mockEndpoints: getSettingsSetEndpoints(
				expectedSettingsRequestJSON,
				false,
				http.StatusOK,
				`true`,
			),
			expected: getExpected[bool]{
				result: true,
			},
		},
		{
			name: "set user setting success",
			request: setSettingsRequest{
				userID: "user1",
				name:   "foo",
				value:  "bar",
			},
			mockEndpoints: getSettingsSetEndpoints(
				expectedUserSettingsRequestJSON,
				false,
				http.StatusOK,
				`true`,
			),
			expected: getExpected[bool]{
				result: true,
			},
		},
		{
			name:    "set tenant setting success",
			request: setSettingsRequest{tenant: true, name: "foo", value: "bar"},
			mockEndpoints: getSettingsSetEndpoints(
				expectedSettingsRequestJSON,
				true,
				http.StatusOK,
				`true`,
			),
			expected: getExpected[bool]{
				result: true,
			},
		},
		{
			name:    "set user setting error",
			request: setSettingsRequest{name: "foo", value: "bar"},
			mockEndpoints: getSettingsSetEndpoints(
				expectedSettingsRequestJSON,
				false,
				http.StatusInternalServerError,
				`fail`,
			),
			expected: getExpected[bool]{
				err: "POST ${URL}/postgrest/rpc/set_user_setting failed\n Caused by:\n\t* 500 fail",
			},
		},
		{
			name:    "set tenant setting error",
			request: setSettingsRequest{tenant: true, name: "foo", value: "bar"},
			mockEndpoints: getSettingsSetEndpoints(
				expectedSettingsRequestJSON,
				true,
				http.StatusInternalServerError,
				`fail`,
			),
			expected: getExpected[bool]{
				err: "POST ${URL}/postgrest/rpc/set_tenant_setting failed\n Caused by:\n\t* 500 fail",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv, _ := stubs.NewStubHTTPServer(t, tt.mockEndpoints)
			defer srv.Close()
			svc := workbench.NewWorkbenchService(srv.URL, "access-token", "csrf-token")
			res, err := workbench.SetSetting(
				context.Background(),
				svc,
				tt.request.userID,
				tt.request.name,
				tt.request.value,
				tt.request.tenant,
			)
			if tt.expected.err != "" {
				require.EqualError(t, err, strings.ReplaceAll(tt.expected.err, "${URL}", srv.URL))
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.expected.result, res)
		})
	}
}

func getSettingsGetEndpoints(
	userID string,
	tenant bool,
	responseStatus int,
	responseBody string,
) []stubs.StubHTTPEndpoint {
	url := constants.WorkbenchEndpointWorkspaceGetSettings
	expectedQuery := ""
	if userID != "" {
		url = constants.WorkbenchEndpointWorkspaceGetUserSettings
		expectedQuery = "user_id=" + userID
	}
	if tenant {
		url = constants.WorkbenchEndpointWorkspaceGetTenantSettings
	}
	return []stubs.StubHTTPEndpoint{
		{
			Request: stubs.StubHTTPRequest{
				Method:          http.MethodGet,
				URL:             postgrest.Endpoint(url),
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

// nolint:unparam
func getSettingsSetEndpoints(
	expectedBody string,
	tenant bool,
	responseStatus int,
	responseBody string,
) []stubs.StubHTTPEndpoint {
	url := constants.WorkbenchEndpointWorkspaceSetUserSetting
	if tenant {
		url = constants.WorkbenchEndpointWorkspaceSetTenantSetting
	}
	return []stubs.StubHTTPEndpoint{
		{
			Request: stubs.StubHTTPRequest{
				Method:          http.MethodPost,
				URL:             postgrest.Endpoint(url),
				ExpectedHeaders: authHeaders,
				ExpectedBody:    expectedBody,
			},
			Response: stubs.StubHTTPResponse{
				Status: responseStatus,
				Body:   responseBody,
			},
		},
	}
}
