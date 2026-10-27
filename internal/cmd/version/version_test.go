// Copyright © 2026, SAS Institute Inc., Cary, NC, USA.  All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package version_test

import (
	"context"
	"fmt"
	"runtime/debug"
	"testing"

	"github.com/sassoftware/wbctl/internal/cmd"
	"github.com/sassoftware/wbctl/internal/cmd/flags"
	"github.com/sassoftware/wbctl/internal/cmd/version"
	"github.com/sassoftware/wbctl/internal/constants"
	"github.com/sassoftware/wbctl/internal/debugutil"
	"github.com/sassoftware/wbctl/internal/stubs"
	"github.com/spf13/afero"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
)

type versionInput struct {
	outputFormat string
}

type versionResponse struct {
	stubUI *stubs.StubUI
}

type versionExpected struct {
	displays []string
}

type versionTestCase struct {
	name     string
	input    versionInput
	response versionResponse
	expected versionExpected
}

const versionJSONOutput = `{
    "clientVersion": {
        "version": "%s",
        "gitCommit": "%s",
        "goVersion": "%s",
        "platform": "%s"
    }
}`

func TestVersionCommand(t *testing.T) {
	buildInfo, ok := debug.ReadBuildInfo()
	require.True(t, ok)
	tests := []versionTestCase{
		{
			name: "basic",
			expected: versionExpected{
				displays: []string{buildInfo.Main.Version},
			},
		},
		{
			name: "json",
			input: versionInput{
				outputFormat: constants.JSONOutputValue,
			},
			expected: versionExpected{
				displays: []string{fmt.Sprintf(
					versionJSONOutput,
					buildInfo.Main.Version,
					debugutil.GetSetting(buildInfo.Settings, constants.BuildSettingVCSRevisionKey),
					buildInfo.GoVersion,
					debugutil.GetSetting(buildInfo.Settings, constants.BuildSettingGOOSKey)+"/"+
						debugutil.GetSetting(buildInfo.Settings, constants.BuildSettingGOARCHKey),
				)},
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
			mockCmd := version.Command(context.Background(), afero.NewMemMapFs(), stubUI)
			cmd.AddRootFlags(mockCmd)
			mockCmd.SetArgs(getVersionArgs(tt))
			mockCmd.PersistentPreRun = flags.BindToViper
			err := mockCmd.Execute()
			require.NoError(t, err)
			require.ElementsMatch(t, tt.expected.displays, stubUI.Displays)
		})
	}
}

func getVersionArgs(
	tt versionTestCase,
) []string {
	args := make([]string, 0, 1)
	if tt.input.outputFormat != "" {
		args = append(args, "--"+constants.OutputFlag, tt.input.outputFormat)
	}
	return args
}
