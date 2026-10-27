// Copyright © 2026, SAS Institute Inc., Cary, NC, USA.  All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package show_test

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/sassoftware/wbctl/internal/cmd"
	"github.com/sassoftware/wbctl/internal/cmd/flags"
	profileshow "github.com/sassoftware/wbctl/internal/cmd/profile/show"
	"github.com/sassoftware/wbctl/internal/constants"
	"github.com/sassoftware/wbctl/internal/osutil"
	"github.com/sassoftware/wbctl/internal/stubs"
	"github.com/spf13/afero"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
)

func TestShowCommand(t *testing.T) {
	tests := []struct {
		name             string
		args             []string
		stubFS           afero.Fs
		noHome           bool
		configContent    string
		expectedErr      string
		expectedDisplays []string
	}{
		{
			name: "show Default",
			configContent: `
{
  "Default": {
    "endpoint": "https://old.example.com",
    "auth-mode": "pkce"
  },
  "verbose": "trace",
  "explicit null": null
}
`,
			expectedDisplays: []string{`auth-mode:       pkce
endpoint:        https://old.example.com
explicit null:   <nil>
verbose:         trace
`},
		},
		{
			name: "show Default as json",
			args: []string{"--output", "json"},
			configContent: `
{
  "Default": {
    "endpoint": "https://old.example.com",
    "auth-mode": "pkce"
  },
  "verbose": "trace"
}
`,
			expectedDisplays: []string{
				`{
    "auth-mode": "pkce",
    "endpoint": "https://old.example.com",
    "verbose": "trace"
}`,
			},
		},
		{
			name: "show other",
			args: []string{"--profile", "other"},
			configContent: `
{
  "Default": {
    "endpoint": "https://old.example.com",
    "auth-mode": "pkce"
  },
  "other": {
    "endpoint": "https://other.example.com"
  },
  "verbose": "trace"
}
`,
			expectedDisplays: []string{"endpoint:   https://other.example.com\nverbose:    trace\n"},
		},
		{
			name:             "invalid file",
			args:             []string{"--profile", "other"},
			configContent:    `invalid`,
			expectedDisplays: []string{"endpoint:   https://other.example.com\nverbose:    trace\n"},
			expectedErr: "Failed to read config file '${HOME}${SEP}.wbctl${SEP}config.json'\n" +
				" Caused by:\n" +
				"\t* invalid character 'i' looking for beginning of value",
		},
		{
			name: "invalid shape",
			configContent: `
{
  "Default": "not an object"
}
`,
			expectedDisplays: []string{"endpoint:   https://other.example.com\nverbose:    trace\n"},
			expectedErr: "Failed to read config file '${HOME}${SEP}.wbctl${SEP}config.json'\n" +
				" Caused by:\n" +
				"\t* invalid configuration file: '${HOME}${SEP}.wbctl${SEP}config.json'",
		},
		{
			name: "stat error",
			stubFS: stubs.StubFS{
				Fs: afero.NewMemMapFs(),
				StatFunc: func(_ afero.Fs, _ string) (os.FileInfo, error) {
					return nil, errors.New("stat error")
				},
			},
			expectedErr: "Failed to read config file '${HOME}${SEP}.wbctl${SEP}config.json'\n" +
				" Caused by:\n" +
				"\t* stat error",
		},
		{
			name:   "no home dir",
			noHome: true,
			expectedErr: "Failed to read config file ''\n" +
				" Caused by:\n" +
				"\t* " + osutil.HomeEnvVarName() + " is not defined",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer viper.Reset()
			home := filepath.Clean("/home/user")
			if tt.noHome {
				home = ""
			}
			t.Setenv(osutil.HomeEnvVar(), home)
			configPath := filepath.Join(home, constants.DefaultConfigFileName)

			fsys := tt.stubFS
			if fsys == nil {
				fsys = afero.NewMemMapFs()
			}
			if tt.configContent != "" {
				require.NoError(t, fsys.MkdirAll(filepath.Join(home, ".wbctl"), 0o700))
				require.NoError(t, afero.WriteFile(fsys, configPath, []byte(tt.configContent), 0o600))
			}

			stubUI := &stubs.StubUI{}
			mockCmd := profileshow.Command(context.Background(), fsys, stubUI)
			if tt.args == nil {
				tt.args = []string{}
			}
			mockCmd.SetArgs(tt.args)
			cmd.AddRootFlags(mockCmd)
			mockCmd.PersistentPreRun = flags.BindToViper
			err := mockCmd.Execute()

			replacements := map[string]string{
				"${HOME}": home,
				"${SEP}":  string(os.PathSeparator),
			}
			if tt.expectedErr != "" {
				expectedErr, actualErr := stubs.Sanitize(tt.expectedErr, err.Error(), replacements)
				require.Equal(t, expectedErr, actualErr)
				return
			}
			require.NoError(t, err)

			require.Equal(t, tt.expectedDisplays, stubUI.Displays)
		})
	}
}
