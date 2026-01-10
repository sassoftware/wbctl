// Copyright © 2026, SAS Institute Inc., Cary, NC, USA.  All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package settings

import (
	"context"
	"encoding/json"
	"strconv"

	"github.com/sas-institute-rnd-saas/wbctl/internal/cmd/auth"
	cmdtypes "github.com/sas-institute-rnd-saas/wbctl/internal/cmd/types"
	"github.com/sas-institute-rnd-saas/wbctl/internal/constants"
	"github.com/sas-institute-rnd-saas/wbctl/internal/services/workbench"
	"github.com/sas-institute-rnd-saas/wbctl/internal/ui"
	"github.com/sassoftware/sas-ggdk/pkg/jsonutils"
	"github.com/sassoftware/sas-ggdk/pkg/result"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// newSettingsFunction creates a new cli.ActionFunc for the setting command.
func newSettingsFunction(
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
		set := v.GetString(constants.SetFlag)
		value := v.GetString(constants.ValueFlag)
		userID := v.GetString(constants.UserIDFlag)
		tenant := v.GetBool(constants.TenantFlag)
		if set == "" {
			resources, err := workbench.GetSettings(ctx, svc, userID, tenant)
			if err != nil {
				return err
			}
			if constants.JSONOutputValue == v.GetString(constants.OutputFlag) {
				response, err := result.Split(jsonutils.ToJSON(resources))
				if err == nil {
					_ui.Display(response)
				}
				return err
			}
			return workbench.PrintSettings(_ui, resources)
		}
		valueAny := fixType(value)
		success, err := workbench.SetSetting(ctx, svc, userID, set, valueAny, tenant)
		if err != nil {
			return err
		}
		if success {
			_ui.Display(set)
		}
		return nil
	}
}

func fixType(value string) interface{} {
	// If it parses as an int then return that
	i, err := strconv.ParseInt(value, 10, 64)
	if err == nil {
		return i
	}
	// If it parses as a bool then return that
	b, err := strconv.ParseBool(value)
	if err == nil {
		return b
	}
	// If it parses as a JSON object then return that
	o := map[string]interface{}{}
	err = json.Unmarshal([]byte(value), &o)
	if err == nil {
		return o
	}
	// If it parses as a JSON array then return that
	a := []interface{}{}
	err = json.Unmarshal([]byte(value), &a)
	if err == nil {
		return a
	}
	// Otherwise return it as is
	return value
}
