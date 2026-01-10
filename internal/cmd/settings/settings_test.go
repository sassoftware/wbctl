// Copyright © 2026, SAS Institute Inc., Cary, NC, USA.  All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package settings_test

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/sas-institute-rnd-saas/wbctl/internal/cmd"
	"github.com/sas-institute-rnd-saas/wbctl/internal/cmd/settings"
	"github.com/sas-institute-rnd-saas/wbctl/internal/constants"
	"github.com/sas-institute-rnd-saas/wbctl/internal/postgrest"
	"github.com/sas-institute-rnd-saas/wbctl/internal/stubs"
	"github.com/spf13/afero"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
)

const (
	userSettingsReponseBody = `{
    "max_concurrent_jobs_per_user": 10,
    "max_storage_gb_per_user": 50,
    "max_workbenches_per_user": 7,
    "nfs_retain_user_sub_dir": "{{.user_name}}",
    "watchdog_timeout": 30
}`

	tenantSettingsReponseBody = `{
    "max_concurrent_jobs_per_user": 100,
    "max_storage_gb_per_user": 500,
    "max_workbenches_per_user": 70,
    "nfs_retain_user_sub_dir": "{{.user_name}}",
    "watchdog_timeout": 300
}`

	userSettingsTable = "NAME\tVALUE\n" +
		"max_concurrent_jobs_per_user\t10\n" +
		"max_storage_gb_per_user\t50\n" +
		"max_workbenches_per_user\t7\n" +
		"nfs_retain_user_sub_dir\t\"{{.user_name}}\"\n" +
		"watchdog_timeout\t30"

	tenantSettingsTable = "NAME\tVALUE\n" +
		"max_concurrent_jobs_per_user\t100\n" +
		"max_storage_gb_per_user\t500\n" +
		"max_workbenches_per_user\t70\n" +
		"nfs_retain_user_sub_dir\t\"{{.user_name}}\"\n" +
		"watchdog_timeout\t300"

	expectedUserPostBody = `{
  "name" : "user name",
  "value" : "user value"
}`

	expectedTenantPostBody = `{
  "name" : "tenant name",
  "value" : "tenant value"
}`

	expectedUserIntPostBody = `{
  "name" : "user name",
  "value" : 3
}`

	expectedUserArrayPostBody = `{
  "name" : "user name",
  "value" : [3]
}`

	expectedUserObjectPostBody = `{
  "name" : "user name",
  "value" : {"a":true}
}`

	expectedUserBoolPostBody = `{
  "name" : "user name",
  "value" : true
}`

	expectedUserError = "GET %s/postgrest/rpc/get_settings failed\n Caused by:\n\t* 400 "

	expectedTenantError = "GET %s/postgrest/rpc/get_tenant_settings failed\n Caused by:\n\t* 400 "

	expectedUserSetError = "POST %s/postgrest/rpc/set_user_setting failed\n Caused by:\n\t* 400 "

	expectedTenantSetError = "POST %s/postgrest/rpc/set_tenant_setting failed\n Caused by:\n\t* 400 "

	expectedInvalidEndpointErr = "failed to get username/password token" +
		"\n Caused by:" +
		"\n\t* parse \"a\\nb\": net/url: invalid control character in URL"
)

type settingsInput struct {
	endpoint string
	set      string
	value    string
	tenant   bool
	output   string
}

type settingsExpected struct {
	err          string
	displays     []string
	warns        []string
	tableStrings []string
}

type settingsTestCase struct {
	name          string
	input         settingsInput
	mockEndpoints []stubs.StubHTTPEndpoint
	expected      settingsExpected
}

