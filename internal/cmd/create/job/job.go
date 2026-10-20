// Copyright © 2026, SAS Institute Inc., Cary, NC, USA.  All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package job

import (
	"context"
	"errors"
	"strings"

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

// CreateCommand returns the create job Command.
func CreateCommand(
	ctx context.Context,
	fsys afero.Fs,
	_ui ui.UI,
) *cobra.Command {
	cmd := &cobra.Command{
		Use:     constants.JobCommandName,
		Aliases: []string{constants.JobsCommandName},
		Args:    cobra.NoArgs,
		RunE:    newCreateJobFunction(ctx, fsys, _ui),
		Short: i18n.Format(
			i18n.CmdCreateTypeUsage,
			i18n.A{constants.I18nArgTypeKey: cmdtypes.Titler.String(constants.JobType)},
		),
		Long: i18n.Format(
			i18n.CmdCreateTypeDescription,
			i18n.A{constants.I18nArgTypeKey: cmdtypes.Titler.String(constants.JobType)},
		),
		PreRun: flags.BindToViper,
	}
	flags.AddDryRunFlag(cmd)
	flags.AddFilenameFlag(cmd, constants.JobType)
	flags.AddDataplaneFlag(cmd, constants.JobType)
	flags.AddNodepoolFlag(cmd, constants.JobType)
	flags.AddNameFlag(cmd, constants.JobType)
	flags.AddDescriptionFlag(cmd, constants.JobType)
	flags.AddContextFlag(cmd, constants.JobType)
	flags.AddLabelFlag(cmd, constants.JobType)
	flags.AddCommandFlag(cmd)
	flags.AddSubPathFlag(cmd, constants.JobType)
	flags.AddScheduleFlag(cmd)
	flags.AddVolumeFlag(cmd, constants.JobType)
	flags.AddApplicationFlag(cmd, constants.JobType)
	flags.AddServiceFlag(cmd, constants.JobType)
	flags.AddEnvironmentFlag(cmd, constants.JobType)
	return cmd
}

// UpdateCommand returns the update job Command.
func UpdateCommand(
	ctx context.Context,
	fsys afero.Fs,
	_ui ui.UI,
) *cobra.Command {
	cmd := &cobra.Command{
		Use:     constants.JobsCommandName,
		Aliases: []string{constants.JobCommandName},
		Args:    cobra.ArbitraryArgs,
		RunE:    newUpdateJobFunction(ctx, fsys, _ui),
		Short: i18n.Format(
			i18n.CmdUpdateTypeUsage,
			i18n.A{constants.I18nArgTypeKey: cmdtypes.Titler.String(constants.JobType)},
		),
		Long: i18n.Format(
			i18n.CmdUpdateTypeDescription,
			i18n.A{constants.I18nArgTypeKey: cmdtypes.Titler.String(constants.JobType)},
		),
		PreRun: flags.BindToViper,
	}
	flags.AddDryRunFlag(cmd)
	flags.AddFilenameFlag(cmd, constants.JobType)
	flags.AddNodepoolFlag(cmd, constants.JobType)
	flags.AddNameFlag(cmd, constants.JobType)
	flags.AddDescriptionFlag(cmd, constants.JobType)
	flags.AddContextFlag(cmd, constants.JobType)
	flags.AddLabelFlag(cmd, constants.JobType)
	flags.AddCommandFlag(cmd)
	flags.AddSubPathFlag(cmd, constants.JobType)
	flags.AddScheduleFlag(cmd)
	flags.AddVolumeFlag(cmd, constants.JobType)
	flags.AddApplicationFlag(cmd, constants.JobType)
	flags.AddServiceFlag(cmd, constants.JobType)
	flags.AddEnvironmentFlag(cmd, constants.JobType)
	return cmd
}

// newCreateJobFunction creates a new cli.ActionFunc for the `create job`
// command.
func newCreateJobFunction(
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
		body, err := getCreateJobRequestBody(ctx, v, fsys, _ui, svc)
		if err != nil {
			return err
		}
		dryRun := v.GetBool(constants.DryRunFlag)
		if dryRun {
			response, err := wbctljson.MakeMarshaler[types.Job](wbctljson.EditOptions)(body)
			if err == nil {
				_ui.Display(string(response))
			}
			return err
		}
		res, err := workbench.CreateJob(ctx, svc, body)
		if err == nil {
			_ui.Display(res.ID)
		}
		return err
	}
}

// newUpdateJobFunction creates a new cli.ActionFunc for the `update job`
// command.
func newUpdateJobFunction(
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
		body, err := getJobRequestBody(ctx, v, fsys, svc)
		if err != nil {
			return err
		}
		ids, err := resolveUpdateJobIDs(args, body)
		if err != nil {
			return err
		}

		dryRun := v.GetBool(constants.DryRunFlag)
		if dryRun {
			payload := makeUpdateJobPayload(ids, body)
			response, err := wbctljson.MakeMarshaler[map[string]any](wbctljson.EditOptions)(payload)
			if err == nil {
				_ui.Display(string(response))
			}
			return err
		}
		res, err := workbench.BulkUpdateJobs(ctx, svc, ids, body)
		if err == nil {
			for _, resource := range res {
				_ui.Display(resource.ID)
			}
		}
		return err
	}
}

