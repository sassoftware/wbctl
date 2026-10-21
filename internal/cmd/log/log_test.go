// Copyright © 2026, SAS Institute Inc., Cary, NC, USA.  All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package log_test

import (
	"context"
	"net/http"
	"strings"
	"testing"

	"github.com/sassoftware/wbctl/internal/cmd"
	cmdlog "github.com/sassoftware/wbctl/internal/cmd/log"
	"github.com/sassoftware/wbctl/internal/constants"
	"github.com/sassoftware/wbctl/internal/postgrest"
	"github.com/sassoftware/wbctl/internal/stubs"
	"github.com/spf13/afero"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
)

const (
	logImageBuildsResponseBody = `[
		{"id":"image-build-1","log_url":"http://${HOST}/logs/image-build-1"},
		{"id":"image-build-2","log_url":""}
	]`
	logJobSubmissionsResponseBody = `[
		{"id":"job-submission-1","log_url":"http://${HOST}/logs/job-submission-1"},
		{"id":"job-submission-2","log_url":""}
	]`
	expectedInvalidEndpointErr = "failed to get device token" +
		"\n Caused by:" +
		"\n\t* parse \"a\\nb\": net/url: invalid control character in URL"
)

type logInput struct {
	endpoint     string
	resourceIDs  []string
	resourceType string
}

type logExpected struct {
	err           string
	displays      []string
	warns         []string
	rawCalls      int
	resourceCalls int
}

type logTestCase struct {
	name          string
	input         logInput
	mockEndpoints []stubs.StubHTTPEndpoint
	expected      logExpected
}

