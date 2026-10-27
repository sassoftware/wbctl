// Copyright © 2026, SAS Institute Inc., Cary, NC, USA.  All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package get

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

// Command returns the get Command.
func Command(
	ctx context.Context,
	fsys afero.Fs,
	_ui ui.UI,
) *cobra.Command {
	cmd := cobra.Command{
		Use:   constants.GetCommandName,
		Short: i18n.Format(i18n.CmdGetUsage),
		Long:  i18n.Format(i18n.CmdGetDescription),
		Args:  cobra.NoArgs,
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
			Use:     constants.ApplicationsCommandName,
			Aliases: []string{constants.ApplicationCommandName},
			Args:    cobra.ArbitraryArgs,
			RunE:    newGetFunction(ctx, fsys, _ui, constants.ApplicationType),
			Short: i18n.Format(
				i18n.CmdGetTypeUsage,
				i18n.A{constants.I18nArgTypeKey: types.Titler.String(constants.ApplicationType)},
			),
			Long: i18n.Format(
				i18n.CmdGetTypeDescription,
				i18n.A{constants.I18nArgTypeKey: types.Titler.String(constants.ApplicationType)},
			),
		},
		{
			Use:     constants.DataplanesCommandName,
			Aliases: []string{constants.DataplaneCommandName},
			Args:    cobra.ArbitraryArgs,
			RunE:    newGetFunction(ctx, fsys, _ui, constants.DataplaneType),
			Short: i18n.Format(
				i18n.CmdGetTypeUsage,
				i18n.A{constants.I18nArgTypeKey: types.Titler.String(constants.DataplaneType)},
			),
			Long: i18n.Format(
				i18n.CmdGetTypeDescription,
				i18n.A{constants.I18nArgTypeKey: types.Titler.String(constants.DataplaneType)},
			),
		},
		{
			Use:     constants.EventsCommandName,
			Aliases: []string{constants.EventCommandName},
			Args:    cobra.ArbitraryArgs,
			RunE:    newGetFunction(ctx, fsys, _ui, constants.EventType),
			Short: i18n.Format(
				i18n.CmdGetTypeUsage,
				i18n.A{constants.I18nArgTypeKey: types.Titler.String(constants.EventType)},
			),
			Long: i18n.Format(
				i18n.CmdGetTypeDescription,
				i18n.A{constants.I18nArgTypeKey: types.Titler.String(constants.EventType)},
			),
		},
		{
			Use:     constants.ImagesCommandName,
			Aliases: []string{constants.ImageCommandName},
			Args:    cobra.ArbitraryArgs,
			RunE:    newGetFunction(ctx, fsys, _ui, constants.ImageType),
			Short: i18n.Format(
				i18n.CmdGetTypeUsage,
				i18n.A{constants.I18nArgTypeKey: types.Titler.String(constants.ImageType)},
			),
			Long: i18n.Format(
				i18n.CmdGetTypeDescription,
				i18n.A{constants.I18nArgTypeKey: types.Titler.String(constants.ImageType)},
			),
		},
		{
			Use:     constants.ImageBuildsCommandName,
			Aliases: []string{constants.ImageBuildCommandName},
			Args:    cobra.ArbitraryArgs,
			RunE:    newGetFunction(ctx, fsys, _ui, constants.ImageBuildType),
			Short: i18n.Format(
				i18n.CmdGetTypeUsage,
				i18n.A{constants.I18nArgTypeKey: types.Titler.String(constants.ImageBuildType)},
			),
			Long: i18n.Format(
				i18n.CmdGetTypeDescription,
				i18n.A{constants.I18nArgTypeKey: types.Titler.String(constants.ImageBuildType)},
			),
		},
		{
			Use:     constants.InitializersCommandName,
			Aliases: []string{constants.InitializerCommandName},
			Args:    cobra.ArbitraryArgs,
			RunE:    newGetFunction(ctx, fsys, _ui, constants.InitializerType),
			Short: i18n.Format(
				i18n.CmdGetTypeUsage,
				i18n.A{constants.I18nArgTypeKey: types.Titler.String(constants.InitializerType)},
			),
			Long: i18n.Format(
				i18n.CmdGetTypeDescription,
				i18n.A{constants.I18nArgTypeKey: types.Titler.String(constants.InitializerType)},
			),
		},
		{
			Use:     constants.JobsCommandName,
			Aliases: []string{constants.JobCommandName},
			Args:    cobra.ArbitraryArgs,
			RunE:    newGetFunction(ctx, fsys, _ui, constants.JobType),
			Short: i18n.Format(
				i18n.CmdGetTypeUsage,
				i18n.A{constants.I18nArgTypeKey: types.Titler.String(constants.JobType)},
			),
			Long: i18n.Format(
				i18n.CmdGetTypeDescription,
				i18n.A{constants.I18nArgTypeKey: types.Titler.String(constants.JobType)},
			),
		},
		{
			Use:     constants.JobSubmissionsCommandName,
			Aliases: []string{constants.JobSubmissionCommandName},
			Args:    cobra.ArbitraryArgs,
			RunE:    newGetFunction(ctx, fsys, _ui, constants.JobSubmissionType),
			Short: i18n.Format(
				i18n.CmdGetTypeUsage,
				i18n.A{constants.I18nArgTypeKey: types.Titler.String(constants.JobSubmissionType)},
			),
			Long: i18n.Format(
				i18n.CmdGetTypeDescription,
				i18n.A{constants.I18nArgTypeKey: types.Titler.String(constants.JobSubmissionType)},
			),
		},
		{
			Use:     constants.NodepoolsCommandName,
			Aliases: []string{constants.NodepoolCommandName},
			Args:    cobra.ArbitraryArgs,
			RunE:    newGetFunction(ctx, fsys, _ui, constants.NodepoolType),
			Short: i18n.Format(
				i18n.CmdGetTypeUsage,
				i18n.A{constants.I18nArgTypeKey: types.Titler.String(constants.NodepoolType)},
			),
			Long: i18n.Format(
				i18n.CmdGetTypeDescription,
				i18n.A{constants.I18nArgTypeKey: types.Titler.String(constants.NodepoolType)},
			),
		},
		{
			Use:     constants.PingsCommandName,
			Aliases: []string{constants.PingCommandName},
			Args:    cobra.ArbitraryArgs,
			RunE:    newGetFunction(ctx, fsys, _ui, constants.PingType),
			Short: i18n.Format(
				i18n.CmdGetTypeUsage,
				i18n.A{constants.I18nArgTypeKey: types.Titler.String(constants.PingType)},
			),
			Long: i18n.Format(
				i18n.CmdGetTypeDescription,
				i18n.A{constants.I18nArgTypeKey: types.Titler.String(constants.PingType)},
			),
		},
		{
			Use:     constants.ServicesCommandName,
			Aliases: []string{constants.ServiceCommandName},
			Args:    cobra.ArbitraryArgs,
			RunE:    newGetFunction(ctx, fsys, _ui, constants.ServiceType),
			Short: i18n.Format(
				i18n.CmdGetTypeUsage,
				i18n.A{constants.I18nArgTypeKey: types.Titler.String(constants.ServiceType)},
			),
			Long: i18n.Format(
				i18n.CmdGetTypeDescription,
				i18n.A{constants.I18nArgTypeKey: types.Titler.String(constants.ServiceType)},
			),
		},
		{
			Use:     constants.SnapshotsCommandName,
			Aliases: []string{constants.SnapshotCommandName},
			Args:    cobra.ArbitraryArgs,
			RunE:    newGetFunction(ctx, fsys, _ui, constants.SnapshotType),
			Short: i18n.Format(
				i18n.CmdGetTypeUsage,
				i18n.A{constants.I18nArgTypeKey: types.Titler.String(constants.SnapshotType)},
			),
			Long: i18n.Format(
				i18n.CmdGetTypeDescription,
				i18n.A{constants.I18nArgTypeKey: types.Titler.String(constants.SnapshotType)},
			),
		},
		{
			Use:     constants.StorageclassesCommandName,
			Aliases: []string{constants.StorageclassCommandName},
			Args:    cobra.ArbitraryArgs,
			RunE:    newGetFunction(ctx, fsys, _ui, constants.StorageclassType),
			Short: i18n.Format(
				i18n.CmdGetTypeUsage,
				i18n.A{constants.I18nArgTypeKey: types.Titler.String(constants.StorageclassType)},
			),
			Long: i18n.Format(
				i18n.CmdGetTypeDescription,
				i18n.A{constants.I18nArgTypeKey: types.Titler.String(constants.StorageclassType)},
			),
		},
		{
			Use:     constants.UsersCommandName,
			Aliases: []string{constants.UserCommandName},
			Args:    cobra.ArbitraryArgs,
			RunE:    newGetFunction(ctx, fsys, _ui, constants.UserType),
			Short: i18n.Format(
				i18n.CmdGetTypeUsage,
				i18n.A{constants.I18nArgTypeKey: types.Titler.String(constants.UserType)},
			),
			Long: i18n.Format(
				i18n.CmdGetTypeDescription,
				i18n.A{constants.I18nArgTypeKey: types.Titler.String(constants.UserType)},
			),
		},
		{
			Use:     constants.WorkbenchesCommandName,
			Aliases: []string{constants.WorkbenchCommandName},
			Args:    cobra.ArbitraryArgs,
			RunE:    newGetFunction(ctx, fsys, _ui, constants.WorkbenchType),
			Short: i18n.Format(
				i18n.CmdGetTypeUsage,
				i18n.A{constants.I18nArgTypeKey: types.Titler.String(constants.WorkbenchType)},
			),
			Long: i18n.Format(
				i18n.CmdGetTypeDescription,
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
			RunE: newGetFunction(ctx, fsys, _ui, constants.VolumeConflictType),
			Short: i18n.Format(
				i18n.CmdGetTypeUsage,
				i18n.A{constants.I18nArgTypeKey: types.Titler.String(constants.VolumeConflictType)},
			),
			Long: i18n.Format(
				i18n.CmdGetTypeDescription,
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
			RunE: newGetFunction(ctx, fsys, _ui, constants.VolumeType),
			Short: i18n.Format(
				i18n.CmdGetTypeUsage,
				i18n.A{constants.I18nArgTypeKey: types.Titler.String(constants.VolumeType)},
			),
			Long: i18n.Format(
				i18n.CmdGetTypeDescription,
				i18n.A{constants.I18nArgTypeKey: types.Titler.String(constants.VolumeType)},
			),
		},
	}
	for _, cmd := range cmds {
		flags.AddOutputFlag(cmd)
		flags.AddSortFlag(cmd)
	}
	return cmds
}
