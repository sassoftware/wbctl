// Copyright © 2026, SAS Institute Inc., Cary, NC, USA.  All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package whoami

import (
	"context"

	"github.com/sassoftware/wbctl/internal/cmd/flags"
	"github.com/sassoftware/wbctl/internal/constants"
	"github.com/sassoftware/wbctl/internal/i18n"
	"github.com/sassoftware/wbctl/internal/ui"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
)

// Command returns the whoami Command.
func Command(
	ctx context.Context,
	fsys afero.Fs,
	_ui ui.UI,
) *cobra.Command {
	cmd := &cobra.Command{
		Use:    constants.WhoamiCommandName,
		Short:  i18n.Format(i18n.CmdWhoamiUsage),
		Long:   i18n.Format(i18n.CmdWhoamiDescription),
		Args:   cobra.NoArgs,
		RunE:   newWhoamiFunction(ctx, fsys, _ui),
		PreRun: flags.BindToViper,
	}
	flags.AddOutputFlag(cmd)
	return cmd
}
