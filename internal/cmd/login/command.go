// Copyright © 2026, SAS Institute Inc., Cary, NC, USA.  All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package login

import (
	"context"

	"github.com/sassoftware/wbctl/internal/cmd/flags"
	"github.com/sassoftware/wbctl/internal/constants"
	"github.com/sassoftware/wbctl/internal/i18n"
	"github.com/sassoftware/wbctl/internal/ui"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
)

// Command returns the login Command.
func Command(
	ctx context.Context,
	fsys afero.Fs,
	_ui ui.UI,
) *cobra.Command {
	cmd := &cobra.Command{
		Use:    constants.LoginCommandName,
		Short:  i18n.Format(i18n.CmdLoginUsage),
		Long:   i18n.Format(i18n.CmdLoginDescription),
		Args:   cobra.NoArgs,
		RunE:   newLoginFunction(ctx, fsys, _ui),
		PreRun: flags.BindToViper,
	}
	return cmd
}
