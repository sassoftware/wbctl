// Copyright © 2026, SAS Institute Inc., Cary, NC, USA.  All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Package delete provides the delete command.
// This package name collides with the builtin delete function but context is
// enough to distinguish it.
// nolint: predeclared
package delete

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

// Command returns the delete Command.
func Command(
	ctx context.Context,
	fsys afero.Fs,
	_ui ui.UI,
) *cobra.Command {
	cmd := cobra.Command{
		Use:              constants.DeleteCommandName,
		Short:            i18n.Format(i18n.CmdDeleteUsage),
		Long:             i18n.Format(i18n.CmdDeleteDescription),
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
	cmds := []*cobra.Command{
		{
			Use:     constants.JobsCommandName,
			Aliases: []string{constants.JobCommandName},
			Args:    cobra.ArbitraryArgs,
			RunE:    newDeleteFunction(ctx, fsys, _ui, constants.JobType),
			Short: i18n.Format(
				i18n.CmdDeleteTypeUsage,
				i18n.A{constants.I18nArgTypeKey: types.Titler.String(constants.JobType)},
			),
			Long: i18n.Format(
				i18n.CmdDeleteTypeDescription,
				i18n.A{constants.I18nArgTypeKey: types.Titler.String(constants.JobType)},
			),
		},
		{
			Use:     constants.JobSubmissionsCommandName,
			Aliases: []string{constants.JobSubmissionCommandName},
			Args:    cobra.ArbitraryArgs,
			RunE:    newDeleteFunction(ctx, fsys, _ui, constants.JobSubmissionType),
			Short: i18n.Format(
				i18n.CmdDeleteTypeUsage,
				i18n.A{constants.I18nArgTypeKey: types.Titler.String(constants.JobSubmissionType)},
			),
			Long: i18n.Format(
				i18n.CmdDeleteTypeDescription,
				i18n.A{constants.I18nArgTypeKey: types.Titler.String(constants.JobSubmissionType)},
			),
		},
		{
			Use:     constants.WorkbenchesCommandName,
			Aliases: []string{constants.WorkbenchCommandName},
			Args:    cobra.ArbitraryArgs,
			RunE:    newDeleteFunction(ctx, fsys, _ui, constants.WorkbenchType),
			Short: i18n.Format(
				i18n.CmdDeleteTypeUsage,
				i18n.A{constants.I18nArgTypeKey: types.Titler.String(constants.WorkbenchType)},
			),
			Long: i18n.Format(
				i18n.CmdDeleteTypeDescription,
				i18n.A{constants.I18nArgTypeKey: types.Titler.String(constants.WorkbenchType)},
			),
		},
		{
			Use: constants.VolumesCommandName,
			Aliases: []string{
				constants.VolumeCommandName,
				constants.WorkspaceCommandName,
				constants.WorkspacesCommandName,
			},
			Args: cobra.ArbitraryArgs,
			RunE: newDeleteFunction(ctx, fsys, _ui, constants.VolumeType),
			Short: i18n.Format(
				i18n.CmdDeleteTypeUsage,
				i18n.A{constants.I18nArgTypeKey: types.Titler.String(constants.VolumeType)},
			),
			Long: i18n.Format(
				i18n.CmdDeleteTypeDescription,
				i18n.A{constants.I18nArgTypeKey: types.Titler.String(constants.VolumeType)},
			),
		},
		{
			Use:     constants.ImagesCommandName,
			Aliases: []string{constants.ImageCommandName},
			Args:    cobra.ArbitraryArgs,
			RunE:    newDeleteFunction(ctx, fsys, _ui, constants.ImageType),
			Short: i18n.Format(
				i18n.CmdDeleteTypeUsage,
				i18n.A{constants.I18nArgTypeKey: types.Titler.String(constants.ImageType)},
			),
			Long: i18n.Format(
				i18n.CmdDeleteTypeDescription,
				i18n.A{constants.I18nArgTypeKey: types.Titler.String(constants.ImageType)},
			),
		},
		{
			Use:     constants.SnapshotsCommandName,
			Aliases: []string{constants.SnapshotCommandName},
			Args:    cobra.ArbitraryArgs,
			RunE:    newDeleteFunction(ctx, fsys, _ui, constants.SnapshotType),
			Short: i18n.Format(
				i18n.CmdDeleteTypeUsage,
				i18n.A{constants.I18nArgTypeKey: types.Titler.String(constants.SnapshotType)},
			),
			Long: i18n.Format(
				i18n.CmdDeleteTypeDescription,
				i18n.A{constants.I18nArgTypeKey: types.Titler.String(constants.SnapshotType)},
			),
		},
	}
	return cmds
}
