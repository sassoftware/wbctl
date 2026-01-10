// Copyright © 2026, SAS Institute Inc., Cary, NC, USA.  All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package describe

import (
	"context"

	"github.com/sas-institute-rnd-saas/wbctl/internal/cmd/flags"
	"github.com/sas-institute-rnd-saas/wbctl/internal/cmd/types"
	"github.com/sas-institute-rnd-saas/wbctl/internal/constants"
	"github.com/sas-institute-rnd-saas/wbctl/internal/i18n"
	"github.com/sas-institute-rnd-saas/wbctl/internal/ui"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
)

// Command returns the describe Command.
func Command(
	ctx context.Context,
	fsys afero.Fs,
	_ui ui.UI,
) *cobra.Command {
	cmd := cobra.Command{
		Use:              constants.DescribeCommandName,
		Short:            i18n.Format(i18n.CmdDescribeUsage),
		Long:             i18n.Format(i18n.CmdDescribeDescription),
		Args:             cobra.NoArgs,
		PersistentPreRun: flags.BindToViper,
	}
	cmd.AddCommand(getSubcommands(ctx, fsys, _ui)...)
	return &cmd
}

// nolint:funlen
func getSubcommands(
	ctx context.Context,
	fsys afero.Fs,
	_ui ui.UI,
) []*cobra.Command {
	cmds := []*cobra.Command{
		{
			Use:     constants.DataplanesCommandName,
			Aliases: []string{constants.DataplaneCommandName},
			Args:    cobra.ArbitraryArgs,
			RunE:    newDescribeFunction(ctx, fsys, _ui, constants.DataplaneType),
			Short: i18n.Format(
				i18n.CmdDescribeTypeUsage,
				i18n.A{constants.I18nArgTypeKey: types.Titler.String(constants.DataplaneType)},
			),
			Long: i18n.Format(
				i18n.CmdDescribeTypeDescription,
				i18n.A{constants.I18nArgTypeKey: types.Titler.String(constants.DataplaneType)},
			),
		},
		{
			Use:     constants.ImagesCommandName,
			Aliases: []string{constants.ImageCommandName},
			Args:    cobra.ArbitraryArgs,
			RunE:    newDescribeFunction(ctx, fsys, _ui, constants.ImageType),
			Short: i18n.Format(
				i18n.CmdDescribeTypeUsage,
				i18n.A{constants.I18nArgTypeKey: types.Titler.String(constants.ImageType)},
			),
			Long: i18n.Format(
				i18n.CmdDescribeTypeDescription,
				i18n.A{constants.I18nArgTypeKey: types.Titler.String(constants.ImageType)},
			),
		},
		{
			Use:     constants.ImageBuildsCommandName,
			Aliases: []string{constants.ImageBuildCommandName},
			Args:    cobra.ArbitraryArgs,
			RunE:    newDescribeFunction(ctx, fsys, _ui, constants.ImageBuildType),
			Short: i18n.Format(
				i18n.CmdDescribeTypeUsage,
				i18n.A{constants.I18nArgTypeKey: types.Titler.String(constants.ImageBuildType)},
			),
			Long: i18n.Format(
				i18n.CmdDescribeTypeDescription,
				i18n.A{constants.I18nArgTypeKey: types.Titler.String(constants.ImageBuildType)},
			),
		},
		{
			Use:     constants.JobsCommandName,
			Aliases: []string{constants.JobCommandName},
			Args:    cobra.ArbitraryArgs,
			RunE:    newDescribeFunction(ctx, fsys, _ui, constants.JobType),
			Short: i18n.Format(
				i18n.CmdDescribeTypeUsage,
				i18n.A{constants.I18nArgTypeKey: types.Titler.String(constants.JobType)},
			),
			Long: i18n.Format(
				i18n.CmdDescribeTypeDescription,
				i18n.A{constants.I18nArgTypeKey: types.Titler.String(constants.JobType)},
			),
		},
		{
			Use:     constants.JobSubmissionsCommandName,
			Aliases: []string{constants.JobSubmissionCommandName},
			Args:    cobra.ArbitraryArgs,
			RunE:    newDescribeFunction(ctx, fsys, _ui, constants.JobSubmissionType),
			Short: i18n.Format(
				i18n.CmdDescribeTypeUsage,
				i18n.A{constants.I18nArgTypeKey: types.Titler.String(constants.JobSubmissionType)},
			),
			Long: i18n.Format(
				i18n.CmdDescribeTypeDescription,
				i18n.A{constants.I18nArgTypeKey: types.Titler.String(constants.JobSubmissionType)},
			),
		},
		{
			Use:     constants.SnapshotsCommandName,
			Aliases: []string{constants.SnapshotCommandName},
			Args:    cobra.ArbitraryArgs,
			RunE:    newDescribeFunction(ctx, fsys, _ui, constants.SnapshotType),
			Short: i18n.Format(
				i18n.CmdDescribeTypeUsage,
				i18n.A{constants.I18nArgTypeKey: types.Titler.String(constants.SnapshotType)},
			),
			Long: i18n.Format(
				i18n.CmdDescribeTypeDescription,
				i18n.A{constants.I18nArgTypeKey: types.Titler.String(constants.SnapshotType)},
			),
		},
		{
			Use:     constants.UsersCommandName,
			Aliases: []string{constants.UserCommandName},
			Args:    cobra.ArbitraryArgs,
			RunE:    newDescribeFunction(ctx, fsys, _ui, constants.UserType),
			Short: i18n.Format(
				i18n.CmdDescribeTypeUsage,
				i18n.A{constants.I18nArgTypeKey: types.Titler.String(constants.UserType)},
			),
			Long: i18n.Format(
				i18n.CmdDescribeTypeDescription,
				i18n.A{constants.I18nArgTypeKey: types.Titler.String(constants.UserType)},
			),
		},
		{
			Use:     constants.WorkbenchesCommandName,
			Aliases: []string{constants.WorkbenchCommandName},
			Args:    cobra.ArbitraryArgs,
			RunE:    newDescribeFunction(ctx, fsys, _ui, constants.WorkbenchType),
			Short: i18n.Format(
				i18n.CmdDescribeTypeUsage,
				i18n.A{constants.I18nArgTypeKey: types.Titler.String(constants.WorkbenchType)},
			),
			Long: i18n.Format(
				i18n.CmdDescribeTypeDescription,
				i18n.A{constants.I18nArgTypeKey: types.Titler.String(constants.WorkbenchType)},
			),
		},
		{
			Use: constants.VolumeConflictsCommandName,
			Aliases: []string{
				constants.VolumeConflictCommandName,
				constants.WorkspaceConflictCommandName,
				constants.WorkspaceConflictsCommandName,
			},
			Args: cobra.ArbitraryArgs,
			RunE: newDescribeFunction(ctx, fsys, _ui, constants.VolumeConflictType),
			Short: i18n.Format(
				i18n.CmdDescribeTypeUsage,
				i18n.A{constants.I18nArgTypeKey: types.Titler.String(constants.VolumeConflictType)},
			),
			Long: i18n.Format(
				i18n.CmdDescribeTypeDescription,
				i18n.A{constants.I18nArgTypeKey: types.Titler.String(constants.VolumeConflictType)},
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
			RunE: newDescribeFunction(ctx, fsys, _ui, constants.VolumeType),
			Short: i18n.Format(
				i18n.CmdDescribeTypeUsage,
				i18n.A{constants.I18nArgTypeKey: types.Titler.String(constants.VolumeType)},
			),
			Long: i18n.Format(
				i18n.CmdDescribeTypeDescription,
				i18n.A{constants.I18nArgTypeKey: types.Titler.String(constants.VolumeType)},
			),
		},
	}
	return cmds
}
