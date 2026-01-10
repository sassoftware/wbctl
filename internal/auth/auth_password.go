// Copyright © 2026, SAS Institute Inc., Cary, NC, USA.  All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package auth

import (
	"context"
	"net/http"
	"net/url"
	"strings"

	"github.com/sas-institute-rnd-saas/wbctl/internal/constants"
	"github.com/sas-institute-rnd-saas/wbctl/internal/i18n"
	"github.com/sas-institute-rnd-saas/wbctl/internal/ui"
	"github.com/sassoftware/sas-ggdk/pkg/errors"
	"github.com/sassoftware/sas-ggdk/pkg/httputils"
	"github.com/sassoftware/sas-ggdk/pkg/result"
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
	data := buildPasswordGrantValues(clientID, username, password)
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
) url.Values {
	data := url.Values{}
	data.Set("grant_type", "password")
	data.Set("audience", constants.OAuthAudience)
	data.Set("client_id", clientID)
	data.Set("username", username)
	data.Set("password", password)
	data.Set("scope", constants.OAuthScope)
	return data
}
