// Copyright © 2026, SAS Institute Inc., Cary, NC, USA.  All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Package delete provides the delete command.
// This package name collides with the builtin delete function but context is
// enough to distinguish it.
// nolint: predeclared
package delete

import (
	"context"
	"errors"
	"strings"

	"github.com/sassoftware/wbctl/internal/cmd/auth"
	cmdtypes "github.com/sassoftware/wbctl/internal/cmd/types"
	"github.com/sassoftware/wbctl/internal/constants"
	"github.com/sassoftware/wbctl/internal/i18n"
	"github.com/sassoftware/wbctl/internal/services/workbench"
	"github.com/sassoftware/wbctl/internal/types"
	"github.com/sassoftware/wbctl/internal/ui"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// SupportedResourceTypes returns the resource types supported by the delete
// command.
func SupportedResourceTypes() []string {
	return types.SupportedResourceTypes(deleteFunctions)
}

type deleteF func(
	context.Context,
	[]string,
	ui.UI,
	*workbench.WorkbenchService,
) error

var deleteFunctions = []types.FunctionMetadata[deleteF]{
	{
		Singular: constants.ImageType,
		Plural:   constants.ImagesType,
		F:        getImpl(workbench.BulkDeleteImages),
	},
	{
		Singular: constants.JobType,
		Plural:   constants.JobsType,
		F:        getImpl(workbench.BulkDeleteJobs),
	},
	{
		Singular: constants.JobSubmissionType,
		Plural:   constants.JobSubmissionsType,
		F:        getImpl(workbench.BulkDeleteJobSubmissions),
	},
	{
		Singular: constants.SnapshotType,
		Plural:   constants.SnapshotsType,
		F:        getImpl(workbench.BulkDeleteSnapshots),
	},
	{
		Singular: constants.WorkbenchType,
		Plural:   constants.WorkbenchesType,
		F:        getImpl(workbench.BulkDeleteWorkbenches),
	},
	{
		Singular: constants.VolumeType,
		Plural:   constants.VolumesType,
		F:        getImpl(workbench.BulkDeleteWorkspaces),
	},
}

// newDeleteFunction creates a new cli.ActionFunc for the delete command.
func newDeleteFunction(
	ctx context.Context,
	fsys afero.Fs,
	_ui ui.UI,
	resourceType string,
) cmdtypes.ActionFunc {
	return func(
		_ *cobra.Command,
		args []string,
	) error {
		deleter, ok := types.GetFunction(resourceType, deleteFunctions)
		if !ok {
			supportedTypes := SupportedResourceTypes()
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
		return deleter(ctx, args, _ui, svc)
	}
}

type resourceDeleter func(
	ctx context.Context,
	s *workbench.WorkbenchService,
	id []string,
) ([]string, error)

func getImpl(
	deleter resourceDeleter,
) deleteF {
	return func(
		ctx context.Context,
		args []string,
		_ui ui.UI,
		svc *workbench.WorkbenchService,
	) error {
		ids := args
		if len(ids) == 0 {
			return errors.New(
				i18n.Format(
					i18n.AtLeastOneResourceIDForCommand,
					i18n.A{constants.I18nArgKeyCommandName: constants.DeleteCommandName},
				),
			)
		}
		deletedIDs, err := deleter(ctx, svc, ids)
		if err != nil {
			return err
		}
		_ui.Display(strings.Join(deletedIDs, "\n"))
		return nil
	}
}
