// Copyright © 2026, SAS Institute Inc., Cary, NC, USA.  All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package auth

import (
	"context"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/sassoftware/sas-ggdk/pkg/errors"
	"github.com/sassoftware/sas-ggdk/pkg/httputils"
	"github.com/sassoftware/sas-ggdk/pkg/result"
	"github.com/sassoftware/wbctl/internal/constants"
	"github.com/sassoftware/wbctl/internal/i18n"
	"github.com/sassoftware/wbctl/internal/logutil"
	"github.com/sassoftware/wbctl/internal/sliceutil"
	"github.com/sassoftware/wbctl/internal/ui"
	"github.com/spf13/afero"
	"github.com/spf13/viper"
)

// nolint:cyclop
func getAccessTokenByDeviceGrant(
	ctx context.Context,
	_ui ui.UI,
	fsys afero.Fs,
	v *viper.Viper,
	params GetAccessTokenParams,
) (string, error) {
	logutil.GetLogger().Info(
		"getting access token by device grant",
		"params", params,
	)
	if params.Organization == "" {
		err := errors.New(i18n.Format(i18n.AuthModeDeviceRequiresOrganization))
		return "", errors.Wrap(err, i18n.Format(i18n.FailedToGetDeviceToken))
	}
	var tokenResp tokenResponse
	var accessToken string
	var err error
	defer deferSaveTokenToCache(_ui, fsys, v, params.ControlPlaneEndpoint, &tokenResp, &err)

	cfg, err := getWellKnownConfig(ctx, params.ControlPlaneEndpoint)
	if err != nil {
		return "", errors.Wrap(err, i18n.Format(i18n.FailedToGetDeviceToken))
	}

	// Check for a cached access token.
	if !params.IgnoreTokenCache {
		tokenResp, accessToken = getCachedOrRefreshedToken(
			ctx,
			cfg,
			fsys,
			v,
			params.ControlPlaneEndpoint,
			params.IngressHost,
		)
		if accessToken != "" {
			return accessToken, nil
		}
	}

	// Fall back to device authorization
	logutil.GetLogger().Info("falling back to full authorization")
	deviceAuthResponse, err := startDeviceAuthorizationAtEndpoint(ctx, params.ClientID, params.Organization, cfg)
	if err != nil {
		return "", errors.Wrap(err, i18n.Format(i18n.FailedToGetDeviceToken))
	}
	tokenResp, err = authorizationPolling(ctx, _ui, params.ClientID, params.Organization, deviceAuthResponse, cfg)
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
	organization string,
	cfg wellKnownConfigResponse,
) (deviceAuthorizationResponse, error) {
	if clientID == "" {
		clientID = cfg.Workbench.NativeClientID
	}
	data := buildDeviceGrantValues(clientID, organization, cfg.Workbench.SPAAudience)
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
	if clientID == "" {
		clientID = cfg.Workbench.NativeClientID
	}
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
	clientID string,
	organization string,
	d deviceAuthorizationResponse,
	cfg wellKnownConfigResponse,
) {
	// If we have been given a client ID then assume the user wants normal device auth.
	if clientID != "" {
		if d.VerificationURIComplete != "" {
			_ui.Display(i18n.Format(i18n.DeviceOpenThisURLInYourBrowser, i18n.A{
				constants.I18nArgKeyURL:  d.VerificationURIComplete,
				constants.I18nArgKeyCode: d.UserCode,
			}))
			return
		}
		_ui.Display(i18n.Format(i18n.DeviceOpenURLAndEnterCode, i18n.A{
			constants.I18nArgKeyURL:  d.VerificationURI,
			constants.I18nArgKeyCode: d.UserCode,
		}))
		return
	}
	// Otherwise assume we need to go to the SPA endpoint first
	var msg string
	data := url.Values{}
	data.Set(constants.QueryParameterClientID, cfg.Workbench.SPAClientID)
	data.Set(constants.QueryParameterResponseType, constants.PKCEResponseType)
	data.Set(constants.QueryParameterScope, constants.OAuthScope)
	if organization != "" {
		data.Set(constants.QueryParameterOrganization, organization)
	}
	if d.VerificationURIComplete != "" {
		data.Set(constants.QueryParameterRedirectURI, d.VerificationURIComplete)
		msg = i18n.Format(i18n.DeviceOpenThisURLInYourBrowser, i18n.A{
			constants.I18nArgKeyURL:  cfg.AuthorizationEndpoint + "?" + data.Encode(),
			constants.I18nArgKeyCode: d.UserCode,
		})
	} else {
		data.Set(constants.QueryParameterRedirectURI, d.VerificationURI)
		msg = i18n.Format(i18n.DeviceOpenURLAndEnterCode, i18n.A{
			constants.I18nArgKeyURL:  cfg.AuthorizationEndpoint + "?" + data.Encode(),
			constants.I18nArgKeyCode: d.UserCode,
		})
	}
	_ui.Display(msg)
}

func authorizationPolling(
	ctx context.Context,
	_ui ui.UI,
	clientID string,
	organization string,
	deviceAuthResponse deviceAuthorizationResponse,
	cfg wellKnownConfigResponse,
) (tokenResponse, error) {
	displayDeviceInstructions(_ui, clientID, organization, deviceAuthResponse, cfg)
	return pollForDeviceTokenAuthorization(ctx, clientID, deviceAuthResponse, cfg)
}

func buildDeviceGrantValues(
	clientID string,
	organization string,
	audience string,
) url.Values {
	data := url.Values{}
	data.Set(constants.QueryParameterAudience, sliceutil.Coalesce(audience, constants.OAuthAudience))
	data.Set(constants.QueryParameterClientID, clientID)
	scope := constants.OAuthScope
	if organization != "" {
		scope += " organization:" + organization
	}
	data.Set(constants.QueryParameterScope, scope)
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
