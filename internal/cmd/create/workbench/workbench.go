// Copyright © 2026, SAS Institute Inc., Cary, NC, USA.  All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package workbench

import (
	"context"
	"errors"

	"github.com/sassoftware/wbctl/internal/cmd/auth"
	"github.com/sassoftware/wbctl/internal/cmd/flags"
	cmdtypes "github.com/sassoftware/wbctl/internal/cmd/types"
	"github.com/sassoftware/wbctl/internal/constants"
	"github.com/sassoftware/wbctl/internal/i18n"
	"github.com/sassoftware/wbctl/internal/jsonutil"
	"github.com/sassoftware/wbctl/internal/maputil"
	"github.com/sassoftware/wbctl/internal/services/workbench"
	"github.com/sassoftware/wbctl/internal/types"
	"github.com/sassoftware/wbctl/internal/ui"
	"github.com/sassoftware/wbctl/internal/wbctljson"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// CreateCommand returns the create workbench Command.
func CreateCommand(
	ctx context.Context,
	fsys afero.Fs,
	_ui ui.UI,
) *cobra.Command {
	cmd := &cobra.Command{
		Use:     constants.WorkbenchCommandName,
		Aliases: []string{constants.WorkbenchesCommandName},
		Args:    cobra.NoArgs,
		RunE:    newCreateWorkbenchFunction(ctx, fsys, _ui),
		Short: i18n.Format(
			i18n.CmdCreateTypeUsage,
			i18n.A{constants.I18nArgTypeKey: cmdtypes.Titler.String(constants.WorkbenchType)},
		),
		Long: i18n.Format(
			i18n.CmdCreateTypeDescription,
			i18n.A{constants.I18nArgTypeKey: cmdtypes.Titler.String(constants.WorkbenchType)},
		),
		PreRun: flags.BindToViper,
	}
	flags.AddFilenameFlag(cmd, constants.WorkbenchType)
	flags.AddDryRunFlag(cmd)
	flags.AddDataplaneFlag(cmd, constants.WorkbenchType)
	flags.AddNodepoolFlag(cmd, constants.WorkbenchType)
	flags.AddNameFlag(cmd, constants.WorkbenchType)
	flags.AddLabelFlag(cmd, constants.WorkbenchType)
	flags.AddVolumeFlag(cmd, constants.WorkbenchType)
	flags.AddApplicationFlag(cmd, constants.WorkbenchType)
	flags.AddServiceFlag(cmd, constants.WorkbenchType)
	flags.AddDesiredStateFlag(cmd, constants.WorkbenchType)
	return cmd
}

// UpdateCommand returns the update workbench Command.
func UpdateCommand(
	ctx context.Context,
	fsys afero.Fs,
	_ui ui.UI,
) *cobra.Command {
	cmd := &cobra.Command{
		Use:     constants.WorkbenchesCommandName,
		Aliases: []string{constants.WorkbenchCommandName},
		Args:    cobra.ArbitraryArgs,
		RunE:    newUpdateWorkbenchFunction(ctx, fsys, _ui),
		Short: i18n.Format(
			i18n.CmdUpdateTypeUsage,
			i18n.A{constants.I18nArgTypeKey: cmdtypes.Titler.String(constants.WorkbenchType)},
		),
		Long: i18n.Format(
			i18n.CmdUpdateTypeDescription,
			i18n.A{constants.I18nArgTypeKey: cmdtypes.Titler.String(constants.WorkbenchType)},
		),
		PreRun: flags.BindToViper,
	}
	flags.AddFilenameFlag(cmd, constants.WorkbenchType)
	flags.AddDryRunFlag(cmd)
	flags.AddNodepoolFlag(cmd, constants.WorkbenchType)
	flags.AddNameFlag(cmd, constants.WorkbenchType)
	flags.AddLabelFlag(cmd, constants.WorkbenchType)
	flags.AddVolumeFlag(cmd, constants.WorkbenchType)
	flags.AddApplicationFlag(cmd, constants.WorkbenchType)
	flags.AddServiceFlag(cmd, constants.WorkbenchType)
	flags.AddDesiredStateFlag(cmd, constants.WorkbenchType)
	return cmd
}

// newCreateWorkbenchFunction creates a new cli.ActionFunc for the `create
// workbench` command.
func newCreateWorkbenchFunction(
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
		body, err := getCreateWorkbenchRequestBody(ctx, v, fsys, _ui, svc)
		if err != nil {
			return err
		}
		dryRun := v.GetBool(constants.DryRunFlag)
		if dryRun {
			response, err := wbctljson.MakeMarshaler[types.Workbench](wbctljson.EditOptions)(body)
			if err == nil {
				_ui.Display(string(response))
			}
			return err
		}
		res, err := workbench.CreateWorkbench(ctx, svc, body)
		if err == nil {
			_ui.Display(res.ID)
		}
		return err
	}
}

