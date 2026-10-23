// Copyright © 2026, SAS Institute Inc., Cary, NC, USA.  All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package workbench

import (
	"context"
	"maps"
	"net/url"
	"slices"

	"github.com/sassoftware/sas-ggdk/pkg/jsonutils"
	"github.com/sassoftware/sas-ggdk/pkg/result"
	"github.com/sassoftware/wbctl/internal/constants"
	"github.com/sassoftware/wbctl/internal/logutil"
	"github.com/sassoftware/wbctl/internal/postgrest"
	"github.com/sassoftware/wbctl/internal/types"
	"github.com/sassoftware/wbctl/internal/ui"
)

// GetSettings returns the Settings currently present in the Workbench Service.
func GetSettings(
	ctx context.Context,
	s *WorkbenchService,
	userID string,
	tenant bool,
) (types.Settings, error) {
	logutil.GetLogger().Info("getting settings", "userID", userID, "tenant", tenant)
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
	logutil.GetLogger().Info("setting setting", "userid", userID, "name", name, "value", value, "tenant", tenant)
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
