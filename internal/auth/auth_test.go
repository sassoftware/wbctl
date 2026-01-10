// Copyright © 2026, SAS Institute Inc., Cary, NC, USA.  All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package auth_test

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/sas-institute-rnd-saas/wbctl/internal/auth"
	"github.com/sas-institute-rnd-saas/wbctl/internal/constants"
	"github.com/sas-institute-rnd-saas/wbctl/internal/osutil"
	"github.com/sas-institute-rnd-saas/wbctl/internal/postgrest"
	"github.com/sas-institute-rnd-saas/wbctl/internal/stubs"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	authTestEndpointRequiredErr                 = "--endpoint must be specified"
	authTestDeviceAuthorizationFailureErrFormat = "failed to get device token\n" +
		" Caused by:\n" +
		"\t* POST ${URL}/device failed\n" +
		"\t* 400 {\"error\":\"invalid_client\"}"
	authTestDeviceAuthorizationBody = `
{
  "device_code": "dc",
  "user_code": "uc",
  "verification_uri": "https://verify",
  "expires_in": 1,
  "interval": 1
}
`
	authTestDeviceAuthorizationBodyWithVerificationURIComplete = `
{
  "device_code": "dc",
  "user_code": "uc",
  "verification_uri": "https://verify",
  "verification_uri_complete": "https://verify.complete",
  "expires_in": 1,
  "interval": 1
}
`
	authTestDeviceAuthorizationBodyWithZeroInterval = `
{
  "device_code": "dc",
  "user_code": "uc",
  "verification_uri": "https://verify",
  "expires_in": 5,
  "interval": 0
}
`
	authTestDeviceInvalidClientBody   = `{"error":"invalid_client"}`
	authTestDiscoveryFailureErrFormat = "failed to get pkce token\n" +
		" Caused by:\n" +
		"\t* GET ${URL}/auth0Admin/internal/.well-known/openid-configuration failed\n" +
		"\t* 500 {\"error\":\"discovery\"}"
	authTestDiscoveryFailurePasswordErrFormat = "failed to get username/password token\n" +
		" Caused by:\n" +
		"\t* GET ${URL}/auth0Admin/internal/.well-known/openid-configuration failed\n" +
		"\t* 500 {\"error\":\"discovery\"}"
	authTestGetCSRFFailureErrFormat = "failed to get CSRF token\n" +
		" Caused by:\n" +
		"\t* HEAD ${URL} failed\n" +
		"\t* 500 "
	authTestInvalidAuthModeErr            = "invalid --auth-mode: \"invalid\" (allowed: auto, password, device)"
	authTestPasswordGrantFailureErrFormat = "failed to get username/password token\n" +
		" Caused by:\n" +
		"\t* POST ${URL}/token failed\n" +
		"\t* 401 {\"error\":\"bad\"}"
	authTestPasswordModeRequiresUsername = "--auth-mode=password requires --username" // nolint:gosec
	authTestPollingAccessDeniedErrFormat = "failed to get device token\n" +
		" Caused by:\n" +
		"\t* POST ${URL}/token failed\n" +
		"\t* 400 {\"error\":\"access_denied\"}"
	authTestPollingCancelledErr = "failed to get device token\n" +
		" Caused by:\n" +
		"\t* device token poll cancelled\n" +
		"\t* context canceled"
	authTestDeviceAuthorizationTimeoutErr = "failed to get device token\n" +
		" Caused by:\n" +
		"\t* device authorization timeout"
	authTestPasswordModeEmptyAccessTokenErr = "failed to get username/password token\n" +
		" Caused by:\n" +
		"\t* token response missing access_token"
	authTestDeviceModeEmptyAccessTokenErr = "failed to get device token\n" +
		" Caused by:\n" +
		"\t* token response missing access_token"
	authTestPasswordModeInvalidTokenURLErr = "failed to get username/password token\n" +
		" Caused by:\n" +
		"\t* parse \"a\\nb\": net/url: invalid control character in URL"
	authTestDeviceModeInvalidDeviceURLErr = "failed to get device token\n" +
		" Caused by:\n" +
		"\t* parse \"a\\nb\": net/url: invalid control character in URL"
	authTestClientCredentialsModeInvalidControlPlaneURLErr = "failed to get client credentials token\n" +
		" Caused by:\n" +
		"\t* parse \"a\\nb\": net/url: invalid control character in URL"
	authTestClientCredentialsModeInvalidTokenEndpointURLErr = "failed to get client credentials token\n" +
		" Caused by:\n" +
		"\t* parse \"a\\x7fb\": net/url: invalid control character in URL"
	authTestDeviceModeInvalidControlPlaneURLErr = "failed to get device token\n" +
		" Caused by:\n" +
		"\t* parse \"a\\nb\": net/url: invalid control character in URL"
	authTestPKCEModeInvalidControlPlaneURLErr = "failed to get pkce token\n" +
		" Caused by:\n" +
		"\t* parse \"a\\nb\": net/url: invalid control character in URL"
	authTestDeviceModeInvalidRefreshURLErr = "failed to get device token\n" +
		" Caused by:\n" +
		"\t* POST  failed\n" +
		"\t* Post \"\": unsupported protocol scheme \"\""
	authTestInvalidControlPlaneEndpoint = "a\nb"
)

