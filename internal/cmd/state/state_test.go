// Copyright © 2026, SAS Institute Inc., Cary, NC, USA.  All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package state_test

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/sas-institute-rnd-saas/wbctl/internal/cmd"
	cmdstate "github.com/sas-institute-rnd-saas/wbctl/internal/cmd/state"
	"github.com/sas-institute-rnd-saas/wbctl/internal/constants"
	"github.com/sas-institute-rnd-saas/wbctl/internal/postgrest"
	"github.com/sas-institute-rnd-saas/wbctl/internal/stubs"
	"github.com/sas-institute-rnd-saas/wbctl/internal/ui"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
)

const (
	stateExpectedBodyFormat    = `{"ids":%s,"workbench":{"state":"%s"}}`
	stateExpectedDisplayWB1    = "wb-1"
	stateExpectedDisplayWB2    = "wb-2"
	stateExpectedDisplayWB3    = "wb-3"
	stateGetResponseBodyPlural = `[{
		"id":"wb-3",
		"controller_id":"controller-3"
	}]`
	stateGetResponseBodySingular = `[{
		"id":"wb-1",
		"controller_id":"controller-1"
	},{
		"id":"wb-2",
		"controller_id":"controller-2"
	}]`
	statePostResponseBodyPlural = `[{
		"id":"wb-3"
	}]`
	statePostResponseBodySingular = `[{
		"id":"wb-1"
	},{
		"id":"wb-2"
	}]`
	expectedInvalidEndpointErr = "failed to get username/password token" +
		"\n Caused by:" +
		"\n\t* parse \"a\\nb\": net/url: invalid control character in URL"
	expectedInvalidResourceErr = "" +
		"GET ${URL}/postgrest/workbenches?controller_id=in.%28controller-1%29 failed" +
		"\n Caused by:" +
		"\n\t* invalid character 'i' looking for beginning of value"
	expectedInvalidUpdateResponseErr = "" +
		"POST ${URL}/postgrest/rpc/workbench_bulk_update failed" +
		"\n Caused by:" +
		"\n\t* invalid character 'i' looking for beginning of value"
)

type stateInput struct {
	endpoint     string
	resourceIDs  []string
	resourceType string
}

type stateExpected struct {
	err       string
	displays  []string
	warns     []string
	getCalls  int
	postCalls int
}

type stateTestCase struct {
	name          string
	input         stateInput
	mockEndpoints []stubs.StubHTTPEndpoint
	expected      stateExpected
}

