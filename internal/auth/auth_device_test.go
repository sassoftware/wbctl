// Copyright © 2026, SAS Institute Inc., Cary, NC, USA.  All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package auth_test

import (
	"fmt"
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/sas-institute-rnd-saas/wbctl/internal/constants"
	"github.com/sas-institute-rnd-saas/wbctl/internal/osutil"
	"github.com/sas-institute-rnd-saas/wbctl/internal/stubs"
)

func TestGetAccessTokenDevice(t *testing.T) {
	tests := []authTestCase{
		{
			name: "invalid control plane endpoint returns discovery error - device",
			input: authInput{
				authMode:             constants.AuthModeDevice,
				clientID:             constants.OAuthClientID,
				controlPlaneEndpoint: authTestInvalidControlPlaneEndpoint,
			},
			expected: authExpected{
				errFormat: authTestDeviceModeInvalidControlPlaneURLErr,
			},
		},
		{
			name: "valid cached token short-circuits and returns cached token - device",
			input: authInput{
				authMode: constants.AuthModeDevice,
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
			name: "first time through with no cached token",
			input: authInput{
				authMode: constants.AuthModeDevice,
				clientID: constants.OAuthClientID,
			},
			mockEndpoints: getDeviceEndpoints(deviceConfig{
				discoveryConfig: successfulDeviceAuthDiscovery,
				deviceConfig:    successfulDevice,
				tokenConfigs: []stubs.StubHTTPPartialPostConfig{{
					ExpectedRequestBody: expectedTokenRequestBody,
					ResponseStatus:      http.StatusOK,
					ResponseBody:        `{"access_token":"device"}`,
				}},
			}),
			expected: authExpected{
				token:    "device",
				displays: []string{"Open https://verify and enter code: uc"},
				cachedToken: &expectedTokenCache{
					accessToken: "device",
				},
			},
		},
		{
			name: "refresh failure falls back to device authorization and returns device token",
			input: authInput{
				authMode: constants.AuthModeDevice,
				cache: fmt.Sprintf(
					`{"access_token":"old","refresh_token":"rt","expires_at":"%s"}`,
					time.Now().Add(-time.Hour).UTC().Format(time.RFC3339Nano),
				),
				clientID: constants.OAuthClientID,
			},
			mockEndpoints: getDeviceEndpoints(deviceConfig{
				discoveryConfig: successfulDeviceAuthDiscovery,
				deviceConfig:    successfulDevice,
				tokenConfigs: []stubs.StubHTTPPartialPostConfig{
					{
						ExpectedRequestBody: `audience=` + url.QueryEscape(constants.OAuthAudience) +
							`&client_id=` + url.QueryEscape(constants.OAuthClientID) +
							`&grant_type=refresh_token&refresh_token=rt`,
						ResponseStatus: http.StatusUnauthorized,
						ResponseBody:   `{"error":"bad"}`,
					}, {

						ExpectedRequestBody: expectedTokenRequestBody,
						ResponseStatus:      http.StatusOK,
						ResponseBody:        `{"access_token":"device-ok","refresh_token":"drt","expires_in":3600}`,
					},
				},
			}),
			expected: authExpected{
				token:    "device-ok",
				displays: []string{"Open https://verify and enter code: uc"},
				warns:    []string{"falling back to full reauthorization"},
				cachedToken: &expectedTokenCache{
					accessToken:  "device-ok",
					refreshToken: "drt",
				},
			},
		},
		{
			name: "refresh empty token falls back to device - if the refresh response doesn't contain an access token, " +
				"treat it as a refresh failure and fall back to device flow (auto mode with no username)",
			input: authInput{
				authMode: constants.AuthModeDevice,
				cache: fmt.Sprintf(
					`{"access_token":"old","refresh_token":"rt","expires_at":"%s"}`,
					time.Now().Add(-time.Hour).UTC().Format(time.RFC3339Nano),
				),
				clientID: constants.OAuthClientID,
			},
			mockEndpoints: getDeviceEndpoints(deviceConfig{
				discoveryConfig: successfulDeviceAuthDiscovery,
				deviceConfig:    successfulDevice,
				tokenConfigs: []stubs.StubHTTPPartialPostConfig{
					{
						ExpectedRequestBody: `audience=` + url.QueryEscape(constants.OAuthAudience) +
							`&client_id=` + url.QueryEscape(constants.OAuthClientID) +
							`&grant_type=refresh_token&refresh_token=rt`,
						ResponseStatus: http.StatusOK,
						ResponseBody:   `{"access_token":"","refresh_token":"rt2","expires_in":3600}`,
					}, {

						ExpectedRequestBody: expectedTokenRequestBody,
						ResponseStatus:      http.StatusOK,
						ResponseBody:        `{"access_token":"device2","refresh_token":"d2","expires_in":3600}`,
					},
				},
			}),
			expected: authExpected{
				token:    "device2",
				displays: []string{"Open https://verify and enter code: uc"},
				cachedToken: &expectedTokenCache{
					accessToken:  "device2",
					refreshToken: "d2",
				},
				warns: []string{"falling back to full reauthorization"},
			},
		},
		{
			name: "device authorization endpoint failure - an expired cached token causes " +
				"device flow",
			input: authInput{
				authMode: constants.AuthModeDevice,
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
						`&grant_type=refresh_token&refresh_token=rt`,
					ResponseBody:   `{"error":"bad"}`,
					ResponseStatus: http.StatusUnauthorized,
				}},
				deviceConfig: stubs.StubHTTPPartialPostConfig{
					ExpectedRequestBody: `audience=` + url.QueryEscape(constants.OAuthAudience) +
						`&client_id=` + url.QueryEscape(constants.OAuthClientID) +
						`&scope=` + url.QueryEscape(constants.OAuthScope),
					ResponseBody:   authTestDeviceInvalidClientBody,
					ResponseStatus: http.StatusBadRequest,
				},
			}),
			expected: authExpected{
				errFormat: authTestDeviceAuthorizationFailureErrFormat,
				warns:     []string{"falling back to full reauthorization"},
			},
		},
		// This is a fake token so not a security issue.
		//nolint:gosec
		{
			name: "device mode first time through with no cached token - explicit device flow",
			input: authInput{
				authMode: constants.AuthModeDevice,
				clientID: constants.OAuthClientID,
			},
			mockEndpoints: getDeviceEndpoints(deviceConfig{
				discoveryConfig: successfulDeviceAuthDiscovery,
				deviceConfig:    successfulDevice,
				tokenConfigs: []stubs.StubHTTPPartialPostConfig{{
					ExpectedRequestBody: expectedTokenRequestBody,
					ResponseStatus:      http.StatusOK,
					ResponseBody:        `{"access_token":"device-mode-token"}`,
				}},
			}),
			expected: authExpected{
				token:    "device-mode-token",
				displays: []string{"Open https://verify and enter code: uc"},
			},
		},
		{
			name: "device flow displays verification_uri_complete when present",
			input: authInput{
				authMode: constants.AuthModeDevice,
				clientID: constants.OAuthClientID,
			},
			mockEndpoints: getDeviceEndpoints(deviceConfig{
				discoveryConfig: successfulDeviceAuthDiscovery,
				tokenConfigs: []stubs.StubHTTPPartialPostConfig{{
					ExpectedRequestBody: expectedTokenRequestBody,
					ResponseStatus:      http.StatusOK,
					ResponseBody:        `{"access_token":"device-verify-complete"}`,
				}},
				deviceConfig: stubs.StubHTTPPartialPostConfig{
					ExpectedRequestBody: `audience=` + url.QueryEscape(constants.OAuthAudience) +
						`&client_id=` + url.QueryEscape(constants.OAuthClientID) +
						`&scope=` + url.QueryEscape(constants.OAuthScope),
					ResponseStatus: http.StatusOK,
					ResponseBody:   authTestDeviceAuthorizationBodyWithVerificationURIComplete,
				},
			}),
			expected: authExpected{
				token: "device-verify-complete",
				displays: []string{
					"Open this URL in your browser and complete the authorization:\nhttps://verify.complete",
				},
			},
		},
		{
			name: "device polling terminal oauth error fails fast",
			input: authInput{
				authMode: constants.AuthModeDevice,
				clientID: constants.OAuthClientID,
			},
			mockEndpoints: getDeviceEndpoints(deviceConfig{
				discoveryConfig: successfulDeviceAuthDiscovery,
				deviceConfig:    successfulDevice,
				tokenConfigs: []stubs.StubHTTPPartialPostConfig{{
					ExpectedRequestBody: expectedTokenRequestBody,
					ResponseStatus:      http.StatusBadRequest,
					ResponseBody:        `{"error":"access_denied"}`,
				}},
			}),
			expected: authExpected{
				errFormat: authTestPollingAccessDeniedErrFormat,
				pollCount: 1,
				displays:  []string{"Open https://verify and enter code: uc"},
			},
		},
		{
			// This test just drives the execution pointer through the backoff
			// code. It doesn't validate that backoff was handled correctly.
			name: "device polling slow_down increases interval then times out",
			input: authInput{
				authMode: constants.AuthModeDevice,
				clientID: constants.OAuthClientID,
			},
			mockEndpoints: getDeviceEndpoints(deviceConfig{
				discoveryConfig: successfulDeviceAuthDiscovery,
				deviceConfig:    successfulDevice,
				tokenConfigs: []stubs.StubHTTPPartialPostConfig{{
					ExpectedRequestBody: expectedTokenRequestBody,
					ResponseStatus:      http.StatusBadRequest,
					ResponseBody:        `{"error":"slow_down"}`,
				}},
			}),
			expected: authExpected{
				errFormat:    authTestDeviceAuthorizationTimeoutErr,
				minPollCount: 1,
				displays:     []string{"Open https://verify and enter code: uc"},
			},
		},
		{
			name: "device polling non-terminal oauth error continues until timeout",
			input: authInput{
				authMode: constants.AuthModeDevice,
				clientID: constants.OAuthClientID,
			},
			mockEndpoints: getDeviceEndpoints(deviceConfig{
				discoveryConfig: successfulDeviceAuthDiscovery,
				deviceConfig:    successfulDevice,
				tokenConfigs: []stubs.StubHTTPPartialPostConfig{{
					ExpectedRequestBody: expectedTokenRequestBody,
					ResponseStatus:      http.StatusBadRequest,
					ResponseBody:        `{"error":"authorization_pending"}`,
				}},
			}),
			expected: authExpected{
				errFormat:    authTestDeviceAuthorizationTimeoutErr,
				minPollCount: 1,
				displays:     []string{"Open https://verify and enter code: uc"},
			},
		},
		{
			name: "device polling cancelled context returns cancellation error",
			input: authInput{
				authMode:      constants.AuthModeDevice,
				cancelContext: true,
				clientID:      constants.OAuthClientID,
			},
			mockEndpoints: getDeviceEndpoints(deviceConfig{
				discoveryConfig: successfulDeviceAuthDiscovery,
				deviceConfig: stubs.StubHTTPPartialPostConfig{
					ExpectedRequestBody: `audience=` + url.QueryEscape(constants.OAuthAudience) +
						`&client_id=` + url.QueryEscape(constants.OAuthClientID) +
						`&scope=` + url.QueryEscape(constants.OAuthScope),
					ResponseBody: `{"device_code":"dc","user_code":"uc",` +
						`"verification_uri":"https://verify","expires_in":5,"interval":1}`,
					ResponseStatus: http.StatusOK,
				},
				tokenConfigs: []stubs.StubHTTPPartialPostConfig{{
					ExpectedRequestBody: expectedTokenRequestBody,
					ResponseStatus:      http.StatusBadRequest,
					ResponseBody:        `{"error":"authorization_pending"}`,
				}},
			}),
			expected: authExpected{
				errFormat: authTestPollingCancelledErr,
				pollCount: 1,
				displays:  []string{"Open https://verify and enter code: uc"},
			},
		},
		{
			name: "device polling with invalid token endpoint still hits poll request",
			input: authInput{
				authMode:      constants.AuthModeDevice,
				cancelContext: true,
				clientID:      constants.OAuthClientID,
			},
			mockEndpoints: getDeviceEndpoints(deviceConfig{
				discoveryConfig: stubs.StubHTTPPartialGetConfig{
					ResponseStatus: http.StatusOK,
					ResponseBody:   `{"token_endpoint":"a\nb", "device_authorization_endpoint":"http://${HOST}/device"}`,
				},
				deviceConfig: successfulDevice,
			}),
			expected: authExpected{
				errFormat: authTestPollingCancelledErr,
				displays:  []string{"Open https://verify and enter code: uc"},
			},
		},
		{
			name: "device polling uses default interval when interval is zero",
			input: authInput{
				authMode: constants.AuthModeDevice,
				clientID: constants.OAuthClientID,
			},
			mockEndpoints: getDeviceEndpoints(deviceConfig{
				discoveryConfig: successfulDeviceAuthDiscovery,
				deviceConfig: stubs.StubHTTPPartialPostConfig{
					ExpectedRequestBody: `audience=` + url.QueryEscape(constants.OAuthAudience) +
						`&client_id=` + url.QueryEscape(constants.OAuthClientID) +
						`&scope=` + url.QueryEscape(constants.OAuthScope),
					ResponseStatus: http.StatusOK,
					ResponseBody:   authTestDeviceAuthorizationBodyWithZeroInterval,
				},
				tokenConfigs: []stubs.StubHTTPPartialPostConfig{{
					ExpectedRequestBody: expectedTokenRequestBody,
					ResponseStatus:      http.StatusOK,
					ResponseBody:        `{"access_token":"device-default-interval"}`,
				}},
			}),
			expected: authExpected{
				token: "device-default-interval",
				displays: []string{
					"Open https://verify and enter code: uc",
				},
			},
		},
		{
			name: "modern discovery 404 falls back to legacy discovery path",
			input: authInput{
				authMode: constants.AuthModeDevice,
				clientID: constants.OAuthClientID,
			},
			mockEndpoints: getLegacyDeviceEndpoints(deviceConfig{
				discoveryConfig: successfulDeviceAuthDiscovery,
				deviceConfig:    successfulDevice,
				tokenConfigs: []stubs.StubHTTPPartialPostConfig{{
					ExpectedRequestBody: expectedTokenRequestBody,
					ResponseStatus:      http.StatusOK,
					ResponseBody:        `{"access_token":"device-legacy","refresh_token":"drt","expires_in":3600}`,
				}},
			}),
			expected: authExpected{
				token:               "device-legacy",
				displays:            []string{"Open https://verify and enter code: uc"},
				modernDiscoveryHits: 1,
				legacyDiscoveryHits: 1,
			},
		},
		{
			name: "load cache fail for invalid home still returns token",
			input: authInput{
				authMode: constants.AuthModeDevice,
				clientID: constants.OAuthClientID,
				home:     homeAsFile,
			},
			mockEndpoints: getDeviceEndpoints(deviceConfig{
				discoveryConfig: successfulDeviceAuthDiscovery,
				deviceConfig:    successfulDevice,
				tokenConfigs: []stubs.StubHTTPPartialPostConfig{{
					ExpectedRequestBody: expectedTokenRequestBody,
					ResponseStatus:      http.StatusOK,
					ResponseBody:        `{"access_token":"device3","refresh_token":"drt","expires_in":3600}`,
				}},
			}),
			expected: authExpected{
				token:    "device3",
				displays: []string{"Open https://verify and enter code: uc"},
				warns: []string{
					"failed to load token cache" +
						"\n Caused by:" +
						"\n\t* open ${HOME}" + osutil.PathSeparator() + ".wbctl" + osutil.PathSeparator() + "token.json: not a directory",
					"failed to save device token: failed to save token cache" +
						"\n Caused by:" +
						"\n\t* failed to load token cache" +
						"\n\t* open ${HOME}" + osutil.PathSeparator() + ".wbctl" + osutil.PathSeparator() + "token.json: not a directory",
				},
			},
		},
		{
			name: "load cache fail for unset home still returns token",
			input: authInput{
				authMode: constants.AuthModeDevice,
				clientID: constants.OAuthClientID,
				home:     homeAsUnset,
			},
			mockEndpoints: getDeviceEndpoints(deviceConfig{
				discoveryConfig: successfulDeviceAuthDiscovery,
				deviceConfig:    successfulDevice,
				tokenConfigs: []stubs.StubHTTPPartialPostConfig{{
					ExpectedRequestBody: expectedTokenRequestBody,
					ResponseStatus:      http.StatusOK,
					ResponseBody:        `{"access_token":"device3","refresh_token":"drt","expires_in":3600}`,
				}},
			}),
			expected: authExpected{
				token:    "device3",
				displays: []string{"Open https://verify and enter code: uc"},
				warns: []string{
					osutil.HomeEnvVarName() + " is not defined",
					"failed to save device token: failed to save token cache\n Caused by:\n\t* " +
						osutil.HomeEnvVarName() + " is not defined",
				},
			},
		},
		{
			name: "device mode with invalid device url is error",
			input: authInput{
				authMode: constants.AuthModeDevice,
				clientID: constants.OAuthClientID,
			},
			mockEndpoints: getDeviceEndpoints(deviceConfig{
				discoveryConfig: stubs.StubHTTPPartialGetConfig{
					ResponseStatus: http.StatusOK,
					ResponseBody:   `{"device_authorization_endpoint":"a\nb"}`,
				},
			}),
			expected: authExpected{
				errFormat: authTestDeviceModeInvalidDeviceURLErr,
			},
		},
		{
			name: "device mode with empty access token is error",
			input: authInput{
				authMode: constants.AuthModeDevice,
				clientID: constants.OAuthClientID,
			},
			mockEndpoints: getDeviceEndpoints(deviceConfig{
				discoveryConfig: successfulDeviceAuthDiscovery,
				deviceConfig:    successfulDevice,
				tokenConfigs: []stubs.StubHTTPPartialPostConfig{{
					ExpectedRequestBody: expectedTokenRequestBody,
					ResponseStatus:      http.StatusOK,
					ResponseBody:        `{"access_token":""}`,
				}},
			}),
			expected: authExpected{
				errFormat: authTestDeviceModeEmptyAccessTokenErr,
				displays:  []string{"Open https://verify and enter code: uc"},
			},
		},
		{
			name: "device mode invalid token endpoint is error",
			input: authInput{
				authMode: constants.AuthModeDevice,
				cache: fmt.Sprintf(
					`{"access_token":"old","refresh_token":"rt","expires_at":"%s"}`,
					time.Now().Add(-time.Hour).UTC().Format(time.RFC3339Nano),
				),
				clientID: constants.OAuthClientID,
			},
			mockEndpoints: getDeviceEndpoints(deviceConfig{
				discoveryConfig: stubs.StubHTTPPartialGetConfig{
					ResponseStatus: http.StatusOK,
					ResponseBody:   `{"token_endpoint":"a\nb"}`,
				},
			}),
			expected: authExpected{
				warns: []string{
					"falling back to full reauthorization",
				},
				errFormat: authTestDeviceModeInvalidRefreshURLErr,
			},
		},
		{
			name: "device discovered audience and client id",
			input: authInput{
				authMode: constants.AuthModeDevice,
			},
			mockEndpoints: getDeviceEndpoints(deviceConfig{
				discoveryConfig: successfulDeviceAuthDiscoveryWithClientID,
				deviceConfig: stubs.StubHTTPPartialPostConfig{
					ExpectedRequestBody: `audience=discovered_audience` +
						`&client_id=discovered_client_id` +
						`&scope=` + url.QueryEscape(constants.OAuthScope),
					ResponseStatus: http.StatusOK,
					ResponseBody:   authTestDeviceAuthorizationBody,
				},
				tokenConfigs: []stubs.StubHTTPPartialPostConfig{{
					ExpectedRequestBody: `audience=discovered_audience` +
						`&client_id=discovered_client_id` +
						`&device_code=dc&grant_type=urn%3Aietf%3Aparams%3Aoauth%3Agrant-type%3Adevice_code`,
					ResponseStatus: http.StatusOK,
					ResponseBody:   `{"access_token":"device"}`,
				}},
			}),
			expected: authExpected{
				token:    "device",
				displays: []string{"Open https://verify and enter code: uc"},
				cachedToken: &expectedTokenCache{
					accessToken: "device",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, makeAuthTestFunc(tt))
	}
}