func TestSettingsCommand(t *testing.T) {
	tests := []settingsTestCase{
		{
			name: "get user settings",
			mockEndpoints: getGetUserSettingsEndpoint(stubs.StubHTTPPartialGetConfig{
				ResponseBody:   userSettingsReponseBody,
				ResponseStatus: http.StatusOK,
			}),
			expected: settingsExpected{tableStrings: []string{userSettingsTable}},
		},
		{
			name:  "get user settings json",
			input: settingsInput{output: "json"},
			mockEndpoints: getGetUserSettingsEndpoint(stubs.StubHTTPPartialGetConfig{
				ResponseBody:   userSettingsReponseBody,
				ResponseStatus: http.StatusOK,
			}),
			expected: settingsExpected{displays: []string{userSettingsReponseBody}},
		},
		{
			name: "get user settings error",
			mockEndpoints: getGetUserSettingsEndpoint(stubs.StubHTTPPartialGetConfig{
				ResponseStatus: http.StatusBadRequest,
			}),
			expected: settingsExpected{err: expectedUserError},
		},
		{
			name:  "get tenant settings",
			input: settingsInput{tenant: true},
			mockEndpoints: getGetTenantSettingsEndpoint(stubs.StubHTTPPartialGetConfig{
				ResponseStatus: http.StatusOK,
				ResponseBody:   tenantSettingsReponseBody,
			}),
			expected: settingsExpected{tableStrings: []string{tenantSettingsTable}},
		},
		{
			name:  "get tenant settings json",
			input: settingsInput{tenant: true, output: "json"},
			mockEndpoints: getGetTenantSettingsEndpoint(stubs.StubHTTPPartialGetConfig{
				ResponseStatus: http.StatusOK,
				ResponseBody:   tenantSettingsReponseBody,
			}),
			expected: settingsExpected{displays: []string{tenantSettingsReponseBody}},
		},
		{
			name:  "get tenant settings error",
			input: settingsInput{tenant: true},
			mockEndpoints: getGetTenantSettingsEndpoint(stubs.StubHTTPPartialGetConfig{
				ResponseStatus: http.StatusBadRequest,
			}),
			expected: settingsExpected{err: expectedTenantError},
		},
		{
			name:  "set user settings",
			input: settingsInput{set: "user name", value: "user value"},
			mockEndpoints: getSetUserSettingsEndpoint(stubs.StubHTTPPartialPostConfig{
				ExpectedRequestBody: expectedUserPostBody,
				ResponseStatus:      http.StatusOK,
				ResponseBody:        "true",
			}),
			expected: settingsExpected{displays: []string{"user name"}},
		},
		{
			name:  "set user settings - int",
			input: settingsInput{set: "user name", value: "3"},
			mockEndpoints: getSetUserSettingsEndpoint(stubs.StubHTTPPartialPostConfig{
				ExpectedRequestBody: expectedUserIntPostBody,
				ResponseStatus:      http.StatusOK,
				ResponseBody:        "true",
			}),
			expected: settingsExpected{displays: []string{"user name"}},
		},
		{
			name:  "set user settings - array",
			input: settingsInput{set: "user name", value: "[3]"},
			mockEndpoints: getSetUserSettingsEndpoint(stubs.StubHTTPPartialPostConfig{
				ExpectedRequestBody: expectedUserArrayPostBody,
				ResponseStatus:      http.StatusOK,
				ResponseBody:        "true",
			}),
			expected: settingsExpected{displays: []string{"user name"}},
		},
		{
			name:  "set user settings - object",
			input: settingsInput{set: "user name", value: `{"a":true}`},
			mockEndpoints: getSetUserSettingsEndpoint(stubs.StubHTTPPartialPostConfig{
				ExpectedRequestBody: expectedUserObjectPostBody,
				ResponseStatus:      http.StatusOK,
				ResponseBody:        "true",
			}),
			expected: settingsExpected{displays: []string{"user name"}},
		},
		{
			name:  "set user settings - bool",
			input: settingsInput{set: "user name", value: `true`},
			mockEndpoints: getSetUserSettingsEndpoint(stubs.StubHTTPPartialPostConfig{
				ExpectedRequestBody: expectedUserBoolPostBody,
				ResponseStatus:      http.StatusOK,
				ResponseBody:        "true",
			}),
			expected: settingsExpected{displays: []string{"user name"}},
		},
		{
			name:  "set tenant settings",
			input: settingsInput{set: "tenant name", value: "tenant value", tenant: true},
			mockEndpoints: getSetTenantSettingsEndpoint(stubs.StubHTTPPartialPostConfig{
				ExpectedRequestBody: expectedTenantPostBody,
				ResponseStatus:      http.StatusOK,
				ResponseBody:        "true",
			}),
			expected: settingsExpected{displays: []string{"tenant name"}},
		},
		{
			name:  "set user settings error",
			input: settingsInput{set: "user name", value: "user value"},
			mockEndpoints: getSetUserSettingsEndpoint(stubs.StubHTTPPartialPostConfig{
				ExpectedRequestBody: expectedUserPostBody,
				ResponseStatus:      http.StatusBadRequest,
			}),
			expected: settingsExpected{err: expectedUserSetError},
		},
		{
			name:  "set tenant settings error",
			input: settingsInput{set: "tenant name", value: "tenant value", tenant: true},
			mockEndpoints: getSetTenantSettingsEndpoint(stubs.StubHTTPPartialPostConfig{
				ExpectedRequestBody: expectedTenantPostBody,
				ResponseStatus:      http.StatusBadRequest,
			}),
			expected: settingsExpected{err: expectedTenantSetError},
		},
		{
			name: "invalid endpoint",
			input: settingsInput{
				endpoint: "a\nb",
			},
			expected: settingsExpected{
				err: expectedInvalidEndpointErr,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer viper.Reset()
			srv, _ := stubs.NewStubHTTPServer(t, tt.mockEndpoints)
			defer srv.Close()

			stubUI := &stubs.StubUI{}
			if tt.input.endpoint == "" {
				tt.input.endpoint = srv.URL
			}
			mockCmd := settings.Command(context.Background(), afero.NewMemMapFs(), stubUI)
			cmd.AddRootFlags(mockCmd)
			mockCmd.SetArgs(getSettingsArgs(tt))
			err := mockCmd.Execute()
			if tt.expected.err != "" {
				if strings.Contains(tt.expected.err, "%s") {
					require.EqualError(t, err, fmt.Sprintf(tt.expected.err, srv.URL))
				} else {
					require.EqualError(t, err, tt.expected.err)
				}
			} else {
				require.NoError(t, err)
			}

			require.Len(t, stubUI.Tables, len(tt.expected.tableStrings))
			for i, expected := range tt.expected.tableStrings {
				actual := stubUI.Tables[i].String()
				require.Equal(t, expected, actual, actual)
				require.Equal(t, 1, stubUI.Tables[i].PrintCalls)
			}
			require.Equal(t, tt.expected.displays, stubUI.Displays, strings.Join(stubUI.Displays, "\n"))
			require.Equal(t, tt.expected.warns, stubUI.Warns, strings.Join(stubUI.Warns, "\n"))
		})
	}
}

