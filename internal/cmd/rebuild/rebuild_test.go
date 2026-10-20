// Copyright © 2026, SAS Institute Inc., Cary, NC, USA.  All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package rebuild_test

import (
	"context"
	"net/http"
	"strings"
	"testing"

	"github.com/sassoftware/wbctl/internal/cmd"
	"github.com/sassoftware/wbctl/internal/cmd/rebuild"
	"github.com/sassoftware/wbctl/internal/constants"
	"github.com/sassoftware/wbctl/internal/postgrest"
	"github.com/sassoftware/wbctl/internal/stubs"
	"github.com/spf13/afero"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
)

const (
	rebuildImagesResponseBody = `[
	{"id":"image-1"},
	{"id":"image-2"}
]`
	expectedInvalidEndpointErr = "failed to get device token" +
		"\n Caused by:" +
		"\n\t* parse \"a\\nb\": net/url: invalid control character in URL"
	expectedInvalidResourceErr = "" +
		"GET ${URL}/postgrest/images?controller_id=in.%28controller-1%29 failed" +
		"\n Caused by:" +
		"\n\t* invalid character 'i' looking for beginning of value"
)

type rebuildInput struct {
	endpoint     string
	resourceIDs  []string
	resourceType string
}

type rebuildExpected struct {
	err       string
	displays  []string
	warns     []string
	getCalls  int
	postCalls int
}

type rebuildTestCase struct {
	name          string
	input         rebuildInput
	mockEndpoints []stubs.StubHTTPEndpoint
	expected      rebuildExpected
}

func TestRebuildCommand(t *testing.T) {
	tests := []rebuildTestCase{
		{
			name:  "rebuild images by singular type",
			input: rebuildInput{resourceType: constants.ImageType, resourceIDs: []string{"controller-1"}},
			mockEndpoints: getRebuildEndpoints(rebuildConfig{
				getConfig: stubs.StubHTTPPartialGetConfig{
					ExpectedQuery:  "controller_id=in.(controller-1)",
					ResponseStatus: http.StatusOK,
					ResponseBody:   rebuildImagesResponseBody,
				},
				postConfig: []stubs.StubHTTPPartialPostConfig{
					{ExpectedRequestBody: `{"id":"image-1"}`, ResponseStatus: http.StatusOK, ResponseBody: "true"},
					{ExpectedRequestBody: `{"id":"image-2"}`, ResponseStatus: http.StatusOK, ResponseBody: "true"},
				},
			}),
			expected: rebuildExpected{
				displays:  []string{stubs.ExpectedDeviceAuthDisplay, "image-1", "image-2"},
				getCalls:  1,
				postCalls: 2,
			},
		},
		{
			name:  "rebuild images by plural type",
			input: rebuildInput{resourceType: constants.ImagesType, resourceIDs: []string{"controller-2"}},
			mockEndpoints: getRebuildEndpoints(rebuildConfig{
				getConfig: stubs.StubHTTPPartialGetConfig{
					ExpectedQuery:  "controller_id=in.(controller-2)",
					ResponseStatus: http.StatusOK,
					ResponseBody:   rebuildImagesResponseBody,
				},
				postConfig: []stubs.StubHTTPPartialPostConfig{
					{ExpectedRequestBody: `{"id":"image-1"}`, ResponseStatus: http.StatusOK, ResponseBody: "true"},
					{ExpectedRequestBody: `{"id":"image-2"}`, ResponseStatus: http.StatusOK, ResponseBody: "true"},
				},
			}),
			expected: rebuildExpected{
				displays:  []string{stubs.ExpectedDeviceAuthDisplay, "image-1", "image-2"},
				getCalls:  1,
				postCalls: 2,
			},
		},
		{
			name:  "returns error when rebuild call fails",
			input: rebuildInput{resourceType: constants.ImagesType, resourceIDs: []string{"controller-2"}},
			mockEndpoints: getRebuildEndpoints(rebuildConfig{
				getConfig: stubs.StubHTTPPartialGetConfig{
					ExpectedQuery:  "controller_id=in.(controller-2)",
					ResponseStatus: http.StatusOK,
					ResponseBody:   rebuildImagesResponseBody,
				},
				postConfig: []stubs.StubHTTPPartialPostConfig{
					{ExpectedRequestBody: `{"id":"image-1"}`, ResponseStatus: http.StatusBadRequest},
				},
			}),
			expected: rebuildExpected{
				displays:  []string{stubs.ExpectedDeviceAuthDisplay},
				err:       "POST ${URL}/postgrest/rpc/image_rebuild failed\n Caused by:\n\t* 400 ",
				getCalls:  1,
				postCalls: 1,
			},
		},
		{
			name: "invalid endpoint",
			input: rebuildInput{
				endpoint:     "a\nb",
				resourceType: constants.ImageType,
				resourceIDs:  []string{"controller-1"},
			},
			expected: rebuildExpected{
				err: expectedInvalidEndpointErr,
			},
		},
		{
			name: "invalid image response",
			input: rebuildInput{
				resourceType: constants.ImageType,
				resourceIDs:  []string{"controller-1"},
			},
			mockEndpoints: getRebuildEndpoints(rebuildConfig{
				getConfig: stubs.StubHTTPPartialGetConfig{
					ExpectedQuery:  "controller_id=in.(controller-1)",
					ResponseStatus: http.StatusOK,
					ResponseBody:   `[invalid`,
				},
			}),
			expected: rebuildExpected{
				displays: []string{stubs.ExpectedDeviceAuthDisplay},
				getCalls: 1,
				err:      expectedInvalidResourceErr,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer viper.Reset()
			srv, hitCounts := stubs.NewStubHTTPServer(t, tt.mockEndpoints)
			defer srv.Close()

			stubUI := &stubs.StubUI{}
			if tt.input.endpoint == "" {
				tt.input.endpoint = srv.URL
			}
			mockCmd := rebuild.Command(context.Background(), afero.NewMemMapFs(), stubUI)
			cmd.AddRootFlags(mockCmd)
			mockCmd.SetArgs(getRebuildArgs(tt))
			err := mockCmd.Execute()
			replacements := map[string]string{
				"${PORT}": stubs.Port(t, srv.URL),
				"${HOST}": stubs.Host(t, srv.URL),
				"${URL}":  srv.URL,
			}
			if tt.expected.err != "" {
				expectedErr, actualErr := stubs.Sanitize(tt.expected.err, err.Error(), replacements)
				require.Equal(t, expectedErr, actualErr)
			} else {
				require.NoError(t, err)
			}

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
			require.Equal(t, tt.expected.warns, stubUI.Warns, strings.Join(stubUI.Warns, "\n"))
			require.Empty(t, stubUI.Tables)

			require.Equal(
				t,
				tt.expected.getCalls,
				hitCounts.Count("GET"+postgrest.Endpoint(constants.WorkbenchEndpointImages)),
				"get calls",
			)
			require.Equal(
				t,
				tt.expected.postCalls,
				hitCounts.Count("POST"+postgrest.Endpoint(constants.WorkbenchEndpointRPCImageRebuild)),
				"update calls",
			)
		})
	}
}

