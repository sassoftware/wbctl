// Copyright © 2026, SAS Institute Inc., Cary, NC, USA.  All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package list_test

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/sassoftware/wbctl/internal/cmd"
	"github.com/sassoftware/wbctl/internal/cmd/flags"
	profilelist "github.com/sassoftware/wbctl/internal/cmd/profile/list"
	"github.com/sassoftware/wbctl/internal/constants"
	"github.com/sassoftware/wbctl/internal/osutil"
	"github.com/sassoftware/wbctl/internal/stubs"
	"github.com/spf13/afero"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
)

func TestListCommand(t *testing.T) {
	tests := []struct {
		name             string
		stubFS           afero.Fs
		noHome           bool
		configContent    string
		expectedErr      string
		expectedDisplays []string
	}{
		{
			name: "list profiles",
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
			expectedDisplays: []string{"Available profiles:\nmyprofile, other"},
		},
		{
			name: "stat error",
			stubFS: stubs.StubFS{
				Fs: afero.NewMemMapFs(),
				StatFunc: func(_ afero.Fs, _ string) (os.FileInfo, error) {
					return nil, errors.New("stat error")
				},
			},
			expectedErr: "Failed to read config file '${SEP}home${SEP}user${SEP}.wbctl${SEP}config.json'\n" +
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
			home := "/home/user"
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
			mockCmd := profilelist.Command(context.Background(), fsys, stubUI)
			mockCmd.SetArgs([]string{})
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