var (
	expectedTokenRequestBody = `audience=` + url.QueryEscape(constants.OAuthAudience) +
		`&client_id=` + url.QueryEscape(constants.OAuthClientID) +
		`&device_code=dc&grant_type=urn%3Aietf%3Aparams%3Aoauth%3Agrant-type%3Adevice_code`
	successfulPasswordAuthDiscovery = stubs.StubHTTPPartialGetConfig{
		ResponseStatus: http.StatusOK,
		ResponseBody:   `{"token_endpoint":"http://${HOST}/token"}`,
	}
	successfulDeviceAuthDiscovery = stubs.StubHTTPPartialGetConfig{
		ResponseStatus: http.StatusOK,
		ResponseBody:   `{"token_endpoint":"http://${HOST}/token", "device_authorization_endpoint":"http://${HOST}/device"}`,
	}
	successfulPKCEAuthDiscovery = stubs.StubHTTPPartialGetConfig{
		ResponseStatus: http.StatusOK,
		ResponseBody:   `{"token_endpoint":"http://${HOST}/token", "authorization_endpoint":"http://${HOST}/authorization"}`,
	}
	successfulDevice = stubs.StubHTTPPartialPostConfig{
		ExpectedRequestBody: `audience=` + url.QueryEscape(constants.OAuthAudience) +
			`&client_id=` + url.QueryEscape(constants.OAuthClientID) +
			`&scope=` + url.QueryEscape(constants.OAuthScope),
		ResponseStatus: http.StatusOK,
		ResponseBody:   authTestDeviceAuthorizationBody,
	}
)

