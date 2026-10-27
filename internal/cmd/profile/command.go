// Copyright © 2026, SAS Institute Inc., Cary, NC, USA.  All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package profile

import (
	"context"

	profileinit "github.com/sassoftware/wbctl/internal/cmd/profile/init"
	profilelist "github.com/sassoftware/wbctl/internal/cmd/profile/list"
	profileshow "github.com/sassoftware/wbctl/internal/cmd/profile/show"
	"github.com/sassoftware/wbctl/internal/constants"
	"github.com/sassoftware/wbctl/internal/i18n"
	"github.com/sassoftware/wbctl/internal/ui"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
)

// Command returns the profile Command.
func Command(
	ctx context.Context,
	fsys afero.Fs,
	_ui ui.UI,
) *cobra.Command {
	cmd := &cobra.Command{
		Use:     constants.ProfileCommandName,
		Aliases: []string{constants.ProfileAliasName},
		Short:   i18n.Format(i18n.CmdProfileUsage),
		Long:    i18n.Format(i18n.CmdProfileDescription),
		Args:    cobra.NoArgs,
	}
	cmd.AddCommand([]*cobra.Command{
		profileinit.Command(ctx, fsys, _ui),
		profilelist.Command(ctx, fsys, _ui),
		profileshow.Command(ctx, fsys, _ui),
	}...)
	return cmd
}
