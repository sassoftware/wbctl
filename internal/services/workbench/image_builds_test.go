// Copyright © 2026, SAS Institute Inc., Cary, NC, USA.  All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package workbench_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/sassoftware/wbctl/internal/constants"
	"github.com/sassoftware/wbctl/internal/postgrest"
	"github.com/sassoftware/wbctl/internal/services/workbench"
	"github.com/sassoftware/wbctl/internal/stubs"
	"github.com/sassoftware/wbctl/internal/timeutil"
	"github.com/sassoftware/wbctl/internal/types"
	"github.com/stretchr/testify/require"
)

const (
	imageBuildsEmptyResourcesJSON = `[]`

	imageBuildsGetOneResponseJSON = `
[
  {
    "id": "ib-1",
    "name": "build-1",
    "controller_id": "ctrl-1",
    "created_by": "alice",
    "image_id": "img-1"
  }
]
`

	imageBuildsGetMixedIDsResponseJSON = `
[
  {
    "id": "123e4567-e89b-12d3-a456-426614174000",
    "name": "build-uuid",
    "controller_id": "ctrl-uuid",
    "created_by": "bob",
    "image_id": "img-uuid"
  },
  {
    "id": "ib-controller",
    "name": "build-controller",
    "controller_id": "controller-build-2",
    "created_by": "carol",
    "image_id": "img-controller"
  }
]
`

	imageBuildsPrintOneResourcesJSON = `
[
  {
    "id": "ib-1",
    "name": "build-1",
    "controller_id": "",
    "image_id": "img-1",
    "created_at": null
  }
]
`

	imageBuildsPrintTwoResourcesJSON = `
[
  {
	"id": "ib-1",
	"name": "build-1",
	"controller_id": "",
	"image_id": "img-1",
	"created_at": null
  },
  {
	"id": "ib-2",
	"name": "build-2",
	"controller_id": "ctrl-2",
	"image_id": "img-2",
	"created_at": null
  }
]
`

	imageBuildsDescribeOneResourcesJSON = `
[
  {
    "id": "ib-1",
    "name": "build-1",
    "controller_id": "ctrl-1",
    "created_by": "alice",
    "created_at": "2024-01-02T03:04:05Z",
    "image_id": "img-1"
  }
]
`
	imageBuildsDescribeOneResourcesWithStatusJSON = `
[
  {
    "id": "ib-2",
    "name": "build-2",
    "controller_id": "ctrl-2",
    "created_by": "bob",
    "created_at": "2024-01-02T03:04:05Z",
    "image_id": "img-2",
    "pod_status": [
      {
        "name": "pod-1",
        "status": {
          "phase": "Running",
          "message": "pod-ok",
          "reason": "scheduled",
          "nominatedNodeName": "node-a",
          "podIP": "10.0.0.1"
        }
      }
    ]
  }
]
`

	imageBuildsDescribeTwoResourcesJSON = `
[
  {
	"id": "ib-1",
	"name": "build-1",
	"controller_id": "ctrl-1",
	"created_by": "alice",
	"created_at": "2024-01-02T03:04:05Z",
	"image_id": "img-1"
  },
  {
	"id": "ib-2",
	"name": "build-2",
	"controller_id": "ctrl-2",
	"created_by": "bob",
	"created_at": "2024-01-03T03:04:05Z",
	"image_id": "img-2"
  }
]
`

	expectedImageBuildsEmptyTable = "ID\tCONTROLLER_ID\tNAME\tIMAGE_ID\tAGE"

	expectedImageBuildsOneRowTable = "ID\tCONTROLLER_ID\tNAME\tIMAGE_ID\tAGE\n" +
		"ib-1\t<unknown>\tbuild-1\timg-1\t"

	expectedImageBuildsTwoRowsTable = "ID\tCONTROLLER_ID\tNAME\tIMAGE_ID\tAGE\n" +
		"ib-1\t<unknown>\tbuild-1\timg-1\t\n" +
		"ib-2\tctrl-2\tbuild-2\timg-2\t"

	expectedDescribeImageBuildOneDisplay = `ID:              ib-1
Controller ID:   ctrl-1
Name:            build-1
Created By:      alice
Created At:      2024-01-02T03:04:05Z (2y126d ago)
Image ID:        img-1
`

	expectedDescribeImageBuildMultipleDisplay = `ID:              ib-1
Controller ID:   ctrl-1
Name:            build-1
Created By:      alice
Created At:      2024-01-02T03:04:05Z (2y126d ago)
Image ID:        img-1


ID:              ib-2
Controller ID:   ctrl-2
Name:            build-2
Created By:      bob
Created At:      2024-01-03T03:04:05Z (2y125d ago)
Image ID:        img-2
`

	expectedDescribeImageBuildWithPodStatusDisplay = `ID:              ib-2
Controller ID:   ctrl-2
Name:            build-2
Created By:      bob
Created At:      2024-01-02T03:04:05Z (2y126d ago)
Image ID:        img-2
Pod pod-1:
  Phase:            Running
  Message:          pod-ok
  Reason:           scheduled
  Nominated node:   node-a
  IPs:              10.0.0.1
`
)

