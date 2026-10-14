// Copyright © 2026, SAS Institute Inc., Cary, NC, USA.  All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package edit

import (
	"context"

	"github.com/sassoftware/wbctl/internal/cmd/flags"
	"github.com/sassoftware/wbctl/internal/cmd/types"
	"github.com/sassoftware/wbctl/internal/constants"
	"github.com/sassoftware/wbctl/internal/i18n"
	"github.com/sassoftware/wbctl/internal/ui"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
)

// Command returns the edit Command.
func Command(
	ctx context.Context,
	fsys afero.Fs,
	_ui ui.UI,
) *cobra.Command {
	cmd := cobra.Command{
		Use:              constants.EditCommandName,
		Short:            i18n.Format(i18n.CmdEditUsage),
		Long:             i18n.Format(i18n.CmdEditDescription),
		Args:             cobra.NoArgs,
		PersistentPreRun: flags.BindToViper,
	}
	cmd.AddCommand(getSubcommands(ctx, fsys, _ui)...)
	return &cmd
}

func getSubcommands(
	ctx context.Context,
	fsys afero.Fs,
	_ui ui.UI,
) []*cobra.Command {
	return []*cobra.Command{
		{
			Use:  constants.JobCommandName,
			Args: cobra.ExactArgs(1),
			RunE: newEditFunction(ctx, fsys, _ui, constants.JobType),
			Short: i18n.Format(
				i18n.CmdEditTypeUsage,
				i18n.A{constants.I18nArgTypeKey: types.Titler.String(constants.JobType)},
			),
			Long: i18n.Format(
				i18n.CmdEditTypeDescription,
				i18n.A{constants.I18nArgTypeKey: types.Titler.String(constants.JobType)},
			),
		},
		{
			Use:  constants.WorkbenchCommandName,
			Args: cobra.ExactArgs(1),
			RunE: newEditFunction(ctx, fsys, _ui, constants.WorkbenchType),
			Short: i18n.Format(
				i18n.CmdEditTypeUsage,
				i18n.A{constants.I18nArgTypeKey: types.Titler.String(constants.WorkbenchType)},
			),
			Long: i18n.Format(
				i18n.CmdEditTypeDescription,
				i18n.A{constants.I18nArgTypeKey: types.Titler.String(constants.WorkbenchType)},
			),
		},
		{
			Use: constants.VolumeCommandName,
			Aliases: []string{
				constants.WorkspaceCommandName,
			},
			Args: cobra.ExactArgs(1),
			RunE: newEditFunction(ctx, fsys, _ui, constants.VolumeType),
			Short: i18n.Format(
				i18n.CmdEditTypeUsage,
				i18n.A{constants.I18nArgTypeKey: types.Titler.String(constants.VolumeType)},
			),
			Long: i18n.Format(
				i18n.CmdEditTypeDescription,
				i18n.A{constants.I18nArgTypeKey: types.Titler.String(constants.VolumeType)},
			),
		},
		{
			Use:  constants.ImageCommandName,
			Args: cobra.ExactArgs(1),
			RunE: newEditFunction(ctx, fsys, _ui, constants.ImageType),
			Short: i18n.Format(
				i18n.CmdEditTypeUsage,
				i18n.A{constants.I18nArgTypeKey: types.Titler.String(constants.ImageType)},
			),
			Long: i18n.Format(
				i18n.CmdEditTypeDescription,
				i18n.A{constants.I18nArgTypeKey: types.Titler.String(constants.ImageType)},
			),
		},
		{
			Use:  constants.SnapshotCommandName,
			Args: cobra.ExactArgs(1),
			RunE: newEditFunction(ctx, fsys, _ui, constants.SnapshotType),
			Short: i18n.Format(
				i18n.CmdEditTypeUsage,
				i18n.A{constants.I18nArgTypeKey: types.Titler.String(constants.SnapshotType)},
			),
			Long: i18n.Format(
				i18n.CmdEditTypeDescription,
				i18n.A{constants.I18nArgTypeKey: types.Titler.String(constants.SnapshotType)},
			),
		},
	}
}
