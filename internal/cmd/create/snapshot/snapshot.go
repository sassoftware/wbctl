// Copyright © 2026, SAS Institute Inc., Cary, NC, USA.  All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package snapshot

import (
	"context"
	"errors"
	"fmt"

	snapshot1 "github.com/kubernetes-csi/external-snapshotter/client/v8/apis/volumesnapshot/v1"
	"github.com/sassoftware/sas-ggdk/pkg/pointer"
	"github.com/sassoftware/wbctl/internal/cmd/auth"
	"github.com/sassoftware/wbctl/internal/cmd/flags"
	cmdtypes "github.com/sassoftware/wbctl/internal/cmd/types"
	"github.com/sassoftware/wbctl/internal/constants"
	"github.com/sassoftware/wbctl/internal/i18n"
	"github.com/sassoftware/wbctl/internal/jsonutil"
	"github.com/sassoftware/wbctl/internal/services/workbench"
	"github.com/sassoftware/wbctl/internal/sliceutil"
	"github.com/sassoftware/wbctl/internal/types"
	"github.com/sassoftware/wbctl/internal/ui"
	"github.com/sassoftware/wbctl/internal/wbctljson"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// CreateCommand returns the create snapshot Command.
func CreateCommand(
	ctx context.Context,
	fsys afero.Fs,
	_ui ui.UI,
) *cobra.Command {
	cmd := &cobra.Command{
		Use:     constants.SnapshotCommandName,
		Aliases: []string{constants.SnapshotsCommandName},
		Args:    cobra.NoArgs,
		RunE:    newCreateSnapshotFunction(ctx, fsys, _ui),
		Short: i18n.Format(
			i18n.CmdCreateTypeUsage,
			i18n.A{constants.I18nArgTypeKey: cmdtypes.Titler.String(constants.SnapshotType)},
		),
		Long: i18n.Format(
			i18n.CmdCreateTypeDescription,
			i18n.A{constants.I18nArgTypeKey: cmdtypes.Titler.String(constants.SnapshotType)},
		),
	}
	flags.AddFilenameFlag(cmd, constants.SnapshotType)
	flags.AddDryRunFlag(cmd)
	flags.AddDataplaneFlag(cmd, constants.SnapshotType)
	flags.AddNameFlag(cmd, constants.SnapshotType)
	flags.AddDescriptionFlag(cmd, constants.SnapshotType)
	flags.AddSnapshotVolumeFlag(cmd)
	return cmd
}

// UpdateCommand returns the update snapshot Command.
func UpdateCommand(
	ctx context.Context,
	fsys afero.Fs,
	_ui ui.UI,
) *cobra.Command {
	cmd := &cobra.Command{
		Use:     constants.SnapshotsCommandName,
		Aliases: []string{constants.SnapshotCommandName},
		Args:    cobra.ArbitraryArgs,
		RunE:    newUpdateSnapshotFunction(ctx, fsys, _ui),
		Short: i18n.Format(
			i18n.CmdUpdateTypeUsage,
			i18n.A{constants.I18nArgTypeKey: cmdtypes.Titler.String(constants.SnapshotType)},
		),
		Long: i18n.Format(
			i18n.CmdUpdateTypeDescription,
			i18n.A{constants.I18nArgTypeKey: cmdtypes.Titler.String(constants.SnapshotType)},
		),
	}
	flags.AddFilenameFlag(cmd, constants.SnapshotType)
	flags.AddDryRunFlag(cmd)
	flags.AddNameFlag(cmd, constants.SnapshotType)
	flags.AddDescriptionFlag(cmd, constants.SnapshotType)
	flags.AddSnapshotVolumeFlag(cmd)
	return cmd
}

// newCreateSnapshotFunction creates a new cli.ActionFunc for the `create
// snapshot` command.
func newCreateSnapshotFunction(
	ctx context.Context,
	fsys afero.Fs,
	_ui ui.UI,
) cmdtypes.ActionFunc {
	return func(
		_ *cobra.Command,
		_ []string,
	) error {
		v := viper.GetViper()
		endpoint := v.GetString(constants.EndpointFlag)
		accessToken, csrfToken, err := auth.GetAccessTokens(ctx, v, _ui, fsys)
		if err != nil {
			return err
		}
		svc := workbench.NewWorkbenchService(endpoint, accessToken, csrfToken)
		body, err := getCreateSnapshotRequestBody(ctx, v, fsys, svc)
		if err != nil {
			return err
		}
		dryRun := v.GetBool(constants.DryRunFlag)
		if dryRun {
			response, err := wbctljson.MakeMarshaler[types.Snapshot](wbctljson.EditOptions)(body)
			if err == nil {
				_ui.Display(string(response))
			}
			return err
		}
		res, err := workbench.CreateSnapshot(ctx, svc, body)
		if err == nil {
			_ui.Display(res.ID)
		}
		return err
	}
}

// newUpdateSnapshotFunction creates a new cli.ActionFunc for the `update
// snapshot` command.
func newUpdateSnapshotFunction(
	ctx context.Context,
	fsys afero.Fs,
	_ui ui.UI,
) cmdtypes.ActionFunc {
	return func(
		_ *cobra.Command,
		args []string,
	) error {
		v := viper.GetViper()
		endpoint := v.GetString(constants.EndpointFlag)
		accessToken, csrfToken, err := auth.GetAccessTokens(ctx, v, _ui, fsys)
		if err != nil {
			return err
		}
		svc := workbench.NewWorkbenchService(endpoint, accessToken, csrfToken)
		body, err := getSnapshotRequestBody(ctx, v, fsys, svc)
		if err != nil {
			return err
		}
		ids, err := resolveUpdateSnapshotIDs(args, body)
		if err != nil {
			return err
		}
		dryRun := v.GetBool(constants.DryRunFlag)
		if dryRun {
			payload := makeUpdateSnapshotPayload(ids, body)
			response, err := wbctljson.MakeMarshaler[map[string]any](wbctljson.EditOptions)(payload)
			if err == nil {
				_ui.Display(string(response))
			}
			return err
		}
		res, err := workbench.BulkUpdateSnapshots(ctx, svc, ids, body)
		if err == nil {
			for _, resource := range res {
				_ui.Display(resource.ID)
			}
		}
		return err
	}
}

func getSnapshotRequestBody(
	ctx context.Context,
	v *viper.Viper,
	fsys afero.Fs,
	svc *workbench.WorkbenchService,
) (types.Snapshot, error) {
	filename := v.GetString(constants.FilenameFlag)
	if filename != "" {
		return jsonutil.GetFromFileOrStdin[types.Snapshot](fsys, filename)
	}
	name := v.GetString(constants.NameFlag)
	description := v.GetString(constants.DescriptionFlag)
	volumeID := v.GetString(constants.VolumeFlag)
	if volumeID == "" {
		volumeID = v.GetString(constants.WorkspaceFlag)
	}

	controllerID, err := getWorkspaceControllerID(ctx, svc, volumeID)
	if err != nil {
		return types.Snapshot{}, err
	}
	snap := types.Snapshot{
		Name:        name,
		Description: description,
	}
	if controllerID != "" {
		snap.Spec = &snapshot1.VolumeSnapshotSpec{
			Source: snapshot1.VolumeSnapshotSource{
				PersistentVolumeClaimName: pointer.Ptr(controllerID),
			},
		}
	}
	return snap, nil
}

func getCreateSnapshotRequestBody(
	ctx context.Context,
	v *viper.Viper,
	fsys afero.Fs,
	svc *workbench.WorkbenchService,
) (types.Snapshot, error) {
	snapshot, err := getSnapshotRequestBody(ctx, v, fsys, svc)
	if err != nil {
		return types.Snapshot{}, err
	}
	// Dataplane is not supported for update but is for create. So we have to
	// resolve the dataplane id here instead of in the common implementation
	// function.
	dataplaneID := v.GetString(constants.DataplaneFlag)
	snapshot.DataplaneID, err = workbench.GetValidDataplaneID(ctx, svc, dataplaneID)
	if err != nil {
		return types.Snapshot{}, err
	}
	if snapshot.Name == "" {
		return types.Snapshot{}, errors.New(
			i18n.Format(
				i18n.FlagIsRequired,
				i18n.A{constants.I18nArgKeyFlagName: constants.NameFlag},
			),
		)
	}
	if snapshot.Spec == nil {
		return types.Snapshot{},
			errors.New(
				i18n.Format(
					i18n.FlagIsRequired,
					i18n.A{constants.I18nArgKeyFlagName: constants.VolumeFlag},
				),
			)
	}
	return snapshot, nil
}

func getWorkspaceControllerID(
	ctx context.Context,
	svc *workbench.WorkbenchService,
	workspaceID string,
) (string, error) {
	if workspaceID == "" {
		return "", nil
	}
	workspaces, err := workbench.GetWorkspaces(ctx, svc, []string{workspaceID}, "")
	if err != nil {
		return "", fmt.Errorf(
			"%s: %w",
			i18n.Format(i18n.FailedToGetWorkspace, i18n.A{constants.I18nArgKeyID: workspaceID}),
			err,
		)
	}
	workspace, err := sliceutil.GetAndEnsureOnlyOne[types.Workspace](workspaces)
	if err != nil {
		return "", fmt.Errorf(
			"%s: %w",
			i18n.Format(i18n.FailedToGetWorkspace, i18n.A{constants.I18nArgKeyID: workspaceID}),
			err,
		)
	}
	return workspace.ControllerID, nil
}

func resolveUpdateSnapshotIDs(
	args []string,
	body types.Snapshot,
) ([]string, error) {
	ids := args
	if len(ids) > 0 {
		return ids, nil
	}
	if body.ID == "" {
		return nil, errors.New(
			i18n.Format(
				i18n.AtLeastOneResourceIDForCommand,
				i18n.A{constants.I18nArgKeyCommandName: constants.UpdateCommandName},
			),
		)
	}
	return []string{body.ID}, nil
}

func makeUpdateSnapshotPayload(
	ids []string,
	body types.Snapshot,
) map[string]interface{} {
	return map[string]interface{}{
		constants.RPCBodyKeyIDs:   ids,
		constants.RPCBodyKeyParam: body,
	}
}
