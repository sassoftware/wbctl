// Copyright © 2026, SAS Institute Inc., Cary, NC, USA.  All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package stubs

import (
	"fmt"
	"net/http"

	"github.com/sassoftware/wbctl/internal/constants"
	"github.com/sassoftware/wbctl/internal/postgrest"
)

const (
	expectedWorkspaceQueryHome        = "controller_id=in.(ws-home)"
	expectedWorkspaceQueryData        = "controller_id=in.(ws-data)"
	expectedWorkspaceResponseBodyHome = `[{"id":"ws-home","controller_id":"ctrl-home","name":"home",` +
		`"properties":{"HOME_DIRECTORY":"true"}}]`
	expectedWorkspaceResponseBodyData = `[{"id":"ws-data","controller_id":"ctrl-data","name":"data",` +
		`"properties":{}}]`
)

var (
	// StubAuthHeaders represents the headers expected from test clients when
	// making requests to Workbench API endpoints.
	StubAuthHeaders = map[string]string{
		constants.AuthorizationHeaderKey:      "Bearer access-token",
		constants.PostgrestCSRFTokenHeaderKey: "csrf-token",
		constants.CookieHeaderKey:             constants.PostgrestCSRFTokenKey + "=csrf-token;",
	}
)

// GetStubAuthEndpoints returns default endpoints that mock an authentication
// server. Used when the auth server endpoints themselves are not under test.
func GetStubAuthEndpoints() []StubHTTPEndpoint {
	csrfEndpoint := postgrest.Endpoint(constants.WorkbenchEndpointRoot)
	return []StubHTTPEndpoint{
		{
			Request: StubHTTPRequest{
				Method: http.MethodGet,
				URL:    constants.AuthWellKnownPath,
			},
			Response: StubHTTPResponse{
				Body:   fmt.Sprintf(`{"token_endpoint":%q}`, "http://${HOST}/token"), //nolint:gosec
				Status: http.StatusOK,
				BodyOptions: []StubHTTPBodyOption{
					StubHTTPBodyReplaceHOST(),
				},
			},
		},
		{
			Request: StubHTTPRequest{
				Method: http.MethodPost,
				URL:    "/token",
				ExpectedBody: `audience=https%3A%2F%2Fworkbench.sas.com%2Fpostgrest&` +
					`client_id=wLDaVQp7dck7ryFttmHpruyIVZIHVhNY&grant_type=password&` +
					`password=pass&scope=openid+profile+offline_access&username=user`,
			},
			Response: StubHTTPResponse{
				Body:   `{"access_token":"access-token"}`,
				Status: http.StatusOK,
			},
		},
		{
			Request: StubHTTPRequest{
				Method: http.MethodHead,
				URL:    csrfEndpoint,
			},
			Response: StubHTTPResponse{
				Status: http.StatusOK,
				Headers: map[string][]string{
					constants.SetCookieHeaderKey: {
						constants.PostgrestCSRFTokenKey + "=csrf-token; Path=/",
					},
				},
			},
		},
	}
}

// StubHTTPPartialGetConfig represents a single get request and response that
// does not require additional config like headers.
type StubHTTPPartialGetConfig struct {
	ExpectedQuery  string
	ResponseStatus int
	ResponseBody   string
}

// StubHTTPPartialPostConfig represents a single post request and response that
// does not require additional config like headers.
type StubHTTPPartialPostConfig struct {
	ExpectedRequestBody string
	ResponseStatus      int
	ResponseBody        string
}

