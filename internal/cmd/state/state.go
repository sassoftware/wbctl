// Copyright © 2026, SAS Institute Inc., Cary, NC, USA.  All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package state

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

// SupportedResourceTypes returns the resource types supported by the start and
// stop commands.
func SupportedResourceTypes() []string {
	return types.SupportedResourceTypes(stateFunctions)
}

type stateF func(
	context.Context,
	ui.UI,
	[]string,
	string,
	*workbench.WorkbenchService,
) error

var stateFunctions = []types.FunctionMetadata[stateF]{
	{
		Singular: constants.WorkbenchType,
		Plural:   constants.WorkbenchesType,
		F:        setWorkbenchDesiredState,
	},
}

func desiredStateToCommand(desiredState string) string {
	switch desiredState {
	case "RUNNING":
		return constants.StartCommandName
	default:
		return constants.StopCommandName
	}
}

func newUpdateStateFunction(
	ctx context.Context,
	fsys afero.Fs,
	_ui ui.UI,
	resourceType string,
	desiredState string,
) cmdtypes.ActionFunc {
	return func(
		_ *cobra.Command,
		args []string,
	) error {
		stater, ok := types.GetFunction(resourceType, stateFunctions)
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
		ids := args
		if len(ids) == 0 {
			return errors.New(
				i18n.Format(
					i18n.AtLeastOneResourceIDForCommand,
					i18n.A{constants.I18nArgKeyCommandName: desiredStateToCommand(desiredState)},
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
		return stater(ctx, _ui, ids, desiredState, svc)
	}
}

func setWorkbenchDesiredState(
	ctx context.Context,
	_ui ui.UI,
	ids []string,
	desiredState string,
	svc *workbench.WorkbenchService,
) error {
	workbenches, err := workbench.GetWorkbenches(ctx, svc, ids, "")
	if err != nil {
		return err
	}
	resolvedIDs := getIDs(workbenches)
	changedIDs, err := bulkUpdateState(ctx, svc, desiredState, resolvedIDs)
	if err != nil {
		return err
	}
	for _, id := range changedIDs {
		_ui.Display(id)
	}
	return nil
}

func bulkUpdateState(
	ctx context.Context,
	svc *workbench.WorkbenchService,
	desiredState string,
	ids []string,
) ([]string, error) {
	responses, err := workbench.BulkUpdateWorkbenches(ctx, svc, ids, types.Workbench{State: desiredState})
	if err != nil {
		return nil, err
	}
	return getIDs(responses), nil
}

func getIDs(wb types.Workbenches) []string {
	ids := make([]string, 0, len(wb))
	for _, w := range wb {
		ids = append(ids, w.ID)
	}
	return ids
}
