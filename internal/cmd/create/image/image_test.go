// Copyright © 2026, SAS Institute Inc., Cary, NC, USA.  All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package image_test

import (
	"context"
	"net/http"
	"strings"
	"testing"

	"github.com/sassoftware/wbctl/internal/cmd"
	"github.com/sassoftware/wbctl/internal/cmd/create/image"
	"github.com/sassoftware/wbctl/internal/constants"
	"github.com/sassoftware/wbctl/internal/osutil"
	"github.com/sassoftware/wbctl/internal/postgrest"
	"github.com/sassoftware/wbctl/internal/stubs"
	"github.com/spf13/afero"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
)

const (
	expectedCreateImageDryRunOutput = `{
    "context": "build/src",
    "dataplane_id": "dp-1",
    "description": "image desc",
    "image_name": "myimage",
    "image_registry": "registry.example.com",
    "image_tag": "v1",
    "name": "image-1"
}`
	expectedCreateImageRequestBody = `{"param":{"context":"build/src","dataplane_id":"dp-1",` +
		`"description":"image desc","image_name":"myimage",` +
		`"image_registry":"registry.example.com","image_tag":"v1","name":"image-1"}}`
	expectedCreateImageSpecRequestBody = `{"param":{"context":"build/src","dataplane_id":"dp-1",` +
		`"description":"image desc","image_name":"myimage",` +
		`"image_registry":"registry.example.com","image_tag":"v1",` +
		`"name":"image-1","spec":{"version":"3"}}}`
	expectedCreateImageSpecFileRequestBody = `{"param":{"context":"build/src","dataplane_id":"dp-1",` +
		`"description":"image desc","image_name":"myimage",` +
		`"image_registry":"registry.example.com","image_tag":"v1",` +
		`"name":"image-1","spec":{"version":"4"}}}`
	expectedCreateImageFileContent = `{"dataplane_id":"dp-1","name":"image-1",` +
		`"description":"image desc","context":"build/src",` +
		`"image_name":"myimage","image_registry":"registry.example.com","image_tag":"v1"}`
	expectedCreateImageIncompleteFileContent = `{"dataplane_id":"dp-1",` +
		`"description":"image desc","context":"build/src",` +
		`"image_name":"myimage","image_registry":"registry.example.com","image_tag":"v1"}`
	expectedCreateImageIncompleteFileErr = "--name is required"
	expectedCreateImageSpecFileContent   = `{"version":"4"}`
	expectedCreateImageID                = "img-1"
	expectedCreateImageNameError         = "--name is required"
	expectedCreateImageContextErr        = "either --context or --spec is required"
	expectedCreateImageServerErr         = "POST ${URL}/postgrest/rpc/image_create failed\n" +
		" Caused by:\n" +
		"\t* 401 unauthorized"
	expectedUpdateImageDryRunOutput = `{
    "ids": [
        "img-cli-1",
        "img-cli-2"
    ],
    "param": {
        "context": "build/src",
        "description": "image update desc",
        "image_name": "myimage",
        "image_registry": "registry.example.com",
        "image_tag": "v2",
        "spec": {
            "version": "3"
        },
        "sub_path": "src"
    }
}`
	expectedUpdateImageRequestBodyMapping = `{"ids":["img-cli-1","img-cli-2"],` +
		`"param":{"context":"build/src",` +
		`"description":"image update desc","image_name":"myimage",` +
		`"image_registry":"registry.example.com","image_tag":"v2",` +
		`"spec":{"version":"3"},"sub_path":"src"}}`
	expectedUpdateImageRequestBodyIDsOnly = `{"ids":["img-cli-1"],"param":{}}`
	expectedUpdateImageFileContent        = `{"id":"img-file","description":"from file"}`
	expectedUpdateImageFileDryRunOutput   = `{
    "ids": [
        "img-file"
    ],
    "param": {
        "description": "from file"
    }
}`
	expectedUpdateImageMissingIDErr = "at least one resource ID must be provided for update"
	expectedUpdateImageServerErr    = "POST ${URL}/postgrest/rpc/image_bulk_update failed\n" +
		" Caused by:\n" +
		"\t* 401 unauthorized"
	expectedUpdateImageRequestBodyAllFlags = `{"ids":["img-cli-1"],` +
		`"param":{"context":"build/src",` +
		`"description":"image update desc","image_name":"myimage",` +
		`"image_registry":"registry.example.com","image_tag":"v2","labels":{"team":"core"},` +
		`"name":"img-name","spec":{"version":"3"},"sub_path":"src"}}`
	expectedUpdateImageLongFlagsDryRunOutput = `{
    "ids": [
        "img-cli-1"
    ],
    "param": {
        "context": "build/src",
        "description": "image update desc",
        "image_name": "myimage",
        "image_registry": "registry.example.com",
        "image_tag": "v2",
        "labels": {
            "team": "core"
        },
        "name": "img-name",
        "spec": {
            "version": "3"
        },
        "sub_path": "src"
    }
}`
	expectedInvalidEndpointErr = "failed to get device token" +
		"\n Caused by:" +
		"\n\t* parse \"a\\nb\": net/url: invalid control character in URL"
	expectedInvlidDataplaneErr = "GET ${URL}/postgrest/dataplanes?id=in.%28invalid%29 failed" +
		"\n Caused by:" +
		"\n\t* 400 error"
	expectedInvalidSpecErr            = "invalid character 'i' looking for beginning of value"
	expectedCreateImageInvalidFileErr = "unexpected end of JSON input"
)

