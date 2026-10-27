// Copyright © 2026, SAS Institute Inc., Cary, NC, USA.  All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package logout

import (
	"github.com/sassoftware/wbctl/internal/auth"
	cmdtypes "github.com/sassoftware/wbctl/internal/cmd/types"
	"github.com/sassoftware/wbctl/internal/constants"
	"github.com/sassoftware/wbctl/internal/i18n"
	"github.com/sassoftware/wbctl/internal/ui"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// newLogoutFunction creates a command action that authenticates and caches an
// access token. Any required value not supplied via flags is prompted for
// interactively.
func newLogoutFunction(
	fsys afero.Fs,
	_ui ui.UI,
) cmdtypes.ActionFunc {
	return func(
		_ *cobra.Command,
		_ []string,
	) error {
		v := viper.GetViper()
		endpoint := v.GetString(constants.EndpointFlag)
		endpoint = promptForMissing(_ui, endpoint)
		return auth.RemoveAccessToken(
			fsys,
			v,
			endpoint,
		)
	}
}

// promptForMissing interactively fills in any blank credential values before
// the auth call is made.
// nolint:cyclop
func promptForMissing(
	_ui ui.UI,
	endpoint string,
) string {
	if endpoint == "" {
		endpoint = _ui.Ask(i18n.Format(i18n.EndpointPrompt))
	}
	return endpoint
}
