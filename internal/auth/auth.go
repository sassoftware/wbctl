// Copyright © 2026, SAS Institute Inc., Cary, NC, USA.  All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package auth

import (
	"context"
	"encoding/base64"
	"io"
	"net/http"
	"net/url"
	"slices"
	"strings"
	"time"

	"github.com/sassoftware/sas-ggdk/pkg/errors"
	"github.com/sassoftware/sas-ggdk/pkg/httputils"
	"github.com/sassoftware/sas-ggdk/pkg/jsonutils"
	"github.com/sassoftware/sas-ggdk/pkg/result"
	"github.com/sassoftware/wbctl/internal/constants"
	"github.com/sassoftware/wbctl/internal/httputil"
	"github.com/sassoftware/wbctl/internal/i18n"
	"github.com/sassoftware/wbctl/internal/postgrest"
	"github.com/sassoftware/wbctl/internal/sliceutil"
	"github.com/sassoftware/wbctl/internal/ui"
	"github.com/spf13/afero"
)

type wellKnownConfigResponse struct {
	TokenEndpoint               string                   `json:"token_endpoint"`
	DeviceAuthorizationEndpoint string                   `json:"device_authorization_endpoint"`
	AuthorizationEndpoint       string                   `json:"authorization_endpoint"`
	Workbench                   wellKnownWorkbenchConfig `json:"workbench"`
}

type wellKnownWorkbenchConfig struct {
	SPAAudience    string `json:"spa_audience"`
	SPAClientID    string `json:"spa_client_id"`
	NativeClientID string `json:"native_client_id"`
}

type tokenResponse struct {
	AccessToken      string `json:"access_token"`
	ExpiresIn        int    `json:"expires_in"`
	IDToken          string `json:"id_token"`
	NotBeforePolicy  int    `json:"not-before-policy"`
	RefreshExpiresIn int    `json:"refresh_expires_in"`
	RefreshToken     string `json:"refresh_token"`
	Scope            string `json:"scope"`
	SessionState     string `json:"session_state"`
	TokenType        string `json:"token_type"`
}

type deviceAuthorizationResponse struct {
	DeviceCode              string `json:"device_code"`
	UserCode                string `json:"user_code"`
	VerificationURI         string `json:"verification_uri"`
	VerificationURIComplete string `json:"verification_uri_complete"`
	ExpiresIn               int    `json:"expires_in"`
	Interval                int    `json:"interval"`
}

var defaultClient = &http.Client{
	Timeout: time.Second * 10,
}

// GetCSRFToken fetches the CSRF token cookie from postgrest.
func GetCSRFToken(
	ctx context.Context,
	controlPlaneEndpoint,
	accessToken string,
) (string, error) {
	resourceURLString, err :=
		url.JoinPath(controlPlaneEndpoint, postgrest.Endpoint(constants.WorkbenchEndpointRoot))
	if err != nil {
		return "", errors.Wrap(err, i18n.Format(i18n.FailedToGetCSRFToken))
	}
	//  url.JoinPath succeeds only after it has already parsed/validated the
	//  base URL internally. Its result is a URL string generated from a url.URL
	//  value, so feeding that same string into http.NewRequestWithContext
	//  should not fail. There is no way to unit test the err if we capture it
	//  here.
	request, _ := http.NewRequestWithContext(ctx, http.MethodHead, resourceURLString, nil)
	request = httputil.AddHeaders(map[string]string{
		constants.AuthorizationHeaderKey: "Bearer " + accessToken,
	}, request)
	// Lint can't tell that this body is closed below.
	// nolint:bodyclose
	resp, err := result.Split(httputils.Do(defaultClient, request))
	if err != nil {
		return "", errors.Wrap(err, i18n.Format(i18n.FailedToGetCSRFToken))
	}
	defer func() { _ = resp.Body.Close() }()
	for _, cookie := range resp.Cookies() {
		if cookie.Name == constants.PostgrestCSRFTokenKey {
			return cookie.Value, nil
		}
	}
	return "", nil
}