type createInput struct {
	endpoint    string
	filename    string
	specFile    string
	fileContent string
	spec        string
	dryRun      bool
	dataplane   string
	name        string
	description string
	context     string
	image       string
}

type createExpected struct {
	displays       []string
	warns          []string
	err            string
	createCalls    int
	dataplaneCalls int
}

type createImageTestCase struct {
	name          string
	input         createInput
	mockEndpoints []stubs.StubHTTPEndpoint
	expected      createExpected
}

func TestCreateImageCommand(t *testing.T) {
	tests := []createImageTestCase{
		{
			name: "dry run inline body",
			input: createInput{
				dryRun:      true,
				dataplane:   "dp-1",
				name:        "image-1",
				description: "image desc",
				context:     "build/src",
				image:       "registry.example.com/myimage:v1",
			},
			mockEndpoints: getCreateImageEndpoints(imageEndpointsConfig{
				expectedDataplanesQuery:  `id=in.(dp-1)`,
				dataplanesResponseStatus: http.StatusOK,
				dataplanesResponseBody:   `[{"id":"dp-1"}]`,
			}),
			expected: createExpected{
				displays:       []string{stubs.ExpectedDeviceAuthDisplay, expectedCreateImageDryRunOutput},
				dataplaneCalls: 1,
			},
		},
		{
			name: "dry run from file",
			input: createInput{
				filename:    "/input.json",
				fileContent: expectedCreateImageFileContent,
				dryRun:      true,
			},
			mockEndpoints: getCreateImageEndpoints(imageEndpointsConfig{
				dataplanesResponseStatus: http.StatusOK,
				dataplanesResponseBody:   `[{"id":"dp-1"}]`,
			}),
			expected: createExpected{
				dataplaneCalls: 1,
				displays:       []string{stubs.ExpectedDeviceAuthDisplay, expectedCreateImageDryRunOutput},
			},
		},
		{
			name: "dry run from incomplete file",
			input: createInput{
				filename:    "/input.json",
				fileContent: expectedCreateImageIncompleteFileContent,
				dryRun:      true,
			},
			mockEndpoints: getCreateImageEndpoints(imageEndpointsConfig{
				dataplanesResponseStatus: http.StatusOK,
				dataplanesResponseBody:   `[{"id":"dp-1"}]`,
			}),
			expected: createExpected{
				dataplaneCalls: 1,
				err:            expectedCreateImageIncompleteFileErr,
				displays:       []string{stubs.ExpectedDeviceAuthDisplay},
			},
		},
		{
			name: "create from invalid",
			input: createInput{
				name:        "image-1",
				filename:    "/input.json",
				fileContent: `[`,
			},
			mockEndpoints: stubs.GetStubAuthEndpoints(),
			expected: createExpected{
				err:      expectedCreateImageInvalidFileErr,
				displays: []string{stubs.ExpectedDeviceAuthDisplay},
			},
		},
		{
			name: "create success",
			input: createInput{
				dataplane:   "dp-1",
				name:        "image-1",
				description: "image desc",
				context:     "build/src",
				image:       "registry.example.com/myimage:v1",
			},
			mockEndpoints: getCreateImageEndpoints(imageEndpointsConfig{
				expectedDataplanesQuery:        `id=in.(dp-1)`,
				dataplanesResponseStatus:       http.StatusOK,
				dataplanesResponseBody:         `[{"id":"dp-1"}]`,
				expectedImageCreateRequestBody: expectedCreateImageRequestBody,
				imageCreateResponseStatus:      http.StatusOK,
				imageCreateResponseBody:        `{"id":"img-1"}`,
			}),
			expected: createExpected{
				displays:       []string{stubs.ExpectedDeviceAuthDisplay, expectedCreateImageID},
				createCalls:    1,
				dataplaneCalls: 1,
			},
		},
		{
			name: "create success with literal spec",
			input: createInput{
				dataplane:   "dp-1",
				name:        "image-1",
				description: "image desc",
				context:     "build/src",
				image:       "registry.example.com/myimage:v1",
				spec:        `{"version":"3"}`,
			},
			mockEndpoints: getCreateImageEndpoints(imageEndpointsConfig{
				expectedDataplanesQuery:        `id=in.(dp-1)`,
				dataplanesResponseStatus:       http.StatusOK,
				dataplanesResponseBody:         `[{"id":"dp-1"}]`,
				expectedImageCreateRequestBody: expectedCreateImageSpecRequestBody,
				imageCreateResponseStatus:      http.StatusOK,
				imageCreateResponseBody:        `{"id":"img-1"}`,
			}),
			expected: createExpected{
				displays:       []string{stubs.ExpectedDeviceAuthDisplay, expectedCreateImageID},
				createCalls:    1,
				dataplaneCalls: 1,
			},
		},
		{
			name: "create success with spec file",
			input: createInput{
				dataplane:   "dp-1",
				name:        "image-1",
				description: "image desc",
				context:     "build/src",
				image:       "registry.example.com/myimage:v1",
				specFile:    "/spec.json",
				fileContent: expectedCreateImageSpecFileContent,
				spec:        "/spec.json",
			},
			mockEndpoints: getCreateImageEndpoints(imageEndpointsConfig{
				expectedDataplanesQuery:        `id=in.(dp-1)`,
				dataplanesResponseStatus:       http.StatusOK,
				dataplanesResponseBody:         `[{"id":"dp-1"}]`,
				expectedImageCreateRequestBody: expectedCreateImageSpecFileRequestBody,
				imageCreateResponseStatus:      http.StatusOK,
				imageCreateResponseBody:        `{"id":"img-1"}`,
			}),
			expected: createExpected{
				displays:       []string{stubs.ExpectedDeviceAuthDisplay, expectedCreateImageID},
				createCalls:    1,
				dataplaneCalls: 1,
			},
		},
		{
			name: "create error",
			input: createInput{
				dataplane:   "dp-1",
				name:        "image-1",
				description: "image desc",
				context:     "build/src",
				image:       "registry.example.com/myimage:v1",
			},
			mockEndpoints: getCreateImageEndpoints(imageEndpointsConfig{
				expectedDataplanesQuery:        `id=in.(dp-1)`,
				dataplanesResponseStatus:       http.StatusOK,
				dataplanesResponseBody:         `[{"id":"dp-1"}]`,
				expectedImageCreateRequestBody: expectedCreateImageRequestBody,
				imageCreateResponseStatus:      http.StatusUnauthorized,
				imageCreateResponseBody:        `unauthorized`,
			}),
			expected: createExpected{
				err:            expectedCreateImageServerErr,
				createCalls:    1,
				dataplaneCalls: 1,
				displays:       []string{stubs.ExpectedDeviceAuthDisplay},
			},
		},
		{
			name: "name required",
			input: createInput{
				context: "build/src",
				image:   "registry.example.com/myimage:v1",
			},
			mockEndpoints: getCreateImageEndpoints(imageEndpointsConfig{
				dataplanesResponseStatus: http.StatusOK,
				dataplanesResponseBody:   `[{"id":"dp-1"}]`,
			}),
			expected: createExpected{
				dataplaneCalls: 1,
				displays:       []string{stubs.ExpectedDeviceAuthDisplay},
				err:            expectedCreateImageNameError,
			},
		},
		{
			name: "context or spec required",
			input: createInput{
				name:  "image-1",
				image: "registry.example.com/myimage:v1",
			},
			mockEndpoints: getCreateImageEndpoints(imageEndpointsConfig{
				dataplanesResponseStatus: http.StatusOK,
				dataplanesResponseBody:   `[{"id":"dp-1"}]`,
			}),
			expected: createExpected{
				err:            expectedCreateImageContextErr,
				dataplaneCalls: 1,
				displays:       []string{stubs.ExpectedDeviceAuthDisplay},
			},
		},
		{
			name: "invalid endpoint",
			input: createInput{
				endpoint: "a\nb",
			},
			expected: createExpected{
				err: expectedInvalidEndpointErr,
			},
		},
		{
			name: "invalid dataplane id",
			input: createInput{
				dryRun:    true,
				dataplane: "invalid",
			},
			mockEndpoints: getCreateImageEndpoints(imageEndpointsConfig{
				expectedDataplanesQuery:  "id=in.(invalid)",
				dataplanesResponseStatus: http.StatusBadRequest,
				dataplanesResponseBody:   `error`,
			}),
			expected: createExpected{
				err:            expectedInvlidDataplaneErr,
				displays:       []string{stubs.ExpectedDeviceAuthDisplay},
				dataplaneCalls: 1,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer viper.Reset()
			t.Setenv(osutil.HomeEnvVar(), t.TempDir())

			srv, hitCounts := stubs.NewStubHTTPServer(t, tt.mockEndpoints)
			defer srv.Close()

			fsys := afero.NewMemMapFs()
			if tt.input.filename != "" {
				require.NoError(t, afero.WriteFile(fsys, tt.input.filename, []byte(tt.input.fileContent), 0o600))
			}
			if tt.input.specFile != "" {
				require.NoError(t, afero.WriteFile(fsys, tt.input.specFile, []byte(tt.input.fileContent), 0o600))
			}

			stubUI := &stubs.StubUI{}
			if tt.input.endpoint == "" {
				tt.input.endpoint = srv.URL
			}
			mockCmd := image.CreateCommand(context.Background(), fsys, stubUI)
			cmd.AddRootFlags(mockCmd)
			mockCmd.SetArgs(getCreateArgs(tt))
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
			require.Equal(t, tt.expected.createCalls,
				hitCounts.Count("POST"+postgrest.Endpoint(constants.WorkbenchEndpointRPCImageCreate)),
				"create calls")
			require.Equal(t, tt.expected.dataplaneCalls,
				hitCounts.Count("GET"+postgrest.Endpoint(constants.WorkbenchEndpointDataplanes)),
				"dataplane calls")
		})
	}
}

