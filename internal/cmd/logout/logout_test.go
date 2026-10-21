// Copyright © 2026, SAS Institute Inc., Cary, NC, USA.  All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package logout_test

import (
	"context"
	"testing"

	"github.com/sassoftware/wbctl/internal/cmd"
	"github.com/sassoftware/wbctl/internal/cmd/logout"
	"github.com/sassoftware/wbctl/internal/constants"
	"github.com/sassoftware/wbctl/internal/stubs"
	"github.com/spf13/afero"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
)

type logoutInput struct {
	authMode string
	clientID string
	endpoint string
}

type logoutResponse struct {
	stubUI *stubs.StubUI
}

type logoutExpected struct {
	asks            []string
	askForPasswords []string
	err             string
}

type logoutTestCase struct {
	name     string
	input    logoutInput
	response logoutResponse
	expected logoutExpected
}

func TestLogoutCommand(t *testing.T) {
	tests := []logoutTestCase{
		{
			name: "nothing provided by user",
			expected: logoutExpected{
				asks: []string{
					"Endpoint URL",
				},
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

			mockCmd := logout.Command(context.Background(), afero.NewMemMapFs(), stubUI)
			cmd.AddRootFlags(mockCmd)
			mockCmd.SetArgs(getLogoutArgs(tt))
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

func getLogoutArgs(
	tt logoutTestCase,
) []string {
	args := make([]string, 0, 8)
	// Explicitly assign even empty values to test prompting
	args = append(args, "--"+constants.ClientIDFlag, tt.input.clientID)
	args = append(args, "--"+constants.EndpointFlag, tt.input.endpoint)
	args = append(args, "--"+constants.AuthModeFlag, tt.input.authMode)
	return args
}
