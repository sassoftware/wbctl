// Copyright © 2026, SAS Institute Inc., Cary, NC, USA.  All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package cancel

import (
	"context"
	"errors"
	"strings"

	"github.com/sassoftware/sas-ggdk/pkg/result"
	"github.com/sassoftware/wbctl/internal/cmd/auth"
	cmdtypes "github.com/sassoftware/wbctl/internal/cmd/types"
	"github.com/sassoftware/wbctl/internal/constants"
	"github.com/sassoftware/wbctl/internal/i18n"
	"github.com/sassoftware/wbctl/internal/postgrest"
	"github.com/sassoftware/wbctl/internal/services/workbench"
	"github.com/sassoftware/wbctl/internal/types"
	"github.com/sassoftware/wbctl/internal/ui"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// SupportedResourceTypes returns the resource types supported by the cancel
// command.
func SupportedResourceTypes() []string {
	return types.SupportedResourceTypes(cancelFunctions)
}

type cancelF func(
	context.Context,
	[]string,
	ui.UI,
	*workbench.WorkbenchService,
) error

var cancelFunctions = []types.FunctionMetadata[cancelF]{
	{
		Singular: constants.JobSubmissionType,
		Plural:   constants.JobSubmissionsType,
		F:        cancelJobSubmissions,
	},
}

// newCancelFunction creates a new cli.ActionFunc for the cancel command.
func newCancelFunction(
	ctx context.Context,
	fsys afero.Fs,
	_ui ui.UI,
	resourceType string,
) cmdtypes.ActionFunc {
	return func(
		_ *cobra.Command,
		args []string,
	) error {
		canceler, ok := types.GetFunction[cancelF](resourceType, cancelFunctions)
		if !ok {
			supportedTypes := types.SupportedResourceTypes(cancelFunctions)
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
		ids := args
		if len(ids) == 0 {
			return errors.New(
				i18n.Format(
					i18n.AtLeastOneResourceIDForCommand,
					i18n.A{constants.I18nArgKeyCommandName: constants.CancelCommandName},
				),
			)
		}
		v := viper.GetViper()
		endpoint := v.GetString(constants.EndpointFlag)
		accessToken, csrfToken, err := auth.GetAccessTokens(ctx, v, _ui, fsys)
		if err != nil {
			return err
		}
		svc := workbench.NewWorkbenchService(endpoint, accessToken, csrfToken)
		return canceler(ctx, ids, _ui, svc)
	}
}

func cancelJobSubmissions(
	ctx context.Context,
	ids []string,
	_ui ui.UI,
	svc *workbench.WorkbenchService,
) error {
	jobSubmissions, err := workbench.GetJobSubmissions(ctx, svc, ids, "")
	if err != nil {
		return err
	}
	resolvedIDs := getIDs(jobSubmissions)
	canceledIDs, err := bulkCancel(ctx, svc, resolvedIDs)
	if err != nil {
		return err
	}
	for _, id := range canceledIDs {
		_ui.Display(id)
	}
	return nil
}

func bulkCancel(
	ctx context.Context,
	svc *workbench.WorkbenchService,
	ids []string,
) ([]string, error) {
	return result.Split(workbench.Post[[]string](
		ctx,
		svc,
		postgrest.Endpoint(constants.WorkbenchEndpointRPCJobRunBulkCancel),
		map[string]interface{}{
			"ids": ids,
		}))
}

func getIDs(jobSubmissions types.JobSubmissions) []string {
	ids := make([]string, 0, len(jobSubmissions))
	for _, jr := range jobSubmissions {
		ids = append(ids, jr.ID)
	}
	return ids
}