func TestStateCommands(t *testing.T) {
	actionTests := []struct {
		actionName     string
		desiredState   string
		newCommandFunc func(
			ctx context.Context,
			fsys afero.Fs,
			_ui ui.UI,
		) *cobra.Command
	}{
		{
			actionName:     "start",
			desiredState:   "RUNNING",
			newCommandFunc: cmdstate.StartCommand,
		},
		{
			actionName:     "stop",
			desiredState:   "STOPPED",
			newCommandFunc: cmdstate.StopCommand,
		},
	}

	tests := []stateTestCase{
		{
			name: "updates state by singular resource type",
			input: stateInput{
				resourceType: constants.WorkbenchType,
				resourceIDs:  []string{"controller-1", "controller-2"},
			},
			mockEndpoints: getStateEndpoints(stateConfig{
				getConfig: stubs.StubHTTPPartialGetConfig{
					ExpectedQuery:  "controller_id=in.(controller-1,controller-2)",
					ResponseStatus: http.StatusOK,
					ResponseBody:   stateGetResponseBodySingular,
				},
				postConfig: stubs.StubHTTPPartialPostConfig{
					ExpectedRequestBody: fmt.Sprintf(stateExpectedBodyFormat, `["wb-1","wb-2"]`, "${desiredState}"),
					ResponseStatus:      http.StatusOK,
					ResponseBody:        statePostResponseBodySingular,
				},
			}),
			expected: stateExpected{
				getCalls:  1,
				postCalls: 1,
				displays:  []string{stateExpectedDisplayWB1, stateExpectedDisplayWB2},
			},
		},
		{
			name: "updates state by plural resource type",
			input: stateInput{
				resourceType: constants.WorkbenchesType,
				resourceIDs:  []string{"controller-3"},
			},
			mockEndpoints: getStateEndpoints(stateConfig{
				getConfig: stubs.StubHTTPPartialGetConfig{
					ExpectedQuery:  "controller_id=in.(controller-3)",
					ResponseStatus: http.StatusOK,
					ResponseBody:   stateGetResponseBodyPlural,
				},
				postConfig: stubs.StubHTTPPartialPostConfig{
					ExpectedRequestBody: fmt.Sprintf(stateExpectedBodyFormat, `["wb-3"]`, "${desiredState}"),
					ResponseStatus:      http.StatusOK,
					ResponseBody:        statePostResponseBodyPlural,
				},
			}),
			expected: stateExpected{
				getCalls:  1,
				postCalls: 1,
				displays:  []string{stateExpectedDisplayWB3},
			},
		},
		{
			name: "invalid endpoint",
			input: stateInput{
				resourceType: constants.WorkbenchType,
				resourceIDs:  []string{"controller-1"},
				endpoint:     "a\nb",
			},
			expected: stateExpected{
				err: expectedInvalidEndpointErr,
			},
		},
		{
			name: "invalid get workbench response",
			input: stateInput{
				resourceType: constants.WorkbenchesType,
				resourceIDs:  []string{"controller-1"},
			},
			mockEndpoints: getStateEndpoints(stateConfig{
				getConfig: stubs.StubHTTPPartialGetConfig{
					ExpectedQuery:  "controller_id=in.(controller-1)",
					ResponseStatus: http.StatusOK,
					ResponseBody:   `[invalid`,
				},
			}),
			expected: stateExpected{
				getCalls: 1,
				err:      expectedInvalidResourceErr,
			},
		},
		{
			name: "invalid update workbench response",
			input: stateInput{
				resourceType: constants.WorkbenchesType,
				resourceIDs:  []string{"controller-3"},
			},
			mockEndpoints: getStateEndpoints(stateConfig{
				getConfig: stubs.StubHTTPPartialGetConfig{
					ExpectedQuery:  "controller_id=in.(controller-3)",
					ResponseStatus: http.StatusOK,
					ResponseBody:   stateGetResponseBodyPlural,
				},
				postConfig: stubs.StubHTTPPartialPostConfig{
					ExpectedRequestBody: fmt.Sprintf(stateExpectedBodyFormat, `["wb-3"]`, "${desiredState}"),
					ResponseStatus:      http.StatusOK,
					ResponseBody:        `[invalid`,
				},
			}),
			expected: stateExpected{
				getCalls:  1,
				postCalls: 1,
				err:       expectedInvalidUpdateResponseErr,
			},
		},
	}

	for _, action := range actionTests {
		for _, tt := range tests {
			t.Run(action.actionName+" "+tt.name, func(t *testing.T) {
				defer viper.Reset()
				srv, hitCounts := stubs.NewStubHTTPServer(
					t,
					updateMockEndpointsExpectedDesiredState(action.desiredState, tt.mockEndpoints),
				)
				defer srv.Close()

				stubUI := &stubs.StubUI{}
				if tt.input.endpoint == "" {
					tt.input.endpoint = srv.URL
				}
				mockCmd := action.newCommandFunc(context.Background(), afero.NewMemMapFs(), stubUI)
				cmd.AddRootFlags(mockCmd)
				mockCmd.SetArgs(getStateArgs(tt))
				err := mockCmd.Execute()
				if tt.expected.err != "" {
					expectedErr := strings.ReplaceAll(tt.expected.err, "${action}", action.actionName)
					expectedErr = strings.ReplaceAll(expectedErr, "${URL}", srv.URL)
					require.EqualError(t, err, expectedErr)
				} else {
					require.NoError(t, err)
				}

				require.Equal(t, tt.expected.displays, stubUI.Displays, strings.Join(stubUI.Displays, "\n"))
				require.Equal(t, tt.expected.warns, stubUI.Warns, strings.Join(stubUI.Warns, "\n"))
				require.Empty(t, stubUI.Tables)

				require.Equal(
					t,
					tt.expected.getCalls,
					hitCounts.Count("GET"+postgrest.Endpoint(constants.WorkbenchEndpointWorkbenches)),
					"get calls",
				)
				require.Equal(
					t,
					tt.expected.postCalls,
					hitCounts.Count("POST"+postgrest.Endpoint(constants.WorkbenchEndpointRPCWorkbenchBulkUpdate)),
					"update calls",
				)
			})
		}
	}
}

