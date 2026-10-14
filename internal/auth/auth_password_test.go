// Copyright © 2026, SAS Institute Inc., Cary, NC, USA.  All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package auth_test

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"testing"
	"time"

	"github.com/sassoftware/sas-ggdk/pkg/errors"
	"github.com/sassoftware/wbctl/internal/constants"
	"github.com/sassoftware/wbctl/internal/stubs"
	"github.com/spf13/afero"
)

func TestGetAccessTokenPassword(t *testing.T) {
	tests := []authTestCase{
		{
			name: "discovery failure - on password flow causes error",
			input: authInput{
				authMode: constants.AuthModePassword,
				clientID: constants.OAuthClientID,
				username: "user",
				password: "pass",
			},
			mockEndpoints: getPasswordEndpoints(passwordConfig{
				discoveryConfig: stubs.StubHTTPPartialGetConfig{
					ResponseStatus: http.StatusInternalServerError,
					ResponseBody:   `{"error":"discovery"}`,
				},
			}),
			expected: authExpected{
				errFormat: authTestDiscoveryFailurePasswordErrFormat,
			},
		},
		{
			name: "password mode requires username",
			input: authInput{
				authMode: constants.AuthModePassword,
				clientID: constants.OAuthClientID,
			},
			expected: authExpected{
				errFormat: authTestPasswordModeRequiresUsername,
			},
		},
		{
			name: "password mode with empty access token is error",
			input: authInput{
				authMode: constants.AuthModePassword,
				clientID: constants.OAuthClientID,
				username: "user",
				password: "pass",
			},
			mockEndpoints: getPasswordEndpoints(passwordConfig{
				discoveryConfig: successfulPasswordAuthDiscovery,
				tokenConfig: stubs.StubHTTPPartialPostConfig{
					ExpectedRequestBody: `audience=` + url.QueryEscape(constants.OAuthAudience) +
						`&client_id=` + url.QueryEscape(constants.OAuthClientID) +
						`&grant_type=password&password=pass&scope=` + url.QueryEscape(constants.OAuthScope) +
						`&username=user`,
					ResponseStatus: http.StatusOK,
					ResponseBody:   `{"access_token":""}`,
				},
			}),
			expected: authExpected{
				errFormat: authTestPasswordModeEmptyAccessTokenErr,
			},
		},
		// This is a fake token so not a security issue.
		//nolint:gosec
		{
			name: "password mode bypasses device flow",
			input: authInput{
				authMode: constants.AuthModePassword,
				clientID: constants.OAuthClientID,
				username: "user",
				password: "pass",
			},
			mockEndpoints: getPasswordEndpoints(passwordConfig{
				discoveryConfig: successfulPasswordAuthDiscovery,
				tokenConfig: stubs.StubHTTPPartialPostConfig{
					ExpectedRequestBody: `audience=` + url.QueryEscape(constants.OAuthAudience) +
						`&client_id=` + url.QueryEscape(constants.OAuthClientID) +
						`&grant_type=password&password=pass&scope=` + url.QueryEscape(constants.OAuthScope) +
						`&username=user`,
					ResponseStatus: http.StatusOK,
					ResponseBody:   `{"access_token":"pw-device-test-token"}`,
				},
			}),
			expected: authExpected{
				token: "pw-device-test-token",
			},
		},
		// This is a fake token so not a security issue.
		//nolint:gosec
		{
			name: "password mode token failure",
			input: authInput{
				authMode: constants.AuthModePassword,
				clientID: constants.OAuthClientID,
				username: "user",
				password: "pass",
			},
			mockEndpoints: getPasswordEndpoints(passwordConfig{
				discoveryConfig: successfulPasswordAuthDiscovery,
				tokenConfig: stubs.StubHTTPPartialPostConfig{
					ExpectedRequestBody: `audience=` + url.QueryEscape(constants.OAuthAudience) +
						`&client_id=` + url.QueryEscape(constants.OAuthClientID) +
						`&grant_type=password&password=pass&scope=` + url.QueryEscape(constants.OAuthScope) +
						`&username=user`,
					ResponseStatus: http.StatusUnauthorized,
					ResponseBody:   `{"error":"bad"}`,
				},
			}),
			expected: authExpected{
				errFormat: authTestPasswordGrantFailureErrFormat,
			},
		},
		// This is a fake token so not a security issue.
		//nolint:gosec
		{
			name: "password mode cache save failure warns and returns token",
			input: authInput{
				authMode: constants.AuthModePassword,
				clientID: constants.OAuthClientID,
				username: "user",
				password: "pass",
				cache: fmt.Sprintf(
					`{"access_token":"old","refresh_token":"rt","expires_at":"%s"}`,
					time.Now().Add(-time.Hour).UTC().Format(time.RFC3339Nano),
				),
				readOnlyCache: true,
			},
			mockEndpoints: getPasswordEndpoints(passwordConfig{
				discoveryConfig: successfulPasswordAuthDiscovery,
				tokenConfig: stubs.StubHTTPPartialPostConfig{
					ExpectedRequestBody: `audience=` + url.QueryEscape(constants.OAuthAudience) +
						`&client_id=` + url.QueryEscape(constants.OAuthClientID) +
						`&grant_type=password&password=pass&scope=` + url.QueryEscape(constants.OAuthScope) +
						`&username=user`,
					ResponseStatus: http.StatusOK,
					ResponseBody:   `{"access_token":"pw-cache-fail","expires_in":3600}`,
				},
			}),
			expected: authExpected{
				token: "pw-cache-fail",
				warns: []string{
					"failed to save device token: failed to save token cache\n" +
						" Caused by:\n" +
						"\t* operation not permitted",
				},
			},
		},
		// This is a fake token so not a security issue.
		//nolint:gosec
		{
			// saveTokenToCache failure path 1: afero.TempFile fails because
			// OpenFile with O_EXCL returns an error. The deferred save in
			// getAccessTokenByPassword must warn and still return the token.
			name: "password mode save failure - temp file creation error warns and returns token",
			input: authInput{
				authMode: constants.AuthModePassword,
				clientID: constants.OAuthClientID,
				username: "user",
				password: "pass",
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
			mockEndpoints: getPasswordEndpoints(passwordConfig{
				discoveryConfig: successfulPasswordAuthDiscovery,
				tokenConfig: stubs.StubHTTPPartialPostConfig{
					ExpectedRequestBody: `audience=` + url.QueryEscape(constants.OAuthAudience) +
						`&client_id=` + url.QueryEscape(constants.OAuthClientID) +
						`&grant_type=password&password=pass&scope=` + url.QueryEscape(constants.OAuthScope) +
						`&username=user`,
					ResponseStatus: http.StatusOK,
					ResponseBody:   `{"access_token":"pw-save-create-fail","expires_in":3600}`,
				},
			}),
			expected: authExpected{
				token: "pw-save-create-fail",
				warns: []string{
					"failed to save device token: failed to save token cache\n" +
						" Caused by:\n" +
						"\t* tempfile create error",
				},
			},
		},
		// This is a fake token so not a security issue.
		//nolint:gosec
		{
			// saveTokenToCache failure path 2: afero.TempFile succeeds but the
			// subsequent Close on the temp file returns an error. The deferred
			// save must warn and still return the token.
			name: "password mode save failure - temp file close error warns and returns token",
			input: authInput{
				authMode: constants.AuthModePassword,
				clientID: constants.OAuthClientID,
				username: "user",
				password: "pass",
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
			mockEndpoints: getPasswordEndpoints(passwordConfig{
				discoveryConfig: successfulPasswordAuthDiscovery,
				tokenConfig: stubs.StubHTTPPartialPostConfig{
					ExpectedRequestBody: `audience=` + url.QueryEscape(constants.OAuthAudience) +
						`&client_id=` + url.QueryEscape(constants.OAuthClientID) +
						`&grant_type=password&password=pass&scope=` + url.QueryEscape(constants.OAuthScope) +
						`&username=user`,
					ResponseStatus: http.StatusOK,
					ResponseBody:   `{"access_token":"pw-save-close-fail","expires_in":3600}`,
				},
			}),
			expected: authExpected{
				token: "pw-save-close-fail",
				warns: []string{
					"failed to save device token: failed to save token cache\n" +
						" Caused by:\n" +
						"\t* tempfile close error",
				},
			},
		},
		// This is a fake token so not a security issue.
		//nolint:gosec
		{
			// saveTokenToCache failure path 3: TempFile and Close both succeed
			// but the subsequent jsonutils.SaveWith / afero.WriteFile call
			// fails because OpenFile with O_TRUNC returns an error. The
			// deferred save must warn and still return the token.
			name: "password mode save failure - write error warns and returns token",
			input: authInput{
				authMode: constants.AuthModePassword,
				clientID: constants.OAuthClientID,
				username: "user",
				password: "pass",
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
			mockEndpoints: getPasswordEndpoints(passwordConfig{
				discoveryConfig: successfulPasswordAuthDiscovery,
				tokenConfig: stubs.StubHTTPPartialPostConfig{
					ExpectedRequestBody: `audience=` + url.QueryEscape(constants.OAuthAudience) +
						`&client_id=` + url.QueryEscape(constants.OAuthClientID) +
						`&grant_type=password&password=pass&scope=` + url.QueryEscape(constants.OAuthScope) +
						`&username=user`,
					ResponseStatus: http.StatusOK,
					ResponseBody:   `{"access_token":"pw-save-write-fail","expires_in":3600}`,
				},
			}),
			expected: authExpected{
				token: "pw-save-write-fail",
				warns: []string{
					"failed to save device token: failed to save token cache\n" +
						" Caused by:\n" +
						"\t* tempfile write error",
				},
			},
		},
		{
			name: "password mode with invalid token url is error",
			input: authInput{
				authMode: constants.AuthModePassword,
				clientID: constants.OAuthClientID,
				username: "user",
				password: "pass",
			},
			mockEndpoints: getPasswordEndpoints(passwordConfig{
				discoveryConfig: stubs.StubHTTPPartialGetConfig{
					ResponseStatus: http.StatusOK,
					ResponseBody:   `{"token_endpoint":"a\nb"}`,
				},
			}),
			expected: authExpected{
				errFormat: authTestPasswordModeInvalidTokenURLErr,
			},
		},
		// This is a fake token so not a security issue.
		//nolint:gosec
		{
			name: "password mode discovered audience and client id",
			input: authInput{
				authMode: constants.AuthModePassword,
				username: "user",
				password: "pass",
			},
			mockEndpoints: getPasswordEndpoints(passwordConfig{
				discoveryConfig: successfulPasswordAuthDiscoveryWithClientID,
				tokenConfig: stubs.StubHTTPPartialPostConfig{
					ExpectedRequestBody: `audience=discovered_audience` +
						`&client_id=discovered_client_id` +
						`&grant_type=password&password=pass&scope=` + url.QueryEscape(constants.OAuthScope) +
						`&username=user`,
					ResponseStatus: http.StatusOK,
					ResponseBody:   `{"access_token":"pw-device-test-token"}`,
				},
			}),
			expected: authExpected{
				token: "pw-device-test-token",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, makeAuthTestFunc(tt))
	}
}