func getCreateArgs(tt createImageTestCase) []string {
	args := []string{}
	args = append(args, "--"+constants.AuthModeFlag, constants.AuthModeDevice)
	args = append(args, "--"+constants.OrganizationFlag, "org-1")
	if tt.input.endpoint != "" {
		args = append(args, "--"+constants.EndpointFlag, tt.input.endpoint)
	}
	if tt.input.dryRun {
		args = append(args, "--"+constants.DryRunFlag)
	}
	if tt.input.filename != "" {
		args = append(args, "--"+constants.FilenameFlag, tt.input.filename)
	}
	if tt.input.spec != "" {
		args = append(args, "--"+constants.SpecFlag, tt.input.spec)
	}
	if tt.input.dataplane != "" {
		args = append(args, "--"+constants.DataplaneFlag, tt.input.dataplane)
	}
	if tt.input.name != "" {
		args = append(args, "--"+constants.NameFlag, tt.input.name)
	}
	if tt.input.description != "" {
		args = append(args, "--"+constants.DescriptionFlag, tt.input.description)
	}
	if tt.input.context != "" {
		args = append(args, "--"+constants.ContextFlag, tt.input.context)
	}
	if tt.input.image != "" {
		args = append(args, "--"+constants.ImageFlag, tt.input.image)
	}
	return args
}