func TestSupportedResourceTypes(t *testing.T) {
	require.Equal(t, []string{constants.ImagesType}, rebuild.SupportedResourceTypes())
}

func getRebuildArgs(
	tt rebuildTestCase,
) []string {
	args := []string{}
	args = append(args, tt.input.resourceType)
	args = append(args, "--"+constants.AuthModeFlag, constants.AuthModeDevice)
	if tt.input.endpoint != "" {
		args = append(args, "--"+constants.EndpointFlag, tt.input.endpoint)
	}
	args = append(args, tt.input.resourceIDs...)
	return args
}

type rebuildConfig struct {
	getConfig  stubs.StubHTTPPartialGetConfig
	postConfig []stubs.StubHTTPPartialPostConfig
}

func getRebuildEndpoints(
	config rebuildConfig,
) []stubs.StubHTTPEndpoint {
	endpoints := stubs.GetStubAuthEndpoints()
	endpoints = append(endpoints,
		stubs.StubHTTPEndpoint{
			Request: stubs.StubHTTPRequest{
				Method:          http.MethodGet,
				URL:             postgrest.Endpoint(constants.WorkbenchEndpointImages),
				ExpectedHeaders: stubs.StubAuthHeaders,
				ExpectedQuery:   config.getConfig.ExpectedQuery,
			},
			Response: stubs.StubHTTPResponse{
				Status: config.getConfig.ResponseStatus,
				Body:   config.getConfig.ResponseBody,
			},
		},
	)
	for _, cfg := range config.postConfig {
		endpoints = append(endpoints, stubs.StubHTTPEndpoint{
			Request: stubs.StubHTTPRequest{
				Method:          http.MethodPost,
				URL:             postgrest.Endpoint(constants.WorkbenchEndpointRPCImageRebuild),
				ExpectedHeaders: stubs.StubAuthHeaders,
				ExpectedBody:    cfg.ExpectedRequestBody,
			},
			Response: stubs.StubHTTPResponse{
				Status: cfg.ResponseStatus,
				Body:   cfg.ResponseBody,
			},
		},
		)
	}
	return endpoints
}
