// Copyright © 2026, SAS Institute Inc., Cary, NC, USA.  All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package init_test

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/sassoftware/wbctl/internal/cmd"
	"github.com/sassoftware/wbctl/internal/cmd/flags"
	profileinit "github.com/sassoftware/wbctl/internal/cmd/profile/init"
	"github.com/sassoftware/wbctl/internal/constants"
	"github.com/sassoftware/wbctl/internal/osutil"
	"github.com/sassoftware/wbctl/internal/stubs"
	"github.com/spf13/afero"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
)

func TestInitCommand(t *testing.T) {
	tests := []struct {
		name           string
		args           []string
		stubFS         afero.Fs
		noHome         bool
		askResponses   []string
		configContent  string
		expectedErr    string
		expectedConfig map[string]any
	}{
		{
			name:         "prompts for profile, endpoint, and organization",
			askResponses: []string{"myprofile", "https://example.com", "myorg"},
			args:         []string{},
			expectedConfig: map[string]any{
				"myprofile": map[string]any{
					"endpoint":     "https://example.com",
					"organization": "myorg",
				},
			},
		},
		{
			name: "flags supply values, no prompting needed",
			args: []string{
				"--" + constants.ProfileFlag, "myprofile",
				"--" + constants.EndpointFlag, "https://example.com",
				"--" + constants.OrganizationFlag, "myorg",
			},
			expectedConfig: map[string]any{
				"myprofile": map[string]any{
					"endpoint":     "https://example.com",
					"organization": "myorg",
				},
			},
		},
		{
			name: "organization is optional",
			args: []string{
				"--" + constants.ProfileFlag, "myprofile",
				"--" + constants.EndpointFlag, "https://example.com",
			},
			askResponses: []string{""},
			expectedConfig: map[string]any{
				"myprofile": map[string]any{
					"endpoint": "https://example.com",
				},
			},
		},
		{
			name: "missing profile name is an error",
			args: []string{
				"--" + constants.ProfileFlag, "",
				"--" + constants.EndpointFlag, "https://example.com",
			},
			askResponses: []string{""},
			expectedErr:  "--profile is required",
		},
		{
			name: "missing endpoint is an error",
			args: []string{
				"--" + constants.ProfileFlag, "myprofile",
			},
			askResponses: []string{""},
			expectedErr:  "--endpoint is required",
		},
		{
			name: "explicitly set flags beyond endpoint and organization are saved",
			args: []string{
				"--" + constants.ProfileFlag, "myprofile",
				"--" + constants.EndpointFlag, "https://example.com",
				"--" + constants.AuthModeFlag, "password",
				"--" + constants.VerboseFlag,
			},
			expectedConfig: map[string]any{
				"myprofile": map[string]any{
					"endpoint":  "https://example.com",
					"auth-mode": "password",
					"verbose":   "INFO",
				},
			},
		},
		{
			name: "replaces existing profile, leaves other profiles alone",
			args: []string{
				"--" + constants.ProfileFlag, "myprofile",
				"--" + constants.EndpointFlag, "https://new.example.com",
			},
			configContent: `
{
  "myprofile": {
    "endpoint": "https://old.example.com",
    "auth-mode": "pkce"
  },
  "other": {
    "endpoint": "https://other.example.com"
  }
}
`,
			expectedConfig: map[string]any{
				"myprofile": map[string]any{
					"endpoint": "https://new.example.com",
				},
				"other": map[string]any{
					"endpoint": "https://other.example.com",
				},
			},
		},
		{
			name: "stat error",
			stubFS: stubs.StubFS{
				Fs: afero.NewMemMapFs(),
				StatFunc: func(_ afero.Fs, _ string) (os.FileInfo, error) {
					return nil, errors.New("stat error")
				},
			},
			args: []string{
				"--" + constants.ProfileFlag, "myprofile",
				"--" + constants.EndpointFlag, "https://example.com",
				"--" + constants.OrganizationFlag, "myorg",
			},
			expectedErr: "Failed to read config file '${HOME}${SEP}.wbctl${SEP}config.json'\n" +
				" Caused by:\n" +
				"\t* stat error",
		},
		{
			name: "mkdir all error",
			stubFS: stubs.StubFS{
				Fs: afero.NewMemMapFs(),
				MkdirAllFunc: func(_ afero.Fs, _ string, _ os.FileMode) error {
					return errors.New("mkdirall error")
				},
			},
			args: []string{
				"--" + constants.ProfileFlag, "myprofile",
				"--" + constants.EndpointFlag, "https://example.com",
				"--" + constants.OrganizationFlag, "myorg",
			},
			expectedErr: "Failed to write config file '${HOME}${SEP}.wbctl${SEP}config.json'\n" +
				" Caused by:\n" +
				"\t* mkdirall error",
		},
		{
			name:   "no home dir",
			noHome: true,
			args: []string{
				"--" + constants.ProfileFlag, "myprofile",
				"--" + constants.EndpointFlag, "https://example.com",
				"--" + constants.OrganizationFlag, "myorg",
			},
			expectedErr: "Failed to write config file ''\n" +
				" Caused by:\n" +
				"\t* " + osutil.HomeEnvVarName() + " is not defined",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer viper.Reset()
			fsys := tt.stubFS
			if fsys == nil {
				fsys = afero.NewMemMapFs()
			}
			home, err := afero.TempDir(fsys, "", "")
			require.NoError(t, err)
			if tt.noHome {
				home = ""
			}
			t.Setenv(osutil.HomeEnvVar(), home)
			configPath := filepath.Join(home, constants.DefaultConfigFileName)

			if tt.configContent != "" {
				require.NoError(t, fsys.MkdirAll(filepath.Join(home, ".wbctl"), 0o700))
				require.NoError(t, afero.WriteFile(fsys, configPath, []byte(tt.configContent), 0o600))
			}

			stubUI := &stubs.StubUI{AskResponses: tt.askResponses}
			mockCmd := profileinit.Command(context.Background(), fsys, stubUI)
			cmd.AddRootFlags(mockCmd)
			mockCmd.SetArgs(tt.args)
			mockCmd.PersistentPreRun = flags.BindToViper
			err = mockCmd.Execute()

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

			content, err := afero.ReadFile(fsys, configPath)
			require.NoError(t, err)
			var doc map[string]any
			require.NoError(t, json.Unmarshal(content, &doc))
			require.Equal(t, tt.expectedConfig, doc)
		})
	}
}