type updateInput struct {
	endpoint    string
	filename    string
	fileContent string
	dryRun      bool
	name        string
	description string
	context     string
	image       string
	spec        string
	subPath     string
	labels      []string
	longArgs    []string
	args        []string
}

type updateExpected struct {
	displays    []string
	warns       []string
	err         string
	updateCalls int
}

type updateImageTestCase struct {
	name          string
	input         updateInput
	mockEndpoints []stubs.StubHTTPEndpoint
	expected      updateExpected
}

func TestUpdateImageCommand(t *testing.T) {
	tests := []updateImageTestCase{
		{
			name: "dry run with CLI ids and mapped fields",
			input: updateInput{
				dryRun:      true,
				description: "image update desc",
				context:     "build/src",
				image:       "registry.example.com/myimage:v2",
				spec:        `{"version":"3"}`,
				subPath:     "src",
				args:        []string{"img-cli-1", "img-cli-2"},
			},
			mockEndpoints: stubs.GetStubAuthEndpoints(),
			expected: updateExpected{
				displays: []string{stubs.ExpectedDeviceAuthDisplay, expectedUpdateImageDryRunOutput},
			},
		},
		{
			name: "dry run long flags include all mapped fields",
			input: updateInput{
				longArgs: []string{
					"--" + constants.DryRunFlag,
					"--" + constants.NameFlag, "img-name",
					"--" + constants.DescriptionFlag, "image update desc",
					"--" + constants.ContextFlag, "build/src",
					"--" + constants.ImageFlag, "registry.example.com/myimage:v2",
					"--" + constants.SpecFlag, `{"version":"3"}`,
					"--" + constants.SubPathFlag, "src",
					"--" + constants.LabelFlag, "team=core",
				},
				args: []string{"img-cli-1"},
			},
			mockEndpoints: stubs.GetStubAuthEndpoints(),
			expected: updateExpected{
				displays: []string{stubs.ExpectedDeviceAuthDisplay, expectedUpdateImageLongFlagsDryRunOutput},
			},
		},
		{
			name:  "update success prints IDs only",
			input: updateInput{args: []string{"img-cli-1"}},
			mockEndpoints: getUpdateImageEndpoints(
				expectedUpdateImageRequestBodyIDsOnly,
				http.StatusOK,
				`[{"id":"img-a"},{"id":"img-b"}]`,
			),
			expected: updateExpected{
				displays:    []string{stubs.ExpectedDeviceAuthDisplay, "img-a", "img-b"},
				updateCalls: 1,
			},
		},
		{
			name:  "server error",
			input: updateInput{args: []string{"img-cli-1"}},
			mockEndpoints: getUpdateImageEndpoints(
				expectedUpdateImageRequestBodyIDsOnly,
				http.StatusUnauthorized,
				`unauthorized`,
			),
			expected: updateExpected{
				err:         expectedUpdateImageServerErr,
				displays:    []string{stubs.ExpectedDeviceAuthDisplay},
				updateCalls: 1,
			},
		},
		{
			name: "ids fallback from file body id",
			input: updateInput{
				filename:    "/update-image.json",
				fileContent: expectedUpdateImageFileContent,
				dryRun:      true,
			},
			mockEndpoints: stubs.GetStubAuthEndpoints(),
			expected: updateExpected{
				displays: []string{stubs.ExpectedDeviceAuthDisplay, expectedUpdateImageFileDryRunOutput},
			},
		},
		{
			name:          "error when no ID source",
			input:         updateInput{description: "image update desc"},
			mockEndpoints: stubs.GetStubAuthEndpoints(),
			expected: updateExpected{
				err:      expectedUpdateImageMissingIDErr,
				displays: []string{stubs.ExpectedDeviceAuthDisplay},
			},
		},
		{
			name:  "ids only with empty body accepted",
			input: updateInput{args: []string{"img-cli-1"}},
			mockEndpoints: getUpdateImageEndpoints(
				expectedUpdateImageRequestBodyIDsOnly,
				http.StatusOK,
				`[{"id":"img-cli-1"}]`,
			),
			expected: updateExpected{
				displays:    []string{stubs.ExpectedDeviceAuthDisplay, "img-cli-1"},
				updateCalls: 1,
			},
		},
		{
			name: "mapping path uses image and spec",
			input: updateInput{
				description: "image update desc",
				context:     "build/src",
				image:       "registry.example.com/myimage:v2",
				spec:        `{"version":"3"}`,
				subPath:     "src",
				args:        []string{"img-cli-1", "img-cli-2"},
			},
			mockEndpoints: getUpdateImageEndpoints(
				expectedUpdateImageRequestBodyMapping,
				http.StatusOK,
				`[{"id":"img-cli-1"},{"id":"img-cli-2"}]`,
			),
			expected: updateExpected{
				displays:    []string{stubs.ExpectedDeviceAuthDisplay, "img-cli-1", "img-cli-2"},
				updateCalls: 1,
			},
		},
		{
			name: "mapping path uses remaining update flags",
			input: updateInput{
				name:        "img-name",
				description: "image update desc",
				context:     "build/src",
				image:       "registry.example.com/myimage:v2",
				spec:        `{"version":"3"}`,
				subPath:     "src",
				labels:      []string{"team=core"},
				args:        []string{"img-cli-1"},
			},
			mockEndpoints: getUpdateImageEndpoints(
				expectedUpdateImageRequestBodyAllFlags,
				http.StatusOK,
				`[{"id":"img-cli-1"}]`,
			),
			expected: updateExpected{
				displays:    []string{stubs.ExpectedDeviceAuthDisplay, "img-cli-1"},
				updateCalls: 1,
			},
		},
		{
			name: "invalid endpoint",
			input: updateInput{
				endpoint: "a\nb",
			},
			expected: updateExpected{
				err: expectedInvalidEndpointErr,
			},
		},
		{
			name: "invalid spec",
			input: updateInput{
				name:    "test",
				context: "test",
				spec:    "[invalid",
			},
			mockEndpoints: stubs.GetStubAuthEndpoints(),
			expected: updateExpected{
				err:      expectedInvalidSpecErr,
				displays: []string{stubs.ExpectedDeviceAuthDisplay},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer viper.Reset()
			t.Setenv(osutil.HomeEnvVar(), t.TempDir())

			srv, hitCounts := stubs.NewStubHTTPServer(t, tt.mockEndpoints)
			defer srv.Close()

			fsys := afero.NewMemMapFs()
			if tt.input.filename != "" {
				require.NoError(t, afero.WriteFile(fsys, tt.input.filename, []byte(tt.input.fileContent), 0o600))
			}

			stubUI := &stubs.StubUI{}
			if tt.input.endpoint == "" {
				tt.input.endpoint = srv.URL
			}
			mockCmd := image.UpdateCommand(context.Background(), fsys, stubUI)
			cmd.AddRootFlags(mockCmd)
			mockCmd.SetArgs(getUpdateArgs(tt))
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
			require.Equal(
				t,
				tt.expected.updateCalls,
				hitCounts.Count("POST"+postgrest.Endpoint(constants.WorkbenchEndpointRPCImageBulkUpdate)),
				"update calls",
			)
		})
	}
}