var (
	expectedResponseOneImageBuild = types.ImageBuilds{
		{ID: "ib-1", Name: "build-1", ControllerID: "ctrl-1", CreatedBy: "alice", ImageID: "img-1"},
	}
	expectedResponseMixedIDsResponseJSON = types.ImageBuilds{
		{
			ID:           "123e4567-e89b-12d3-a456-426614174000",
			Name:         "build-uuid",
			ControllerID: "ctrl-uuid",
			CreatedBy:    "bob",
			ImageID:      "img-uuid",
		},
		{
			ID:           "ib-controller",
			Name:         "build-controller",
			ControllerID: "controller-build-2",
			CreatedBy:    "carol",
			ImageID:      "img-controller",
		},
	}
)

func TestGetImageBuilds(t *testing.T) {
	tests := []getTestCase[types.ImageBuilds]{
		{
			name:          "get image builds success without ids query",
			request:       getRequest{},
			mockEndpoints: getImageBuildsGetEndpoints("", http.StatusOK, imageBuildsGetOneResponseJSON),
			expected: getExpected[types.ImageBuilds]{
				result: expectedResponseOneImageBuild,
			},
		},
		{
			name: "get image builds success with mixed uuid and controller ids query",
			request: getRequest{
				ids:  []string{"123e4567-e89b-12d3-a456-426614174000", "controller-build-2"},
				sort: "id.asc",
			},
			mockEndpoints: getImageBuildsGetEndpoints(
				"or=(id.in.(123e4567-e89b-12d3-a456-426614174000),controller_id.in.(controller-build-2))&order=id.asc",
				http.StatusOK,
				imageBuildsGetMixedIDsResponseJSON,
			),
			expected: getExpected[types.ImageBuilds]{
				result: expectedResponseMixedIDsResponseJSON,
			},
		},
		{
			name: "get image builds sort by controller_id",
			request: getRequest{
				sort: "controller_id",
			},
			mockEndpoints: getImageBuildsGetEndpoints("order=controller_id", http.StatusOK, imageBuildsGetOneResponseJSON),
			expected: getExpected[types.ImageBuilds]{
				result: expectedResponseOneImageBuild,
			},
		},
		{
			name: "get image builds sort by name",
			request: getRequest{
				sort: "name",
			},
			mockEndpoints: getImageBuildsGetEndpoints("order=name", http.StatusOK, imageBuildsGetOneResponseJSON),
			expected: getExpected[types.ImageBuilds]{
				result: expectedResponseOneImageBuild,
			},
		},
		{
			name: "get image builds sort by image_id",
			request: getRequest{
				sort: "image_id",
			},
			mockEndpoints: getImageBuildsGetEndpoints("order=image_id", http.StatusOK, imageBuildsGetOneResponseJSON),
			expected: getExpected[types.ImageBuilds]{
				result: expectedResponseOneImageBuild,
			},
		},
		{
			name: "get image builds sort by age",
			request: getRequest{
				sort: "age",
			},
			mockEndpoints: getImageBuildsGetEndpoints("order=created_at", http.StatusOK, imageBuildsGetOneResponseJSON),
			expected: getExpected[types.ImageBuilds]{
				result: expectedResponseOneImageBuild,
			},
		},
		{
			name: "get image builds sort by age and name",
			request: getRequest{
				sort: "age.desc,name.asc",
			},
			mockEndpoints: getImageBuildsGetEndpoints(
				"order=created_at.desc,name.asc",
				http.StatusOK,
				imageBuildsGetOneResponseJSON,
			),
			expected: getExpected[types.ImageBuilds]{
				result: expectedResponseOneImageBuild,
			},
		},
		{
			name: "get image builds invalid json response",
			request: getRequest{
				ids: []string{"id-a"},
			},
			mockEndpoints: getImageBuildsGetEndpoints("controller_id=in.(id-a)", http.StatusOK, `[`),
			expected: getExpected[types.ImageBuilds]{
				err: "GET %s/postgrest/image_builds?controller_id=in.%%28id-a%%29 failed\n" +
					" Caused by:\n" +
					"\t* unexpected EOF",
			},
		},
		{
			name: "get image builds non-200 from server",
			request: getRequest{
				ids: []string{"id-a"},
			},
			mockEndpoints: getImageBuildsGetEndpoints("controller_id=in.(id-a)", http.StatusUnauthorized, `unauthorized`),
			expected: getExpected[types.ImageBuilds]{
				err: "GET %s/postgrest/image_builds?controller_id=in.%%28id-a%%29 failed\n" +
					" Caused by:\n" +
					"\t* 401 unauthorized",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv, _ := stubs.NewStubHTTPServer(t, tt.mockEndpoints)
			defer srv.Close()
			svc := workbench.NewWorkbenchService(srv.URL, "access-token", "csrf-token")

			actual, err := workbench.GetImageBuilds(context.Background(), svc, tt.request.ids, tt.request.sort)
			if tt.expected.err != "" {
				require.EqualError(t, err, fmt.Sprintf(tt.expected.err, srv.URL))
				return
			}

			require.NoError(t, err)
			require.Equal(t, tt.expected.result, actual)
		})
	}
}