// newUpdateWorkbenchFunction creates a new cli.ActionFunc for the `update
// workbench` command.
func newUpdateWorkbenchFunction(
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
		body, err := getWorkbenchRequestBody(ctx, v, fsys, svc)
		if err != nil {
			return err
		}
		ids, err := resolveUpdateWorkbenchIDs(args, body)
		if err != nil {
			return err
		}
		dryRun := v.GetBool(constants.DryRunFlag)
		if dryRun {
			payload := makeUpdateWorkbenchPayload(ids, body)
			response, err := wbctljson.MakeMarshaler[map[string]any](wbctljson.EditOptions)(payload)
			if err == nil {
				_ui.Display(string(response))
			}
			return err
		}
		res, err := workbench.BulkUpdateWorkbenches(ctx, svc, ids, body)
		if err == nil {
			for _, resource := range res {
				_ui.Display(resource.ID)
			}
		}
		return err
	}
}

func getWorkbenchRequestBody(
	ctx context.Context,
	v *viper.Viper,
	fsys afero.Fs,
	svc *workbench.WorkbenchService,
) (types.Workbench, error) {
	filename := v.GetString(constants.FilenameFlag)
	if filename != "" {
		return jsonutil.GetFromFileOrStdin[types.Workbench](fsys, filename)
	}
	name := v.GetString(constants.NameFlag)
	application := v.GetString(constants.ApplicationFlag)
	nodepool := v.GetString(constants.NodepoolFlag)
	desiredState := v.GetString(constants.DesiredStateFlag)
	labels := v.GetStringSlice(constants.LabelFlag)
	volumes := v.GetStringSlice(constants.VolumeFlag)
	if len(volumes) == 0 {
		volumes = v.GetStringSlice(constants.WorkspaceFlag)
	}
	services := v.GetStringSlice(constants.ServiceFlag)

	serviceRequests, err := workbench.GetServiceRequests(ctx, svc, application, services)
	if err != nil {
		return types.Workbench{}, err
	}
	workspaceRequests, err := workbench.GetValidWorkspaceRequests(ctx, svc, volumes)
	if err != nil {
		return types.Workbench{}, err
	}
	wkbnch := types.Workbench{
		Name:          name,
		Labels:        maputil.FromEqualDelimitedLines(labels),
		Nodepool:      nodepool,
		Workspaces:    make([]types.WorkspaceRequest, 0),
		Services:      serviceRequests,
		State:         desiredState,
		ApplicationID: application,
	}
	setWorkspaceRequests(&wkbnch, workspaceRequests)
	return wkbnch, nil
}

func getCreateWorkbenchRequestBody(
	ctx context.Context,
	v *viper.Viper,
	fsys afero.Fs,
	_ui ui.UI,
	svc *workbench.WorkbenchService,
) (types.Workbench, error) {
	wkbnch, err := getWorkbenchRequestBody(ctx, v, fsys, svc)
	if err != nil {
		return types.Workbench{}, err
	}
	// Dataplane is not supported for update but is for create. So we have to
	// resolve the dataplane id here instead of in the common implementation
	// function.
	dataplaneID := v.GetString(constants.DataplaneFlag)
	wkbnch.DataplaneID, err = workbench.GetValidDataplaneID(ctx, svc, dataplaneID)
	if err != nil {
		return types.Workbench{}, err
	}
	if wkbnch.State == "" {
		wkbnch.State = constants.DesiredStateStopped
	}
	if wkbnch.Name == "" {
		return types.Workbench{}, errors.New(
			i18n.Format(
				i18n.FlagIsRequired,
				i18n.A{constants.I18nArgKeyFlagName: constants.NameFlag},
			),
		)
	}
	if wkbnch.ApplicationID == "" && len(wkbnch.Services) == 0 {
		return types.Workbench{}, errors.New(i18n.Format(i18n.EitherApplicationOrServiceIsRequired))
	}
	if wkbnch.Nodepool == "" {
		return types.Workbench{}, errors.New(
			i18n.Format(
				i18n.FlagIsRequired,
				i18n.A{constants.I18nArgKeyFlagName: constants.NodepoolFlag},
			),
		)
	}
	if len(wkbnch.Workspaces) == 0 {
		_ui.Warn(i18n.Format(i18n.WarningNoVolumeSpecified))
	}
	return wkbnch, nil
}

func resolveUpdateWorkbenchIDs(
	args []string,
	body types.Workbench,
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

func makeUpdateWorkbenchPayload(
	ids []string,
	body types.Workbench,
) map[string]interface{} {
	return map[string]interface{}{
		constants.RPCBodyKeyIDs:       ids,
		constants.RPCBodyKeyWorkbench: body,
	}
}

func setWorkspaceRequests(
	wkbnch *types.Workbench,
	workspaceRequests []types.HomeDirAwareWorkspaceRequest,
) {
	defaultMountName := "" // the first non-home workspace
	for _, workspaceRequest := range workspaceRequests {
		wkbnch.Workspaces = append(wkbnch.Workspaces, workspaceRequest.WorkspaceRequest)
		if !workspaceRequest.HomeDir && defaultMountName == "" {
			defaultMountName = workspaceRequest.MountName
		}
	}
	if defaultMountName != "" {
		wkbnch.Services = workbench.AddEnvsToServiceRequests(
			[]string{"DEFAULT_MOUNTNAME=" + defaultMountName},
			wkbnch.Services,
		)
	}
}