type imageEndpointsConfig struct {
	expectedDataplanesQuery        string
	dataplanesResponseStatus       int
	dataplanesResponseBody         string
	expectedImageCreateRequestBody string
	imageCreateResponseStatus      int
	imageCreateResponseBody        string
}

func getCreateImageEndpoints(config imageEndpointsConfig) []stubs.StubHTTPEndpoint {
	return append(stubs.GetStubAuthEndpoints(),
		[]stubs.StubHTTPEndpoint{
			{
				Request: stubs.StubHTTPRequest{
					Method:          http.MethodGet,
					URL:             postgrest.Endpoint(constants.WorkbenchEndpointDataplanes),
					ExpectedHeaders: stubs.StubAuthHeaders,
					ExpectedQuery:   config.expectedDataplanesQuery,
				},
				Response: stubs.StubHTTPResponse{
					Status: config.dataplanesResponseStatus,
					Body:   config.dataplanesResponseBody,
				},
			},
			{
				Request: stubs.StubHTTPRequest{
					Method:          http.MethodPost,
					URL:             postgrest.Endpoint(constants.WorkbenchEndpointRPCImageCreate),
					ExpectedHeaders: stubs.StubAuthHeaders,
					ExpectedBody:    config.expectedImageCreateRequestBody,
				},
				Response: stubs.StubHTTPResponse{
					Status: config.imageCreateResponseStatus,
					Body:   config.imageCreateResponseBody,
				},
			},
		}...)
}

