// Copyright © 2026, SAS Institute Inc., Cary, NC, USA.  All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package auth_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sassoftware/wbctl/internal/cmd"
	"github.com/sassoftware/wbctl/internal/cmd/auth"
	"github.com/sassoftware/wbctl/internal/cmd/flags"
	"github.com/sassoftware/wbctl/internal/constants"
	"github.com/sassoftware/wbctl/internal/postgrest"
	"github.com/sassoftware/wbctl/internal/stubs"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
)

const (
	cmdAuthEndpointRequiredErr          = "--endpoint must be specified"
	cmdAuthExpectedCSRFFailureErrFormat = "failed to get CSRF token\n" +
		" Caused by:\n" +
		"\t* HEAD ${URL}/postgrest/ failed\n" +
		"\t* 500 "
	cmdAuthInvalidAuthModeErr = "invalid --auth-mode: \"invalid\" (allowed: auto, password, device)"
)

type authInput struct {
	authMode     string
	clientID     string
	endpoint     string
	organization string
}

type authEndpointsConfig struct {
	tokenStatus         int
	tokenBody           string
	csrfStatus          int
	csrfSetCookieHeader string
	clientID            string
}

type authExpected struct {
	accessToken string
	csrfToken   string
	err         string
	displays    []string
}

type authTestCase struct {
	name          string
	input         authInput
	mockEndpoints []stubs.StubHTTPEndpoint
	expected      authExpected
}

// TestGetAccessTokens verifies GetAccessTokens can pull out the required cli
// flags and pass them to GetAccessToken and GetCSRFToken in the internal/auth
// package.
func TestGetAccessTokens(t *testing.T) {
	tests := []authTestCase{
		{
			name: "endpoint required",
			input: authInput{
				authMode: constants.AuthModeDevice,
			},
			expected: authExpected{
				err: cmdAuthEndpointRequiredErr,
			},
		},
		{
			name: "csrf request failure",
			input: authInput{
				authMode:     constants.AuthModeDevice,
				organization: "org-1",
			},
			mockEndpoints: getAuthEndpoints(
				authEndpointsConfig{
					tokenStatus: http.StatusOK,
					tokenBody:   `{"access_token":"at"}`,
					csrfStatus:  http.StatusInternalServerError,
				}),
			expected: authExpected{
				accessToken: "at",
				err:         cmdAuthExpectedCSRFFailureErrFormat,
				displays: []string{"Open this URL in your browser," +
					" complete the authorization, and verify the code 'user-code':\n" +
					"${URL}/authorization?client_id=discovered_spa_client_id&" +
					"organization=org-1&redirect_uri=verification_uri_complete&" +
					"response_type=code&scope=openid+profile+offline_access"},
			},
		},
		{
			name: "invalid auth mode",
			input: authInput{
				authMode: "invalid",
			},
			expected: authExpected{
				err: cmdAuthInvalidAuthModeErr,
			},
		},
		{
			name: "explicit client id honored",
			input: authInput{
				authMode:     constants.AuthModeDevice,
				clientID:     "explicit_client_id",
				organization: "org-1",
			},
			mockEndpoints: getAuthEndpoints(
				authEndpointsConfig{
					tokenStatus: http.StatusOK,
					tokenBody:   `{"access_token":"at"}`,
					csrfStatus:  http.StatusOK,
					clientID:    "explicit_client_id",
				}),
			expected: authExpected{
				accessToken: "at",
				displays: []string{"Open this URL in your browser, " +
					"complete the authorization, and verify the code 'user-code':\n" +
					"verification_uri_complete"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var srv *httptest.Server
			if tt.input.endpoint == "" && tt.name != "endpoint required" {
				srv, _ = stubs.NewStubHTTPServer(t, tt.mockEndpoints)
				defer srv.Close()
				tt.input.endpoint = srv.URL
			}

			stubUI := &stubs.StubUI{}
			_ = bindFlags(
				t,
				tt,
			)
			accessToken, csrfToken, err := auth.GetAccessTokens(
				context.Background(),
				viper.GetViper(),
				stubUI,
				afero.NewMemMapFs(),
			)

			var replacements map[string]string
			if srv != nil {
				replacements = map[string]string{
					"${PORT}": stubs.Port(t, srv.URL),
					"${HOST}": stubs.Host(t, srv.URL),
					"${URL}":  srv.URL,
				}
			}
			if tt.expected.err != "" {
				expectedErr, actualErr := stubs.Sanitize(tt.expected.err, err.Error(), replacements)
				require.Equal(t, expectedErr, actualErr)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.expected.accessToken, accessToken)
				require.Equal(t, tt.expected.csrfToken, csrfToken)
			}

			expectedDisplays, actualDisplays := stubs.SanitizeSlice(
				tt.expected.displays,
				stubUI.Displays,
				replacements,
			)
			require.Equal(t,
				expectedDisplays,
				actualDisplays,
				"unexpected display messages: %v",
				stubUI.Displays,
			)
			require.Empty(t, stubUI.Warns)
			require.Empty(t, stubUI.Tables)
		})
	}
}

