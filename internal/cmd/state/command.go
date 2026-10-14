// Copyright © 2026, SAS Institute Inc., Cary, NC, USA.  All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package state

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

// StopCommand returns the stop Command.
func StopCommand(
	ctx context.Context,
	fsys afero.Fs,
	_ui ui.UI,
) *cobra.Command {
	cmd := &cobra.Command{
		Use:              constants.StopCommandName,
		Short:            i18n.Format(i18n.CmdStopUsage),
		Long:             i18n.Format(i18n.CmdStopDescription),
		Args:             cobra.NoArgs,
		PersistentPreRun: flags.BindToViper,
	}
	cmd.AddCommand(getStopSubcommands(ctx, fsys, _ui)...)
	return cmd
}

// StartCommand returns the start Command.
func StartCommand(
	ctx context.Context,
	fsys afero.Fs,
	_ui ui.UI,
) *cobra.Command {
	cmd := &cobra.Command{
		Use:              constants.StartCommandName,
		Short:            i18n.Format(i18n.CmdStartUsage),
		Long:             i18n.Format(i18n.CmdStartDescription),
		Args:             cobra.NoArgs,
		PersistentPreRun: flags.BindToViper,
	}
	cmd.AddCommand(getStartSubcommands(ctx, fsys, _ui)...)
	return cmd
}

func getStopSubcommands(
	ctx context.Context,
	fsys afero.Fs,
	_ui ui.UI,
) []*cobra.Command {
	return []*cobra.Command{
		{
			Use:     constants.WorkbenchesCommandName,
			Aliases: []string{constants.WorkbenchCommandName},
			Args:    cobra.MinimumNArgs(1),
			RunE:    newUpdateStateFunction(ctx, fsys, _ui, constants.WorkbenchType, constants.DesiredStateStopped),
			Short: i18n.Format(
				i18n.CmdStopTypeUsage,
				i18n.A{constants.I18nArgTypeKey: types.Titler.String(constants.WorkbenchType)},
			),
			Long: i18n.Format(
				i18n.CmdStopTypeDescription,
				i18n.A{constants.I18nArgTypeKey: types.Titler.String(constants.WorkbenchType)},
			),
		},
	}
}

func getStartSubcommands(
	ctx context.Context,
	fsys afero.Fs,
	_ui ui.UI,
) []*cobra.Command {
	return []*cobra.Command{
		{
			Use:     constants.WorkbenchesCommandName,
			Aliases: []string{constants.WorkbenchCommandName},
			Args:    cobra.MinimumNArgs(1),
			RunE:    newUpdateStateFunction(ctx, fsys, _ui, constants.WorkbenchType, constants.DesiredStateRunning),
			Short: i18n.Format(
				i18n.CmdStartTypeUsage,
				i18n.A{constants.I18nArgTypeKey: types.Titler.String(constants.WorkbenchType)},
			),
			Long: i18n.Format(
				i18n.CmdStartTypeDescription,
				i18n.A{constants.I18nArgTypeKey: types.Titler.String(constants.WorkbenchType)},
			),
		},
	}
}