func getUpdateArgs(
	tt updateImageTestCase,
) []string {
	args := []string{}
	args = append(args, "--"+constants.AuthModeFlag, constants.AuthModeDevice)
	args = append(args, "--"+constants.OrganizationFlag, "org-1")
	if tt.input.endpoint != "" {
		args = append(args, "--"+constants.EndpointFlag, tt.input.endpoint)
	}
	if tt.input.dryRun {
		args = append(args, "--"+constants.DryRunFlag)
	}
	if tt.input.filename != "" {
		args = append(args, "--"+constants.FilenameFlag, tt.input.filename)
	}
	if tt.input.name != "" {
		args = append(args, "--"+constants.NameFlag, tt.input.name)
	}
	if tt.input.description != "" {
		args = append(args, "--"+constants.DescriptionFlag, tt.input.description)
	}
	if tt.input.context != "" {
		args = append(args, "--"+constants.ContextFlag, tt.input.context)
	}
	if tt.input.image != "" {
		args = append(args, "--"+constants.ImageFlag, tt.input.image)
	}
	if tt.input.spec != "" {
		args = append(args, "--"+constants.SpecFlag, tt.input.spec)
	}
	if tt.input.subPath != "" {
		args = append(args, "--"+constants.SubPathFlag, tt.input.subPath)
	}
	for _, label := range tt.input.labels {
		args = append(args, "--"+constants.LabelFlag, label)
	}
	args = append(args, tt.input.longArgs...)
	args = append(args, tt.input.args...)
	return args
}

func getUpdateImageEndpoints(
	expectedRequestBody string,
	responseStatus int,
	responseBody string,
) []stubs.StubHTTPEndpoint {
	return append(stubs.GetStubAuthEndpoints(),
		[]stubs.StubHTTPEndpoint{
			{
				Request: stubs.StubHTTPRequest{
					Method:          http.MethodPost,
					URL:             postgrest.Endpoint(constants.WorkbenchEndpointRPCImageBulkUpdate),
					ExpectedHeaders: stubs.StubAuthHeaders,
					ExpectedBody:    expectedRequestBody,
				},
				Response: stubs.StubHTTPResponse{
					Status: responseStatus,
					Body:   responseBody,
				},
			},
		}...)
}
