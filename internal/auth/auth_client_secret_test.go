// Copyright © 2026, SAS Institute Inc., Cary, NC, USA.  All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package auth_test

import (
	"net/http"
	"net/url"
	"testing"

	"github.com/sas-institute-rnd-saas/wbctl/internal/constants"
	"github.com/sas-institute-rnd-saas/wbctl/internal/stubs"
)

func TestGetAccessTokenClientCredentials(t *testing.T) {
	tests := []authTestCase{
		{
			name: "invalid control plane endpoint returns discovery error - client credentials",
			input: authInput{
				authMode:             constants.AuthModeClientCredentials,
				clientID:             "cid",
				clientSecret:         "secret",
				controlPlaneEndpoint: authTestInvalidControlPlaneEndpoint,
			},
			expected: authExpected{
				errFormat: authTestClientCredentialsModeInvalidControlPlaneURLErr,
			},
		},
		{
			name: "client credentials success",
			input: authInput{
				authMode:     constants.AuthModeClientCredentials,
				clientID:     "cid",
				clientSecret: "secret",
			},
			mockEndpoints: getPasswordEndpoints(passwordConfig{
				discoveryConfig: successfulPasswordAuthDiscovery,
				tokenConfig: stubs.StubHTTPPartialPostConfig{
					ExpectedRequestBody: `audience=` + url.QueryEscape(constants.OAuthAudience) +
						`&client_id=cid&client_secret=secret&grant_type=client_credentials`,
					ResponseStatus: http.StatusOK,
					ResponseBody:   `{"access_token":"cc-token"}`,
				},
			}),
			expected: authExpected{
				token: "cc-token",
			},
		},
		{
			name: "client credentials missing client id",
			input: authInput{
				authMode:     constants.AuthModeClientCredentials,
				clientSecret: "secret",
			},
			mockEndpoints: getPasswordEndpoints(passwordConfig{
				discoveryConfig: successfulPasswordAuthDiscovery,
				tokenConfig: stubs.StubHTTPPartialPostConfig{
					ExpectedRequestBody: `audience=` + url.QueryEscape(constants.OAuthAudience) +
						`&client_id=cid&client_secret=secret&grant_type=client_credentials`,
					ResponseStatus: http.StatusOK,
					ResponseBody:   `{"access_token":"cc-token"}`,
				},
			}),
			expected: authExpected{
				errFormat: "--client-id must be specified",
			},
		},
		// Error from token endpoint (non-200)
		{
			name: "client credentials token endpoint error",
			input: authInput{
				authMode:     constants.AuthModeClientCredentials,
				clientID:     "cid",
				clientSecret: "secret",
			},
			mockEndpoints: getPasswordEndpoints(passwordConfig{
				discoveryConfig: successfulPasswordAuthDiscovery,
				tokenConfig: stubs.StubHTTPPartialPostConfig{
					ExpectedRequestBody: `audience=` + url.QueryEscape(constants.OAuthAudience) +
						`&client_id=cid&client_secret=secret&grant_type=client_credentials`,
					ResponseStatus: http.StatusUnauthorized,
					ResponseBody:   `{"error":"bad"}`,
				},
			}),
			expected: authExpected{
				errFormat: "failed to get client credentials token\n" +
					" Caused by:\n" +
					"\t* POST ${URL}/token failed\n\t* 401 {\"error\":\"bad\"}",
			},
		},
		// Token response missing access_token
		{
			name: "client credentials missing access_token in response",
			input: authInput{
				authMode:     constants.AuthModeClientCredentials,
				clientID:     "cid",
				clientSecret: "secret",
			},
			mockEndpoints: getPasswordEndpoints(passwordConfig{
				discoveryConfig: successfulPasswordAuthDiscovery,
				tokenConfig: stubs.StubHTTPPartialPostConfig{
					ExpectedRequestBody: `audience=` + url.QueryEscape(constants.OAuthAudience) +
						`&client_id=cid&client_secret=secret&grant_type=client_credentials`,
					ResponseStatus: http.StatusOK,
					ResponseBody:   `{}`,
				},
			}),
			expected: authExpected{
				errFormat: "failed to get client credentials token\n Caused by:\n\t* token response missing access_token",
			},
		},
		// Discovery returns token_endpoint containing a control character -> NewRequestWithContext fails
		{
			name: "client credentials discovery returns invalid token endpoint",
			input: authInput{
				authMode:     constants.AuthModeClientCredentials,
				clientID:     "cid",
				clientSecret: "secret",
			},
			mockEndpoints: getPasswordEndpoints(passwordConfig{
				discoveryConfig: stubs.StubHTTPPartialGetConfig{
					ResponseStatus: http.StatusOK,
					ResponseBody:   "{\"token_endpoint\":\"a\x7fb\"}",
				},
				tokenConfig: stubs.StubHTTPPartialPostConfig{
					ExpectedRequestBody: `audience=` + url.QueryEscape(constants.OAuthAudience) +
						`&client_id=cid&client_secret=secret&grant_type=client_credentials`,
					ResponseStatus: http.StatusOK,
					ResponseBody:   `{"access_token":"cc-token"}`,
				},
			}),
			expected: authExpected{
				errFormat: authTestClientCredentialsModeInvalidTokenEndpointURLErr,
			},
		},
		{
			name: "client credentials discovered audience",
			input: authInput{
				authMode:     constants.AuthModeClientCredentials,
				clientID:     "cid",
				clientSecret: "secret",
			},
			mockEndpoints: getPasswordEndpoints(passwordConfig{
				discoveryConfig: successfulPasswordAuthDiscoveryWithClientID,
				tokenConfig: stubs.StubHTTPPartialPostConfig{
					ExpectedRequestBody: `audience=discovered_audience` +
						`&client_id=cid&client_secret=secret&grant_type=client_credentials`,
					ResponseStatus: http.StatusOK,
					ResponseBody:   `{"access_token":"cc-token"}`,
				},
			}),
			expected: authExpected{
				token: "cc-token",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, makeAuthTestFunc(tt))
	}
}
