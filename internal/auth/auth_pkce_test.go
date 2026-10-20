// Copyright © 2026, SAS Institute Inc., Cary, NC, USA.  All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package auth_test

import (
	"math"
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/sassoftware/sas-ggdk/pkg/pointer"
	"github.com/sassoftware/wbctl/internal/auth"
	"github.com/sassoftware/wbctl/internal/constants"
	"github.com/sassoftware/wbctl/internal/stubs"
)

func TestGetAccessTokenPKCE(t *testing.T) {
	tests := []authTestCase{
		{
			name: "invalid control plane endpoint returns discovery error - pkce",
			input: authInput{
				authMode:             constants.AuthModePKCE,
				controlPlaneEndpoint: authTestInvalidControlPlaneEndpoint,
			},
			expected: authExpected{
				errFormat: authTestPKCEModeInvalidControlPlaneURLErr,
			},
		},
		{
			name: "pkce with no listener available fails state validation",
			input: authInput{
				authMode:     constants.AuthModePKCE,
				noListen:     true,
				organization: "org-1",
				askResponses: []string{"http://example.com?code=abc"},
			},
			mockEndpoints: getPKCEEndpoints(pkceConfig{
				discoveryConfig: successfulPKCEAuthDiscovery,
			}),
			expected: authExpected{
				displays: []string{
					"Open this URL in your browser and complete the authorization:\n" +
						"${URL}/authorization?" +
						"audience=discovered_audience" +
						"&client_id=discovered_client_id" +
						"{SANITIZE:&code_challenge=.*?&:&code_challenge=SANITIZED&}" +
						"code_challenge_method=S256" +
						"&ingressHost=${HOST}%3A${PORT}" +
						"&organization=org-1" +
						"&redirect_uri=http%3A%2F%2Flocalhost%3A20000%2F" +
						"&response_type=code" +
						"&scope=openid+profile+offline_access" +
						"{SANITIZE:&state=.*:&state=SANITIZED}",
					"After completing authorization, paste the full redirect URL here:",
				},
				warns: []string{
					"NOTE: You will see a connection error in your browser. This is expected.",
				},
				asks: []string{
					"Browser URL",
				},
				errFormat: "failed to get pkce token\n" +
					" Caused by:\n" +
					"\t* state mismatch in authentication response",
			},
		},
		{
			name: "pkce with no listener available prompts and honors explicit clientid",
			input: authInput{
				authMode:           constants.AuthModePKCE,
				clientID:           "explicit_client_id",
				noListen:           true,
				organization:       "org-1",
				askResponses:       []string{"http://example.com?code=abc"},
				pkceStateValidator: func(string, string) bool { return true },
			},
			mockEndpoints: getPKCEEndpoints(pkceConfig{
				discoveryConfig: successfulPKCEAuthDiscovery,
				tokenConfig: stubs.StubHTTPPartialPostConfig{
					ExpectedRequestBody: `client_id=explicit_client_id` +
						`&code=abc` +
						`{SANITIZE:&code_verifier=.*?&:&code_verifier=SANITIZED&}` +
						`grant_type=authorization_code` +
						`&redirect_uri=` + url.QueryEscape(constants.PKCECallbackURL),
					ResponseStatus: http.StatusOK,
					ResponseBody:   `{"access_token":"pkce-token"}`,
				},
			}),
			expected: authExpected{
				displays: []string{
					"Open this URL in your browser and complete the authorization:\n" +
						"${URL}/authorization?" +
						"audience=discovered_audience" +
						"&client_id=explicit_client_id" +
						"{SANITIZE:&code_challenge=.*?&:&code_challenge=SANITIZED&}" +
						"code_challenge_method=S256" +
						"&ingressHost=${HOST}%3A${PORT}" +
						"&organization=org-1" +
						"&redirect_uri=http%3A%2F%2Flocalhost%3A20000%2F" +
						"&response_type=code" +
						"&scope=openid+profile+offline_access" +
						"{SANITIZE:&state=.*:&state=SANITIZED}",
					"After completing authorization, paste the full redirect URL here:",
				},
				warns: []string{
					"NOTE: You will see a connection error in your browser. This is expected.",
				},
				asks: []string{
					"Browser URL",
				},
				token: "pkce-token",
			},
		},
		{
			name: "pkce with no listener invalid copy/paste",
			input: authInput{
				authMode:           constants.AuthModePKCE,
				noListen:           true,
				organization:       "org-1",
				askResponses:       []string{"not-\x7fvalid"},
				pkceStateValidator: func(string, string) bool { return true },
			},
			mockEndpoints: getPKCEEndpoints(pkceConfig{
				discoveryConfig: successfulPKCEAuthDiscovery,
			}),
			expected: authExpected{
				errFormat: "failed to get pkce token\n" +
					" Caused by:\n" +
					"\t* parse \"not-\\x7fvalid\": net/url: invalid control character in URL",
				displays: []string{
					"Open this URL in your browser and complete the authorization:\n" +
						"${URL}/authorization?" +
						"audience=discovered_audience" +
						"&client_id=discovered_client_id" +
						"{SANITIZE:&code_challenge=.*?&:&code_challenge=SANITIZED&}" +
						"code_challenge_method=S256" +
						"&ingressHost=${HOST}%3A${PORT}" +
						"&organization=org-1" +
						"&redirect_uri=http%3A%2F%2Flocalhost%3A20000%2F" +
						"&response_type=code" +
						"&scope=openid+profile+offline_access" +
						"{SANITIZE:&state=.*:&state=SANITIZED}",
					"After completing authorization, paste the full redirect URL here:",
				},
				warns: []string{
					"NOTE: You will see a connection error in your browser. This is expected.",
				},
				asks: []string{
					"Browser URL",
				},
				token: "pkce-token",
			},
		},
		{
			name: "pkce with no listener available error in copy/paste url",
			input: authInput{
				authMode:     constants.AuthModePKCE,
				noListen:     true,
				organization: "org-1",
				askResponses: []string{
					"http://example.com?code=abc&error=access_denied&error_description=invalid%20parameter",
				},
				pkceStateValidator: func(string, string) bool { return true },
			},
			mockEndpoints: getPKCEEndpoints(pkceConfig{
				discoveryConfig: successfulPKCEAuthDiscovery,
			}),
			expected: authExpected{
				displays: []string{
					"Open this URL in your browser and complete the authorization:\n" +
						"${URL}/authorization?" +
						"audience=discovered_audience" +
						"&client_id=discovered_client_id" +
						"{SANITIZE:&code_challenge=.*?&:&code_challenge=SANITIZED&}" +
						"code_challenge_method=S256" +
						"&ingressHost=${HOST}%3A${PORT}" +
						"&organization=org-1" +
						"&redirect_uri=http%3A%2F%2Flocalhost%3A20000%2F" +
						"&response_type=code" +
						"&scope=openid+profile+offline_access" +
						"{SANITIZE:&state=.*:&state=SANITIZED}",
					"After completing authorization, paste the full redirect URL here:",
				},
				warns: []string{
					"NOTE: You will see a connection error in your browser. This is expected.",
				},
				asks: []string{
					"Browser URL",
				},
				errFormat: "failed to get pkce token\n" +
					" Caused by:\n" +
					"\t* 'access_denied': 'invalid parameter'",
			},
		},
		{
			name: "pkce expected callback flow - honors explicit clientid",
			input: authInput{
				authMode:                  constants.AuthModePKCE,
				clientID:                  "explicit-clientid",
				organization:              "org-1",
				pkceStateValidator:        func(string, string) bool { return true },
				pkceAuthorizationResponse: "code=abc",
			},
			mockEndpoints: getPKCEEndpoints(pkceConfig{
				discoveryConfig: successfulPKCEAuthDiscovery,
				tokenConfig: stubs.StubHTTPPartialPostConfig{
					ExpectedRequestBody: `client_id=explicit-clientid` +
						`&code=abc` +
						`{SANITIZE:&code_verifier=.*?&:&code_verifier=SANITIZED&}` +
						`grant_type=authorization_code` +
						`&redirect_uri=` + url.QueryEscape(constants.PKCECallbackURL),
					ResponseStatus: http.StatusOK,
					ResponseBody:   `{"access_token":"pkce-token"}`,
				},
			}),
			expected: authExpected{
				displays: []string{
					"Open this URL in your browser and complete the authorization:\n" +
						"${URL}/authorization?" +
						"audience=discovered_audience" +
						"&client_id=explicit-clientid" +
						"{SANITIZE:&code_challenge=.*?&:&code_challenge=SANITIZED&}" +
						"code_challenge_method=S256" +
						"&ingressHost=${HOST}%3A${PORT}" +
						"&organization=org-1" +
						"&redirect_uri=http%3A%2F%2Flocalhost%3A20000%2F" +
						"&response_type=code" +
						"&scope=openid+profile+offline_access" +
						"{SANITIZE:&state=.*:&state=SANITIZED}",
				},
				token: "pkce-token",
			},
		},
		{
			name: "pkce expected callback flow fails state validation",
			input: authInput{
				authMode:                  constants.AuthModePKCE,
				organization:              "org-1",
				pkceAuthorizationResponse: "code=abc",
				pkceListenPort:            20001,
			},
			mockEndpoints: getPKCEEndpoints(pkceConfig{
				discoveryConfig: successfulPKCEAuthDiscovery,
			}),
			expected: authExpected{
				displays: []string{
					"Open this URL in your browser and complete the authorization:\n" +
						"${URL}/authorization?" +
						"audience=discovered_audience" +
						"&client_id=discovered_client_id" +
						"{SANITIZE:&code_challenge=.*?&:&code_challenge=SANITIZED&}" +
						"code_challenge_method=S256" +
						"&ingressHost=${HOST}%3A${PORT}" +
						"&organization=org-1" +
						"&redirect_uri=http%3A%2F%2Flocalhost%3A20001%2F" +
						"&response_type=code" +
						"&scope=openid+profile+offline_access" +
						"{SANITIZE:&state=.*:&state=SANITIZED}",
				},
				errFormat: "failed to get pkce token\n" +
					" Caused by:\n" +
					"\t* state mismatch in authentication response",
			},
		},
		{
			name: "pkce expected callback flow error response",
			input: authInput{
				authMode:                  constants.AuthModePKCE,
				organization:              "org-1",
				pkceAuthorizationResponse: "code=abc&error=access_denied&error_description=invalid%20parameter",
				pkceListenPort:            20002,
			},
			mockEndpoints: getPKCEEndpoints(pkceConfig{
				discoveryConfig: successfulPKCEAuthDiscovery,
			}),
			expected: authExpected{
				displays: []string{
					"Open this URL in your browser and complete the authorization:\n" +
						"${URL}/authorization?" +
						"audience=discovered_audience" +
						"&client_id=discovered_client_id" +
						"{SANITIZE:&code_challenge=.*?&:&code_challenge=SANITIZED&}" +
						"code_challenge_method=S256" +
						"&ingressHost=${HOST}%3A${PORT}" +
						"&organization=org-1" +
						"&redirect_uri=http%3A%2F%2Flocalhost%3A20002%2F" +
						"&response_type=code" +
						"&scope=openid+profile+offline_access" +
						"{SANITIZE:&state=.*:&state=SANITIZED}",
				},
				errFormat: "failed to get pkce token\n" +
					" Caused by:\n" +
					"\t* 'access_denied': 'invalid parameter'",
			},
		},
		{
			name: "pkce expected callback flow failed exchange",
			input: authInput{
				authMode:                  constants.AuthModePKCE,
				organization:              "org-1",
				pkceAuthorizationResponse: "code=abc",
				pkceListenPort:            20003,
				pkceStateValidator:        func(string, string) bool { return true },
			},
			mockEndpoints: getPKCEEndpoints(pkceConfig{
				discoveryConfig: successfulPKCEAuthDiscovery,
				tokenConfig: stubs.StubHTTPPartialPostConfig{
					ExpectedRequestBody: `client_id=discovered_client_id` +
						`&code=abc` +
						`{SANITIZE:&code_verifier=.*?&:&code_verifier=SANITIZED&}` +
						`grant_type=authorization_code` +
						`&redirect_uri=` + url.QueryEscape(auth.PKCECallbackURL(20003)),
					ResponseStatus: http.StatusForbidden,
					ResponseBody:   `Not allowed`,
				},
			}),
			expected: authExpected{
				displays: []string{
					"Open this URL in your browser and complete the authorization:\n" +
						"${URL}/authorization?" +
						"audience=discovered_audience" +
						"&client_id=discovered_client_id" +
						"{SANITIZE:&code_challenge=.*?&:&code_challenge=SANITIZED&}" +
						"code_challenge_method=S256" +
						"&ingressHost=${HOST}%3A${PORT}" +
						"&organization=org-1" +
						"&redirect_uri=http%3A%2F%2Flocalhost%3A20003%2F" +
						"&response_type=code" +
						"&scope=openid+profile+offline_access" +
						"{SANITIZE:&state=.*:&state=SANITIZED}",
				},
				errFormat: "failed to get pkce token\n" +
					" Caused by:\n" +
					"\t* POST ${URL}/token failed\n\t* 403 Not allowed",
			},
		},
		{
			name: "pkce expected callback flow invalid exchange response body",
			input: authInput{
				authMode:                  constants.AuthModePKCE,
				organization:              "org-1",
				pkceAuthorizationResponse: "code=abc",
				pkceListenPort:            20004,
				pkceStateValidator:        func(string, string) bool { return true },
			},
			mockEndpoints: getPKCEEndpoints(pkceConfig{
				discoveryConfig: successfulPKCEAuthDiscovery,
				tokenConfig: stubs.StubHTTPPartialPostConfig{
					ExpectedRequestBody: `client_id=discovered_client_id` +
						`&code=abc` +
						`{SANITIZE:&code_verifier=.*?&:&code_verifier=SANITIZED&}` +
						`grant_type=authorization_code` +
						`&redirect_uri=` + url.QueryEscape(auth.PKCECallbackURL(20004)),
					ResponseStatus: http.StatusOK,
					ResponseBody:   `invalid`,
				},
			}),
			expected: authExpected{
				displays: []string{
					"Open this URL in your browser and complete the authorization:\n" +
						"${URL}/authorization?" +
						"audience=discovered_audience" +
						"&client_id=discovered_client_id" +
						"{SANITIZE:&code_challenge=.*?&:&code_challenge=SANITIZED&}" +
						"code_challenge_method=S256" +
						"&ingressHost=${HOST}%3A${PORT}" +
						"&organization=org-1" +
						"&redirect_uri=http%3A%2F%2Flocalhost%3A20004%2F" +
						"&response_type=code" +
						"&scope=openid+profile+offline_access" +
						"{SANITIZE:&state=.*:&state=SANITIZED}",
				},
				errFormat: "failed to get pkce token\n" +
					" Caused by:\n" +
					"\t* POST ${URL}/token failed\n" +
					"\t* invalid character 'i' looking for beginning of value",
			},
		},
		{
			name: "pkce expected callback flow unavailable listen port",
			input: authInput{
				authMode:                  constants.AuthModePKCE,
				organization:              "org-1",
				pkceAuthorizationResponse: "code=abc",
				pkceListenPort:            math.MaxInt,
				pkceStateValidator:        func(string, string) bool { return true },
			},
			mockEndpoints: getPKCEEndpoints(pkceConfig{
				discoveryConfig: successfulPKCEAuthDiscovery,
			}),
			expected: authExpected{
				errFormat: "failed to get pkce token\n" +
					" Caused by:\n" +
					"\t* listen tcp: address -9223372036854755811: invalid port",
			},
		},
		{
			name: "pkce expected callback flow invalid token endpoint",
			input: authInput{
				authMode:                  constants.AuthModePKCE,
				organization:              "org-1",
				pkceAuthorizationResponse: "code=abc",
				pkceListenPort:            20005,
				pkceStateValidator:        func(string, string) bool { return true },
			},
			mockEndpoints: getPKCEEndpoints(pkceConfig{
				discoveryConfig: stubs.StubHTTPPartialGetConfig{
					ResponseStatus: http.StatusOK,
					ResponseBody: `
						{
						  "token_endpoint": "http://${HOST}/` + "\x7f" + `/invalid",
						  "authorization_endpoint": "http://${HOST}/authorization",
						  "workbench": {
							"spa_audience": "discovered_audience",
							"spa_client_id": "discovered_client_id"
						  }
						}
						`,
				},
			}),
			expected: authExpected{
				displays: []string{
					"Open this URL in your browser and complete the authorization:\n" +
						"${URL}/authorization?" +
						"audience=discovered_audience" +
						"&client_id=discovered_client_id" +
						"{SANITIZE:&code_challenge=.*?&:&code_challenge=SANITIZED&}" +
						"code_challenge_method=S256" +
						"&ingressHost=${HOST}%3A${PORT}" +
						"&organization=org-1" +
						"&redirect_uri=http%3A%2F%2Flocalhost%3A20005%2F" +
						"&response_type=code" +
						"&scope=openid+profile+offline_access" +
						"{SANITIZE:&state=.*:&state=SANITIZED}",
				},
				errFormat: "failed to get pkce token\n" +
					" Caused by:\n" +
					"\t* parse \"${URL}/\\x7f/invalid\": net/url: invalid control character in URL",
			},
		},
		{
			name: "pkce expected callback flow timeout",
			input: authInput{
				authMode:                  constants.AuthModePKCE,
				organization:              "org-1",
				pkceAuthorizationResponse: "code=abc",
				pkceListenPort:            20006,
				pkceStateValidator:        func(string, string) bool { return true },
				pkceCallbackTimeout:       pointer.Ptr(0 * time.Second),
			},
			mockEndpoints: getPKCEEndpoints(pkceConfig{
				discoveryConfig: successfulPKCEAuthDiscovery,
			}),
			expected: authExpected{
				displays: []string{
					"Open this URL in your browser and complete the authorization:\n" +
						"${URL}/authorization?" +
						"audience=discovered_audience" +
						"&client_id=discovered_client_id" +
						"{SANITIZE:&code_challenge=.*?&:&code_challenge=SANITIZED&}" +
						"code_challenge_method=S256" +
						"&ingressHost=${HOST}%3A${PORT}" +
						"&organization=org-1" +
						"&redirect_uri=http%3A%2F%2Flocalhost%3A20006%2F" +
						"&response_type=code" +
						"&scope=openid+profile+offline_access" +
						"{SANITIZE:&state=.*:&state=SANITIZED}",
				},
				errFormat: "failed to get pkce token\n" +
					" Caused by:\n" +
					"\t* pkce authorization timeout",
			},
		},
		{
			name: "pkce discovered audience and client id",
			input: authInput{
				authMode:                  constants.AuthModePKCE,
				organization:              "org-1",
				pkceStateValidator:        func(string, string) bool { return true },
				pkceAuthorizationResponse: "code=abc",
			},
			mockEndpoints: getPKCEEndpoints(pkceConfig{
				discoveryConfig: successfulPKCEAuthDiscovery,
				tokenConfig: stubs.StubHTTPPartialPostConfig{
					ExpectedRequestBody: `client_id=discovered_client_id` +
						`&code=abc` +
						`{SANITIZE:&code_verifier=.*?&:&code_verifier=SANITIZED&}` +
						`grant_type=authorization_code` +
						`&redirect_uri=` + url.QueryEscape(constants.PKCECallbackURL),
					ResponseStatus: http.StatusOK,
					ResponseBody:   `{"access_token":"pkce-token"}`,
				},
			}),
			expected: authExpected{
				displays: []string{
					"Open this URL in your browser and complete the authorization:\n" +
						"${URL}/authorization?" +
						"audience=discovered_audience" +
						"&client_id=discovered_client_id" +
						"{SANITIZE:&code_challenge=.*?&:&code_challenge=SANITIZED&}" +
						"code_challenge_method=S256" +
						"&ingressHost=${HOST}%3A${PORT}" +
						"&organization=org-1" +
						"&redirect_uri=http%3A%2F%2Flocalhost%3A20000%2F" +
						"&response_type=code" +
						"&scope=openid+profile+offline_access" +
						"{SANITIZE:&state=.*:&state=SANITIZED}",
				},
				token: "pkce-token",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, makeAuthTestFunc(tt))
	}
}