func TestLogCommand(t *testing.T) {
	tests := []logTestCase{
		{
			name: "new log function image build list non-200 from server",
			input: logInput{
				resourceType: constants.ImageBuildsType,
				resourceIDs:  []string{"controller-1"},
			},
			mockEndpoints: getLogEndpoints(logConfig{
				resourceURL: postgrest.Endpoint(constants.WorkbenchEndpointImageBuilds),
				getResourceConfig: stubs.StubHTTPPartialGetConfig{
					ExpectedQuery:  "controller_id=in.(controller-1)",
					ResponseStatus: http.StatusInternalServerError,
					ResponseBody:   `resource-error`,
				},
			}),
			expected: logExpected{
				displays: []string{stubs.ExpectedDeviceAuthDisplay},
				err: "GET ${URL}/postgrest/image_builds?controller_id=in.%28controller-1%29 failed\n" +
					" Caused by:\n" +
					"\t* 500 resource-error",
				resourceCalls: 1},
		},
		{
			name: "new log function image build raw log non-200 from server",
			input: logInput{
				resourceType: constants.ImageBuildsType,
				resourceIDs:  []string{"controller-1"},
			},
			mockEndpoints: getLogEndpoints(logConfig{
				resourceURL: postgrest.Endpoint(constants.WorkbenchEndpointImageBuilds),
				getResourceConfig: stubs.StubHTTPPartialGetConfig{
					ExpectedQuery:  "controller_id=in.(controller-1)",
					ResponseStatus: http.StatusOK,
					ResponseBody:   logImageBuildsResponseBody,
				},
				logURL: "/logs/raw/image-build-1",
				getLogConfig: stubs.StubHTTPPartialGetConfig{
					ResponseStatus: http.StatusForbidden,
					ResponseBody:   "forbidden",
				},
			}),
			expected: logExpected{
				displays: []string{stubs.ExpectedDeviceAuthDisplay},
				err: "GET ${URL}/logs/raw/image-build-1 failed\n" +
					" Caused by:\n" +
					"\t* 403 forbidden",
				rawCalls: 1, resourceCalls: 1},
		},
		{
			name: "new log function job submission list non-200 from server",
			input: logInput{
				resourceType: constants.JobSubmissionsType,
				resourceIDs:  []string{"controller-1"},
			},
			mockEndpoints: getLogEndpoints(logConfig{
				resourceURL: postgrest.Endpoint(constants.WorkbenchEndpointJobRuns),
				getResourceConfig: stubs.StubHTTPPartialGetConfig{
					ExpectedQuery:  "controller_id=in.(controller-1)",
					ResponseStatus: http.StatusInternalServerError,
					ResponseBody:   "resource-error",
				},
			}),
			expected: logExpected{
				displays: []string{stubs.ExpectedDeviceAuthDisplay},
				err: "GET ${URL}/postgrest/job_runs?controller_id=in.%28controller-1%29 failed\n" +
					" Caused by:\n" +
					"\t* 500 resource-error",
				resourceCalls: 1},
		},
		{
			name: "new log function job submission raw log non-200 from server",
			input: logInput{
				resourceType: constants.JobSubmissionsType,
				resourceIDs:  []string{"controller-1"},
			},
			mockEndpoints: getLogEndpoints(logConfig{
				resourceURL: postgrest.Endpoint(constants.WorkbenchEndpointJobRuns),
				getResourceConfig: stubs.StubHTTPPartialGetConfig{
					ExpectedQuery:  "controller_id=in.(controller-1)",
					ResponseStatus: http.StatusOK,
					ResponseBody:   logJobSubmissionsResponseBody,
				},
				logURL: "/logs/raw/job-submission-1",
				getLogConfig: stubs.StubHTTPPartialGetConfig{
					ResponseStatus: http.StatusForbidden,
					ResponseBody:   "forbidden",
				},
			}),
			expected: logExpected{
				displays: []string{stubs.ExpectedDeviceAuthDisplay},
				err: "GET ${URL}/logs/raw/job-submission-1 failed\n" +
					" Caused by:\n" +
					"\t* 403 forbidden",
				rawCalls: 1, resourceCalls: 1},
		},
		{
			name:  "new log function prints image build logs by singular type",
			input: logInput{resourceType: constants.ImageBuildType, resourceIDs: []string{"controller-1"}},
			mockEndpoints: getLogEndpoints(logConfig{
				resourceURL: postgrest.Endpoint(constants.WorkbenchEndpointImageBuilds),
				getResourceConfig: stubs.StubHTTPPartialGetConfig{
					ExpectedQuery:  "controller_id=in.(controller-1)",
					ResponseStatus: http.StatusOK,
					ResponseBody:   logImageBuildsResponseBody,
				},
				logURL: "/logs/raw/image-build-1",
				getLogConfig: stubs.StubHTTPPartialGetConfig{
					ResponseStatus: http.StatusOK,
					ResponseBody:   "image-build-log-1\n",
				},
			}),
			expected: logExpected{
				displays:      []string{stubs.ExpectedDeviceAuthDisplay, "image-build-log-1\n"},
				rawCalls:      1,
				resourceCalls: 1,
			},
		},
		{
			name: "new log function prints image build logs by plural type",
			input: logInput{
				resourceType: constants.ImageBuildsType,
				resourceIDs:  []string{"controller-1"},
			},
			mockEndpoints: getLogEndpoints(logConfig{
				resourceURL: postgrest.Endpoint(constants.WorkbenchEndpointImageBuilds),
				getResourceConfig: stubs.StubHTTPPartialGetConfig{
					ExpectedQuery:  "controller_id=in.(controller-1)",
					ResponseStatus: http.StatusOK,
					ResponseBody:   logImageBuildsResponseBody,
				},
				logURL: "/logs/raw/image-build-1",
				getLogConfig: stubs.StubHTTPPartialGetConfig{
					ResponseStatus: http.StatusOK,
					ResponseBody:   "image-build-log-1\n",
				},
			}),
			expected: logExpected{
				displays:      []string{stubs.ExpectedDeviceAuthDisplay, "image-build-log-1\n"},
				rawCalls:      1,
				resourceCalls: 1,
			},
		},
		{
			name:  "new log function prints job submission logs by singular type",
			input: logInput{resourceType: constants.JobSubmissionType, resourceIDs: []string{"controller-2"}},
			mockEndpoints: getLogEndpoints(logConfig{
				resourceURL: postgrest.Endpoint(constants.WorkbenchEndpointJobRuns),
				getResourceConfig: stubs.StubHTTPPartialGetConfig{
					ExpectedQuery:  "controller_id=in.(controller-2)",
					ResponseStatus: http.StatusOK,
					ResponseBody:   logJobSubmissionsResponseBody,
				},
				logURL: "/logs/raw/job-submission-1",
				getLogConfig: stubs.StubHTTPPartialGetConfig{
					ResponseStatus: http.StatusOK,
					ResponseBody:   "job-submission-log-1\n",
				},
			}),
			expected: logExpected{
				displays:      []string{stubs.ExpectedDeviceAuthDisplay, "job-submission-log-1\n"},
				rawCalls:      1,
				resourceCalls: 1,
			},
		},
		{
			name:  "new log function prints job submission logs by plural type",
			input: logInput{resourceType: constants.JobSubmissionsType, resourceIDs: []string{"controller-2"}},
			mockEndpoints: getLogEndpoints(logConfig{
				resourceURL: postgrest.Endpoint(constants.WorkbenchEndpointJobRuns),
				getResourceConfig: stubs.StubHTTPPartialGetConfig{
					ExpectedQuery:  "controller_id=in.(controller-2)",
					ResponseStatus: http.StatusOK,
					ResponseBody:   logJobSubmissionsResponseBody,
				},
				logURL: "/logs/raw/job-submission-1",
				getLogConfig: stubs.StubHTTPPartialGetConfig{
					ResponseStatus: http.StatusOK,
					ResponseBody:   "job-submission-log-1\n",
				},
			}),
			expected: logExpected{
				displays:      []string{stubs.ExpectedDeviceAuthDisplay, "job-submission-log-1\n"},
				rawCalls:      1,
				resourceCalls: 1,
			},
		},
		{
			name: "invalid endpoint",
			input: logInput{
				endpoint:     "a\nb",
				resourceType: constants.JobSubmissionType,
				resourceIDs:  []string{"controller-1"},
			},
			expected: logExpected{
				err: expectedInvalidEndpointErr,
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
			mockCmd := cmdlog.Command(context.Background(), afero.NewMemMapFs(), stubUI)
			cmd.AddRootFlags(mockCmd)
			mockCmd.SetArgs(getLogArgs(tt))
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
				tt.expected.resourceCalls,
				hitCounts.CountPrefix("GET"+constants.PostgrestAPIEndpoint),
				"resource calls",
			)
			require.Equal(
				t,
				tt.expected.rawCalls,
				hitCounts.CountPrefix("GET/logs/raw"),
				"raw calls",
			)
		})
	}
}

