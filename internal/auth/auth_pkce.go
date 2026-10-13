// Copyright © 2026, SAS Institute Inc., Cary, NC, USA.  All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/sas-institute-rnd-saas/wbctl/internal/constants"
	"github.com/sas-institute-rnd-saas/wbctl/internal/i18n"
	"github.com/sas-institute-rnd-saas/wbctl/internal/sliceutil"
	"github.com/sas-institute-rnd-saas/wbctl/internal/ui"
	"github.com/sassoftware/sas-ggdk/pkg/errors"
	"github.com/sassoftware/sas-ggdk/pkg/httputils"
	"github.com/sassoftware/sas-ggdk/pkg/pointer"
	"github.com/sassoftware/sas-ggdk/pkg/result"
	"github.com/spf13/afero"
)

type pkceAuthState struct {
	audience             string
	callbackURL          string
	challenge            string
	clientID             string
	code                 string
	controlPlaneEndpoint string
	ingressHost          string
	organization         string
	state                string
	verifier             string
}

// getAccessTokenByPKCE implements a PKCE + authorization-code flow.
func getAccessTokenByPKCE(
	ctx context.Context,
	_ui ui.UI,
	fsys afero.Fs,
	params GetAccessTokenParams,
) (string, error) {
	var tokenResp tokenResponse
	var accessToken string
	var err error
	authState := pkceAuthState{
		clientID:             params.ClientID,
		ingressHost:          params.IngressHost,
		organization:         params.Organization,
		controlPlaneEndpoint: params.ControlPlaneEndpoint,
	}
	defer deferSaveTokenToCache(_ui, fsys, authState.controlPlaneEndpoint, &tokenResp, &err)

	if authState.ingressHost == "" {
		endpoint, err := url.Parse(authState.controlPlaneEndpoint)
		if err != nil {
			return "", errors.Wrap(err, i18n.Format(i18n.FailedToGetPKCEToken))
		}
		authState.ingressHost = endpoint.Host
	}
	cfg, err := getWellKnownConfig(ctx, authState.controlPlaneEndpoint)
	if err != nil {
		return "", errors.Wrap(err, i18n.Format(i18n.FailedToGetPKCEToken))
	}
	authState.clientID = sliceutil.Coalesce(authState.clientID, cfg.Workbench.SPAClientID)
	authState.audience = cfg.Workbench.SPAAudience

	// Check for a cached access token.
	if !params.IgnoreTokenCache {
		tokenResp, accessToken = getCachedOrRefreshedToken(
			ctx,
			_ui,
			authState.clientID,
			cfg,
			fsys,
			authState.controlPlaneEndpoint,
			authState.ingressHost,
		)
		if accessToken != "" {
			return accessToken, nil
		}
	}

	// Generate PKCE params
	authState.verifier = generateCodeVerifier()
	authState.challenge = codeChallengeS256(authState.verifier)
	authState.state = generateState()

	// Try loopback flow unless noBrowser requested
	if params.NoListen {
		authState.code, err = getCodeFromPKCEManualFlow(_ui, cfg, &authState)
	} else {
		authState.code, err = getCodeFromPKCELoopbackFlow(ctx, _ui, cfg, &authState)
	}
	if err != nil {
		return "", errors.Wrap(err, i18n.Format(i18n.FailedToGetPKCEToken))
	}

	tokenResp, accessToken, err = exchangePKCECodeForToken(ctx, cfg, &authState)

	if err != nil {
		err = errors.Wrap(err, i18n.Format(i18n.FailedToGetPKCEToken))
	}
	return accessToken, err
}

