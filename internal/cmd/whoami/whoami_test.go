// Copyright © 2026, SAS Institute Inc., Cary, NC, USA.  All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package whoami_test

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/sassoftware/wbctl/internal/cmd"
	"github.com/sassoftware/wbctl/internal/cmd/flags"
	"github.com/sassoftware/wbctl/internal/cmd/whoami"
	"github.com/sassoftware/wbctl/internal/constants"
	"github.com/sassoftware/wbctl/internal/osutil"
	"github.com/sassoftware/wbctl/internal/stubs"
	"github.com/spf13/afero"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
)

const (
	fakeCache = `
{
  "Default.https://example.com": {
    "access_token": "access-token",
    "expires_at": "2026-10-26T21:02:17.516241Z",
    "id_token": "header.ewogICJhdWQiOiAibXktYXVkaWVuY2UiLAogICJleHAiOiAxNzkzMDg0M` +
		`jM3LAogICJpYXQiOiAxNzkzMDQ4MjM3LAogICJpc3MiOiAiaXNzdWVyIiwKICAibmFtZSI6ICJ1c2Vy` +
		`LW5hbWUiLAogICJuaWNrbmFtZSI6ICJuaWNrbmFtZSIsCiAgInBpY3R1cmUiOiAicGljdHVyZSB1cmwi` +
		`LAogICJzYXMiOiB7CiAgICAiYXV0aDBfb3JnX2lkIjogImF1dGgwLW9yZy1pZCIsCiAgICAiYXV0aDBf` +
		`dXNlcl9pZCI6ICJhdXRoMC11c2VyLWlkIiwKICAgICJvcmdfaWQiOiAib3JnLWlkIiwKICAgICJvcmdf` +
		`bmFtZSI6ICJvcmctbmFtZSIsCiAgICAicm9sZXMiOiBbCiAgICAgICJvcmdfYWRtaW4iLAogICAgICAi` +
		`d29ya2JlbmNoX2FkbWluIgogICAgXSwKICAgICJ1c2VyX2lkIjogInVzZXItaWQiCiAgfSwKICAic3Vi` +
		`IjogInN1YiIsCiAgInVwZGF0ZWRfYXQiOiAiMjAyNi0xMC0yM1QxMzoyNjowMC44ODJaIgp9Cg==.sig` +
		`nature",
    "refresh_token": "refresh-token",
    "token_type": "Bearer"
  }
}
`
	invalidCache = `
{
  "Default.https://example.com": {
    "access_token": "access-token",
    "expires_at": "2026-10-26T21:02:17.516241Z",
    "id_token": "header.aW52YWxpZC1wYXlsb2FkCg==.signature",
    "refresh_token": "refresh-token",
    "token_type": "Bearer"
  }
}
`
	expectedIDDisplay = `User:           user-name (user-id)
Organization:   org-name (org-id)
Roles:          org_admin, workbench_admin
Issued:         2026-10-26T20:57:17Z
Expires:        2026-10-27T06:57:17Z (EXPIRED)
`
	expectedIDDisplayJSON = `{
    "aud": "my-audience",
    "exp": 1793084237,
    "iat": 1793048237,
    "iss": "issuer",
    "name": "user-name",
    "nickname": "nickname",
    "picture": "picture url",
    "sas": {
        "auth0_org_id": "auth0-org-id",
        "auth0_user_id": "auth0-user-id",
        "org_id": "org-id",
        "org_name": "org-name",
        "roles": [
            "org_admin",
            "workbench_admin"
        ],
        "user_id": "user-id"
    },
    "sub": "sub",
    "updated_at": "2026-10-23T13:26:00.882Z"
}`
)

type whoamiInput struct {
	cacheContent string
	jsonOutput   bool
	endpoint     string
}

type whoamiResponse struct {
	stubUI *stubs.StubUI
}

type whoamiExpected struct {
	displays []string
	err      string
}

type whoamiTestCase struct {
	name     string
	input    whoamiInput
	response whoamiResponse
	expected whoamiExpected
}

func TestWhoamiCommand(t *testing.T) {
	tests := []whoamiTestCase{
		{
			name: "no cached token",
			expected: whoamiExpected{
				displays: []string{"Not authenticated"},
			},
		},
		{
			name: "no cached token - json",
			input: whoamiInput{
				jsonOutput: true,
			},
			expected: whoamiExpected{
				displays: []string{"{}"},
			},
		},
		{
			name: "fake cached token",
			input: whoamiInput{
				endpoint:     "https://example.com",
				cacheContent: fakeCache,
			},
			expected: whoamiExpected{
				displays: []string{expectedIDDisplay},
			},
		},
		{
			name: "fake cached token - json",
			input: whoamiInput{
				jsonOutput:   true,
				endpoint:     "https://example.com",
				cacheContent: fakeCache,
			},
			expected: whoamiExpected{
				displays: []string{expectedIDDisplayJSON},
			},
		},
		{
			name: "invalid cached token",
			input: whoamiInput{
				endpoint:     "https://example.com",
				cacheContent: invalidCache,
			},
			expected: whoamiExpected{
				err: `failed to marshal jwt payload
 Caused by:
	* invalid character 'i' looking for beginning of value`},
		},
		{
			name: "invaid cached token - json",
			input: whoamiInput{
				jsonOutput:   true,
				endpoint:     "https://example.com",
				cacheContent: invalidCache,
			},
			expected: whoamiExpected{
				err: `failed to marshal jwt payload
 Caused by:
	* invalid character 'i' looking for beginning of value`,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer viper.Reset()

			stubUI := tt.response.stubUI
			if stubUI == nil {
				stubUI = &stubs.StubUI{}
			}

			fsys := stubs.StubFS{Fs: afero.NewMemMapFs()}
			home, err := afero.TempDir(fsys, "", "")
			require.NoError(t, err)
			t.Setenv(osutil.HomeEnvVar(), home)
			if tt.input.cacheContent != "" {
				tokenFilePath := filepath.Join(home, constants.DefaultTokenFileName)
				err = afero.WriteFile(fsys, tokenFilePath, []byte(tt.input.cacheContent), 0644)
				require.NoError(t, err)
			}
			mockCmd := whoami.Command(context.Background(), fsys, stubUI)
			cmd.AddRootFlags(mockCmd)
			mockCmd.SetArgs(getWhoamiArgs(tt))
			mockCmd.PersistentPreRun = flags.BindToViper
			err = mockCmd.Execute()
			if tt.expected.err != "" {
				require.EqualError(t, err, tt.expected.err)
			} else {
				require.NoError(t, err)
			}
			require.ElementsMatch(t, tt.expected.displays, stubUI.Displays)
		})
	}
}

func getWhoamiArgs(
	tt whoamiTestCase,
) []string {
	args := make([]string, 0, 8)
	// Explicitly assign even empty values to test prompting
	if tt.input.jsonOutput {
		args = append(args, "--"+constants.OutputFlag, constants.JSONOutputValue)
	}
	if tt.input.endpoint != "" {
		args = append(args, "--"+constants.EndpointFlag, tt.input.endpoint)
	}
	return args
}
