// Copyright © 2026, SAS Institute Inc., Cary, NC, USA.  All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package describe

import (
	"context"
	"errors"
	"strings"

	"github.com/sas-institute-rnd-saas/wbctl/internal/cmd/auth"
	cmdtypes "github.com/sas-institute-rnd-saas/wbctl/internal/cmd/types"
	"github.com/sas-institute-rnd-saas/wbctl/internal/constants"
	"github.com/sas-institute-rnd-saas/wbctl/internal/i18n"
	"github.com/sas-institute-rnd-saas/wbctl/internal/services/workbench"
	"github.com/sas-institute-rnd-saas/wbctl/internal/types"
	"github.com/sas-institute-rnd-saas/wbctl/internal/ui"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// SupportedResourceTypes returns the resource types supported by the describe
// command.
func SupportedResourceTypes() []string {
	return types.SupportedResourceTypes(describeFunctions)
}

type describeF func(
	context.Context,
	[]string,
	ui.UI,
	*workbench.WorkbenchService,
) error

var describeFunctions = []types.FunctionMetadata[describeF]{
	{
		Singular: constants.DataplaneType,
		Plural:   constants.DataplanesType,
		F:        describeDataplanes,
	},
	{
		Singular: constants.ImageType,
		Plural:   constants.ImagesType,
		F:        describeImages,
	},
	{
		Singular: constants.ImageBuildType,
		Plural:   constants.ImageBuildsType,
		F:        describeImageBuilds,
	},
	{
		Singular: constants.JobType,
		Plural:   constants.JobsType,
		F:        describeJobs,
	},
	{
		Singular: constants.JobSubmissionType,
		Plural:   constants.JobSubmissionsType,
		F:        describeJobSubmissions,
	},
	{
		Singular: constants.SnapshotType,
		Plural:   constants.SnapshotsType,
		F:        describeSnapshots,
	},
	{
		Singular: constants.UserType,
		Plural:   constants.UsersType,
		F:        describeUsers,
	},
	{
		Singular: constants.WorkbenchType,
		Plural:   constants.WorkbenchesType,
		F:        describeWorkbenches,
	},
	{
		Singular: constants.VolumeType,
		Plural:   constants.VolumesType,
		F:        describeWorkspaces,
	},
	{
		Singular: constants.VolumeConflictType,
		Plural:   constants.VolumeConflictsType,
		F:        describeConflicts,
	},
}

// newDescribeFunction creates a new cli.ActionFunc for the describe command.
func newDescribeFunction(
	ctx context.Context,
	fsys afero.Fs,
	_ui ui.UI,
	resourceType string,
) cmdtypes.ActionFunc {
	return func(
		_ *cobra.Command,
		args []string,
	) error {
		describer, ok := types.GetFunction(resourceType, describeFunctions)
		if !ok {
			supportedTypes := types.SupportedResourceTypes(describeFunctions)
			msg := i18n.Format(i18n.ResourceTypeIsRequired)
			if resourceType != "" {
				msg = i18n.Format(
					i18n.UnknownResourceType,
					i18n.A{constants.I18nArgKeyType: resourceType},
				)
			}
			supportedTypesMsg := i18n.Format(
				i18n.SupportedTypes,
				i18n.A{constants.I18nArgKeyTypes: strings.Join(supportedTypes, "\n    ")},
			)
			return errors.New(msg + "\n  " + supportedTypesMsg)
		}
		v := viper.GetViper()
		endpoint := v.GetString(constants.EndpointFlag)
		accessToken, csrfToken, err := auth.GetAccessTokens(ctx, v, _ui, fsys)
		if err != nil {
			return err
		}
		svc := workbench.NewWorkbenchService(endpoint, accessToken, csrfToken)
		return describer(ctx, args, _ui, svc)
	}
}

func describeDataplanes(
	ctx context.Context,
	args []string,
	_ui ui.UI,
	svc *workbench.WorkbenchService,
) error {
	resources, err := workbench.GetDataplanes(ctx, svc, args, "")
	if err != nil {
		return err
	}
	return workbench.DescribeDataplanes(_ui, resources)
}

func describeWorkspaces(
	ctx context.Context,
	args []string,
	_ui ui.UI,
	svc *workbench.WorkbenchService,
) error {
	resources, err := workbench.GetWorkspaces(ctx, svc, args, "")
	if err != nil {
		return err
	}
	return workbench.DescribeWorkspaces(_ui, resources)
}

func describeWorkbenches(
	ctx context.Context,
	args []string,
	_ui ui.UI,
	svc *workbench.WorkbenchService,
) error {
	resources, err := workbench.GetWorkbenches(ctx, svc, args, "")
	if err != nil {
		return err
	}
	return workbench.DescribeWorkbenches(_ui, resources)
}

func describeImages(
	ctx context.Context,
	args []string,
	_ui ui.UI,
	svc *workbench.WorkbenchService,
) error {
	resources, err := workbench.GetImages(ctx, svc, args, "")
	if err != nil {
		return err
	}
	return workbench.DescribeImages(_ui, resources)
}

func describeImageBuilds(
	ctx context.Context,
	args []string,
	_ui ui.UI,
	svc *workbench.WorkbenchService,
) error {
	resources, err := workbench.GetImageBuilds(ctx, svc, args, "")
	if err != nil {
		return err
	}
	return workbench.DescribeImageBuilds(_ui, resources)
}

func describeJobs(
	ctx context.Context,
	args []string,
	_ui ui.UI,
	svc *workbench.WorkbenchService,
) error {
	resources, err := workbench.GetJobs(ctx, svc, args, "")
	if err != nil {
		return err
	}
	return workbench.DescribeJobs(_ui, resources)
}

func describeJobSubmissions(
	ctx context.Context,
	args []string,
	_ui ui.UI,
	svc *workbench.WorkbenchService,
) error {
	resources, err := workbench.GetJobSubmissions(ctx, svc, args, "")
	if err != nil {
		return err
	}
	return workbench.DescribeJobSubmissions(_ui, resources)
}

func describeSnapshots(
	ctx context.Context,
	args []string,
	_ui ui.UI,
	svc *workbench.WorkbenchService,
) error {
	resources, err := workbench.GetSnapshots(ctx, svc, args, "")
	if err != nil {
		return err
	}
	return workbench.DescribeSnapshots(_ui, resources)
}

func describeUsers(
	ctx context.Context,
	args []string,
	_ui ui.UI,
	svc *workbench.WorkbenchService,
) error {
	resources, err := workbench.GetUsers(ctx, svc, args, "")
	if err != nil {
		return err
	}
	return workbench.DescribeUsers(_ui, resources)
}

func describeConflicts(
	ctx context.Context,
	args []string,
	_ui ui.UI,
	svc *workbench.WorkbenchService,
) error {
	resources, err := workbench.GetConflicts(ctx, svc, args, "")
	if err != nil {
		return err
	}
	return workbench.DescribeConflicts(_ui, resources)
}
