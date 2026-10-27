// Copyright © 2026, SAS Institute Inc., Cary, NC, USA.  All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package log

import (
	"context"

	"github.com/sassoftware/wbctl/internal/cmd/types"
	"github.com/sassoftware/wbctl/internal/constants"
	"github.com/sassoftware/wbctl/internal/i18n"
	"github.com/sassoftware/wbctl/internal/ui"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
)

// Command returns the log Command.
func Command(
	ctx context.Context,
	fsys afero.Fs,
	_ui ui.UI,
) *cobra.Command {
	cmd := &cobra.Command{
		Use:   constants.LogCommandName,
		Short: i18n.Format(i18n.CmdLogUsage),
		Long:  i18n.Format(i18n.CmdLogDescription),
		Args:  cobra.NoArgs,
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
		{
			Use:     constants.JobSubmissionsCommandName,
			Aliases: []string{constants.JobSubmissionCommandName},
			Args:    cobra.MinimumNArgs(1),
			RunE:    newLogFunction(ctx, fsys, _ui, constants.JobSubmissionType),
			Short: i18n.Format(
				i18n.CmdLogTypeUsage,
				i18n.A{constants.I18nArgTypeKey: types.Titler.String(constants.JobSubmissionType)},
			),
			Long: i18n.Format(
				i18n.CmdLogTypeDescription,
				i18n.A{constants.I18nArgTypeKey: types.Titler.String(constants.JobSubmissionType)},
			),
		},
		{
			Use:     constants.ImageBuildsCommandName,
			Aliases: []string{constants.ImageBuildCommandName},
			Args:    cobra.MinimumNArgs(1),
			RunE:    newLogFunction(ctx, fsys, _ui, constants.ImageBuildType),
			Short: i18n.Format(
				i18n.CmdLogTypeUsage,
				i18n.A{constants.I18nArgTypeKey: types.Titler.String(constants.ImageBuildType)},
			),
			Long: i18n.Format(
				i18n.CmdLogTypeDescription,
				i18n.A{constants.I18nArgTypeKey: types.Titler.String(constants.ImageBuildType)},
			),
		},
	}
}
