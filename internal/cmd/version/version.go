// Copyright © 2026, SAS Institute Inc., Cary, NC, USA.  All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package version

import (
	"runtime/debug"

	cmdtypes "github.com/sassoftware/wbctl/internal/cmd/types"
	"github.com/sassoftware/wbctl/internal/constants"
	"github.com/sassoftware/wbctl/internal/debugutil"
	"github.com/sassoftware/wbctl/internal/i18n"
	"github.com/sassoftware/wbctl/internal/ui"
	"github.com/sassoftware/wbctl/internal/wbctljson"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type version struct {
	ClientVersion clientVersion `json:"clientVersion"`
}

type clientVersion struct {
	Version   string `json:"version"`
	GitCommit string `json:"gitCommit"`
	GoVersion string `json:"goVersion"`
	Platform  string `json:"platform"`
}

// newVersionFunction creates a command action that prints the current version
// of wbctl.
func newVersionFunction(
	_ui ui.UI,
) cmdtypes.ActionFunc {
	return func(
		_ *cobra.Command,
		_ []string,
	) error {
		buildInfo := &debug.BuildInfo{}
		actualBuildInfo, ok := debug.ReadBuildInfo()
		if ok {
			buildInfo = actualBuildInfo
		}
		ver := version{
			ClientVersion: clientVersion{
				Version:   buildInfo.Main.Version,
				GitCommit: debugutil.GetSetting(buildInfo.Settings, constants.BuildSettingVCSRevisionKey),
				GoVersion: buildInfo.GoVersion,
				Platform: debugutil.GetSetting(buildInfo.Settings, constants.BuildSettingGOOSKey) + "/" +
					debugutil.GetSetting(buildInfo.Settings, constants.BuildSettingGOARCHKey),
			},
		}
		v := viper.GetViper()
		if constants.JSONOutputValue == v.GetString(constants.OutputFlag) {
			response, err := wbctljson.MakeMarshaler[version](wbctljson.ViewOptions)(ver)
			if err == nil {
				_ui.Display(string(response))
			}
			return err
		}
		_ui.Display(i18n.Format(i18n.VersionDisplay, i18n.A{"version": ver.ClientVersion.Version}))
		return nil
	}
}