// GetAccessToken retrieves an access token. The cache is checked first and, if
// a valid token is not found, the specified auth mode is used to create one.
func GetAccessToken(
	ctx context.Context,
	_ui ui.UI,
	fsys afero.Fs,
	params GetAccessTokenParams,
) (string, error) {
	if params.ControlPlaneEndpoint == "" {
		return "", errors.New(i18n.Format(i18n.EndpointMustBeSpecified))
	}
	authMode := params.NormalizedAuthMode()
	switch authMode {
	case constants.AuthModeDevice:
		return getAccessTokenByDeviceGrant(ctx, _ui, fsys, params)
	case constants.AuthModePKCE:
		return getAccessTokenByPKCE(ctx, _ui, fsys, params)
	case constants.AuthModeClientCredentials:
		return getAccessTokenByClientSecret(ctx, _ui, fsys, params)
	default:
		return "", errors.New(i18n.Format(i18n.InvalidAuthMode, i18n.A{
			constants.I18nArgKeyAuthMode: authMode,
		}))
	}
}

// RemoveAccessToken removes an access token from the token cache.
func RemoveAccessToken(
	fsys afero.Fs,
	controlPlaneEndpoint string,
) error {
	return removeTokenFromCache(fsys, controlPlaneEndpoint)
}

func deferSaveTokenToCache(
	_ui ui.UI,
	fsys afero.Fs,
	controlPlaneEndpoint string,
	tokenResp *tokenResponse,
	errp *error,
) {
	if tokenResp.AccessToken == "" || *errp != nil {
		return
	}
	err := saveTokenToCache(fsys, controlPlaneEndpoint, *tokenResp)
	if err != nil {
		// We can continue now even if we can't save the token for later.
		_ui.Warn(i18n.Format(i18n.FailedToSaveDeviceToken, i18n.A{constants.I18nArgKeyError: err.Error()}))
	}
}

// getCachedOrRefreshedToken returns both any token response (so it can be
// cached) and any access token (either cached or refreshed).
func getCachedOrRefreshedToken(
	ctx context.Context,
	_ui ui.UI,
	cfg wellKnownConfigResponse,
	fsys afero.Fs,
	controlPlaneEndpoint string,
	ingressHost string,
) (tokenResponse, string) {
	ct, err := getCachedTokenFor(fsys, controlPlaneEndpoint)
	if err != nil {
		_ui.Warn(err.Error())
	} else {
		accessToken, err := getValidAccessTokenFromCachedToken(ct)
		if err == nil && accessToken != "" {
			return tokenResponse{}, accessToken
		}
	}
	// Check for a cached refresh token
	if ct.RefreshToken != "" {
		tokenResp, err := performRefreshForToken(
			ctx,
			clientIDFromAccessToken(ct.AccessToken),
			cfg.TokenEndpoint,
			ct.RefreshToken,
			ingressHost,
			cfg.Workbench.SPAAudience,
		)
		if err == nil {
			var accessToken string
			accessToken, err = getAccessTokenFromTokenResponse(tokenResp)
			if err == nil && accessToken != "" {
				return tokenResp, accessToken
			}
		}
		_ui.Warn(i18n.Format(i18n.FallingBackToFullReauthorization))
	}
	return tokenResponse{}, ""
}

func getAccessTokenFromTokenResponse(
	tr tokenResponse,
) (string, error) {
	if tr.AccessToken == "" {
		return "", errors.New(i18n.Format(i18n.TokenResponseMissingAccessToken))
	}
	return tr.AccessToken, nil
}

func getWellKnownConfig(
	ctx context.Context,
	controlPlaneEndpoint string,
) (wellKnownConfigResponse, error) {
	var firstErr error
	wellKnownPaths := []string{
		constants.AuthWellKnownPath,
		constants.AuthLegacyWellKnownPath,
	}
	for _, wellKnownPath := range wellKnownPaths {
		cfg, err := getWellKnownConfigAtPath(ctx, controlPlaneEndpoint, wellKnownPath)
		if err == nil {
			return cfg, nil
		}
		if firstErr == nil {
			firstErr = err
		}
	}
	return wellKnownConfigResponse{}, firstErr
}

