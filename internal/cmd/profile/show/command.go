// Copyright © 2026, SAS Institute Inc., Cary, NC, USA.  All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package show

import (
	"context"

	"github.com/sassoftware/wbctl/internal/cmd/flags"
	"github.com/sassoftware/wbctl/internal/constants"
	"github.com/sassoftware/wbctl/internal/i18n"
	"github.com/sassoftware/wbctl/internal/ui"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
)

// Command returns the profile init Command. The flags it reads (auth-mode,
// endpoint, ignore-token-cache, no-listen, ingress-host, organization,
// client-id, client-secret, verbose, profile) are the persistent root flags
// inherited from the parent app.
func Command(
	ctx context.Context,
	fsys afero.Fs,
	_ui ui.UI,
) *cobra.Command {
	cmd := &cobra.Command{
		Use:   constants.ShowCommandName,
		Short: i18n.Format(i18n.CmdProfileShowUsage),
		Long:  i18n.Format(i18n.CmdProfileShowDescription),
		Args:  cobra.NoArgs,
		RunE:  newShowFunction(ctx, fsys, _ui),
	}
	flags.AddOutputFlag(cmd)
	return cmd
}