func getSettingsArgs(
	tt settingsTestCase,
) []string {
	args := []string{}
	args = append(args, "--"+constants.AuthModeFlag, constants.AuthModePassword)
	args = append(args, "--"+constants.ClientIDFlag, constants.OAuthClientID)
	args = append(args, "--"+constants.PasswordFlag, "pass")
	args = append(args, "--"+constants.UsernameFlag, "user")
	if tt.input.endpoint != "" {
		args = append(args, "--"+constants.EndpointFlag, tt.input.endpoint)
	}
	if tt.input.set != "" {
		args = append(args, "--"+constants.SetFlag, tt.input.set)
	}
	if tt.input.value != "" {
		args = append(args, "--"+constants.ValueFlag, tt.input.value)
	}
	if tt.input.tenant {
		args = append(args, "--"+constants.TenantFlag)
	}
	if tt.input.output != "" {
		args = append(args, "--"+constants.OutputFlag, tt.input.output)
	}
	return args
}

func getGetUserSettingsEndpoint(
	config stubs.StubHTTPPartialGetConfig,
) []stubs.StubHTTPEndpoint {
	endpoints := stubs.GetStubAuthEndpoints()
	return append(endpoints,
		stubs.StubHTTPEndpoint{
			Request: stubs.StubHTTPRequest{
				Method:          http.MethodGet,
				URL:             postgrest.Endpoint(constants.WorkbenchEndpointWorkspaceGetSettings),
				ExpectedHeaders: stubs.StubAuthHeaders,
				ExpectedQuery:   config.ExpectedQuery,
			},
			Response: stubs.StubHTTPResponse{
				Status: config.ResponseStatus,
				Body:   config.ResponseBody,
			},
		},
	)
}

func getSetUserSettingsEndpoint(
	config stubs.StubHTTPPartialPostConfig,
) []stubs.StubHTTPEndpoint {
	endpoints := stubs.GetStubAuthEndpoints()
	return append(endpoints,
		stubs.StubHTTPEndpoint{
			Request: stubs.StubHTTPRequest{
				Method:          http.MethodPost,
				URL:             postgrest.Endpoint(constants.WorkbenchEndpointWorkspaceSetUserSetting),
				ExpectedHeaders: stubs.StubAuthHeaders,
				ExpectedBody:    config.ExpectedRequestBody,
			},
			Response: stubs.StubHTTPResponse{
				Status: config.ResponseStatus,
				Body:   config.ResponseBody,
			},
		},
	)
}

func getGetTenantSettingsEndpoint(
	config stubs.StubHTTPPartialGetConfig,
) []stubs.StubHTTPEndpoint {
	endpoints := stubs.GetStubAuthEndpoints()
	return append(endpoints,
		stubs.StubHTTPEndpoint{
			Request: stubs.StubHTTPRequest{
				Method:          http.MethodGet,
				URL:             postgrest.Endpoint(constants.WorkbenchEndpointWorkspaceGetTenantSettings),
				ExpectedHeaders: stubs.StubAuthHeaders,
				ExpectedQuery:   config.ExpectedQuery,
			},
			Response: stubs.StubHTTPResponse{
				Status: config.ResponseStatus,
				Body:   config.ResponseBody,
			},
		},
	)
}

func getSetTenantSettingsEndpoint(
	config stubs.StubHTTPPartialPostConfig,
) []stubs.StubHTTPEndpoint {
	endpoints := stubs.GetStubAuthEndpoints()
	return append(endpoints,
		stubs.StubHTTPEndpoint{
			Request: stubs.StubHTTPRequest{
				Method:          http.MethodPost,
				URL:             postgrest.Endpoint(constants.WorkbenchEndpointWorkspaceSetTenantSetting),
				ExpectedHeaders: stubs.StubAuthHeaders,
				ExpectedBody:    config.ExpectedRequestBody,
			},
			Response: stubs.StubHTTPResponse{
				Status: config.ResponseStatus,
				Body:   config.ResponseBody,
			},
		},
	)
}
