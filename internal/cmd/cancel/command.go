// Copyright © 2026, SAS Institute Inc., Cary, NC, USA.  All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package cancel

import (
	"context"

	"github.com/sassoftware/wbctl/internal/cmd/flags"
	cmdtypes "github.com/sassoftware/wbctl/internal/cmd/types"
	"github.com/sassoftware/wbctl/internal/constants"
	"github.com/sassoftware/wbctl/internal/i18n"
	"github.com/sassoftware/wbctl/internal/ui"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
)

// Command returns the cancel Command.
func Command(
	ctx context.Context,
	fsys afero.Fs,
	_ui ui.UI,
) *cobra.Command {
	cmd := &cobra.Command{
		Use:              constants.CancelCommandName,
		Short:            i18n.Format(i18n.CmdCancelUsage),
		Long:             i18n.Format(i18n.CmdCancelDescription),
		Args:             cobra.NoArgs,
		PersistentPreRun: flags.BindToViper,
	}
	cmd.AddCommand(getSubcommands(ctx, fsys, _ui)...)
	return cmd
}

func getSubcommands(
	ctx context.Context,
	fsys afero.Fs,
	_ui ui.UI,
) []*cobra.Command {
	return []*cobra.Command{
		CancelJobSubmissionCommand(ctx, fsys, _ui),
	}
}

// CancelJobSubmissionCommand returns the cancel job-submission command.
// nolint:revive
func CancelJobSubmissionCommand(
	ctx context.Context,
	fsys afero.Fs,
	_ui ui.UI,
) *cobra.Command {
	return &cobra.Command{
		Use:     constants.JobSubmissionCommandName,
		Aliases: []string{constants.JobSubmissionsCommandName},
		Args:    cobra.MinimumNArgs(1),
		RunE:    newCancelFunction(ctx, fsys, _ui, constants.JobSubmissionType),
		Short: i18n.Format(
			i18n.CmdCancelTypeUsage,
			i18n.A{constants.I18nArgTypeKey: cmdtypes.Titler.String(constants.JobSubmissionType)},
		),
		Long: i18n.Format(
			i18n.CmdCancelTypeDescription,
			i18n.A{constants.I18nArgTypeKey: cmdtypes.Titler.String(constants.JobSubmissionType)},
		),
	}
}
