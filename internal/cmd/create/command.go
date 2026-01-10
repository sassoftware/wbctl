// Copyright © 2026, SAS Institute Inc., Cary, NC, USA.  All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package create

import (
	"context"

	"github.com/sas-institute-rnd-saas/wbctl/internal/cmd/create/image"
	"github.com/sas-institute-rnd-saas/wbctl/internal/cmd/create/job"
	"github.com/sas-institute-rnd-saas/wbctl/internal/cmd/create/job_submission"
	"github.com/sas-institute-rnd-saas/wbctl/internal/cmd/create/snapshot"
	"github.com/sas-institute-rnd-saas/wbctl/internal/cmd/create/workbench"
	"github.com/sas-institute-rnd-saas/wbctl/internal/cmd/create/workspace"
	"github.com/sas-institute-rnd-saas/wbctl/internal/constants"
	"github.com/sas-institute-rnd-saas/wbctl/internal/i18n"
	"github.com/sas-institute-rnd-saas/wbctl/internal/ui"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
)

// CreateCommand returns the create Command.
// nolint:revive
func CreateCommand(
	ctx context.Context,
	fsys afero.Fs,
	_ui ui.UI,
) *cobra.Command {
	cmd := cobra.Command{
		Use:   constants.CreateCommandName,
		Short: i18n.Format(i18n.CmdCreateUsage),
		Long:  i18n.Format(i18n.CmdCreateDescription),
		Args:  cobra.NoArgs,
	}
	cmd.AddCommand(getCreateSubcommands(ctx, fsys, _ui)...)
	return &cmd
}

// UpdateCommand returns the update Command.
func UpdateCommand(
	ctx context.Context,
	fsys afero.Fs,
	_ui ui.UI,
) *cobra.Command {
	cmd := cobra.Command{
		Use:   constants.UpdateCommandName,
		Short: i18n.Format(i18n.CmdUpdateUsage),
		Long:  i18n.Format(i18n.CmdUpdateDescription),
		Args:  cobra.NoArgs,
	}
	cmd.AddCommand(getUpdateSubcommands(ctx, fsys, _ui)...)
	return &cmd
}

func getCreateSubcommands(
	ctx context.Context,
	fsys afero.Fs,
	_ui ui.UI,
) []*cobra.Command {
	cmds := []*cobra.Command{
		job.CreateCommand(ctx, fsys, _ui),
		job_submission.CreateCommand(ctx, fsys, _ui),
		workbench.CreateCommand(ctx, fsys, _ui),
		workspace.CreateCommand(ctx, fsys, _ui),
		image.CreateCommand(ctx, fsys, _ui),
		snapshot.CreateCommand(ctx, fsys, _ui),
	}
	return cmds
}

func getUpdateSubcommands(
	ctx context.Context,
	fsys afero.Fs,
	_ui ui.UI,
) []*cobra.Command {
	cmds := []*cobra.Command{
		job.UpdateCommand(ctx, fsys, _ui),
		workbench.UpdateCommand(ctx, fsys, _ui),
		workspace.UpdateCommand(ctx, fsys, _ui),
		image.UpdateCommand(ctx, fsys, _ui),
		snapshot.UpdateCommand(ctx, fsys, _ui),
	}
	return cmds
}
