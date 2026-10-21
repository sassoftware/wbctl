// Copyright © 2026, SAS Institute Inc., Cary, NC, USA.  All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package cmd_test

import (
	"context"
	"path"
	"slices"
	"strings"
	"testing"

	"github.com/sassoftware/sas-ggdk/pkg/pointer"
	"github.com/sassoftware/wbctl/internal/cmd"
	"github.com/sassoftware/wbctl/internal/constants"
	"github.com/sassoftware/wbctl/internal/osutil"
	"github.com/sassoftware/wbctl/internal/stubs"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
)

func commandNames(commands []*cobra.Command) []string {
	names := make([]string, 0, len(commands))
	for _, command := range commands {
		names = append(names, command.Name())
	}
	return names
}

func TestCreateRoot(t *testing.T) {
	fs := afero.NewMemMapFs()
	stubUI := &stubs.StubUI{}
	app := cmd.CreateApp(context.Background(), fs, stubUI)
	require.NotNil(t, app)
	expected := []string{
		constants.GetCommandName,
		constants.DescribeCommandName,
		constants.DeleteCommandName,
		constants.CreateCommandName,
		constants.UpdateCommandName,
		constants.EditCommandName,
		constants.APICommandName,
		constants.LoginCommandName,
		constants.LogoutCommandName,
		constants.LogCommandName,
		constants.StopCommandName,
		constants.StartCommandName,
		constants.CancelCommandName,
		constants.RebuildCommandName,
		constants.SettingsCommandName,
	}
	slices.Sort(expected)
	actual := commandNames(app.Commands())
	slices.Sort(actual)
	require.Equal(t, expected, actual)
}

type cmdTestCase struct {
	name                string
	passConfigOption    bool
	configContent       string
	endpoint            string
	user                string
	password            string
	expectedBeforeError string
	noHomeDir           bool
	fileExtension       *string
}

func TestRootCommand(t *testing.T) {
	tests := []cmdTestCase{
		{
			name: "no config",
		},
		{
			name:             "invalid config",
			passConfigOption: true,
			configContent:    `[`,
			expectedBeforeError: "Failed to read config file '${HOME}/.wbctl/config.json'\n" +
				" Caused by:\n" +
				"\t* While parsing config: unexpected end of JSON input",
		},
		{
			name:             "empty config",
			passConfigOption: true,
			configContent:    `{}`,
		},
		{
			name:             "full config",
			passConfigOption: true,
			configContent: `{
"endpoint": "http://example.com",
"user": "user",
"password": "password"
}`,
			endpoint: "http://example.com",
			user:     "user",
			password: "password",
		},
		{
			name: "default config (no --config option)",
			configContent: `{
"endpoint": "http://example.com",
"user": "user",
"password": "password"
}`,
			endpoint: "http://example.com",
			user:     "user",
			password: "password",
		},
		{
			name:      "no home dir",
			noHomeDir: true,
		},
		{
			name:             "unknown extension",
			fileExtension:    pointer.Ptr(".conf"),
			passConfigOption: true,
			configContent: `{
"endpoint": "http://example.com",
"user": "user",
"password": "password"
}`,
			endpoint: "http://example.com",
			user:     "user",
			password: "password",
		},
		{
			name:             "missing extension",
			fileExtension:    pointer.Ptr(""),
			passConfigOption: true,
			configContent: `{
"endpoint": "http://example.com",
"user": "user",
"password": "password"
}`,
			endpoint: "http://example.com",
			user:     "user",
			password: "password",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer viper.Reset()
			fs := afero.NewMemMapFs()
			viper.SetFs(fs)
			defer viper.Reset()
			stubUI := &stubs.StubUI{}
			configPath := ""
			home := ""
			if !tt.noHomeDir {
				home = t.TempDir()
			}
			t.Setenv(osutil.HomeEnvVar(), home)
			if tt.configContent != "" {
				ext := ".json"
				if tt.fileExtension != nil {
					ext = *tt.fileExtension
				}
				configPath = path.Join(home, ".wbctl/config"+ext)
				err := afero.WriteFile(fs, configPath, []byte(tt.configContent), 0o600)
				require.NoError(t, err)
			}
			if !tt.passConfigOption {
				configPath = ""
			}
			t.Setenv("WBCONFIG", configPath)
			app := cmd.CreateApp(context.Background(), fs, stubUI)
			require.NotNil(t, app)
			app.SetArgs([]string{})
			err := app.PersistentPreRunE(app, app.Flags().Args())
			if tt.expectedBeforeError != "" {
				require.EqualError(t, err, strings.ReplaceAll(tt.expectedBeforeError, "${HOME}", home))
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.endpoint, viper.GetString("endpoint"))
				require.Equal(t, tt.user, viper.GetString("user"))
				require.Equal(t, tt.password, viper.GetString("password"))
			}
		})
	}
}
