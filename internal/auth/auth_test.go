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

	"github.com/sassoftware/sas-ggdk/pkg/errors"
	"github.com/sassoftware/wbctl/internal/auth"
	"github.com/sassoftware/wbctl/internal/constants"
	"github.com/sassoftware/wbctl/internal/osutil"
	"github.com/sassoftware/wbctl/internal/postgrest"
	"github.com/sassoftware/wbctl/internal/stubs"
	"github.com/spf13/afero"
	"github.com/spf13/viper"
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
	authTestGetCSRFFailureErrFormat = "failed to get CSRF token\n" +
		" Caused by:\n" +
		"\t* HEAD ${URL} failed\n" +
		"\t* 500 "
	authTestInvalidAuthModeErr           = "invalid --auth-mode: \"invalid\" (allowed: auto, password, device)"
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
	authTestDeviceModeEmptyAccessTokenErr = "failed to get device token\n" +
		" Caused by:\n" +
		"\t* token response missing access_token"
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
	authTestDeviceModeMissingOrganizationErr = "failed to get device token\n" +
		" Caused by:\n" +
		"\t* --authmode=device requires --organization"
	authTestPKCEModeInvalidControlPlaneURLErr = "failed to get pkce token\n" +
		" Caused by:\n" +
		"\t* parse \"a\\nb\": net/url: invalid control character in URL"
	authTestDeviceModeInvalidRefreshURLErr = "failed to get device token\n" +
		" Caused by:\n" +
		"\t* POST  failed\n" +
		"\t* Post \"\": unsupported protocol scheme \"\""
	authTestInvalidControlPlaneEndpoint = "a\nb"
	// nolint:gosec
	encodedAccessToken = ".eyJhenAiOiJlbmNvZGVkX2NsaWVudF9pZCJ9Cg."
)

