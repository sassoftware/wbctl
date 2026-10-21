// Copyright © 2026, SAS Institute Inc., Cary, NC, USA.  All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package logout

import (
	"context"

	"github.com/sassoftware/wbctl/internal/cmd/flags"
	"github.com/sassoftware/wbctl/internal/constants"
	"github.com/sassoftware/wbctl/internal/i18n"
	"github.com/sassoftware/wbctl/internal/ui"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
)

// Command returns the login Command.
func Command(
	_ context.Context,
	fsys afero.Fs,
	_ui ui.UI,
) *cobra.Command {
	cmd := &cobra.Command{
		Use:    constants.LogoutCommandName,
		Short:  i18n.Format(i18n.CmdLogoutUsage),
		Long:   i18n.Format(i18n.CmdLogoutDescription),
		Args:   cobra.NoArgs,
		RunE:   newLogoutFunction(fsys, _ui),
		PreRun: flags.BindToViper,
	}
	return cmd
}