func TestGetCSRFToken(t *testing.T) {
	tests := []struct {
		name          string
		mockEndpoints []stubs.StubHTTPEndpoint
		expectedToken string
		err           string
	}{
		{
			name: "token present",
			mockEndpoints: []stubs.StubHTTPEndpoint{
				{
					Request: stubs.StubHTTPRequest{
						Method:          http.MethodHead,
						URL:             postgrest.Endpoint(constants.WorkbenchEndpointRoot),
						ExpectedHeaders: map[string]string{constants.AuthorizationHeaderKey: "Bearer at"},
					},
					Response: stubs.StubHTTPResponse{
						Status: http.StatusOK,
						Headers: map[string][]string{
							constants.SetCookieHeaderKey: {"postgrest-csrf-token=abc; Path=/"}},
					},
				},
			},
			expectedToken: "abc",
		},
		{
			name: "token in later Set-Cookie header",
			mockEndpoints: []stubs.StubHTTPEndpoint{
				{
					Request: stubs.StubHTTPRequest{
						Method:          http.MethodHead,
						URL:             postgrest.Endpoint(constants.WorkbenchEndpointRoot),
						ExpectedHeaders: map[string]string{constants.AuthorizationHeaderKey: "Bearer at"},
					},
					Response: stubs.StubHTTPResponse{
						Status: http.StatusOK,
						Headers: map[string][]string{
							constants.SetCookieHeaderKey: {
								"other=value; Path=/",
								"postgrest-csrf-token=later; Path=/",
							},
						},
					},
				},
			},
			expectedToken: "later",
		},
		{
			name:          "token missing",
			expectedToken: "",
			mockEndpoints: []stubs.StubHTTPEndpoint{
				{
					Request: stubs.StubHTTPRequest{
						Method:          http.MethodHead,
						URL:             postgrest.Endpoint(constants.WorkbenchEndpointRoot),
						ExpectedHeaders: map[string]string{constants.AuthorizationHeaderKey: "Bearer at"},
					},
					Response: stubs.StubHTTPResponse{
						Status: http.StatusOK,
						Headers: map[string][]string{
							constants.SetCookieHeaderKey: {"other=value; Path=/"},
						},
					},
				},
			},
		},
		{
			name: "server error",
			err:  authTestGetCSRFFailureErrFormat,
			mockEndpoints: []stubs.StubHTTPEndpoint{
				{
					Request: stubs.StubHTTPRequest{
						Method:          http.MethodHead,
						URL:             postgrest.Endpoint(constants.WorkbenchEndpointRoot),
						ExpectedHeaders: map[string]string{constants.AuthorizationHeaderKey: "Bearer at"},
					},
					Response: stubs.StubHTTPResponse{
						Status: http.StatusInternalServerError,
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv, _ := stubs.NewStubHTTPServer(t, tt.mockEndpoints)
			t.Cleanup(srv.Close)

			res, err := auth.GetCSRFToken(context.Background(), srv.URL, "at")
			if tt.err != "" {
				expectedErr := tt.err
				requestURL := srv.URL + postgrest.Endpoint(constants.WorkbenchEndpointRoot)
				expectedErr = stubs.Replace(expectedErr, map[string]string{"${URL}": requestURL})
				require.EqualError(t, err, expectedErr)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.expectedToken, res)
		})
	}
}

func TestGetCSRFTokenErrorEdges(t *testing.T) {
	tests := []struct {
		name          string
		url           string
		expectedToken string
		expectedErr   string
	}{
		{
			name: "invalid control character in url",
			url:  "a\nb",
			expectedErr: "failed to get CSRF token\n Caused by:\n\t* parse \"a\\nb\":" +
				" net/url: invalid control character in URL",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, err := auth.GetCSRFToken(context.Background(), tt.url, "at")
			if tt.expectedErr != "" {
				require.EqualError(t, err, tt.expectedErr)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.expectedToken, res)
		})
	}
}

func TestGetAccessToken_EndpointRequired(t *testing.T) {
	uiStub := &stubs.StubUI{}
	fsys := afero.NewMemMapFs()
	_, err := auth.GetAccessToken(
		context.Background(),
		uiStub,
		fsys,
		auth.GetAccessTokenParams{
			ControlPlaneEndpoint: "",
			AuthMode:             constants.AuthModeAuto,
			Username:             "user",
			Password:             "pass",
			IgnoreTokenCache:     false,
			NoListen:             false,
		},
	)
	require.EqualError(t, err, authTestEndpointRequiredErr)
}

const (
	homeUnchanged int = iota
	homeAsFile
	homeAsUnset
)

type authInput struct {
	authMode             string
	controlPlaneEndpoint string
	password             string
	username             string
	clientID             string
	clientSecret         string
	cache                string
	askResponses         []string
	readOnlyCache        bool
	cancelContext        bool
	noListen             bool
	organization         string
	home                 int
	// openFileFunc, when non-nil, is called with the passthrough base Fs
	// after all other fsys setup is complete. The returned function is
	// installed as StubFS.OpenFileFunc, intercepting afero.TempFile
	// (O_EXCL), afero.Afero.WriteFile (O_TRUNC), and similar calls so
	// that individual saveTokenToCache failure branches can be exercised
	// without replacing the whole filesystem.
	openFileFunc func(
		afero.Fs,
		string,
		int,
		os.FileMode,
	) (afero.File, error)
	pkceStateValidator        func(string, string) bool
	pkceAuthorizationResponse string
	pkceListenPort            int
	pkceCallbackTimeout       *time.Duration
}

type authExpected struct {
	cachedToken         *expectedTokenCache
	errFormat           string
	legacyDiscoveryHits int
	modernDiscoveryHits int
	pollCount           int
	token               string
	displays            []string
	warns               []string
	asks                []string
	minPollCount        int
}

type authTestCase struct {
	name          string
	input         authInput
	mockEndpoints []stubs.StubHTTPEndpoint
	expected      authExpected
}

func TestGetAccessToken(t *testing.T) {
	tests := []authTestCase{
		{
			name: "invalid control plane endpoint returns discovery error",
			input: authInput{
				authMode:             constants.AuthModeAuto,
				clientID:             constants.OAuthClientID,
				controlPlaneEndpoint: authTestInvalidControlPlaneEndpoint,
			},
			expected: authExpected{
				errFormat: authTestPKCEModeInvalidControlPlaneURLErr,
			},
		},
		{
			name: "invalid auth mode returns validation error",
			input: authInput{
				authMode: "invalid",
				clientID: constants.OAuthClientID,
			},
			expected: authExpected{
				errFormat: authTestInvalidAuthModeErr,
			},
		},
		{
			name: "valid cached token short-circuits and returns cached token - expect pkce",
			input: authInput{
				authMode: constants.AuthModeAuto,
				cache: fmt.Sprintf(
					`{"access_token":"cached","refresh_token":"rt","expires_at":"%s"}`,
					time.Now().Add(time.Hour).UTC().Format(time.RFC3339Nano),
				),
				clientID: constants.OAuthClientID,
			},
			mockEndpoints: []stubs.StubHTTPEndpoint{
				{
					Request: stubs.StubHTTPRequest{
						Method: http.MethodGet,
						URL:    constants.AuthWellKnownPath,
					},
					Response: stubs.StubHTTPResponse{
						Status: http.StatusOK,
						Body:   `{"token_endpoint":"http://ignored", "device_authorization_endpoint":"http://ignored"}`,
					},
				},
			},
			expected: authExpected{
				token: "cached",
			},
		},
		{
			name: "refresh success returns refreshed token without pkce flow",
			input: authInput{
				authMode: constants.AuthModeAuto,
				cache: fmt.Sprintf(
					`{"access_token":"old","refresh_token":"rt","expires_at":"%s"}`,
					time.Now().Add(-time.Hour).UTC().Format(time.RFC3339Nano),
				),
				clientID: constants.OAuthClientID,
			},
			mockEndpoints: getDeviceEndpoints(deviceConfig{
				discoveryConfig: successfulDeviceAuthDiscovery,
				tokenConfigs: []stubs.StubHTTPPartialPostConfig{{
					ExpectedRequestBody: `audience=` + url.QueryEscape(constants.OAuthAudience) +
						`&client_id=` + url.QueryEscape(constants.OAuthClientID) +
						`&grant_type=refresh_token` +
						`&ingressHost={SANITIZE:127.0.0.1%3A\d+:SANITIZED}` +
						`&refresh_token=rt`,
					ResponseStatus: http.StatusOK,
					ResponseBody:   `{"access_token":"refreshed","refresh_token":"rt2","expires_in":3600}`,
				}},
			}),
			expected: authExpected{
				token: "refreshed",
				cachedToken: &expectedTokenCache{
					accessToken:  "refreshed",
					refreshToken: "rt2",
				},
			},
		},
		{
			name: "discovery failure - an expired token causes pkce flow (auto mode with no username, expect pkce)",
			input: authInput{
				authMode: constants.AuthModeAuto,
				cache: fmt.Sprintf(
					`{"access_token":"old","refresh_token":"rt","expires_at":"%s"}`,
					time.Now().Add(-time.Hour).UTC().Format(time.RFC3339Nano),
				),
				clientID: constants.OAuthClientID,
			},
			mockEndpoints: getDeviceEndpoints(deviceConfig{
				discoveryConfig: stubs.StubHTTPPartialGetConfig{
					ResponseStatus: http.StatusInternalServerError,
					ResponseBody:   `{"error":"discovery"}`,
				},
			}),
			expected: authExpected{
				errFormat: authTestDiscoveryFailureErrFormat,
			},
		},
		// This is a fake token so not a security issue.
		//nolint:gosec
		{
			name: "auto mode with username exercises password flow",
			input: authInput{
				authMode: constants.AuthModeAuto,
				clientID: constants.OAuthClientID,
				password: "pass",
				username: "user",
			},
			mockEndpoints: getPasswordEndpoints(passwordConfig{
				discoveryConfig: successfulPasswordAuthDiscovery,
				tokenConfig: stubs.StubHTTPPartialPostConfig{
					ExpectedRequestBody: `audience=` + url.QueryEscape(constants.OAuthAudience) +
						`&client_id=` + url.QueryEscape(constants.OAuthClientID) +
						`&grant_type=password&password=pass&scope=` + url.QueryEscape(constants.OAuthScope) +
						`&username=user`,
					ResponseStatus: http.StatusOK,
					ResponseBody:   `{"access_token":"pw-auto-token"}`,
				},
			}),
			expected: authExpected{
				token: "pw-auto-token",
			},
		},
		{
			name: "refresh succeeds but token cache save fails - warns and returns refreshed token",
			input: authInput{
				authMode: constants.AuthModeAuto,
				cache: fmt.Sprintf(
					`{"access_token":"old","refresh_token":"rt","expires_at":"%s"}`,
					time.Now().Add(-time.Hour).UTC().Format(time.RFC3339Nano),
				),
				clientID:      constants.OAuthClientID,
				readOnlyCache: true,
			},
			mockEndpoints: getDeviceEndpoints(deviceConfig{
				discoveryConfig: successfulDeviceAuthDiscovery,
				tokenConfigs: []stubs.StubHTTPPartialPostConfig{{
					ExpectedRequestBody: `audience=` + url.QueryEscape(constants.OAuthAudience) +
						`&client_id=` + url.QueryEscape(constants.OAuthClientID) +
						`&grant_type=refresh_token` +
						`&ingressHost={SANITIZE:127.0.0.1%3A\d+:SANITIZED}` +
						`&refresh_token=rt`,
					ResponseStatus: http.StatusOK,
					ResponseBody:   `{"access_token":"refreshed-nosave","refresh_token":"rt2","expires_in":3600}`,
				}},
			}),
			expected: authExpected{
				token: "refreshed-nosave",
				warns: []string{
					"failed to save device token: failed to save token cache\n" +
						" Caused by:\n" +
						"\t* operation not permitted",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, makeAuthTestFunc(tt))
	}
}

func makeAuthTestFunc(tt authTestCase) func(t *testing.T) {
	return func(t *testing.T) {
		fsys := stubs.StubFS{Fs: afero.NewMemMapFs()}
		home, err := afero.TempDir(fsys, "", "")
		require.NoError(t, err)
		t.Setenv(osutil.HomeEnvVar(), home)
		if tt.input.home == homeAsFile {
			cachePath := filepath.Join(home, constants.DefaultTokenFileName)
			fsys.OpenFunc = func(
				base afero.Fs,
				path string,
			) (afero.File, error) {
				if path == cachePath {
					return nil, fmt.Errorf("open %s: not a directory", cachePath)
				}
				return base.Open(path)
			}
		}
		if tt.input.home == homeAsUnset {
			t.Setenv(osutil.HomeEnvVar(), "")
		}

		uiStub := &stubs.StubUI{AskResponses: tt.input.askResponses}
		srv, hitCounts := stubs.NewStubHTTPServer(t, tt.mockEndpoints)
		t.Cleanup(srv.Close)

		if tt.input.home == homeUnchanged && tt.input.cache != "" {
			// Populate endpoint-keyed cached token entry when requested.
			cachePath := filepath.Join(home, constants.DefaultTokenFileName)
			require.NoError(t, fsys.MkdirAll(filepath.Dir(cachePath), 0o700))
			cacheEntry := map[string]json.RawMessage{srv.URL: json.RawMessage(tt.input.cache)}
			cacheData, err := json.Marshal(cacheEntry)
			require.NoError(t, err)
			require.NoError(t, afero.WriteFile(fsys, cachePath, cacheData, 0o600))
			if tt.input.readOnlyCache {
				fsys = stubs.StubFS{Fs: afero.NewReadOnlyFs(fsys)}
			}
		}
		if tt.input.openFileFunc != nil {
			fsys.OpenFileFunc = tt.input.openFileFunc
		}

		ctx := context.Background()
		if tt.input.cancelContext {
			cancelledCtx, cancel := context.WithCancel(context.Background())
			time.AfterFunc(50*time.Millisecond, cancel)
			ctx = cancelledCtx
		}
		controlPlaneEndpoint := srv.URL
		if tt.input.controlPlaneEndpoint != "" {
			controlPlaneEndpoint = tt.input.controlPlaneEndpoint
		}
		if tt.input.pkceStateValidator != nil {
			auth.SetStateValidator(tt.input.pkceStateValidator)
			t.Cleanup(func() { auth.SetStateValidator(nil) })
		}
		pkceResponderDoneChan := make(chan any)
		if tt.input.pkceAuthorizationResponse != "" {
			auth.SetListenPort(tt.input.pkceListenPort)
			t.Cleanup(func() { auth.SetListenPort(0) })
			go func() {
				listenPort := tt.input.pkceListenPort
				if listenPort == 0 {
					listenPort = 20000
				}
				time.Sleep(100 * time.Millisecond)
				req, err := http.NewRequestWithContext(context.Background(), http.MethodGet,
					fmt.Sprintf(
						"http://127.0.0.1:%d/callback",
						listenPort,
					)+"?"+tt.input.pkceAuthorizationResponse, nil)
				assert.NoError(t, err)
				res, _ := http.DefaultClient.Do(req)
				if res != nil && res.Body != nil {
					_ = res.Body.Close()
				}
				close(pkceResponderDoneChan)
			}()
		} else {
			close(pkceResponderDoneChan)
		}
		if tt.input.pkceCallbackTimeout != nil {
			auth.SetCallbackTimeout(tt.input.pkceCallbackTimeout)
			t.Cleanup(func() { auth.SetCallbackTimeout(nil) })
		}
		res, err := auth.GetAccessToken(
			ctx,
			uiStub,
			fsys,
			auth.GetAccessTokenParams{
				ControlPlaneEndpoint: controlPlaneEndpoint,
				AuthMode:             tt.input.authMode,
				Username:             tt.input.username,
				Password:             tt.input.password,
				ClientID:             tt.input.clientID,
				ClientSecret:         tt.input.clientSecret,
				NoListen:             tt.input.noListen,
				Organization:         tt.input.organization,
			},
		)
		<-pkceResponderDoneChan

		replacements := map[string]string{
			"${PORT}": stubs.Port(t, srv.URL),
			"${HOST}": stubs.Host(t, srv.URL),
			"${HOME}": home,
			"${URL}":  srv.URL,
		}
		if len(tt.expected.displays) == 0 {
			require.Empty(t, uiStub.Displays, "unexpected display messages: %v", uiStub.Displays)
		} else {
			expectedDisplays, actualDisplays := stubs.SanitizeSlice(
				tt.expected.displays,
				uiStub.Displays,
				replacements,
			)
			require.Equal(t,
				expectedDisplays,
				actualDisplays,
				"unexpected display messages: %v",
				uiStub.Displays,
			)
		}
		if len(tt.expected.warns) == 0 {
			require.Empty(t, uiStub.Warns, "unexpected warn messages: %v", uiStub.Warns)
		} else {
			expectedWarns, actualWarns := stubs.SanitizeSlice(
				tt.expected.warns,
				uiStub.Warns,
				replacements,
			)
			require.Equal(t,
				expectedWarns,
				actualWarns,
				"unexpected warn messages: %v",
				uiStub.Warns,
			)
		}
		if len(tt.expected.asks) == 0 {
			require.Empty(t, uiStub.Asks, "unexpected asks: %v", uiStub.Asks)
		} else {
			expectedAsks, actualAsks := stubs.SanitizeSlice(
				tt.expected.asks,
				uiStub.Asks,
				replacements,
			)
			require.Equal(t,
				expectedAsks,
				actualAsks,
				"unexpected ask: %v",
				uiStub.Asks,
			)
		}
		// Verify all endpoints are hit the expected number of times.
		expectedHits := map[string]int{}
		for _, endpoint := range tt.mockEndpoints {
			key := endpoint.Request.Method + endpoint.Request.URL
			expectedHits[key]++
		}
		for k, v := range *hitCounts {
			assert.Equal(t, expectedHits[k], v, k)
		}
		// Verify any expected error is returned
		if tt.expected.errFormat != "" {
			expectedErr := tt.expected.errFormat
			expectedErr = stubs.Replace(expectedErr, replacements)
			require.EqualError(t, err, expectedErr)
			return
		}
		// Verify any expected success is returned
		require.NoError(t, err)
		require.Equal(t, tt.expected.token, res)
		if tt.expected.cachedToken != nil {
			verifyTokenCacheContents(t, fsys, home, srv.URL, *tt.expected.cachedToken)
		}
	}
}

func TestIsRole(t *testing.T) {
	tests := []struct {
		name     string
		tok      string
		role     string
		expected bool
	}{
		{
			name:     "invalid token format",
			tok:      "abc",
			role:     "r",
			expected: false,
		},
		{
			name:     "role absent",
			tok:      jwtForRoles([]string{"a"}),
			role:     "r",
			expected: false,
		},
		{
			name:     "role present",
			tok:      jwtForRoles([]string{"r"}),
			role:     "r",
			expected: true,
		},
		{
			name:     "invalid json payload",
			tok:      "invalid.json.encoding",
			role:     "r",
			expected: false,
		},
		{
			name:     "invalid base642 payload",
			tok:      "aaa.!!!!.bbb",
			role:     "r",
			expected: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.expected, auth.IsRole(tt.tok, tt.role))
		})
	}
}

func jwtForRoles(roles []string) string {
	payload := map[string]any{
		"sas": map[string]any{
			"roles": roles,
		},
	}
	// This is a test payload that will always marshal.
	// nolint:errchkjson
	b, _ := json.Marshal(payload)
	enc := base64.StdEncoding.EncodeToString(b)
	return "h." + enc + ".s"
}

type expectedTokenCache struct {
	accessToken  string
	refreshToken string
}

// verifyTokenCacheContents reads endpoint-keyed token cache data from
// $HOME/.wbctl/token.json and verifies the entry for endpoint.
func verifyTokenCacheContents(
	t *testing.T,
	fsys afero.Fs,
	home, endpoint string,
	expected expectedTokenCache,
) {
	t.Helper()
	cachePath := filepath.Join(home, constants.DefaultTokenFileName)
	data, err := afero.ReadFile(fsys, cachePath)
	require.NoError(t, err, "failed to read token cache file at %s", cachePath)

	var cached map[string]map[string]any
	err = json.Unmarshal(data, &cached)
	require.NoError(t, err, "failed to unmarshal token cache JSON")

	entry, ok := cached[endpoint]
	require.True(t, ok, "missing endpoint entry in cache: %s", endpoint)

	if expected.accessToken != "" {
		require.Equal(t, expected.accessToken, entry["access_token"], "access_token mismatch in cache")
	}
	if expected.refreshToken != "" {
		require.Equal(t, expected.refreshToken, entry["refresh_token"], "refresh_token mismatch in cache")
	}
}

type passwordConfig struct {
	discoveryConfig stubs.StubHTTPPartialGetConfig
	tokenConfig     stubs.StubHTTPPartialPostConfig
}

func getPasswordEndpoints(
	config passwordConfig,
) []stubs.StubHTTPEndpoint {
	return []stubs.StubHTTPEndpoint{
		{
			Request: stubs.StubHTTPRequest{
				Method: http.MethodGet,
				URL:    constants.AuthWellKnownPath,
			},
			Response: stubs.StubHTTPResponse{
				Status:      config.discoveryConfig.ResponseStatus,
				Body:        config.discoveryConfig.ResponseBody,
				BodyOptions: []stubs.StubHTTPBodyOption{stubs.StubHTTPBodyReplaceHOST()},
			},
		},
		{
			Request: stubs.StubHTTPRequest{
				Method: http.MethodPost,
				URL:    "/token",
				ExpectedHeaders: map[string]string{
					constants.ContentTypeHeaderKey: constants.FormContentType,
				},
				ExpectedBody: config.tokenConfig.ExpectedRequestBody,
			},
			Response: stubs.StubHTTPResponse{
				Status: config.tokenConfig.ResponseStatus,
				Body:   config.tokenConfig.ResponseBody,
			},
		},
	}
}

type deviceConfig struct {
	discoveryConfig stubs.StubHTTPPartialGetConfig
	tokenConfigs    []stubs.StubHTTPPartialPostConfig
	deviceConfig    stubs.StubHTTPPartialPostConfig
}

func getDeviceAuthEndpoints(
	authURL string,
	config deviceConfig,
) []stubs.StubHTTPEndpoint {
	endpoints := make([]stubs.StubHTTPEndpoint, 0, len(config.tokenConfigs)+2)
	endpoints = append(endpoints, []stubs.StubHTTPEndpoint{
		{
			Request: stubs.StubHTTPRequest{
				Method: http.MethodGet,
				URL:    authURL,
			},
			Response: stubs.StubHTTPResponse{
				Status:      config.discoveryConfig.ResponseStatus,
				Body:        config.discoveryConfig.ResponseBody,
				BodyOptions: []stubs.StubHTTPBodyOption{stubs.StubHTTPBodyReplaceHOST()},
			},
		},
		{
			Request: stubs.StubHTTPRequest{
				Method: http.MethodPost,
				URL:    "/device",
				ExpectedHeaders: map[string]string{
					constants.ContentTypeHeaderKey: constants.FormContentType,
				},
				ExpectedBody: config.deviceConfig.ExpectedRequestBody,
			},
			Response: stubs.StubHTTPResponse{
				Status: config.deviceConfig.ResponseStatus,
				Body:   config.deviceConfig.ResponseBody,
			},
		},
	}...)
	for _, tokenConfig := range config.tokenConfigs {
		endpoints = append(endpoints, stubs.StubHTTPEndpoint{
			Request: stubs.StubHTTPRequest{
				Method: http.MethodPost,
				URL:    "/token",
				ExpectedHeaders: map[string]string{
					constants.ContentTypeHeaderKey: constants.FormContentType,
				},
				ExpectedBody: tokenConfig.ExpectedRequestBody,
			},
			Response: stubs.StubHTTPResponse{
				Status: tokenConfig.ResponseStatus,
				Body:   tokenConfig.ResponseBody,
			},
		})
	}
	return endpoints
}

func getDeviceEndpoints(
	config deviceConfig,
) []stubs.StubHTTPEndpoint {
	return getDeviceAuthEndpoints(constants.AuthWellKnownPath, config)
}

func getLegacyDeviceEndpoints(
	config deviceConfig,
) []stubs.StubHTTPEndpoint {
	return getDeviceAuthEndpoints(constants.AuthLegacyWellKnownPath, config)
}
