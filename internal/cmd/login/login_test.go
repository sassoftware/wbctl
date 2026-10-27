// Copyright © 2026, SAS Institute Inc., Cary, NC, USA.  All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package login_test

import (
	"context"
	"testing"

	"github.com/sassoftware/wbctl/internal/cmd"
	"github.com/sassoftware/wbctl/internal/cmd/flags"
	"github.com/sassoftware/wbctl/internal/cmd/login"
	"github.com/sassoftware/wbctl/internal/constants"
	"github.com/sassoftware/wbctl/internal/osutil"
	"github.com/sassoftware/wbctl/internal/stubs"
	"github.com/spf13/afero"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
)

const (
	loginEndpointRequiredErr = "--endpoint must be specified"
)

type loginInput struct {
	authMode string
	clientID string
	endpoint string
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
			name: "nothing provided by user, assumes device",
			expected: loginExpected{
				err: loginEndpointRequiredErr,
				asks: []string{
					"Authentication mode: device or password (default \"device\")",
					"Endpoint URL",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer viper.Reset()
			t.Setenv(osutil.HomeEnvVar(), t.TempDir())

			stubUI := tt.response.stubUI
			if stubUI == nil {
				stubUI = &stubs.StubUI{}
			}

			mockCmd := login.Command(context.Background(), afero.NewMemMapFs(), stubUI)
			cmd.AddRootFlags(mockCmd)
			mockCmd.SetArgs(getLoginArgs(tt))
			mockCmd.PersistentPreRun = flags.BindToViper
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
	return args
}