func getCodeFromPKCEManualFlow(
	_ui ui.UI,
	cfg wellKnownConfigResponse,
	authState *pkceAuthState,
) (string, error) {
	authState.callbackURL = constants.PKCECallbackURL
	authURL := cfg.AuthorizationEndpoint + "?" + buildPKCELoopbackValues(authState).Encode()
	msg := i18n.Format(i18n.OpenThisURLInYourBrowser, i18n.A{
		constants.I18nArgKeyURL: authURL,
	})
	_ui.Display(msg)
	_ui.Display(i18n.Format(i18n.PKCEInstructions))
	_ui.Warn(i18n.Format(i18n.PKCEInstructionsWarning))
	input := _ui.Ask(i18n.Format(i18n.PKCEBrowserURLPrompt))
	u, err := url.Parse(input)
	if err != nil {
		return "", err
	}
	q := u.Query()
	code := q.Get(constants.QueryParameterCode)
	receivedState := q.Get(constants.QueryParameterState)
	errType := q.Get(constants.QueryParameterError)
	errDescription := q.Get(constants.QueryParameterErrorDescription)
	if errType != "" {
		return "", errors.New(i18n.Format(
			i18n.PKCEErrorResponse,
			i18n.A{constants.I18nArgKeyType: errType, constants.I18nArgKeyDescription: errDescription},
		))
	}
	if !stateValidator(authState.state, receivedState) {
		return "", errors.New(i18n.Format(i18n.PKCEReceivedStateMismatch))
	}
	return code, nil
}

func getCodeFromPKCELoopbackFlow(
	ctx context.Context,
	_ui ui.UI,
	cfg wellKnownConfigResponse,
	authState *pkceAuthState,
) (string, error) {
	codeCh, err := startCallbackListener(ctx, authState)
	if err != nil {
		return "", err
	}
	authURL := cfg.AuthorizationEndpoint + "?" + buildPKCELoopbackValues(authState).Encode()

	msg := i18n.Format(i18n.OpenThisURLInYourBrowser, i18n.A{
		constants.I18nArgKeyURL: authURL,
	})
	_ui.Display(msg)

	// Wait for code or timeout
	var code string
	var receivedState string
	var errType string
	var errDescription string
	select {
	case vals := <-codeCh:
		code = vals[constants.QueryParameterCode]
		receivedState = vals[constants.QueryParameterState]
		errType = vals[constants.QueryParameterError]
		errDescription = vals[constants.QueryParameterErrorDescription]
	case <-time.After(callbackTimeout):
		return "", errors.New(i18n.Format(i18n.PKCEAuthorizationTimeout))
	}
	if errType != "" {
		return "", errors.New(i18n.Format(
			i18n.PKCEErrorResponse,
			i18n.A{constants.I18nArgKeyType: errType, constants.I18nArgKeyDescription: errDescription},
		))
	}
	if !stateValidator(authState.state, receivedState) {
		return "", errors.New(i18n.Format(i18n.PKCEReceivedStateMismatch))
	}
	return code, nil
}

func exchangePKCECodeForToken(
	ctx context.Context,
	cfg wellKnownConfigResponse,
	authState *pkceAuthState,
) (tokenResponse, string, error) {
	// Exchange code for token
	data := buildPKCEExchangeValues(authState)
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		cfg.TokenEndpoint,
		strings.NewReader(data.Encode()),
	)
	if err != nil {
		return tokenResponse{}, "", err
	}
	req.Header.Set(constants.ContentTypeHeaderKey, constants.FormContentType)
	res, err := result.Split(httputils.DoAsJSON[tokenResponse](defaultClient, req))
	if err != nil {
		return tokenResponse{}, "", err
	}
	tokenResp := res
	accessToken, err := getAccessTokenFromTokenResponse(tokenResp)
	return tokenResp, accessToken, err
}

func startCallbackListener(ctx context.Context, authState *pkceAuthState) (chan map[string]string, error) {
	codeCh := make(chan map[string]string, 1)
	ctx, cancel := context.WithTimeout(ctx, 300*time.Second)
	cfg := net.ListenConfig{}
	var ln net.Listener
	var err error
	for i := range 19 {
		port := listenPort + (1111 * i)
		// Update the callbackURL with the port.
		authState.callbackURL = PKCECallbackURL(port)
		ln, err = cfg.Listen(ctx,
			constants.PKCECallbackListenProtocol,
			constants.PKCECallbackListenHost+":"+strconv.Itoa(port),
		)
		if err == nil {
			break
		}
	}
	if err != nil {
		cancel()
		return nil, err
	}
	h := http.NewServeMux()
	h.HandleFunc(constants.PKCECallbackEndpoint, func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		vals := map[string]string{
			constants.QueryParameterCode:             q.Get(constants.QueryParameterCode),
			constants.QueryParameterState:            q.Get(constants.QueryParameterState),
			constants.QueryParameterError:            q.Get(constants.QueryParameterError),
			constants.QueryParameterErrorDescription: q.Get(constants.QueryParameterErrorDescription),
		}
		codeCh <- vals
		w.Header().Set(constants.ContentTypeHeaderKey, constants.PKCECallbackResponseContentType)
		_, _ = w.Write([]byte(constants.AuthenticationCompleteHTML))
		// Shutdown after a short delay to allow response to be sent.
		go func() {
			time.Sleep(100 * time.Millisecond)
			_ = ln.Close()
			cancel()
		}()
	})
	srv := &http.Server{
		Handler:           h,
		ReadHeaderTimeout: 30 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       30 * time.Second,
	}
	go func() { _ = srv.Serve(ln) }()
	return codeCh, nil
}

