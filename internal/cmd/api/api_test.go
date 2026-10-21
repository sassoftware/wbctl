// Copyright © 2026, SAS Institute Inc., Cary, NC, USA.  All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package api_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/sassoftware/wbctl/internal/auth0admin"
	"github.com/sassoftware/wbctl/internal/cmd"
	cmdapi "github.com/sassoftware/wbctl/internal/cmd/api"
	"github.com/sassoftware/wbctl/internal/constants"
	"github.com/sassoftware/wbctl/internal/postgrest"
	"github.com/sassoftware/wbctl/internal/stubs"
	"github.com/spf13/afero"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
)

const (
	postgrestAPIServerResponse = `
{
  "swagger": "2.0",
  "paths": {
    "/z": {
      "get": {
        "summary": "Get Z"
      }
    },
    "/a": {
      "get": {
        "summary": "Get A"
      },
      "post": {
        "summary": "Create A"
      }
    }
  }
}
`
	auth0AdminAPIServerResponse = `
{
  "version": 1,
  "links": [
    {
      "method": "GET",
      "rel": "users",
      "href": "/auth0Admin/users",
      "uri": "/auth0Admin/users",
      "type": "application/vnd.sas.collection"
    },
    {
      "method": "POST",
      "rel": "importUsers",
      "href": "/auth0Admin/users",
      "uri": "/auth0Admin/users",
      "type": "application/vnd.sas.collection",
      "responseType": "application/vnd.sas.collection"
    }
  ]
}
`

	expectedAPITable = "PATH\tVERB\tDESCRIPTION\n" +
		"/a\tGET, POST\tGet A, Create A\n" +
		"/auth0Admin/users\tGET, POST\t\n" +
		"/z\tGET\tGet Z"

	expectedJSON = `{
    "swagger": "2.0",
    "paths": {
        "/a": {
            "get": {
                "summary": "Get A"
            },
            "post": {
                "summary": "Create A"
            }
        },
        "/auth0Admin/users": {
            "get": {},
            "post": {}
        },
        "/z": {
            "get": {
                "summary": "Get Z"
            }
        }
    }
}`

	apiExpectedDisplayGetWithQuery = `{
    "ok": true
}`
	apiExpectedDisplayPostWithData = `{
    "created": true
}`
	apiExpectedDisplayPostWithFile = `{
    "accepted": true
}`

	apiServerInvalidJSONResponse = "invalid json"
	apiArgCustom                 = "/custom"
	apiExpectedInvalidJSONErr    = "GET ${URL}/postgrest/custom failed" +
		"\n Caused by:" +
		"\n\t* invalid character 'i' looking for beginning of value"
	apiExpectedInvalidRootJSONErr = "GET ${URL}/postgrest/ failed" +
		"\n Caused by:" +
		"\n\t* invalid character 'i' looking for beginning of value"
	apiExpectedInvalidRootAuth0AdminErr = "GET ${URL}/auth0Admin/ failed" +
		"\n Caused by:" +
		"\n\t* invalid character 'i' looking for beginning of value"
	authEndpointRequiredErr = "--endpoint must be specified"
	apiEndpointInvalidErr   = "failed to get device token\n" +
		" Caused by:\n" +
		"\t* parse \"a\\nb\": net/url: invalid control character in URL"
)

type apiInput struct {
	apiArg       string
	data         string
	endpoint     string
	filename     string
	method       string
	organization string
	output       string
	query        string
}

type apiExpected struct {
	err          string
	displays     []string
	warns        []string
	tableStrings []string
}

type apiTestCase struct {
	name          string
	input         apiInput
	mockEndpoints []stubs.StubHTTPEndpoint
	expected      apiExpected
}

type apiEndpointsConfig struct {
	expectedBody   string
	expectedQuery  string
	expectedMethod string
	expectedURL    string
	responseBody   string
}