func getWellKnownConfigAtPath(
	ctx context.Context,
	controlPlaneEndpoint,
	wellKnownPath string,
) (wellKnownConfigResponse, error) {
	wellKnownURLString, err := url.JoinPath(controlPlaneEndpoint, wellKnownPath)
	if err != nil {
		return wellKnownConfigResponse{}, err
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, wellKnownURLString, nil)
	if err != nil {
		return wellKnownConfigResponse{}, err
	}
	return result.Split(httputils.DoAsJSON[wellKnownConfigResponse](defaultClient, request))
}

func performRefreshForToken(
	ctx context.Context,
	clientID string,
	tokenEndpoint string,
	refreshToken string,
	ingressHost string,
	audience string,
) (tokenResponse, error) {
	if refreshToken == "" {
		return tokenResponse{}, errors.New(i18n.Format(i18n.ProvidedRefreshTokenIsEmpty))
	}
	data := url.Values{}
	data.Set(constants.QueryParameterGrantType, constants.OAuthRefreshGrantType)
	data.Set(constants.QueryParameterAudience, sliceutil.Coalesce(audience, constants.OAuthAudience))
	data.Set(constants.QueryParameterClientID, clientID)
	data.Set(constants.QueryParameterRefreshToken, refreshToken)
	if ingressHost != "" {
		data.Set("ingressHost", ingressHost)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, tokenEndpoint, strings.NewReader(data.Encode()))
	if err != nil {
		return tokenResponse{}, errors.Wrap(err, i18n.Format(i18n.FailedToRefreshToken))
	}
	req.Header.Set(constants.ContentTypeHeaderKey, constants.FormContentType)
	res, err := result.Split(httputils.DoAsJSON[tokenResponse](defaultClient, req))
	if err != nil {
		return tokenResponse{}, errors.Wrap(err, i18n.Format(i18n.FailedToRefreshToken))
	}
	return res, nil
}

// IsRole reports whether the JWT contains the requested role.
func IsRole(
	token string,
	role string,
) bool {
	parsedToken := parseAccessToken(token)
	return slices.Contains(parsedToken.Sas.Roles, role)
}

func clientIDFromAccessToken(token string) string {
	parsedToken := parseAccessToken(token)
	return parsedToken.Azp
}

func parseAccessToken(token string) jwt {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return jwt{}
	}
	payloadPart := parts[1]
	padLen := (4 - (len(payloadPart) % 4)) % 4
	payloadPart += strings.Repeat("=", padLen)
	decoder := base64.NewDecoder(base64.URLEncoding, strings.NewReader(payloadPart))
	payloadBytes, err := io.ReadAll(decoder)
	if err != nil {
		return jwt{}
	}
	jwtPayload, err := result.Split(jsonutils.UnmarshalAs[jwt](payloadBytes))
	if err != nil {
		return jwt{}
	}
	return jwtPayload
}

type workbenchClaims struct {
	Auth0OrgID  string   `json:"auth0_org_id"`
	Auth0UserID string   `json:"auth0_user_id"`
	OrgID       string   `json:"org_id"`
	OrgName     string   `json:"org_name"`
	Roles       []string `json:"roles"`
	UserID      string   `json:"user_id"`
}

type jwt struct {
	Sas     workbenchClaims `json:"sas"`
	Iss     string          `json:"iss"`
	Sub     string          `json:"sub"`
	Aud     []string        `json:"aud"`
	Iat     int             `json:"iat"`
	Exp     int             `json:"exp"`
	Scope   string          `json:"scope"`
	OrgID   string          `json:"org_id"`
	OrgName string          `json:"org_name"`
	Azp     string          `json:"azp"`
}