func buildPKCELoopbackValues(
	authState *pkceAuthState,
) url.Values {
	return url.Values{
		constants.QueryParameterAudience:            {sliceutil.Coalesce(authState.audience, constants.OAuthAudience)},
		constants.QueryParameterClientID:            {authState.clientID},
		constants.QueryParameterCodeChallenge:       {authState.challenge},
		constants.QueryParameterCodeChallengeMethod: {constants.PKCEChallengeMethod},
		constants.QueryParameterIngressHost:         {authState.ingressHost},
		constants.QueryParameterOrganization:        {authState.organization},
		constants.QueryParameterRedirectURI:         {authState.callbackURL},
		constants.QueryParameterResponseType:        {constants.PKCEResponseType},
		constants.QueryParameterScope:               {constants.OAuthScope},
		constants.QueryParameterState:               {authState.state},
	}
}

func buildPKCEExchangeValues(
	authState *pkceAuthState,
) url.Values {
	return url.Values{
		constants.QueryParameterClientID:     {authState.clientID},
		constants.QueryParameterCode:         {authState.code},
		constants.QueryParameterCodeVerifier: {authState.verifier},
		constants.QueryParameterGrantType:    {constants.OAuthAuthCodeGrantType},
		constants.QueryParameterRedirectURI:  {authState.callbackURL},
	}
}

// generateCodeVerifier returns a random code_verifier according to RFC 7636
// (between 43 and 128 characters, URL-safe base64 without padding).
func generateCodeVerifier() string {
	b := make([]byte, 32) // 32 bytes -> 43 chars when base64url-encoded
	// Read is documented to never return an error.
	_, _ = rand.Read(b)
	return base64.RawURLEncoding.EncodeToString(b)
}

// codeChallengeS256 returns the S256 code_challenge for the given verifier.
func codeChallengeS256(verifier string) string {
	h := sha256.Sum256([]byte(verifier))
	return base64.RawURLEncoding.EncodeToString(h[:])
}

// generateState returns a random state string for CSRF protection.
func generateState() string {
	b := make([]byte, 16)
	// Read is documented to never return an error.
	_, _ = rand.Read(b)
	return base64.RawURLEncoding.EncodeToString(b)
}

// PKCECallbackURL returns the pkce callback url for the given port.
func PKCECallbackURL(port int) string {
	return fmt.Sprintf(
		"http://%s%s",
		net.JoinHostPort(constants.PKCECallbackListenHost, strconv.Itoa(port)),
		constants.PKCECallbackEndpoint,
	)
}

// Test hooks

var stateValidator = defaultStateValidator

// SetStateValidator sets the function used to validate returned state. Used in
// tests to override the default behavior.
func SetStateValidator(validator func(string, string) bool) {
	if validator == nil {
		validator = defaultStateValidator
	}
	stateValidator = validator
}

func defaultStateValidator(expected, actual string) bool {
	return expected == actual
}

var listenPort = 20000

// SetListenPort sets the pkce callback listen port. Used in tests to override
// the default behavior.
func SetListenPort(port int) {
	if port == 0 {
		port = 20000
	}
	listenPort = port
}

var callbackTimeout = 300 * time.Second

// SetCallbackTimeout sets the length of time that the pkce callback listener
// waits for an authorization response. Used in tests to override default
// behavior.
func SetCallbackTimeout(timeout *time.Duration) {
	if timeout == nil {
		timeout = pointer.Ptr(300 * time.Second)
	}
	callbackTimeout = *timeout
}
