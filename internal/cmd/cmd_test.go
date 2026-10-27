// Copyright © 2026, SAS Institute Inc., Cary, NC, USA.  All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package cmd_test

import (
	"context"
	"log/slog"
	"path"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/sassoftware/sas-ggdk/pkg/pointer"
	"github.com/sassoftware/wbctl/internal/cmd"
	"github.com/sassoftware/wbctl/internal/constants"
	"github.com/sassoftware/wbctl/internal/logutil"
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
		constants.VersionCommandName,
		constants.WhoamiCommandName,
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
	expectedLogLevel    slog.Level
}

func TestRootCommand(t *testing.T) {
	tests := []cmdTestCase{
		{
			name:             "no config",
			expectedLogLevel: logutil.LevelSilent,
		},
		{
			name:             "invalid config",
			passConfigOption: true,
			configContent:    `[`,
			expectedBeforeError: "Failed to read config file '${HOME}/.wbctl/config.json'\n" +
				" Caused by:\n" +
				"\t* While parsing config: unexpected end of JSON input",
			expectedLogLevel: logutil.LevelSilent,
		},
		{
			name:             "empty config",
			passConfigOption: true,
			configContent:    `{}`,
			expectedLogLevel: logutil.LevelSilent,
		},
		{
			name:             "full config",
			passConfigOption: true,
			configContent: `{
"endpoint": "http://example.com",
"user": "user",
"password": "password"
}`,
			endpoint:         "http://example.com",
			user:             "user",
			password:         "password",
			expectedLogLevel: logutil.LevelSilent,
		},
		{
			name: "default config (no --config option)",
			configContent: `{
"endpoint": "http://example.com",
"user": "user",
"password": "password"
}`,
			endpoint:         "http://example.com",
			user:             "user",
			password:         "password",
			expectedLogLevel: logutil.LevelSilent,
		},
		{
			name:             "no home dir",
			noHomeDir:        true,
			expectedLogLevel: logutil.LevelSilent,
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
			endpoint:         "http://example.com",
			user:             "user",
			password:         "password",
			expectedLogLevel: logutil.LevelSilent,
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
			endpoint:         "http://example.com",
			user:             "user",
			password:         "password",
			expectedLogLevel: logutil.LevelSilent,
		},
		{
			name:             "verbose level",
			fileExtension:    pointer.Ptr(""),
			passConfigOption: true,
			configContent: `{
"endpoint": "http://example.com",
"user": "user",
"password": "password",
"verbose": "trace"
}`,
			endpoint:         "http://example.com",
			user:             "user",
			password:         "password",
			expectedLogLevel: logutil.LevelTrace,
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
			require.Equal(t, tt.expectedLogLevel, logutil.GetLevel())
		})
	}
}

func TestVerboseFlag(t *testing.T) {
	tests := []struct {
		name          string
		args          []string
		expectedLevel slog.Level
	}{
		{
			name:          "no verbose",
			expectedLevel: logutil.LevelSilent,
		},
		{
			name:          "one verbose",
			args:          []string{"-v"},
			expectedLevel: logutil.LevelInfo,
		},
		{
			name:          "two verbose",
			args:          []string{"-vv"},
			expectedLevel: logutil.LevelDebug,
		},
		{
			name:          "three verbose",
			args:          []string{"-vvv"},
			expectedLevel: logutil.LevelTrace,
		},
		{
			name:          "explicit silent",
			args:          []string{"--verbose=silent"},
			expectedLevel: logutil.LevelSilent,
		},
		{
			name:          "explicit error",
			args:          []string{"--verbose=error"},
			expectedLevel: logutil.LevelError,
		},
		{
			name:          "explicit warn",
			args:          []string{"--verbose=warn"},
			expectedLevel: logutil.LevelWarn,
		},
		{
			name:          "explicit info",
			args:          []string{"--verbose=info"},
			expectedLevel: logutil.LevelInfo,
		},
		{
			name:          "explicit debug",
			args:          []string{"--verbose=debug"},
			expectedLevel: logutil.LevelDebug,
		},
		{
			name:          "explicit trace",
			args:          []string{"--verbose=trace"},
			expectedLevel: logutil.LevelTrace,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer viper.Reset()
			defer func() { _ = logutil.UpdateLevel("false") }()
			fs := afero.NewMemMapFs()
			viper.SetFs(fs)
			stubUI := &stubs.StubUI{}
			app := cmd.CreateApp(context.Background(), fs, stubUI)
			require.NotNil(t, app)
			app.SetArgs(tt.args)
			err := app.ParseFlags(tt.args)
			require.NoError(t, err)
			err = app.PersistentPreRunE(app, app.Flags().Args())
			require.NoError(t, err)
			require.Equal(t, tt.expectedLevel, logutil.GetLevel())
		})
	}
}

const (
	config = `
{
  "verbose": "trace",
  "endpoint": "http://example.com"
}
`
	configWithDefaultProfile = `
{
  "verbose": "trace",
  "Default": {
    "endpoint": "http://example.com"
  }
}
`
	configWithNamedProfile = `
{
  "verbose": "trace",
  "profile-name": {
    "endpoint": "http://example.com"
  }
}
`
)

type profileInput struct {
	configContent string
	args          []string
}

type profileExpected struct {
	endpoint string
	verbose  slog.Level
}

type profileTestCase struct {
	name     string
	input    profileInput
	expected profileExpected
}

func TestProfileFlag(t *testing.T) {
	tests := []profileTestCase{
		{
			name: "config without profile",
			input: profileInput{
				configContent: config,
			},
			expected: profileExpected{
				endpoint: "http://example.com",
				verbose:  logutil.LevelTrace,
			},
		},
		{
			name: "config with default profile",
			input: profileInput{
				configContent: configWithDefaultProfile,
			},
			expected: profileExpected{
				endpoint: "http://example.com",
				verbose:  logutil.LevelTrace,
			},
		},
		{
			name: "config with named profile",
			input: profileInput{
				configContent: configWithNamedProfile,
				args:          []string{"--profile", "profile-name"},
			},
			expected: profileExpected{
				endpoint: "http://example.com",
				verbose:  logutil.LevelTrace,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer viper.Reset()

			fsys := stubs.StubFS{Fs: afero.NewMemMapFs()}
			home, err := afero.TempDir(fsys, "", "")
			require.NoError(t, err)
			t.Setenv(osutil.HomeEnvVar(), home)
			if tt.input.configContent != "" {
				tokenFilePath := filepath.Join(home, constants.DefaultConfigFileName)
				err = afero.WriteFile(fsys, tokenFilePath, []byte(tt.input.configContent), 0644)
				require.NoError(t, err)
			}
			viper.SetFs(fsys)
			stubUI := &stubs.StubUI{}
			app := cmd.CreateApp(context.Background(), fsys, stubUI)
			require.NotNil(t, app)
			app.SetArgs(tt.input.args)
			err = app.ParseFlags(tt.input.args)
			require.NoError(t, err)
			err = app.PersistentPreRunE(app, app.Flags().Args())
			require.NoError(t, err)
			require.Equal(t, tt.expected.endpoint, viper.GetString("endpoint"))
			require.Equal(t, tt.expected.verbose, logutil.GetLevel())
		})
	}
}