var (
	expectedTokenRequestBody = `audience=discovered_audience` +
		`&client_id=discovered_native_client_id` +
		`&device_code=dc&grant_type=urn%3Aietf%3Aparams%3Aoauth%3Agrant-type%3Adevice_code`
	successfulPasswordAuthDiscovery = stubs.StubHTTPPartialGetConfig{
		ResponseStatus: http.StatusOK,
		ResponseBody:   `{"token_endpoint":"http://${HOST}/token"}`,
	}
	successfulPasswordAuthDiscoveryWithClientID = stubs.StubHTTPPartialGetConfig{
		ResponseStatus: http.StatusOK,
		ResponseBody: `
{
  "token_endpoint": "http://${HOST}/token",
  "workbench": {
    "spa_audience": "discovered_audience",
    "native_client_id": "discovered_client_id"
  }
}
`,
	}
	successfulDeviceAuthDiscovery = stubs.StubHTTPPartialGetConfig{
		ResponseStatus: http.StatusOK,
		ResponseBody:   `{"token_endpoint":"http://${HOST}/token", "device_authorization_endpoint":"http://${HOST}/device"}`,
	}
	successfulDeviceAuthDiscoveryWithClientID = stubs.StubHTTPPartialGetConfig{
		ResponseStatus: http.StatusOK,
		ResponseBody: `
{
  "authorization_endpoint": "http://${HOST}/authorization",
  "token_endpoint": "http://${HOST}/token",
  "device_authorization_endpoint": "http://${HOST}/device",
  "workbench": {
    "spa_client_id": "discovered_spa_client_id",
    "spa_audience": "discovered_audience",
    "native_client_id": "discovered_native_client_id"
  }
}
`,
	}
	successfulPKCEAuthDiscovery = stubs.StubHTTPPartialGetConfig{
		ResponseStatus: http.StatusOK,
		ResponseBody: `
{
  "token_endpoint": "http://${HOST}/token",
  "authorization_endpoint": "http://${HOST}/authorization",
  "workbench": {
    "spa_audience": "discovered_audience",
    "spa_client_id": "discovered_client_id"
  }
}
`,
	}
	successfulDevice = stubs.StubHTTPPartialPostConfig{
		ExpectedRequestBody: `audience=discovered_audience` +
			`&client_id=discovered_native_client_id` +
			`&scope=` + url.QueryEscape(constants.OAuthScope) + `+organization%3Aorg-1`,
		ResponseStatus: http.StatusOK,
		ResponseBody:   authTestDeviceAuthorizationBody,
	}
	successfulDeviceWithOrg = stubs.StubHTTPPartialPostConfig{
		ExpectedRequestBody: `audience=discovered_audience` +
			`&client_id=discovered_native_client_id` +
			`&scope=` + url.QueryEscape(constants.OAuthScope+" organization:org-1"),
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
	stubUI := &stubs.StubUI{}
	fsys := afero.NewMemMapFs()
	_, err := auth.GetAccessToken(
		context.Background(),
		stubUI,
		fsys,
		viper.New(),
		auth.GetAccessTokenParams{
			ControlPlaneEndpoint: "",
			AuthMode:             constants.AuthModeDevice,
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
	profile                   string
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
				authMode:             constants.AuthModePKCE,
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
			},
			expected: authExpected{
				errFormat: authTestInvalidAuthModeErr,
			},
		},
		{
			name: "valid cached token short-circuits and returns cached token",
			input: authInput{
				authMode: constants.AuthModePKCE,
				cache: fmt.Sprintf(
					`{"access_token":"cached","refresh_token":"rt","expires_at":"%s"}`,
					time.Now().Add(time.Hour).UTC().Format(time.RFC3339Nano),
				),
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
			name: "refresh success returns refreshed token without device flow",
			input: authInput{
				authMode: constants.AuthModePKCE,
				cache: fmt.Sprintf(
					`{"access_token":"`+encodedAccessToken+`","refresh_token":"rt","expires_at":"%s"}`,
					time.Now().Add(-time.Hour).UTC().Format(time.RFC3339Nano),
				),
			},
			mockEndpoints: getAuthEndpoints(authConfig{
				discoveryConfig: successfulDeviceAuthDiscovery,
				tokenConfigs: []stubs.StubHTTPPartialPostConfig{{
					ExpectedRequestBody: `audience=` + url.QueryEscape(constants.OAuthAudience) +
						`&client_id=encoded_client_id` +
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
			name: "discovery failure - an expired token causes device flow",
			input: authInput{
				authMode: constants.AuthModePKCE,
				cache: fmt.Sprintf(
					`{"access_token":"old","refresh_token":"rt","expires_at":"%s"}`,
					time.Now().Add(-time.Hour).UTC().Format(time.RFC3339Nano),
				),
			},
			mockEndpoints: getAuthEndpoints(authConfig{
				discoveryConfig: stubs.StubHTTPPartialGetConfig{
					ResponseStatus: http.StatusInternalServerError,
					ResponseBody:   `{"error":"discovery"}`,
				},
			}),
			expected: authExpected{
				errFormat: authTestDiscoveryFailureErrFormat,
			},
		},
		{
			name: "refresh succeeds but token cache save fails - warns and returns refreshed token",
			input: authInput{
				authMode: constants.AuthModePKCE,
				cache: fmt.Sprintf(
					`{"access_token":"`+encodedAccessToken+`","refresh_token":"rt","expires_at":"%s"}`,
					time.Now().Add(-time.Hour).UTC().Format(time.RFC3339Nano),
				),
				readOnlyCache: true,
			},
			mockEndpoints: getAuthEndpoints(authConfig{
				discoveryConfig: successfulDeviceAuthDiscovery,
				tokenConfigs: []stubs.StubHTTPPartialPostConfig{{
					ExpectedRequestBody: `audience=` + url.QueryEscape(constants.OAuthAudience) +
						`&client_id=encoded_client_id` +
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
		{
			// saveTokenToCache failure path 1: afero.TempFile fails because
			// OpenFile with O_EXCL returns an error. The deferred save in
			// getAccessTokenByPassword must warn and still return the token.
			name: "device mode save failure - temp file creation error warns and returns token",
			input: authInput{
				authMode:     constants.AuthModeDevice,
				organization: "org-1",
				openFileFunc: func(
					base afero.Fs,
					name string,
					flag int,
					perm os.FileMode,
				) (afero.File, error) {
					if flag&os.O_EXCL != 0 {
						return nil, errors.New("tempfile create error")
					}
					return base.OpenFile(name, flag, perm)
				},
			},
			mockEndpoints: stubs.GetStubAuthEndpoints(),
			expected: authExpected{
				token:    "access-token",
				displays: []string{stubs.ExpectedDeviceAuthDisplay},
				warns: []string{
					"failed to save device token: failed to save token cache\n" +
						" Caused by:\n" +
						"\t* tempfile create error",
				},
			},
		},
		{
			// saveTokenToCache failure path 2: afero.TempFile succeeds but the
			// subsequent Close on the temp file returns an error. The deferred
			// save must warn and still return the token.
			name: "device mode save failure - temp file close error warns and returns token",
			input: authInput{
				authMode:     constants.AuthModeDevice,
				organization: "org-1",
				openFileFunc: func(
					base afero.Fs,
					name string,
					flag int,
					perm os.FileMode,
				) (afero.File, error) {
					f, err := base.OpenFile(name, flag, perm)
					if err != nil || flag&os.O_EXCL == 0 {
						return f, err
					}
					// Wrap the real temp file so Close returns an error.
					return stubs.StubFile{File: f, CloseFunc: func(_ afero.File) error {
						return errors.New("tempfile close error")
					}}, nil
				},
			},
			mockEndpoints: stubs.GetStubAuthEndpoints(),
			expected: authExpected{
				token:    "access-token",
				displays: []string{stubs.ExpectedDeviceAuthDisplay},
				warns: []string{
					"failed to save device token: failed to save token cache\n" +
						" Caused by:\n" +
						"\t* tempfile close error",
				},
			},
		},
		{
			// saveTokenToCache failure path 3: TempFile and Close both succeed
			// but the subsequent jsonutils.SaveWith / afero.WriteFile call
			// fails because OpenFile with O_TRUNC returns an error. The
			// deferred save must warn and still return the token.
			name: "device mode save failure - write error warns and returns token",
			input: authInput{
				authMode:     constants.AuthModeDevice,
				organization: "org-1",
				openFileFunc: func(
					base afero.Fs,
					name string,
					flag int,
					perm os.FileMode,
				) (afero.File, error) {
					if flag&os.O_EXCL != 0 {
						// TempFile call – let it succeed.
						return base.OpenFile(name, flag, perm)
					}
					if flag&os.O_TRUNC != 0 {
						// WriteFile call inside jsonutils.SaveWith – fail it.
						return nil, errors.New("tempfile write error")
					}
					return base.OpenFile(name, flag, perm)
				},
			},
			mockEndpoints: stubs.GetStubAuthEndpoints(),
			expected: authExpected{
				token:    "access-token",
				displays: []string{stubs.ExpectedDeviceAuthDisplay},
				warns: []string{
					"failed to save device token: failed to save token cache\n" +
						" Caused by:\n" +
						"\t* tempfile write error",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, makeAuthTestFunc(tt))
	}
}

type removeInput struct {
	profile      string
	endpoint     string
	cacheContent string
}

type removeExpected struct {
	cacheContent string
}

type removeTestCase struct {
	name     string
	input    removeInput
	expected removeExpected
}

func TestRemoveAccessToken(t *testing.T) {
	tests := []removeTestCase{
		{
			name: "simple success",
			input: removeInput{
				endpoint:     "https://example.com",
				cacheContent: `{"https://example.com": {}}`,
			},
			expected: removeExpected{
				cacheContent: `{}`,
			},
		},
		{
			name: "simple success with profile",
			input: removeInput{
				profile:      "my-profile",
				endpoint:     "https://example.com",
				cacheContent: `{"my-profile.https://example.com": {}}`,
			},
			expected: removeExpected{
				cacheContent: `{}`,
			},
		},
		{
			name: "no op",
			input: removeInput{
				endpoint:     "https://does not exist",
				cacheContent: `{"https://example.com": {}}`,
			},
			expected: removeExpected{
				cacheContent: `{"https://example.com": {}}`,
			},
		},
		{
			name: "no op with profile",
			input: removeInput{
				endpoint:     "https://does not exist",
				cacheContent: `{"Default.https://example.com": {}}`,
			},
			expected: removeExpected{
				cacheContent: `{"Default.https://example.com": {}}`,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fsys := stubs.StubFS{Fs: afero.NewMemMapFs()}
			home, err := afero.TempDir(fsys, "", "")
			require.NoError(t, err)
			t.Setenv(osutil.HomeEnvVar(), home)
			tokenFilePath := filepath.Join(home, constants.DefaultTokenFileName)
			err = afero.WriteFile(fsys, tokenFilePath, []byte(tt.input.cacheContent), 0644)
			require.NoError(t, err)
			v := viper.New()
			if tt.input.profile != "" {
				v.Set(constants.ProfileFlag, tt.input.profile)
			}
			err = auth.RemoveAccessToken(fsys, v, tt.input.endpoint)
			require.NoError(t, err)
			actual, err := afero.ReadFile(fsys, tokenFilePath)
			require.NoError(t, err)
			require.Equal(t, tt.expected.cacheContent, string(actual))
		})
	}
}

func makeAuthTestFunc(tt authTestCase) func(t *testing.T) {
	return func(t *testing.T) {
		fsys := stubs.StubFS{Fs: afero.NewMemMapFs()}
		home, err := afero.TempDir(fsys, "", "")
		require.NoError(t, err)
		t.Setenv(osutil.HomeEnvVar(), home)
		if tt.input.profile == "" {
			tt.input.profile = "Default"
		}
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

		stubUI := &stubs.StubUI{AskResponses: tt.input.askResponses}
		srv, hitCounts := stubs.NewStubHTTPServer(t, tt.mockEndpoints)
		t.Cleanup(srv.Close)

		if tt.input.home == homeUnchanged && tt.input.cache != "" {
			// Populate endpoint-keyed cached token entry when requested.
			cachePath := filepath.Join(home, constants.DefaultTokenFileName)
			require.NoError(t, fsys.MkdirAll(filepath.Dir(cachePath), 0o700))
			cacheEntry := map[string]json.RawMessage{tt.input.profile + "." + srv.URL: json.RawMessage(tt.input.cache)}
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
		v := viper.New()
		v.Set(constants.ProfileFlag, tt.input.profile)
		res, err := auth.GetAccessToken(
			ctx,
			stubUI,
			fsys,
			v,
			auth.GetAccessTokenParams{
				ControlPlaneEndpoint: controlPlaneEndpoint,
				AuthMode:             tt.input.authMode,
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
		// Verify any expected error is returned
		if tt.expected.errFormat != "" {
			expectedErr := tt.expected.errFormat
			expectedErr = stubs.Replace(expectedErr, replacements)
			require.EqualError(t, err, expectedErr)
		} else {
			// Verify any expected success is returned
			require.NoError(t, err)
			require.Equal(t, tt.expected.token, res)
			if tt.expected.cachedToken != nil {
				verifyTokenCacheContents(t, fsys, home, tt.input.profile, srv.URL, *tt.expected.cachedToken)
			}
		}
		if len(tt.expected.displays) == 0 {
			require.Empty(t, stubUI.Displays, "unexpected display messages: %v", stubUI.Displays)
		} else {
			expectedDisplays, actualDisplays := stubs.SanitizeSlice(
				tt.expected.displays,
				stubUI.Displays,
				replacements,
			)
			require.Equal(t,
				expectedDisplays,
				actualDisplays,
				"unexpected display messages: %v",
				stubUI.Displays,
			)
		}
		if len(tt.expected.warns) == 0 {
			require.Empty(t, stubUI.Warns, "unexpected warn messages: %v", stubUI.Warns)
		} else {
			expectedWarns, actualWarns := stubs.SanitizeSlice(
				tt.expected.warns,
				stubUI.Warns,
				replacements,
			)
			require.Equal(t,
				expectedWarns,
				actualWarns,
				"unexpected warn messages: %v",
				stubUI.Warns,
			)
		}
		if len(tt.expected.asks) == 0 {
			require.Empty(t, stubUI.Asks, "unexpected asks: %v", stubUI.Asks)
		} else {
			expectedAsks, actualAsks := stubs.SanitizeSlice(
				tt.expected.asks,
				stubUI.Asks,
				replacements,
			)
			require.Equal(t,
				expectedAsks,
				actualAsks,
				"unexpected ask: %v",
				stubUI.Asks,
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

const (
	fakeCache = `
{
  "https://example.com": {
    "access_token": "access-token",
    "expires_at": "2026-10-26T21:02:17.516241Z",
    "id_token": "header.payload.signature",
    "refresh_token": "refresh-token",
    "token_type": "Bearer"
  }
}
`
	fakeCacheWithProfile = `
{
  "profile.https://example.com": {
    "access_token": "access-token",
    "expires_at": "2026-10-26T21:02:17.516241Z",
    "id_token": "header.payload.signature",
    "refresh_token": "refresh-token",
    "token_type": "Bearer"
  }
}
`
	invalidCache = `invalid`
)

type getCachedIdentityTokenInput struct {
	cacheContent string
	endpoint     string
	profile      string
}

type getCachedIdentityTokenExpected struct {
	idToken string
}

type getCachedIdentityTokenTestCase struct {
	name     string
	input    getCachedIdentityTokenInput
	expected getCachedIdentityTokenExpected
}

func TestGetCachedIdentityToken(t *testing.T) {
	tests := []getCachedIdentityTokenTestCase{
		{
			name: "no cached token",
			expected: getCachedIdentityTokenExpected{
				idToken: "",
			},
		},
		{
			name: "fake cached token",
			input: getCachedIdentityTokenInput{
				cacheContent: fakeCache,
				endpoint:     "https://example.com",
			},
			expected: getCachedIdentityTokenExpected{
				idToken: "header.payload.signature",
			},
		},
		{
			name: "fake cached token with profile",
			input: getCachedIdentityTokenInput{
				cacheContent: fakeCacheWithProfile,
				endpoint:     "https://example.com",
				profile:      "profile",
			},
			expected: getCachedIdentityTokenExpected{
				idToken: "header.payload.signature",
			},
		},
		{
			name: "invalid cached token",
			input: getCachedIdentityTokenInput{
				cacheContent: invalidCache,
				endpoint:     "https://example.com",
			},
			expected: getCachedIdentityTokenExpected{
				idToken: "",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fsys := stubs.StubFS{Fs: afero.NewMemMapFs()}
			home, err := afero.TempDir(fsys, "", "")
			require.NoError(t, err)
			t.Setenv(osutil.HomeEnvVar(), home)
			if tt.input.cacheContent != "" {
				tokenFilePath := filepath.Join(home, constants.DefaultTokenFileName)
				err = afero.WriteFile(fsys, tokenFilePath, []byte(tt.input.cacheContent), 0644)
				require.NoError(t, err)
			}
			v := viper.New()
			if tt.input.profile != "" {
				v.Set(constants.ProfileFlag, tt.input.profile)
			}
			actual := auth.GetCachedIdentityToken(fsys, v, tt.input.endpoint)
			require.Equal(t, tt.expected.idToken, actual)
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
	home, profile, endpoint string,
	expected expectedTokenCache,
) {
	t.Helper()
	cachePath := filepath.Join(home, constants.DefaultTokenFileName)
	data, err := afero.ReadFile(fsys, cachePath)
	require.NoError(t, err, "failed to read token cache file at %s", cachePath)

	var cached map[string]map[string]any
	err = json.Unmarshal(data, &cached)
	require.NoError(t, err, "failed to unmarshal token cache JSON")

	key := endpoint
	if profile != "" {
		key = fmt.Sprintf("%s.%s", profile, key)
	}
	entry, ok := cached[key]
	require.True(t, ok, "missing endpoint entry in cache: %s", endpoint)

	if expected.accessToken != "" {
		require.Equal(t, expected.accessToken, entry["access_token"], "access_token mismatch in cache")
	}
	if expected.refreshToken != "" {
		require.Equal(t, expected.refreshToken, entry["refresh_token"], "refresh_token mismatch in cache")
	}
}

type pkceConfig struct {
	discoveryConfig stubs.StubHTTPPartialGetConfig
	tokenConfig     stubs.StubHTTPPartialPostConfig
}

func getPKCEEndpoints(
	config pkceConfig,
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

type authConfig struct {
	discoveryConfig stubs.StubHTTPPartialGetConfig
	tokenConfigs    []stubs.StubHTTPPartialPostConfig
	deviceConfig    stubs.StubHTTPPartialPostConfig
}

func getDeviceAuthEndpoints(
	authURL string,
	config authConfig,
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

func getAuthEndpoints(
	config authConfig,
) []stubs.StubHTTPEndpoint {
	return getDeviceAuthEndpoints(constants.AuthWellKnownPath, config)
}

func getLegacyAuthEndpoints(
	config authConfig,
) []stubs.StubHTTPEndpoint {
	return getDeviceAuthEndpoints(constants.AuthLegacyWellKnownPath, config)
}
