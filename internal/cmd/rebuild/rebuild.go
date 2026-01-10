// Copyright © 2026, SAS Institute Inc., Cary, NC, USA.  All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package rebuild

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

// SupportedResourceTypes returns the resource types supported by the rebuild
// command.
func SupportedResourceTypes() []string {
	return types.SupportedResourceTypes(rebuildFunctions)
}

type rebuildF func(
	context.Context,
	ui.UI,
	[]string,
	*workbench.WorkbenchService,
) error

var rebuildFunctions = []types.FunctionMetadata[rebuildF]{
	{
		Singular: constants.ImageType,
		Plural:   constants.ImagesType,
		F:        rebuildImage,
	},
}

// newRebuildFunction creates a new cli.ActionFunc for the stop command.
func newRebuildFunction(
	ctx context.Context,
	fsys afero.Fs,
	_ui ui.UI,
	resourceType string,
) cmdtypes.ActionFunc {
	return func(
		_ *cobra.Command,
		args []string,
	) error {
		ids := args
		if len(ids) == 0 {
			return errors.New(
				i18n.Format(i18n.AtLeastOneResourceIDForCommand,
					i18n.A{constants.I18nArgKeyCommandName: constants.RebuildCommandName},
				),
			)
		}
		rebuilder, ok := types.GetFunction(resourceType, rebuildFunctions)
		if !ok {
			supportedTypes := types.SupportedResourceTypes(rebuildFunctions)
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
		return rebuilder(ctx, _ui, ids, svc)
	}
}

func rebuildImage(
	ctx context.Context,
	_ui ui.UI,
	ids []string,
	svc *workbench.WorkbenchService,
) error {
	images, err := workbench.GetImages(ctx, svc, ids, "")
	if err != nil {
		return err
	}
	resolvedIDs := getIDs(images)
	rebuiltIDs, err := rebuild(ctx, svc, resolvedIDs)
	if err != nil {
		return err
	}
	for _, id := range rebuiltIDs {
		_ui.Display(id)
	}
	return nil
}

func rebuild(
	ctx context.Context,
	svc *workbench.WorkbenchService,
	ids []string,
) ([]string, error) {
	for _, id := range ids {
		_, err := workbench.RebuildImage(ctx, svc, id)
		if err != nil {
			return nil, err
		}
	}
	return ids, nil
}

func getIDs(images types.Images) []string {
	ids := make([]string, 0, len(images))
	for _, image := range images {
		ids = append(ids, image.ID)
	}
	return ids
}
