// Copyright © 2026, SAS Institute Inc., Cary, NC, USA.  All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package settings

import (
	"context"

	"github.com/sassoftware/wbctl/internal/cmd/flags"
	"github.com/sassoftware/wbctl/internal/constants"
	"github.com/sassoftware/wbctl/internal/i18n"
	"github.com/sassoftware/wbctl/internal/ui"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
)

// Command returns the settings Command.
func Command(
	ctx context.Context,
	fsys afero.Fs,
	_ui ui.UI,
) *cobra.Command {
	cmd := &cobra.Command{
		Use:              constants.SettingsCommandName,
		Aliases:          []string{constants.SettingCommandName},
		Args:             cobra.NoArgs,
		RunE:             newSettingsFunction(ctx, fsys, _ui),
		Short:            i18n.Format(i18n.CmdSettingsUsage),
		Long:             i18n.Format(i18n.CmdSettingsDescription),
		PersistentPreRun: flags.BindToViper,
	}
	flags.AddSetFlag(cmd)
	flags.AddValueFlag(cmd)
	flags.AddTenantFlag(cmd)
	flags.AddUserIDFlag(cmd)
	flags.AddOutputFlag(cmd)
	return cmd
}
