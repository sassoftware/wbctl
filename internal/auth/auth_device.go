// Copyright © 2026, SAS Institute Inc., Cary, NC, USA.  All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package auth

import (
	"context"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/sas-institute-rnd-saas/wbctl/internal/constants"
	"github.com/sas-institute-rnd-saas/wbctl/internal/i18n"
	"github.com/sas-institute-rnd-saas/wbctl/internal/sliceutil"
	"github.com/sas-institute-rnd-saas/wbctl/internal/ui"
	"github.com/sassoftware/sas-ggdk/pkg/errors"
	"github.com/sassoftware/sas-ggdk/pkg/httputils"
	"github.com/sassoftware/sas-ggdk/pkg/result"
	"github.com/spf13/afero"
)

// nolint:cyclop
func getAccessTokenByDeviceGrant(
	ctx context.Context,
	_ui ui.UI,
	fsys afero.Fs,
	params GetAccessTokenParams,
) (string, error) {
	var tokenResp tokenResponse
	var accessToken string
	var err error
	defer deferSaveTokenToCache(_ui, fsys, params.ControlPlaneEndpoint, &tokenResp, &err)

	cfg, err := getWellKnownConfig(ctx, params.ControlPlaneEndpoint)
	if err != nil {
		return "", errors.Wrap(err, i18n.Format(i18n.FailedToGetDeviceToken))
	}
	params.ClientID = sliceutil.Coalesce(params.ClientID, cfg.Workbench.NativeClientID)

	// Check for a cached access token.
	if !params.IgnoreTokenCache {
		tokenResp, accessToken = getCachedOrRefreshedToken(
			ctx,
			_ui,
			params.ClientID,
			cfg,
			fsys,
			params.ControlPlaneEndpoint,
			params.IngressHost,
		)
		if accessToken != "" {
			return accessToken, nil
		}
	}

	// Fall back to device authorization
	deviceAuthResponse, err := startDeviceAuthorizationAtEndpoint(ctx, params.ClientID, cfg)
	if err != nil {
		return "", errors.Wrap(err, i18n.Format(i18n.FailedToGetDeviceToken))
	}
	tokenResp, err = authorizationPolling(ctx, _ui, params.ClientID, deviceAuthResponse, cfg)
	if err != nil {
		return "", errors.Wrap(err, i18n.Format(i18n.FailedToGetDeviceToken))
	}
	accessToken, err = getAccessTokenFromTokenResponse(tokenResp)
	if err != nil {
		return "", errors.Wrap(err, i18n.Format(i18n.FailedToGetDeviceToken))
	}
	return accessToken, nil
}

func startDeviceAuthorizationAtEndpoint(
	ctx context.Context,
	clientID string,
	cfg wellKnownConfigResponse,
) (deviceAuthorizationResponse, error) {
	data := buildDeviceGrantValues(clientID, cfg.Workbench.SPAAudience)
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		cfg.DeviceAuthorizationEndpoint,
		strings.NewReader(data.Encode()),
	)
	if err != nil {
		return deviceAuthorizationResponse{}, err
	}
	req.Header.Set(constants.ContentTypeHeaderKey, constants.FormContentType)
	return result.Split(httputils.DoAsJSON[deviceAuthorizationResponse](defaultClient, req))
}

func pollForDeviceTokenAuthorization(
	ctx context.Context,
	clientID string,
	device deviceAuthorizationResponse,
	cfg wellKnownConfigResponse,
) (tokenResponse, error) {
	interval := normalizePollInterval(device.Interval)
	deadline := time.Now().Add(time.Duration(device.ExpiresIn) * time.Second)
	for time.Now().Before(deadline) {
		// Check if the user has completed authorization and we have a token
		tokenResult, err := checkForAuthorizedDeviceToken(
			ctx,
			clientID,
			cfg.TokenEndpoint,
			device.DeviceCode,
			cfg.Workbench.SPAAudience,
		)
		if err == nil || isDeviceGrantErrorType(err, []string{
			"access_denied",
			"expired_token",
			"invalid_client",
			"invalid_grant",
		}) {
			return tokenResult, err
		}
		if isDeviceGrantErrorType(err, []string{"slow_down"}) {
			interval += 5 * time.Second
		}
		// Not yet, so wait
		select {
		case <-ctx.Done():
			return tokenResponse{}, errors.Wrap(ctx.Err(), i18n.Format(i18n.DeviceTokenPollCancelled))
		case <-time.After(interval):
		}
	}
	return tokenResponse{}, errors.New(i18n.Format(i18n.DeviceAuthorizationTimeout))
}

func isDeviceGrantErrorType(
	err error,
	types []string,
) bool {
	if err == nil {
		return false
	}
	errText := strings.ToLower(err.Error())
	for _, oauthError := range types {
		if strings.Contains(errText, oauthError) {
			return true
		}
	}
	return false
}

func normalizePollInterval(
	seconds int,
) time.Duration {
	if seconds <= 0 {
		return 2 * time.Second
	}
	return time.Duration(seconds) * time.Second
}

func checkForAuthorizedDeviceToken(
	ctx context.Context,
	clientID string,
	tokenEndpoint string,
	deviceCode string,
	audience string,
) (tokenResponse, error) {
	data := buildDeviceCodeValues(clientID, deviceCode, audience)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, tokenEndpoint, strings.NewReader(data.Encode()))
	if err != nil {
		return tokenResponse{}, err
	}
	req.Header.Set(constants.ContentTypeHeaderKey, constants.FormContentType)
	return result.Split(httputils.DoAsJSON[tokenResponse](defaultClient, req))
}

func displayDeviceInstructions(
	_ui ui.UI,
	d deviceAuthorizationResponse,
) {
	if d.VerificationURIComplete != "" {
		_ui.Display(i18n.Format(i18n.OpenThisURLInYourBrowser, i18n.A{
			constants.I18nArgKeyURL: d.VerificationURIComplete,
		}))
	} else {
		_ui.Display(i18n.Format(i18n.OpenURLAndEnterCode, i18n.A{
			constants.I18nArgKeyURL:  d.VerificationURI,
			constants.I18nArgKeyCode: d.UserCode,
		}))
	}
}

func authorizationPolling(
	ctx context.Context,
	_ui ui.UI,
	clientID string,
	deviceAuthResponse deviceAuthorizationResponse,
	cfg wellKnownConfigResponse,
) (tokenResponse, error) {
	displayDeviceInstructions(_ui, deviceAuthResponse)
	return pollForDeviceTokenAuthorization(ctx, clientID, deviceAuthResponse, cfg)
}

func buildDeviceGrantValues(
	clientID string,
	audience string,
) url.Values {
	data := url.Values{}
	data.Set(constants.QueryParameterAudience, sliceutil.Coalesce(audience, constants.OAuthAudience))
	data.Set(constants.QueryParameterClientID, clientID)
	data.Set(constants.QueryParameterScope, constants.OAuthScope)
	return data
}

func buildDeviceCodeValues(
	clientID string,
	deviceCode string,
	audience string,
) url.Values {
	data := url.Values{}
	data.Set(constants.QueryParameterGrantType, constants.OAuthDeviceGrantType)
	data.Set(constants.QueryParameterAudience, sliceutil.Coalesce(audience, constants.OAuthAudience))
	data.Set(constants.QueryParameterClientID, clientID)
	data.Set(constants.QueryParameterDeviceCode, deviceCode)
	return data
}
