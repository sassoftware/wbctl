// Copyright © 2026, SAS Institute Inc., Cary, NC, USA.  All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package api

import (
	"context"

	"github.com/sassoftware/wbctl/internal/cmd/flags"
	"github.com/sassoftware/wbctl/internal/constants"
	"github.com/sassoftware/wbctl/internal/i18n"
	"github.com/sassoftware/wbctl/internal/ui"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
)

// Command returns the api Command.
func Command(
	ctx context.Context,
	fsys afero.Fs,
	_ui ui.UI,
) *cobra.Command {
	cmd := &cobra.Command{
		Use:    constants.APICommandName,
		Short:  i18n.Format(i18n.CmdAPIUsage),
		Args:   cobra.MaximumNArgs(1),
		RunE:   newAPIFunction(ctx, fsys, _ui),
		Hidden: true,
	}
	flags.AddFilenameFlag(cmd, i18n.Format(i18n.APIRequestBodyType))
	flags.AddDataFlag(cmd, i18n.Format(i18n.APIRequestBodyType))
	flags.AddMethodFlag(cmd)
	flags.AddQueryFlag(cmd)
	flags.AddOutputFlag(cmd)
	return cmd
}
