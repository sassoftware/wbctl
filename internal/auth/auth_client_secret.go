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
	"github.com/sas-institute-rnd-saas/wbctl/internal/sliceutil"
	"github.com/sas-institute-rnd-saas/wbctl/internal/ui"
	"github.com/sassoftware/sas-ggdk/pkg/errors"
	"github.com/sassoftware/sas-ggdk/pkg/httputils"
	"github.com/sassoftware/sas-ggdk/pkg/result"
	"github.com/spf13/afero"
)

// nolint:dupl
func getAccessTokenByClientSecret(
	ctx context.Context,
	_ui ui.UI,
	fsys afero.Fs,
	params GetAccessTokenParams,
) (string, error) {
	if params.ClientID == "" {
		return "", errors.New(i18n.Format(i18n.ClientIDMustBeSpecified))
	}
	var tokenResp tokenResponse
	var err error
	defer deferSaveTokenToCache(_ui, fsys, params.ControlPlaneEndpoint, &tokenResp, &err)
	cfg, err := getWellKnownConfig(ctx, params.ControlPlaneEndpoint)
	if err != nil {
		return "", errors.Wrap(err, i18n.Format(i18n.FailedToGetClientSecretToken))
	}
	tokenResp, err = getAccessTokenForClientIDAndSecret(ctx, params.ClientID, params.ClientSecret, cfg)
	if err != nil {
		return "", errors.Wrap(err, i18n.Format(i18n.FailedToGetClientSecretToken))
	}
	accessToken, err := getAccessTokenFromTokenResponse(tokenResp)
	if err != nil {
		return "", errors.Wrap(err, i18n.Format(i18n.FailedToGetClientSecretToken))
	}
	return accessToken, nil
}

func getAccessTokenForClientIDAndSecret(
	ctx context.Context,
	clientID,
	secret string,
	cfg wellKnownConfigResponse,
) (tokenResponse, error) {
	data := buildClientSecretGrantValues(clientID, secret, cfg.Workbench.SPAAudience)
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

func buildClientSecretGrantValues(
	clientID string,
	secret string,
	audience string,
) url.Values {
	data := url.Values{}
	data.Set(constants.QueryParameterGrantType, constants.OAuthClientCredentialsGrantType)
	data.Set(constants.QueryParameterAudience, sliceutil.Coalesce(audience, constants.OAuthAudience))
	data.Set(constants.QueryParameterClientID, clientID)
	data.Set(constants.QueryParameterClientSecret, secret)
	return data
}