func TestPrintImageBuilds(t *testing.T) {
	tests := []struct {
		name               string
		resourcesJSON      string
		expectedTableCount int
		expectedTable      string
		expectedPrintCalls int
	}{
		{
			name:               "print image builds empty resources",
			resourcesJSON:      imageBuildsEmptyResourcesJSON,
			expectedTableCount: 1,
			expectedTable:      expectedImageBuildsEmptyTable,
			expectedPrintCalls: 1,
		},
		{
			name:               "print image builds prints resources and coalesces missing controller id",
			resourcesJSON:      imageBuildsPrintOneResourcesJSON,
			expectedTableCount: 1,
			expectedTable:      expectedImageBuildsOneRowTable,
			expectedPrintCalls: 1,
		},
		{
			name:               "print image builds prints two resources",
			resourcesJSON:      imageBuildsPrintTwoResourcesJSON,
			expectedTableCount: 1,
			expectedTable:      expectedImageBuildsTwoRowsTable,
			expectedPrintCalls: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var resources types.ImageBuilds
			err := json.Unmarshal([]byte(tt.resourcesJSON), &resources)
			require.NoError(t, err)

			u := &stubs.StubUI{}
			err = workbench.PrintImageBuilds(u, "ignored-token", resources)
			require.NoError(t, err)

			require.Len(t, u.Tables, tt.expectedTableCount)
			tbl := u.Tables[0]
			require.Equal(t, tt.expectedTable, tbl.String())
			require.Equal(t, tt.expectedPrintCalls, tbl.PrintCalls)
		})
	}
}

func TestDescribeImageBuilds(t *testing.T) {
	fixedNow := time.Date(2026, time.May, 8, 0, 0, 0, 0, time.UTC)
	timeutil.SetNow(func() time.Time { return fixedNow })
	t.Cleanup(func() { timeutil.SetNow(time.Now) })

	tests := []struct {
		name             string
		resourcesJSON    string
		expectedDisplays []string
		expectedWarns    []string
	}{
		{
			name:             "describe image builds describe empty resources prints nothing",
			resourcesJSON:    imageBuildsEmptyResourcesJSON,
			expectedDisplays: []string{""},
		},
		{
			name:             "describe image builds describe prints one image build",
			resourcesJSON:    imageBuildsDescribeOneResourcesJSON,
			expectedDisplays: []string{expectedDescribeImageBuildOneDisplay},
		},
		{
			name:             "describe image builds describe prints two image builds",
			resourcesJSON:    imageBuildsDescribeTwoResourcesJSON,
			expectedDisplays: []string{expectedDescribeImageBuildMultipleDisplay},
		},
		{
			name:             "describe image builds describe prints pod status details when present",
			resourcesJSON:    imageBuildsDescribeOneResourcesWithStatusJSON,
			expectedDisplays: []string{expectedDescribeImageBuildWithPodStatusDisplay},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var resources types.ImageBuilds
			err := json.Unmarshal([]byte(tt.resourcesJSON), &resources)
			require.NoError(t, err)

			stubUI := &stubs.StubUI{}
			err = workbench.DescribeImageBuilds(stubUI, resources)
			require.NoError(t, err)
			require.Equal(t, tt.expectedDisplays, stubUI.Displays, strings.Join(stubUI.Displays, ""))
			require.Equal(t, tt.expectedWarns, stubUI.Warns, strings.Join(stubUI.Warns, ""))
		})
	}
}

func getImageBuildsGetEndpoints(
	expectedQuery string,
	responseStatus int,
	responseBody string,
) []stubs.StubHTTPEndpoint {
	return []stubs.StubHTTPEndpoint{
		{
			Request: stubs.StubHTTPRequest{
				Method:          http.MethodGet,
				URL:             postgrest.Endpoint(constants.WorkbenchEndpointImageBuilds),
				ExpectedHeaders: authHeaders,
				ExpectedQuery:   expectedQuery,
			},
			Response: stubs.StubHTTPResponse{
				Status: responseStatus,
				Body:   responseBody,
			},
		},
	}
}
