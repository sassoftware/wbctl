// Copyright © 2026, SAS Institute Inc., Cary, NC, USA.  All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package auth

import (
	"context"
	"net/http"
	"net/url"
	"strings"

	"github.com/sassoftware/sas-ggdk/pkg/errors"
	"github.com/sassoftware/sas-ggdk/pkg/httputils"
	"github.com/sassoftware/sas-ggdk/pkg/result"
	"github.com/sassoftware/wbctl/internal/constants"
	"github.com/sassoftware/wbctl/internal/i18n"
	"github.com/sassoftware/wbctl/internal/sliceutil"
	"github.com/sassoftware/wbctl/internal/ui"
	"github.com/spf13/afero"
)

// nolint:dupl
func getAccessTokenByPassword(
	ctx context.Context,
	_ui ui.UI,
	fsys afero.Fs,
	params GetAccessTokenParams,
) (string, error) {
	if params.Username == "" {
		return "", errors.New(i18n.Format(i18n.AuthModePasswordRequiresUsername))
	}
	var tokenResp tokenResponse
	var err error
	defer deferSaveTokenToCache(_ui, fsys, params.ControlPlaneEndpoint, &tokenResp, &err)
	cfg, err := getWellKnownConfig(ctx, params.ControlPlaneEndpoint)
	if err != nil {
		return "", errors.Wrap(err, i18n.Format(i18n.FailedToGetUsernamePasswordToken))
	}
	params.ClientID = sliceutil.Coalesce(params.ClientID, cfg.Workbench.NativeClientID)
	tokenResp, err = getAccessTokenForUsernameAndPassword(ctx, params.ClientID, params.Username, params.Password, cfg)
	if err != nil {
		return "", errors.Wrap(err, i18n.Format(i18n.FailedToGetUsernamePasswordToken))
	}
	accessToken, err := getAccessTokenFromTokenResponse(tokenResp)
	if err != nil {
		return "", errors.Wrap(err, i18n.Format(i18n.FailedToGetUsernamePasswordToken))
	}
	return accessToken, nil
}

func getAccessTokenForUsernameAndPassword(
	ctx context.Context,
	clientID string,
	username,
	password string,
	cfg wellKnownConfigResponse,
) (tokenResponse, error) {
	data := buildPasswordGrantValues(clientID, username, password, cfg.Workbench.SPAAudience)
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		cfg.TokenEndpoint,
		strings.NewReader(data.Encode()),
	)
	if err != nil {
		return tokenResponse{}, err
	}
	req.Header.Set(constants.ContentTypeHeaderKey, constants.FormContentType)
	return result.Split(httputils.DoAsJSON[tokenResponse](defaultClient, req))
}

func buildPasswordGrantValues(
	clientID string,
	username string,
	password string,
	audience string,
) url.Values {
	data := url.Values{}
	data.Set(constants.QueryParameterGrantType, constants.OAuthPasswordGrantType)
	data.Set(constants.QueryParameterAudience, sliceutil.Coalesce(audience, constants.OAuthAudience))
	data.Set(constants.QueryParameterClientID, clientID)
	data.Set(constants.QueryParameterUsername, username)
	data.Set(constants.QueryParameterPassword, password)
	data.Set(constants.QueryParameterScope, constants.OAuthScope)
	return data
}
