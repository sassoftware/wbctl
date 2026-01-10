// Copyright © 2026, SAS Institute Inc., Cary, NC, USA.  All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package job_submission

import (
	"context"
	"errors"

	"github.com/sassoftware/sas-ggdk/pkg/jsonutils"
	"github.com/sassoftware/sas-ggdk/pkg/result"
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
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// CreateCommand returns the create job-submission command.
func CreateCommand(
	ctx context.Context,
	fsys afero.Fs,
	_ui ui.UI,
) *cobra.Command {
	cmd := &cobra.Command{
		Use:     constants.JobSubmissionCommandName,
		Aliases: []string{constants.JobSubmissionsCommandName},
		Args:    cobra.NoArgs,
		RunE:    newCreateJobSubmissionFunction(ctx, fsys, _ui),
		Short: i18n.Format(
			i18n.CmdCreateTypeUsage,
			i18n.A{constants.I18nArgTypeKey: cmdtypes.Titler.String(constants.JobSubmissionType)},
		),
		Long: i18n.Format(
			i18n.CmdCreateTypeDescription,
			i18n.A{constants.I18nArgTypeKey: cmdtypes.Titler.String(constants.JobSubmissionType)},
		),
		PreRun: flags.BindToViper,
	}
	flags.AddFilenameFlag(cmd, constants.JobSubmissionType)
	flags.AddDryRunFlag(cmd)
	flags.AddDataplaneFlag(cmd, constants.JobSubmissionType)
	flags.AddNodepoolFlag(cmd, constants.JobSubmissionType)
	flags.AddJobIDFlag(cmd)
	flags.AddLabelFlag(cmd, constants.JobSubmissionType)
	flags.AddEnvironmentFlag(cmd, constants.JobSubmissionType)
	flags.AddVolumeFlag(cmd, constants.JobSubmissionType)
	return cmd
}

// newCreateJobSubmissionFunction creates a new clitypes.ActionFunc for the
// create job command.
func newCreateJobSubmissionFunction(
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
		body, err := getCreateJobSubmissionRequestBody(ctx, v, fsys, svc)
		if err != nil {
			return err
		}
		dryRun := v.GetBool(constants.DryRunFlag)
		if dryRun {
			response, err := result.Split(jsonutils.ToJSON(body))
			if err == nil {
				_ui.Display(response)
			}
			return err
		}
		res, err := workbench.CreateJobSubmission(ctx, svc, body)
		if err == nil {
			_ui.Display(res.ID)
		}
		return err
	}
}

func getJobSubmissionRequestBody(
	ctx context.Context,
	v *viper.Viper,
	fsys afero.Fs,
	svc *workbench.WorkbenchService,
) (types.JobSubmission, error) {
	filename := v.GetString(constants.FilenameFlag)
	if filename != "" {
		return jsonutil.GetFromFileOrStdin[types.JobSubmission](fsys, filename)
	}
	dataplaneID := v.GetString(constants.DataplaneFlag)
	jobID := v.GetString(constants.JobIDFlag)
	nodepool := v.GetString(constants.NodepoolFlag)
	labels := v.GetStringSlice(constants.LabelFlag)
	volumes := v.GetStringSlice(constants.VolumeFlag)
	if len(volumes) == 0 {
		volumes = v.GetStringSlice(constants.WorkspaceFlag)
	}
	envs := v.GetStringSlice(constants.EnvironmentFlag)

	dataplaneID, err := workbench.GetValidDataplaneID(ctx, svc, dataplaneID)
	if err != nil {
		return types.JobSubmission{}, err
	}
	jobSubmission := types.JobSubmission{
		DataplaneID: dataplaneID,
		JobID:       jobID,
		Labels:      maputil.FromEqualDelimitedLines(labels),
		Env:         maputil.FromEqualDelimitedLines(envs),
		Nodepool:    nodepool,
		Workspaces:  make([]types.WorkspaceRequest, 0),
	}
	workspaceRequests, err := workbench.GetValidWorkspaceRequests(ctx, svc, volumes)
	if err != nil {
		return types.JobSubmission{}, err
	}
	setWorkspaceRequests(&jobSubmission, workspaceRequests)
	return jobSubmission, nil
}

func getCreateJobSubmissionRequestBody(
	ctx context.Context,
	v *viper.Viper,
	fsys afero.Fs,
	svc *workbench.WorkbenchService,
) (types.JobSubmission, error) {
	jobSubmission, err := getJobSubmissionRequestBody(ctx, v, fsys, svc)
	if err != nil {
		return types.JobSubmission{}, err
	}
	if jobSubmission.JobID == "" {
		return types.JobSubmission{}, errors.New(
			i18n.Format(
				i18n.FlagIsRequired,
				i18n.A{constants.I18nArgKeyFlagName: constants.JobIDFlag},
			),
		)
	}
	return jobSubmission, nil
}

func setWorkspaceRequests(
	jobSubmission *types.JobSubmission,
	workspaceRequests []types.HomeDirAwareWorkspaceRequest,
) {
	for _, workspaceRequest := range workspaceRequests {
		jobSubmission.Workspaces = append(jobSubmission.Workspaces, workspaceRequest.WorkspaceRequest)
	}
}