func TestSupportedResourceTypes(t *testing.T) {
	tests := []struct {
		name          string
		expectedTypes []string
	}{
		{
			name:          "returns supported state resource types",
			expectedTypes: []string{constants.WorkbenchesType},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := cmdstate.SupportedResourceTypes()
			require.Equal(t, tt.expectedTypes, actual)
		})
	}
}

func getStateArgs(
	tt stateTestCase,
) []string {
	args := []string{}
	args = append(args, tt.input.resourceType)
	args = append(args, "--"+constants.AuthModeFlag, constants.AuthModePassword)
	args = append(args, "--"+constants.ClientIDFlag, constants.OAuthClientID)
	args = append(args, "--"+constants.PasswordFlag, "pass")
	args = append(args, "--"+constants.UsernameFlag, "user")
	if tt.input.endpoint != "" {
		args = append(args, "--"+constants.EndpointFlag, tt.input.endpoint)
	}
	args = append(args, tt.input.resourceIDs...)
	return args
}

type stateConfig struct {
	getConfig  stubs.StubHTTPPartialGetConfig
	postConfig stubs.StubHTTPPartialPostConfig
}

func getStateEndpoints(
	config stateConfig,
) []stubs.StubHTTPEndpoint {
	endpoints := stubs.GetStubAuthEndpoints()
	endpoints = append(endpoints,
		stubs.StubHTTPEndpoint{
			Request: stubs.StubHTTPRequest{
				Method:          http.MethodGet,
				URL:             postgrest.Endpoint(constants.WorkbenchEndpointWorkbenches),
				ExpectedHeaders: stubs.StubAuthHeaders,
				ExpectedQuery:   config.getConfig.ExpectedQuery,
			},
			Response: stubs.StubHTTPResponse{
				Status: config.getConfig.ResponseStatus,
				Body:   config.getConfig.ResponseBody,
			},
		},
		stubs.StubHTTPEndpoint{
			Request: stubs.StubHTTPRequest{
				Method:          http.MethodPost,
				URL:             postgrest.Endpoint(constants.WorkbenchEndpointRPCWorkbenchBulkUpdate),
				ExpectedHeaders: stubs.StubAuthHeaders,
				ExpectedBody:    config.postConfig.ExpectedRequestBody,
			},
			Response: stubs.StubHTTPResponse{
				Status: config.postConfig.ResponseStatus,
				Body:   config.postConfig.ResponseBody,
			},
		},
	)
	return endpoints
}

func updateMockEndpointsExpectedDesiredState(
	desiredState string,
	endpoints []stubs.StubHTTPEndpoint,
) []stubs.StubHTTPEndpoint {
	updatedEndpoints := make([]stubs.StubHTTPEndpoint, 0, len(endpoints))
	for _, endpoint := range endpoints {
		if endpoint.Request.Method == http.MethodPost {
			endpoint.Request.ExpectedBody = strings.ReplaceAll(endpoint.Request.ExpectedBody, "${desiredState}", desiredState)
		}
		updatedEndpoints = append(updatedEndpoints, endpoint)
	}
	return updatedEndpoints
}
