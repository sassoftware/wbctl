// Copyright © 2026, SAS Institute Inc., Cary, NC, USA.  All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package rebuild

import (
	"context"

	"github.com/sassoftware/wbctl/internal/cmd/types"
	"github.com/sassoftware/wbctl/internal/constants"
	"github.com/sassoftware/wbctl/internal/i18n"
	"github.com/sassoftware/wbctl/internal/ui"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
)

// Command returns the rebuild Command.
func Command(
	ctx context.Context,
	fsys afero.Fs,
	_ui ui.UI,
) *cobra.Command {
	cmd := &cobra.Command{
		Use:   constants.RebuildCommandName,
		Short: i18n.Format(i18n.CmdRebuildUsage),
		Long:  i18n.Format(i18n.CmdRebuildDescription),
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
			Use:     constants.ImagesCommandName,
			Aliases: []string{constants.ImageCommandName},
			Args:    cobra.MinimumNArgs(1),
			RunE:    newRebuildFunction(ctx, fsys, _ui, constants.ImageCommandName),
			Short: i18n.Format(
				i18n.CmdRebuildTypeUsage,
				i18n.A{constants.I18nArgTypeKey: types.Titler.String(constants.ImageType)},
			),
			Long: i18n.Format(
				i18n.CmdRebuildTypeDescription,
				i18n.A{constants.I18nArgTypeKey: types.Titler.String(constants.ImageType)},
			),
		},
	}
}