func TestAPICommand(t *testing.T) {
	tests := []apiTestCase{
		{
			name:  "no api arg prints sorted API table",
			input: apiInput{organization: "org-1"},
			mockEndpoints: getAPIEndpoints([]apiEndpointsConfig{
				{
					expectedMethod: http.MethodGet,
					expectedURL:    postgrest.Endpoint(constants.WorkbenchEndpointRoot),
					responseBody:   postgrestAPIServerResponse,
				},
				{
					expectedMethod: http.MethodGet,
					expectedURL:    auth0admin.Endpoint(constants.Auth0AdminEndpointRoot),
					responseBody:   auth0AdminAPIServerResponse,
				},
			},
			),
			expected: apiExpected{
				tableStrings: []string{
					expectedAPITable,
				},
				displays: []string{stubs.ExpectedDeviceAuthDisplay},
			},
		},
		{
			name:  "no api arg with json output avoids table output",
			input: apiInput{organization: "org-1", output: constants.JSONOutputValue},
			mockEndpoints: getAPIEndpoints([]apiEndpointsConfig{
				{
					expectedMethod: http.MethodGet,
					expectedURL:    postgrest.Endpoint(constants.WorkbenchEndpointRoot),
					responseBody:   postgrestAPIServerResponse,
				},
				{
					expectedMethod: http.MethodGet,
					expectedURL:    auth0admin.Endpoint(constants.Auth0AdminEndpointRoot),
					responseBody:   auth0AdminAPIServerResponse,
				},
			}),
			expected: apiExpected{
				displays: []string{stubs.ExpectedDeviceAuthDisplay, expectedJSON},
			},
		},
		{
			name:  "api GET with query uses default method",
			input: apiInput{apiArg: apiArgCustom, query: "a=1&a=2&b=x", organization: "org-1"},
			mockEndpoints: getAPIEndpoints([]apiEndpointsConfig{
				{
					expectedMethod: http.MethodGet,
					expectedURL:    postgrest.Endpoint(constants.WorkbenchEndpointCustom),
					expectedQuery:  "a=1&a=2&b=x",
					responseBody:   apiExpectedDisplayGetWithQuery,
				},
			}),
			expected: apiExpected{
				displays: []string{stubs.ExpectedDeviceAuthDisplay, apiExpectedDisplayGetWithQuery},
			},
		},
		{
			name: "api POST with data body",
			input: apiInput{
				apiArg: "/submit",
				method: http.MethodPost, data: `{"name":"alpha"}`,
				organization: "org-1",
			},
			mockEndpoints: getAPIEndpoints([]apiEndpointsConfig{
				{
					expectedMethod: http.MethodPost,
					expectedURL:    postgrest.Endpoint(constants.WorkbenchEndpointSubmit),
					expectedBody:   `{"name":"alpha"}`,
					responseBody:   apiExpectedDisplayPostWithData,
				},
			}),
			expected: apiExpected{
				displays: []string{stubs.ExpectedDeviceAuthDisplay, apiExpectedDisplayPostWithData},
			},
		},
		{
			name:          "unsupported method returns exit error",
			input:         apiInput{apiArg: apiArgCustom, method: http.MethodPut, organization: "org-1"},
			mockEndpoints: stubs.GetStubAuthEndpoints(), // just the auth endpoints
			expected: apiExpected{
				err:      "unsupported method: PUT",
				displays: []string{stubs.ExpectedDeviceAuthDisplay},
			},
		},
		{
			name: "api POST with filename body",
			input: apiInput{
				apiArg:       "/submit-file",
				method:       http.MethodPost,
				filename:     `{"count":3}`,
				organization: "org-1",
			},
			mockEndpoints: getAPIEndpoints([]apiEndpointsConfig{
				{
					expectedMethod: http.MethodPost,
					expectedURL:    postgrest.Endpoint(constants.WorkbenchEndpointSubmitFile),
					expectedBody:   `{"count":3}`,
					responseBody:   apiExpectedDisplayPostWithFile,
				},
			}),
			expected: apiExpected{
				displays: []string{stubs.ExpectedDeviceAuthDisplay, apiExpectedDisplayPostWithFile},
			},
		},
		{
			name:          "invalid query returns parse error",
			input:         apiInput{apiArg: apiArgCustom, query: "%zz", organization: "org-1"},
			mockEndpoints: stubs.GetStubAuthEndpoints(), // just the auth endpoints
			expected: apiExpected{
				err:      "invalid URL escape \"%zz\"",
				displays: []string{stubs.ExpectedDeviceAuthDisplay},
			},
		},
		{
			name:          "invalid data json returns unmarshal error",
			input:         apiInput{apiArg: apiArgCustom, data: `{"name":`, organization: "org-1"},
			mockEndpoints: stubs.GetStubAuthEndpoints(), // just the auth endpoints
			expected: apiExpected{
				err:      "unexpected end of JSON input",
				displays: []string{stubs.ExpectedDeviceAuthDisplay},
			},
		},
		{
			name:     "endpoint required",
			input:    apiInput{apiArg: apiArgCustom, endpoint: "", organization: "org-1"},
			expected: apiExpected{err: authEndpointRequiredErr},
		},
		{
			name:     "invalid endpoint",
			input:    apiInput{endpoint: "a\nb", organization: "org-1"},
			expected: apiExpected{err: apiEndpointInvalidErr},
		},
		{
			name:  "no api arg invalid response returns decode error",
			input: apiInput{organization: "org-1"},
			mockEndpoints: getAPIEndpoints([]apiEndpointsConfig{
				{
					expectedMethod: http.MethodGet,
					expectedURL:    postgrest.Endpoint(constants.WorkbenchEndpointRoot),
					responseBody:   apiServerInvalidJSONResponse,
				},
			}),
			expected: apiExpected{
				err:      apiExpectedInvalidRootJSONErr,
				displays: []string{stubs.ExpectedDeviceAuthDisplay},
			},
		},
		{
			name:  "api GET invalid workbench response returns decode error",
			input: apiInput{apiArg: apiArgCustom, organization: "org-1"},
			mockEndpoints: getAPIEndpoints([]apiEndpointsConfig{
				{
					expectedMethod: http.MethodGet,
					expectedURL:    postgrest.Endpoint(constants.WorkbenchEndpointCustom),
					responseBody:   apiServerInvalidJSONResponse,
				},
			}),
			expected: apiExpected{
				err:      apiExpectedInvalidJSONErr,
				displays: []string{stubs.ExpectedDeviceAuthDisplay},
			},
		},
		{
			name:  "api GET invalid auth0Admin response returns decode error",
			input: apiInput{organization: "org-1"},
			mockEndpoints: getAPIEndpoints([]apiEndpointsConfig{
				{
					expectedMethod: http.MethodGet,
					expectedURL:    postgrest.Endpoint(constants.WorkbenchEndpointRoot),
					responseBody:   postgrestAPIServerResponse,
				},
				{
					expectedMethod: http.MethodGet,
					expectedURL:    auth0admin.Endpoint(constants.Auth0AdminEndpointRoot),
					responseBody:   apiServerInvalidJSONResponse,
				},
			}),
			expected: apiExpected{
				err: apiExpectedInvalidRootAuth0AdminErr,
				displays: []string{
					stubs.ExpectedDeviceAuthDisplay,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer viper.Reset()
			stubUI := &stubs.StubUI{}

			var srv *httptest.Server
			if tt.input.endpoint == "" && tt.name != "endpoint required" {
				srv, _ = stubs.NewStubHTTPServer(t, tt.mockEndpoints)
				defer srv.Close()
				tt.input.endpoint = srv.URL
			}

			fsys := afero.NewMemMapFs()
			if tt.input.filename != "" {
				tmpfile := filepath.Join(t.TempDir(), "request.json")
				require.NoError(t, afero.WriteFile(fsys, tmpfile, []byte(tt.input.filename), 0o600))
				tt.input.filename = tmpfile
			}

			mockCmd := cmdapi.Command(context.Background(), fsys, stubUI)
			cmd.AddRootFlags(mockCmd)
			mockCmd.SetArgs(getAPIArgs(tt))
			err := mockCmd.Execute()
			var replacements map[string]string
			if srv != nil {
				replacements = map[string]string{
					"${PORT}": stubs.Port(t, srv.URL),
					"${HOST}": stubs.Host(t, srv.URL),
					"${URL}":  srv.URL,
				}
			}
			if tt.expected.err != "" {
				expectedErr, actualErr := stubs.Sanitize(tt.expected.err, err.Error(), replacements)
				require.Equal(t, expectedErr, actualErr)
			} else {
				require.NoError(t, err)
			}

			require.Len(t, stubUI.Tables, len(tt.expected.tableStrings))
			for i, expected := range tt.expected.tableStrings {
				require.Equal(t, expected, stubUI.Tables[i].String())
				require.Equal(t, 1, stubUI.Tables[i].PrintCalls)
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
		})
	}
}

func getAPIArgs(
	tt apiTestCase,
) []string {
	args := []string{}
	args = append(args, "--"+constants.AuthModeFlag, constants.AuthModeDevice)
	if tt.input.endpoint != "" {
		args = append(args, "--"+constants.EndpointFlag, tt.input.endpoint)
	}
	if tt.input.output != "" {
		args = append(args, "--"+constants.OutputFlag, tt.input.output)
	}
	if tt.input.data != "" {
		args = append(args, "--"+constants.DataFlag, tt.input.data)
	}
	if tt.input.filename != "" {
		args = append(args, "--"+constants.FilenameFlag, tt.input.filename)
	}
	if tt.input.method != "" {
		args = append(args, "--"+constants.MethodFlag, tt.input.method)
	}
	if tt.input.query != "" {
		args = append(args, "--"+constants.QueryFlag, tt.input.query)
	}
	if tt.input.organization != "" {
		args = append(args, "--"+constants.OrganizationFlag, tt.input.organization)
	}
	if tt.input.apiArg != "" {
		args = append(args, tt.input.apiArg)
	}
	return args
}

func getAPIEndpoints(
	configs []apiEndpointsConfig,
) []stubs.StubHTTPEndpoint {
	mockEndpoints := stubs.GetStubAuthEndpoints()
	for _, config := range configs {
		mockEndpoints = append(mockEndpoints, stubs.StubHTTPEndpoint{
			Request: stubs.StubHTTPRequest{
				Method:        config.expectedMethod,
				URL:           config.expectedURL,
				ExpectedQuery: config.expectedQuery,
				ExpectedBody:  config.expectedBody,
			},
			Response: stubs.StubHTTPResponse{
				Status: http.StatusOK,
				Body:   config.responseBody,
			},
		})
	}
	return mockEndpoints
}
