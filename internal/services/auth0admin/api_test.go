// Copyright © 2026, SAS Institute Inc., Cary, NC, USA.  All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package auth0admin_test

import (
	"context"
	"net/http"
	"strings"
	"testing"

	auth0adminhelper "github.com/sas-institute-rnd-saas/wbctl/internal/auth0admin"
	"github.com/sas-institute-rnd-saas/wbctl/internal/constants"
	"github.com/sas-institute-rnd-saas/wbctl/internal/services/auth0admin"
	"github.com/sas-institute-rnd-saas/wbctl/internal/stubs"
	"github.com/stretchr/testify/require"
)

const (
	getAPISuccessResponseJSON = `
{
  "version": 1,
  "links": [
    {
      "method": "GET",
      "rel": "roles",
      "href": "/auth0Admin/roles",
      "uri": "/auth0Admin/roles",
      "type": "application/vnd.sas.collection"
    },
    {
      "method": "GET",
      "rel": "users",
      "href": "/auth0Admin/users",
      "uri": "/auth0Admin/users",
      "type": "application/vnd.sas.collection"
    },
    {
      "method": "POST",
      "rel": "importUsers",
      "href": "/auth0Admin/users",
      "uri": "/auth0Admin/users",
      "type": "application/vnd.sas.collection",
      "responseType": "application/vnd.sas.collection"
    },
    {
      "method": "GET",
      "rel": "connections",
      "href": "/auth0Admin/connections",
      "uri": "/auth0Admin/connections",
      "type": "application/vnd.sas.auth0admin.connection+json"
    },
    {
      "method": "POST",
      "rel": "createConnectionTicket",
      "href": "/auth0Admin/connections",
      "uri": "/auth0Admin/connections",
      "type": "application/vnd.sas.auth0admin.connection.ticket+json",
      "responseType": "text/plain"
    },
    {
      "method": "GET",
      "rel": "settings",
      "href": "/auth0Admin/settings",
      "uri": "/auth0Admin/settings",
      "type": "application/vnd.sas.auth0admin.settings+json"
    },
    {
      "method": "PATCH",
      "rel": "patchSettings",
      "href": "/auth0Admin/settings",
      "uri": "/auth0Admin/settings",
      "type": "application/vnd.sas.auth0admin.settings+json",
      "responseType": "application/vnd.sas.auth0admin.settings+json"
    }
  ]
}
`
)

var (
	expectedAuth0AdminAPIResult = auth0admin.API{
		Version: 1,
		Links: []auth0admin.Link{
			{
				Method:    "GET",
				Rel:       "roles",
				Href:      "/auth0Admin/roles",
				URI:       "/auth0Admin/roles",
				MediaType: "application/vnd.sas.collection",
			},
			{
				Method:    "GET",
				Rel:       "users",
				Href:      "/auth0Admin/users",
				URI:       "/auth0Admin/users",
				MediaType: "application/vnd.sas.collection",
			},
			{
				Method:       "POST",
				Rel:          "importUsers",
				Href:         "/auth0Admin/users",
				URI:          "/auth0Admin/users",
				MediaType:    "application/vnd.sas.collection",
				ResponseType: "application/vnd.sas.collection",
			},
			{
				Method:    "GET",
				Rel:       "connections",
				Href:      "/auth0Admin/connections",
				URI:       "/auth0Admin/connections",
				MediaType: "application/vnd.sas.auth0admin.connection+json",
			},
			{
				Method:       "POST",
				Rel:          "createConnectionTicket",
				Href:         "/auth0Admin/connections",
				URI:          "/auth0Admin/connections",
				MediaType:    "application/vnd.sas.auth0admin.connection.ticket+json",
				ResponseType: "text/plain",
			},
			{
				Method:    "GET",
				Rel:       "settings",
				Href:      "/auth0Admin/settings",
				URI:       "/auth0Admin/settings",
				MediaType: "application/vnd.sas.auth0admin.settings+json",
			},
			{
				Method:       "PATCH",
				Rel:          "patchSettings",
				Href:         "/auth0Admin/settings",
				URI:          "/auth0Admin/settings",
				MediaType:    "application/vnd.sas.auth0admin.settings+json",
				ResponseType: "application/vnd.sas.auth0admin.settings+json",
			},
		},
	}
)

type getExpected[T any] struct {
	err    string
	result T
}

type getTestCase[T any] struct {
	name          string
	mockEndpoints []stubs.StubHTTPEndpoint
	expected      getExpected[T]
}

func TestGetAPI(t *testing.T) {
	tests := []getTestCase[auth0admin.API]{
		{
			name:          "get api success",
			mockEndpoints: getAPIGetEndpoints(http.StatusOK, getAPISuccessResponseJSON),
			expected: getExpected[auth0admin.API]{
				result: expectedAuth0AdminAPIResult,
			},
		},
		{
			name:          "get api invalid json response",
			mockEndpoints: getAPIGetEndpoints(http.StatusOK, `invalid json`),
			expected: getExpected[auth0admin.API]{
				err: "GET ${URL}/auth0Admin/ failed\n" +
					" Caused by:\n" +
					"\t* invalid character 'i' looking for beginning of value",
			},
		},
		{
			name:          "get api non-200 from server",
			mockEndpoints: getAPIGetEndpoints(http.StatusUnauthorized, "unauthorized"),
			expected: getExpected[auth0admin.API]{
				err: "GET ${URL}/auth0Admin/ failed\n" +
					" Caused by:\n" +
					"\t* 401 unauthorized",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv, _ := stubs.NewStubHTTPServer(t, tt.mockEndpoints)
			defer srv.Close()
			svc := auth0admin.NewAuth0AdminService(srv.URL, "access-token")

			actual, err := auth0admin.GetAPI(context.Background(), svc)
			if tt.expected.err != "" {
				require.EqualError(t, err, strings.ReplaceAll(tt.expected.err, "${URL}", srv.URL))
				return
			}

			require.NoError(t, err)
			require.NotNil(t, actual)
			require.Equal(t, tt.expected.result, actual)
		})
	}
}

// Helper endpoints.
var (
	authHeaders = map[string]string{
		constants.AuthorizationHeaderKey: "Bearer access-token",
	}
)

func getAPIGetEndpoints(
	responseStatus int,
	responseBody string,
) []stubs.StubHTTPEndpoint {
	return []stubs.StubHTTPEndpoint{
		{
			Request: stubs.StubHTTPRequest{
				Method:          http.MethodGet,
				URL:             auth0adminhelper.Endpoint(constants.WorkbenchEndpointRoot),
				ExpectedHeaders: authHeaders,
			},
			Response: stubs.StubHTTPResponse{
				Status: responseStatus,
				Body:   responseBody,
			},
		},
	}
}
