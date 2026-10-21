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

// ExpectedDeviceAuthDisplay is the expected display when using the stub
// auth endpoint.
const ExpectedDeviceAuthDisplay = "Open this URL in your browser and complete the authorization:\n" +
	"${URL}/authorization?client_id=discovered_spa_client_id&organization=org-1&" +
	"redirect_uri=verification_uri_complete&response_type=code&scope=openid+profile+offline_access"

// GetStubAuthEndpoints returns default endpoints that mock an authentication
// server. Used when the auth server endpoints themselves are not under test.
func GetStubAuthEndpoints() []StubHTTPEndpoint {
	return []StubHTTPEndpoint{
		{
			Request: StubHTTPRequest{
				Method: http.MethodGet,
				URL:    constants.AuthWellKnownPath,
			},
			Response: StubHTTPResponse{
				Body: fmt.Sprintf(`
					{
						"token_endpoint":%q,
						"device_authorization_endpoint":%q,
						"authorization_endpoint":%q,
						"workbench": {
							"spa_client_id": "discovered_spa_client_id",
							"spa_audience": "discovered_audience",
							"native_client_id": "discovered_native_client_id"
						}
					}
					`, "http://${HOST}/token", "http://${HOST}/device", "http://${HOST}/authorization"), //nolint:gosec
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
				ExpectedBody: `audience=discovered_audience&` +
					`client_id=discovered_native_client_id&` +
					`device_code=device-code&` +
					`grant_type=urn%3Aietf%3Aparams%3Aoauth%3Agrant-type%3Adevice_code`,
			},
			Response: StubHTTPResponse{
				Body:   `{"access_token":"access-token"}`,
				Status: http.StatusOK,
			},
		},
		{
			Request: StubHTTPRequest{
				Method: http.MethodPost,
				URL:    "/device",
				ExpectedBody: `audience=discovered_audience&` +
					`client_id=discovered_native_client_id&` +
					`scope=openid+profile+offline_access+organization%3Aorg-1`,
			},
			Response: StubHTTPResponse{
				Body: `
					{
					  "device_code": "device-code",
					  "user_code": "user-code",
					  "verification_uri": "verification-uri",
					  "verification_uri_complete": "verification_uri_complete",
                      "expires_in": 3600
					}
					`,
				Status: http.StatusOK,
			},
		},
		{
			Request: StubHTTPRequest{
				Method: http.MethodHead,
				URL:    postgrest.Endpoint(constants.WorkbenchEndpointRoot),
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
