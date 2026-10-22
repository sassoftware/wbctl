// Copyright © 2026, SAS Institute Inc., Cary, NC, USA.  All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package version

import (
	"runtime/debug"

	cmdtypes "github.com/sassoftware/wbctl/internal/cmd/types"
	"github.com/sassoftware/wbctl/internal/constants"
	"github.com/sassoftware/wbctl/internal/ui"
	"github.com/sassoftware/wbctl/internal/wbctljson"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type version struct {
	Version string `json:"version"`
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
			Version: buildInfo.Main.Version,
		}
		if constants.JSONOutputValue == viper.GetString(constants.OutputFlag) {
			response, err := wbctljson.MakeMarshaler[version](wbctljson.ViewOptions)(ver)
			if err == nil {
				_ui.Display(string(response))
			}
			return err
		}
		_ui.Display(ver.Version)
		return nil
	}
}