// GetStubWorkspaceEndpoints returns endpoints that represent workspaces. If the
// provided config is empty then two workspaces are represented, one for a home
// dir and one for a regular dir. The home dir must be requested first.
func GetStubWorkspaceEndpoints(config StubHTTPPartialGetConfig) []StubHTTPEndpoint {
	if config.ResponseStatus == 0 {
		return []StubHTTPEndpoint{
			{
				Request: StubHTTPRequest{
					Method:          http.MethodGet,
					URL:             postgrest.Endpoint(constants.WorkbenchEndpointWorkspaces),
					ExpectedHeaders: StubAuthHeaders,
					ExpectedQuery:   expectedWorkspaceQueryHome,
				},
				Response: StubHTTPResponse{
					Status: http.StatusOK,
					Body:   expectedWorkspaceResponseBodyHome,
				},
			},
			{
				Request: StubHTTPRequest{
					Method:          http.MethodGet,
					URL:             postgrest.Endpoint(constants.WorkbenchEndpointWorkspaces),
					ExpectedHeaders: StubAuthHeaders,
					ExpectedQuery:   expectedWorkspaceQueryData,
				},
				Response: StubHTTPResponse{
					Status: http.StatusOK,
					Body:   expectedWorkspaceResponseBodyData,
				},
			},
		}
	}

	return []StubHTTPEndpoint{
		{
			Request: StubHTTPRequest{
				Method:          http.MethodGet,
				URL:             postgrest.Endpoint(constants.WorkbenchEndpointWorkspaces),
				ExpectedHeaders: StubAuthHeaders,
				ExpectedQuery:   config.ExpectedQuery,
			},
			Response: StubHTTPResponse{
				Status: config.ResponseStatus,
				Body:   config.ResponseBody,
			},
		},
	}
}

// GetStubDataplaneEndpoints returns a dataplane endpoint. The first
// request/response is configured with the given config.
func GetStubDataplaneEndpoints(config StubHTTPPartialGetConfig) []StubHTTPEndpoint {
	if config.ResponseStatus == 0 {
		return nil
	}
	return []StubHTTPEndpoint{
		{
			Request: StubHTTPRequest{
				Method:          http.MethodGet,
				URL:             postgrest.Endpoint(constants.WorkbenchEndpointDataplanes),
				ExpectedHeaders: StubAuthHeaders,
				ExpectedQuery:   config.ExpectedQuery,
			},
			Response: StubHTTPResponse{
				Status: config.ResponseStatus,
				Body:   config.ResponseBody,
			},
		},
	}
}

// GetStubApplicationsEndpoints returns an application endpoint. The first
// request/response is configured with the given config.
func GetStubApplicationsEndpoints(config StubHTTPPartialGetConfig) []StubHTTPEndpoint {
	if config.ResponseStatus == 0 {
		return nil
	}
	return []StubHTTPEndpoint{
		{
			Request: StubHTTPRequest{
				Method:          http.MethodGet,
				URL:             postgrest.Endpoint(constants.WorkbenchEndpointDataplaneWorkbenchApplications),
				ExpectedHeaders: StubAuthHeaders,
				ExpectedQuery:   config.ExpectedQuery,
			},
			Response: StubHTTPResponse{
				Status: config.ResponseStatus,
				Body:   config.ResponseBody,
			},
		},
	}
}

// GetStubServicesEndpoints returns a services endpoint. The first
// request/response is configured with the given config.
func GetStubServicesEndpoints(config StubHTTPPartialGetConfig) []StubHTTPEndpoint {
	if config.ResponseStatus == 0 {
		return nil
	}
	return []StubHTTPEndpoint{
		{
			Request: StubHTTPRequest{
				Method:          http.MethodGet,
				URL:             postgrest.Endpoint(constants.WorkbenchEndpointDataplaneWorkbenchServices),
				ExpectedHeaders: StubAuthHeaders,
				ExpectedQuery:   config.ExpectedQuery,
			},
			Response: StubHTTPResponse{
				Status: config.ResponseStatus,
				Body:   config.ResponseBody,
			},
		},
	}
}

// GetStubNodepoolsEndpoints returns a nodepools endpoint. The first
// request/response is configured with the given config.
func GetStubNodepoolsEndpoints(config StubHTTPPartialGetConfig) []StubHTTPEndpoint {
	if config.ResponseStatus == 0 {
		return nil
	}
	return []StubHTTPEndpoint{
		{
			Request: StubHTTPRequest{
				Method:          http.MethodGet,
				URL:             postgrest.Endpoint(constants.WorkbenchEndpointDataplaneNodepools),
				ExpectedHeaders: StubAuthHeaders,
				ExpectedQuery:   config.ExpectedQuery,
			},
			Response: StubHTTPResponse{
				Status: config.ResponseStatus,
				Body:   config.ResponseBody,
			},
		},
	}
}
