// Copyright © 2026, SAS Institute Inc., Cary, NC, USA.  All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package workbench

import (
	"context"
	"maps"
	"net/url"
	"slices"

	"github.com/sas-institute-rnd-saas/wbctl/internal/constants"
	"github.com/sas-institute-rnd-saas/wbctl/internal/postgrest"
	"github.com/sas-institute-rnd-saas/wbctl/internal/types"
	"github.com/sas-institute-rnd-saas/wbctl/internal/ui"
	"github.com/sassoftware/sas-ggdk/pkg/jsonutils"
	"github.com/sassoftware/sas-ggdk/pkg/result"
)

// GetSettings returns the Settings currently present in the Workbench Service.
func GetSettings(
	ctx context.Context,
	s *WorkbenchService,
	userID string,
	tenant bool,
) (types.Settings, error) {
	endpoint := constants.WorkbenchEndpointWorkspaceGetSettings
	query := url.Values{}
	if userID != "" {
		endpoint = constants.WorkbenchEndpointWorkspaceGetUserSettings
		query[constants.RPCBodyKeyUserID] = []string{userID}
	}
	if tenant {
		endpoint = constants.WorkbenchEndpointWorkspaceGetTenantSettings
	}
	return result.Split(Get[types.Settings](ctx, s, postgrest.Endpoint(endpoint), query))
}

// SetSetting sets the given setting to the given value.
func SetSetting(
	ctx context.Context,
	s *WorkbenchService,
	userID string,
	name string,
	value any,
	tenant bool,
) (bool, error) {
	body := map[string]interface{}{
		constants.RPCBodyKeyName:  name,
		constants.RPCBodyKeyValue: value,
	}
	endpoint := constants.WorkbenchEndpointWorkspaceSetUserSetting
	if userID != "" {
		body[constants.RPCBodyKeyUserID] = userID
	}
	if tenant {
		endpoint = constants.WorkbenchEndpointWorkspaceSetTenantSetting
	}
	return result.Split(Post[bool](
		ctx,
		s,
		postgrest.Endpoint(endpoint),
		body))
}

// PrintSettings returns a function that prints the given Settings.
func PrintSettings(_ui ui.UI, resources types.Settings) error {
	keys := slices.Collect(maps.Keys(resources))
	slices.Sort(keys)
	tbl := _ui.Table([]string{
		constants.TableLabelName,
		constants.TableLabelValue,
	})
	for _, name := range keys {
		value := resources[name]
		valueJSON := jsonutils.ToJSON(value)
		if valueJSON.IsError() {
			return valueJSON.Error()
		}
		tbl.Add(
			name,
			valueJSON.MustGet(),
		)
	}
	tbl.Print()
	return nil
}