func TestSupportedResourceTypes(t *testing.T) {
	require.Equal(t, []string{
		constants.ImageBuildsType,
		constants.JobSubmissionsType,
	}, cmdlog.SupportedResourceTypes())
}

func getLogArgs(
	tt logTestCase,
) []string {
	args := []string{}
	args = append(args, tt.input.resourceType)
	args = append(args, "--"+constants.AuthModeFlag, constants.AuthModeDevice)
	args = append(args, "--"+constants.OrganizationFlag, "org-1")
	if tt.input.endpoint != "" {
		args = append(args, "--"+constants.EndpointFlag, tt.input.endpoint)
	}
	args = append(args, tt.input.resourceIDs...)
	return args
}

type logConfig struct {
	resourceURL       string
	getResourceConfig stubs.StubHTTPPartialGetConfig
	logURL            string
	getLogConfig      stubs.StubHTTPPartialGetConfig
}

func getLogEndpoints(
	config logConfig,
) []stubs.StubHTTPEndpoint {
	endpoints := stubs.GetStubAuthEndpoints()
	endpoints = append(endpoints,
		stubs.StubHTTPEndpoint{
			Request: stubs.StubHTTPRequest{
				Method:          http.MethodGet,
				URL:             config.resourceURL,
				ExpectedHeaders: stubs.StubAuthHeaders,
				ExpectedQuery:   config.getResourceConfig.ExpectedQuery,
			},
			Response: stubs.StubHTTPResponse{
				Status:      config.getResourceConfig.ResponseStatus,
				Body:        config.getResourceConfig.ResponseBody,
				BodyOptions: []stubs.StubHTTPBodyOption{stubs.StubHTTPBodyReplaceHOST()},
			},
		},
		stubs.StubHTTPEndpoint{
			Request: stubs.StubHTTPRequest{
				Method:          http.MethodGet,
				URL:             config.logURL,
				ExpectedHeaders: stubs.StubAuthHeaders,
				ExpectedQuery:   config.getLogConfig.ExpectedQuery,
			},
			Response: stubs.StubHTTPResponse{
				Status: config.getLogConfig.ResponseStatus,
				Body:   config.getLogConfig.ResponseBody,
			},
		},
	)
	return endpoints
}