func getJobRequestBody(
	ctx context.Context,
	v *viper.Viper,
	fsys afero.Fs,
	svc *workbench.WorkbenchService,
) (types.Job, error) {
	filename := v.GetString(constants.FilenameFlag)
	if filename != "" {
		return jsonutil.GetFromFileOrStdin[types.Job](fsys, filename)
	}
	var err error
	name := v.GetString(constants.NameFlag)
	description := v.GetString(constants.DescriptionFlag)
	jobContext := v.GetString(constants.ContextFlag)
	command := v.GetString(constants.CommandFlag)
	subPath := v.GetString(constants.SubPathFlag)
	schedule := v.GetString(constants.ScheduleFlag)
	application := v.GetString(constants.ApplicationFlag)
	nodepool := v.GetString(constants.NodepoolFlag)
	labels := v.GetStringSlice(constants.LabelFlag)
	volumes := v.GetStringSlice(constants.VolumeFlag)
	if len(volumes) == 0 {
		volumes = v.GetStringSlice(constants.WorkspaceFlag)
	}
	services := v.GetStringSlice(constants.ServiceFlag)
	envs := v.GetStringSlice(constants.EnvironmentFlag)

	workspaceRequests, err := workbench.GetValidWorkspaceRequests(ctx, svc, volumes)
	if err != nil {
		return types.Job{}, err
	}
	serviceRequests, err := workbench.GetServiceRequests(ctx, svc, application, services)
	if err != nil {
		return types.Job{}, err
	}
	serviceRequests = workbench.AddEnvsToServiceRequests(envs, serviceRequests)
	job := types.Job{
		Name:          name,
		ApplicationID: application,
		Description:   description,
		Context:       jobContext,
		Labels:        maputil.FromEqualDelimitedLines(labels),
		Spec:          getSpec(command),
		SubPath:       subPath,
		Schedule:      schedule,
		Nodepool:      nodepool,
		Workspaces:    make([]types.WorkspaceRequest, 0),
		Services:      serviceRequests,
	}
	setWorkspaceRequests(&job, workspaceRequests)
	properties := map[string]interface{}{
		"isWorkbenchAppCreated": true,
	}
	if job.Schedule != "" {
		properties["schedule"] = "simple"
	}
	if len(job.Workspaces) > 1 {
		properties["primaryVolume"] = map[string]interface{}{
			"id":            workspaceRequests[0].DbID,
			"controller_id": workspaceRequests[0].ID,
			"folderName":    workspaceRequests[0].MountName,
		}
	}
	job.Properties = properties
	return job, nil
}

func getCreateJobRequestBody(
	ctx context.Context,
	v *viper.Viper,
	fsys afero.Fs,
	_ui ui.UI,
	svc *workbench.WorkbenchService,
) (types.Job, error) {
	job, err := getJobRequestBody(ctx, v, fsys, svc)
	if err != nil {
		return types.Job{}, err
	}
	// Dataplane is not supported for update but is for create. So we have to
	// resolve the dataplane id here instead of in the common implementation
	// function.
	dataplaneID := v.GetString(constants.DataplaneFlag)
	job.DataplaneID, err = workbench.GetValidDataplaneID(ctx, svc, dataplaneID)
	if err != nil {
		return types.Job{}, err
	}
	if job.Name == "" {
		return types.Job{},
			errors.New(
				i18n.Format(
					i18n.FlagIsRequired,
					i18n.A{constants.I18nArgKeyFlagName: constants.NameFlag},
				),
			)
	}
	if job.Context == "" && job.Spec == nil {
		return types.Job{}, errors.New(i18n.Format(i18n.EitherFlagNamesAreRequired, i18n.A{
			constants.I18nArgKeyFlagNameOne: constants.ContextFlag,
			constants.I18nArgKeyFlagNameTwo: constants.CommandFlag,
		}))
	}
	if job.ApplicationID == "" && len(job.Services) == 0 {
		validApplicationValues, _ := workbench.GetApplicationIDs(ctx, svc)
		validServiceValues, _ := workbench.GetServiceIDs(ctx, svc)
		return types.Job{}, errors.New(i18n.Format(i18n.EitherApplicationOrServiceIsRequired,
			i18n.A{
				constants.I18nArgKeyApplications: strings.Join(validApplicationValues, ", "),
				constants.I18nArgKeyServices:     strings.Join(validServiceValues, ", "),
			},
		))
	}
	if job.Nodepool == "" {
		validNodepools, _ := workbench.GetNodepoolIDs(ctx, svc)
		return types.Job{}, errors.New(
			i18n.Format(
				i18n.FlagNodepoolIsRequired,
				i18n.A{constants.I18nArgKeyNodepools: strings.Join(validNodepools, ", ")},
			),
		)
	}
	if len(job.Workspaces) == 0 {
		_ui.Warn(i18n.Format(i18n.WarningNoVolumeSpecified))
	}
	return job, nil
}

func setWorkspaceRequests(
	job *types.Job,
	workspaceRequests []types.WorkspaceRequestWithMetadata,
) {
	defaultMountName := "" // the first non-home workspace
	for _, workspaceRequest := range workspaceRequests {
		job.Workspaces = append(job.Workspaces, workspaceRequest.WorkspaceRequest)
		if !workspaceRequest.HomeDir && defaultMountName == "" {
			defaultMountName = workspaceRequest.MountName
		}
	}
	if defaultMountName != "" {
		job.Services = workbench.AddEnvsToServiceRequests([]string{"DEFAULT_MOUNTNAME=" + defaultMountName}, job.Services)
	}
}

func getSpec(command string) map[string]interface{} {
	if command == "" {
		return nil
	}
	return map[string]interface{}{
		"version": "3",
		"tasks": map[string]interface{}{
			"default": map[string]interface{}{
				"cmds": []string{command},
			},
		},
	}
}

func resolveUpdateJobIDs(
	args []string,
	body types.Job,
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

func makeUpdateJobPayload(
	ids []string,
	body types.Job,
) map[string]interface{} {
	return map[string]interface{}{
		constants.RPCBodyKeyIDs:   ids,
		constants.RPCBodyKeyParam: body,
	}
}
