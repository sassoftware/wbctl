// Copyright © 2026, SAS Institute Inc., Cary, NC, USA.  All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package auth

import (
	"context"

	"github.com/sassoftware/wbctl/internal/auth"
	"github.com/sassoftware/wbctl/internal/constants"
	"github.com/sassoftware/wbctl/internal/ui"
	"github.com/spf13/afero"
	"github.com/spf13/viper"
)

// GetAccessTokens returns the access token and CSRF token for the current CLI context.
func GetAccessTokens(
	ctx context.Context,
	v *viper.Viper,
	_ui ui.UI,
	fsys afero.Fs,
) (string, string, error) {
	authMode := v.GetString(constants.AuthModeFlag)
	endpoint := v.GetString(constants.EndpointFlag)
	clientID := v.GetString(constants.ClientIDFlag)
	clientSecret := v.GetString(constants.ClientSecretFlag)
	ingressHost := v.GetString(constants.IngressHostFlag)
	organization := v.GetString(constants.OrganizationFlag)
	ignoreTokenCache := v.GetBool(constants.IgnoreTokenCacheFlag)
	noListen := v.GetBool(constants.NoListenFlag)
	accessToken, err := auth.GetAccessToken(
		ctx,
		_ui,
		fsys,
		v,
		auth.GetAccessTokenParams{
			ControlPlaneEndpoint: endpoint,
			IngressHost:          ingressHost,
			Organization:         organization,
			AuthMode:             authMode,
			ClientID:             clientID,
			ClientSecret:         clientSecret,
			IgnoreTokenCache:     ignoreTokenCache,
			NoListen:             noListen,
		},
	)
	if err != nil {
		return "", "", err
	}
	csrfToken, err := auth.GetCSRFToken(ctx, endpoint, accessToken)
	return accessToken, csrfToken, err
}
