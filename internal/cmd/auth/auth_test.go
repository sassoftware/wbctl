// Copyright © 2026, SAS Institute Inc., Cary, NC, USA.  All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package auth_test

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/sas-institute-rnd-saas/wbctl/internal/cmd"
	"github.com/sas-institute-rnd-saas/wbctl/internal/cmd/auth"
	"github.com/sas-institute-rnd-saas/wbctl/internal/cmd/flags"
	"github.com/sas-institute-rnd-saas/wbctl/internal/constants"
	"github.com/sas-institute-rnd-saas/wbctl/internal/postgrest"
	"github.com/sas-institute-rnd-saas/wbctl/internal/stubs"
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
	// This is not a hardcoded credential value, but rather the expected error
	// messages when options are missing.
	// nolint:gosec
	cmdAuthPasswordUserRequired = "--auth-mode=password requires --username"
)

type authInput struct {
	authMode string
	clientID string
	endpoint string
	password string
	username string
}

type authEndpointsConfig struct {
	tokenStatus         int
	tokenBody           string
	csrfStatus          int
	csrfSetCookieHeader string
}

type authExpected struct {
	accessToken string
	csrfToken   string
	err         string
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
				authMode: constants.AuthModeAuto,
			},
			expected: authExpected{
				err: cmdAuthEndpointRequiredErr,
			},
		},
		{
			name: "password and csrf success",
			input: authInput{
				authMode: constants.AuthModeAuto,
				clientID: constants.OAuthClientID,
				username: "user",
				password: "pass",
			},
			mockEndpoints: getAuthEndpoints(
				authEndpointsConfig{
					tokenStatus:         http.StatusOK,
					tokenBody:           `{"access_token":"at"}`,
					csrfStatus:          http.StatusOK,
					csrfSetCookieHeader: "postgrest-csrf-token=csrf-token; Path=/",
				}),
			expected: authExpected{
				accessToken: "at",
				csrfToken:   "csrf-token",
			},
		},
		{
			name: "csrf request failure",
			input: authInput{
				authMode: constants.AuthModeAuto,
				clientID: constants.OAuthClientID,
				username: "user",
				password: "pass",
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
			},
		},
		{
			name: "invalid auth mode",
			input: authInput{
				authMode: "invalid",
				clientID: constants.OAuthClientID,
			},
			expected: authExpected{
				err: cmdAuthInvalidAuthModeErr,
			},
		},
		{
			name: "password mode requires username",
			input: authInput{
				authMode: constants.AuthModePassword,
				clientID: constants.OAuthClientID,
				endpoint: "https://example.test",
			},
			expected: authExpected{
				err: cmdAuthPasswordUserRequired,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.input.endpoint == "" && tt.name != "endpoint required" {
				srv, _ := stubs.NewStubHTTPServer(t, tt.mockEndpoints)
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

			if tt.expected.err != "" {
				require.EqualError(t, err, strings.ReplaceAll(tt.expected.err, "${URL}", tt.input.endpoint))
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.expected.accessToken, accessToken)
				require.Equal(t, tt.expected.csrfToken, csrfToken)
			}

			require.Empty(t, stubUI.Displays)
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
	if tt.input.clientID != "" {
		args = append(args, "--"+constants.ClientIDFlag, tt.input.clientID)
	}
	if tt.input.password != "" {
		args = append(args, "--"+constants.PasswordFlag, tt.input.password)
	}
	if tt.input.username != "" {
		args = append(args, "--"+constants.UsernameFlag, tt.input.username)
	}
	if tt.input.endpoint != "" {
		args = append(args, "--"+constants.EndpointFlag, tt.input.endpoint)
	}
	require.NoError(t, mockCmd.ParseFlags(args))
	flags.BindToViper(mockCmd, args)
	return mockCmd
}

func getAuthEndpoints(config authEndpointsConfig) []stubs.StubHTTPEndpoint {
	csrfEndpoint := postgrest.Endpoint(constants.WorkbenchEndpointRoot)
	return []stubs.StubHTTPEndpoint{
		{
			Request: stubs.StubHTTPRequest{
				Method: http.MethodGet,
				URL:    constants.AuthWellKnownPath,
			},
			Response: stubs.StubHTTPResponse{
				Body:        fmt.Sprintf(`{"token_endpoint":%q}`, "http://${HOST}/token"), //nolint:gosec
				Status:      http.StatusOK,
				BodyOptions: []stubs.StubHTTPBodyOption{stubs.StubHTTPBodyReplaceHOST()},
			},
		},
		{
			Request: stubs.StubHTTPRequest{
				Method: http.MethodPost,
				URL:    "/token",
				ExpectedBody: `audience=https%3A%2F%2Fworkbench.sas.com%2Fpostgrest&` +
					`client_id=wLDaVQp7dck7ryFttmHpruyIVZIHVhNY&grant_type=password&` +
					`password=pass&scope=openid+profile+offline_access&username=user`,
			},
			Response: stubs.StubHTTPResponse{
				Body:   config.tokenBody,
				Status: config.tokenStatus,
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
