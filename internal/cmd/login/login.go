// Copyright © 2026, SAS Institute Inc., Cary, NC, USA.  All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package login

import (
	"context"

	"github.com/sas-institute-rnd-saas/wbctl/internal/auth"
	cmdtypes "github.com/sas-institute-rnd-saas/wbctl/internal/cmd/types"
	"github.com/sas-institute-rnd-saas/wbctl/internal/constants"
	"github.com/sas-institute-rnd-saas/wbctl/internal/i18n"
	"github.com/sas-institute-rnd-saas/wbctl/internal/ui"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// newLoginFunction creates a command action that authenticates and caches an
// access token. Any required value not supplied via flags is prompted for
// interactively.
func newLoginFunction(
	ctx context.Context,
	fsys afero.Fs,
	_ui ui.UI,
) cmdtypes.ActionFunc {
	return func(
		_ *cobra.Command,
		_ []string,
	) error {
		v := viper.GetViper()
		authMode := ""
		if v.IsSet(constants.AuthModeFlag) {
			authMode = v.GetString(constants.AuthModeFlag)
		}
		endpoint := v.GetString(constants.EndpointFlag)
		password := v.GetString(constants.PasswordFlag)
		username := v.GetString(constants.UsernameFlag)
		clientID := v.GetString(constants.ClientIDFlag)
		clientSecret := v.GetString(constants.ClientSecretFlag)
		ingressHost := v.GetString(constants.IngressHostFlag)
		organization := v.GetString(constants.OrganizationFlag)
		ignoreTokenCache := v.GetBool(constants.IgnoreTokenCacheFlag)
		noListen := v.GetBool(constants.NoListenFlag)
		params := auth.GetAccessTokenParams{
			ControlPlaneEndpoint: endpoint,
			IngressHost:          ingressHost,
			Organization:         organization,
			AuthMode:             authMode,
			Username:             username,
			Password:             password,
			ClientID:             clientID,
			ClientSecret:         clientSecret,
			IgnoreTokenCache:     ignoreTokenCache,
			NoListen:             noListen,
		}
		params = promptForMissing(_ui, params)
		// Call GetAccessToken to log in and cache the access token. We don't
		// need the actual result.
		_, err := auth.GetAccessToken(
			ctx,
			_ui,
			fsys,
			params,
		)
		return err
	}
}

// promptForMissing interactively fills in any blank credential values before
// the auth call is made.
// nolint:cyclop
func promptForMissing(
	_ui ui.UI,
	params auth.GetAccessTokenParams,
) auth.GetAccessTokenParams {
	if params.AuthMode == "" {
		params.AuthMode = _ui.Ask(i18n.Format(i18n.LoginAuthModePrompt))
	}
	if params.AuthMode == constants.AuthModeClientCredentials && params.ClientID == "" {
		params.ClientID = _ui.Ask(i18n.Format(i18n.LoginClientIDPrompt))
	}
	if params.ControlPlaneEndpoint == "" {
		params.ControlPlaneEndpoint = _ui.Ask(i18n.Format(i18n.LoginEndpointPrompt))
	}

	params.AuthMode = params.NormalizedAuthMode()

	switch params.AuthMode {
	case constants.AuthModePassword:
		if params.Username == "" {
			params.Username = _ui.Ask(i18n.Format(i18n.LoginUsernamePrompt))
		}
		if params.Password == "" {
			params.Password = _ui.AskForPassword(i18n.Format(i18n.LoginPasswordPrompt))
		}
	case constants.AuthModeClientCredentials:
		if params.ClientSecret == "" {
			params.ClientSecret = _ui.AskForPassword(i18n.Format(i18n.LoginClientSecretPrompt))
		}
	}

	return params
}