func bindFlags(
	t *testing.T,
	tt authTestCase,
) *cobra.Command {
	t.Helper()
	mockCmd := &cobra.Command{}
	cmd.AddRootFlags(mockCmd)
	args := []string{}
	if tt.input.authMode != "" {
		args = append(args, "--"+constants.AuthModeFlag, tt.input.authMode)
	}
	if tt.input.endpoint != "" {
		args = append(args, "--"+constants.EndpointFlag, tt.input.endpoint)
	}
	if tt.input.clientID != "" {
		args = append(args, "--"+constants.ClientIDFlag, tt.input.clientID)
	}
	if tt.input.organization != "" {
		args = append(args, "--"+constants.OrganizationFlag, tt.input.organization)
	}
	require.NoError(t, mockCmd.ParseFlags(args))
	flags.BindToViper(mockCmd, args)
	return mockCmd
}

func getAuthEndpoints(config authEndpointsConfig) []stubs.StubHTTPEndpoint {
	csrfEndpoint := postgrest.Endpoint(constants.WorkbenchEndpointRoot)
	expectedTokenBody := `audience=discovered_audience&` +
		`client_id=discovered_native_client_id&` +
		`device_code=device-code&` +
		`grant_type=urn%3Aietf%3Aparams%3Aoauth%3Agrant-type%3Adevice_code`
	expectedDeviceBody := `audience=discovered_audience&` +
		`client_id=discovered_native_client_id&` +
		`scope=openid+profile+offline_access+organization%3Aorg-1`
	if config.clientID != "" {
		expectedTokenBody = fmt.Sprintf(
			"audience=discovered_audience&"+
				"client_id=%s&"+
				"device_code=device-code&"+
				"grant_type=urn%%3Aietf%%3Aparams%%3Aoauth%%3Agrant-type%%3Adevice_code",
			config.clientID,
		)
		expectedDeviceBody = fmt.Sprintf(
			"audience=discovered_audience&"+
				"client_id=%s&"+
				"scope=openid+profile+offline_access+organization%%3Aorg-1",
			config.clientID,
		)
	}
	return []stubs.StubHTTPEndpoint{
		{
			Request: stubs.StubHTTPRequest{
				Method: http.MethodGet,
				URL:    constants.AuthWellKnownPath,
			},
			Response: stubs.StubHTTPResponse{
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
				Status:      http.StatusOK,
				BodyOptions: []stubs.StubHTTPBodyOption{stubs.StubHTTPBodyReplaceHOST()},
			},
		},
		{
			Request: stubs.StubHTTPRequest{
				Method:       http.MethodPost,
				URL:          "/token",
				ExpectedBody: expectedTokenBody,
			},
			Response: stubs.StubHTTPResponse{
				Body:   config.tokenBody,
				Status: config.tokenStatus,
			},
		},
		{
			Request: stubs.StubHTTPRequest{
				Method:       http.MethodPost,
				URL:          "/device",
				ExpectedBody: expectedDeviceBody,
			},
			Response: stubs.StubHTTPResponse{
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
			Request: stubs.StubHTTPRequest{
				Method: http.MethodHead,
				URL:    csrfEndpoint,
			},
			Response: stubs.StubHTTPResponse{
				Status:  config.csrfStatus,
				Headers: map[string][]string{constants.SetCookieHeaderKey: {config.csrfSetCookieHeader}},
			},
		},
	}
}
