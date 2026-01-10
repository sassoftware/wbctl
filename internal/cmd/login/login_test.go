// Copyright © 2026, SAS Institute Inc., Cary, NC, USA.  All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package login_test

import (
	"context"
	"testing"

	"github.com/sas-institute-rnd-saas/wbctl/internal/cmd"
	"github.com/sas-institute-rnd-saas/wbctl/internal/cmd/login"
	"github.com/sas-institute-rnd-saas/wbctl/internal/constants"
	"github.com/sas-institute-rnd-saas/wbctl/internal/stubs"
	"github.com/spf13/afero"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
)

const (
	loginEndpointRequiredErr     = "--endpoint must be specified"
	loginPasswordGrantFailureErr = "failed to get username/password token\n" +
		" Caused by:\n" +
		"\t* GET ignored-endpoint/auth0Admin/internal/.well-known/openid-configuration failed\n" +
		"\t* Get \"ignored-endpoint/auth0Admin/internal/.well-known/openid-configuration\": unsupported protocol scheme \"\""
	loginClientCredentialsGrantFailureErr = "failed to get client credentials token\n" +
		" Caused by:\n" +
		"\t* GET ignored-endpoint/auth0Admin/internal/.well-known/openid-configuration failed\n" +
		"\t* Get \"ignored-endpoint/auth0Admin/internal/.well-known/openid-configuration\": unsupported protocol scheme \"\""
)

type loginInput struct {
	authMode string
	clientID string
	endpoint string
	username string
	password string
}

type loginResponse struct {
	stubUI *stubs.StubUI
}

type loginExpected struct {
	asks            []string
	askForPasswords []string
	err             string
}

type loginTestCase struct {
	name     string
	input    loginInput
	response loginResponse
	expected loginExpected
}

func TestLoginCommand(t *testing.T) {
	tests := []loginTestCase{
		{
			name: "nothing provided by user, assumes pkce",
			expected: loginExpected{
				err: loginEndpointRequiredErr,
				asks: []string{
					"Authentication mode: pkce, client_id, password, or device (default \"pkce\")",
					"Client ID",
					"Endpoint URL",
				},
			},
		},
		{
			name: "user provides auth-mode password, endpoint, and username, prompts for password",
			input: loginInput{
				authMode: constants.AuthModePassword,
				clientID: constants.OAuthClientID,
				endpoint: "ignored-endpoint",
				username: "testuser",
			},
			expected: loginExpected{
				askForPasswords: []string{"Password"},
				err:             loginPasswordGrantFailureErr,
			},
		},
		{
			name: "user provides auth-mode password and endpoint, prompts for username and password",
			input: loginInput{
				authMode: constants.AuthModePassword,
				clientID: constants.OAuthClientID,
				endpoint: "ignored-endpoint",
			},
			response: loginResponse{
				stubUI: &stubs.StubUI{
					AskResponses:            []string{"testuser"},
					AskForPasswordResponses: []string{"testpassword"},
				},
			},
			expected: loginExpected{
				asks:            []string{"Username"},
				askForPasswords: []string{"Password"},
				err:             loginPasswordGrantFailureErr,
			},
		},
		{
			name: "user provides auth-mode client-credentials, prompts for client-id and client-secret",
			input: loginInput{
				authMode: constants.AuthModeClientCredentials,
				endpoint: "ignored-endpoint",
			},
			response: loginResponse{
				stubUI: &stubs.StubUI{
					AskResponses:            []string{"testclientid"},
					AskForPasswordResponses: []string{"testclientsecret"},
				},
			},
			expected: loginExpected{
				asks:            []string{"Client ID"},
				askForPasswords: []string{"Client Secret"},
				err:             loginClientCredentialsGrantFailureErr,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer viper.Reset()
			t.Setenv("HOME", t.TempDir())

			stubUI := tt.response.stubUI
			if stubUI == nil {
				stubUI = &stubs.StubUI{}
			}

			mockCmd := login.Command(context.Background(), afero.NewMemMapFs(), stubUI)
			cmd.AddRootFlags(mockCmd)
			mockCmd.SetArgs(getLoginArgs(tt))
			err := mockCmd.Execute()
			if tt.expected.err != "" {
				require.EqualError(t, err, tt.expected.err)
			} else {
				require.NoError(t, err)
			}
			require.ElementsMatch(t, tt.expected.asks, stubUI.Asks)
			require.ElementsMatch(t, tt.expected.askForPasswords, stubUI.AskForPasswords)
		})
	}
}

func getLoginArgs(
	tt loginTestCase,
) []string {
	args := make([]string, 0, 8)
	// Explicitly assign even empty values to test prompting
	args = append(args, "--"+constants.ClientIDFlag, tt.input.clientID)
	args = append(args, "--"+constants.EndpointFlag, tt.input.endpoint)
	args = append(args, "--"+constants.AuthModeFlag, tt.input.authMode)
	args = append(args, "--"+constants.UsernameFlag, tt.input.username)
	args = append(args, "--"+constants.PasswordFlag, tt.input.password)
	return args
}
