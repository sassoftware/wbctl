// Copyright © 2026, SAS Institute Inc., Cary, NC, USA.  All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package workbench_test

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/sas-institute-rnd-saas/wbctl/internal/constants"
	"github.com/sas-institute-rnd-saas/wbctl/internal/postgrest"
	"github.com/sas-institute-rnd-saas/wbctl/internal/services/workbench"
	"github.com/sas-institute-rnd-saas/wbctl/internal/stubs"
	"github.com/sas-institute-rnd-saas/wbctl/internal/timeutil"
	"github.com/sas-institute-rnd-saas/wbctl/internal/types"
	"github.com/stretchr/testify/require"
)

const (
	imagesEmptyResourcesJSON = `[]`

	imagesGetOneResponseJSON = `
[
  {
    "id": "img-1",
    "name": "image-1",
    "image_name": "myapp",
    "image_registry": "registry.example.com",
    "description": "My Application Image"
  }
]
`

	imagesGetTwoResponseJSON = `
[
  {
    "id": "123e4567-e89b-12d3-a456-426614174000",
    "name": "app-uuid",
    "image_name": "app-uuid",
    "image_registry": "registry.uuid.com",
    "controller_id": "ctrl-uuid"
  },
  {
    "id": "img-controller",
    "name": "app-controller",
    "image_name": "app-controller",
    "image_registry": "registry.controller.com",
    "controller_id": "controller-app-2"
  }
]
`

	imagesPrintOneResourcesJSON = `
[
  {
    "id": "img-1",
    "name": "image-1",
    "description": "First image",
    "image_name": "myapp",
    "image_registry": "registry.example.com",
    "image_tag": "v1.0",
    "controller_id": "ctrl-1",
    "latest_image_build_id": "ib-1",
    "current_state": "Ready",
    "message": "ok",
    "created_at": null,
    "username": "alice",
    "email": "alice@example.com"
  }
]
`

	imagesPrintAdminResourcesJSON = `
[
  {
    "id": "img-1",
    "name": "image-1",
    "description": "First image",
    "image_name": "myapp",
    "image_registry": "registry.example.com",
    "image_tag": "v1.0",
    "controller_id": "ctrl-1",
    "latest_image_build_id": "ib-1",
    "current_state": "Ready",
    "message": "ok",
    "created_at": null,
    "username": "alice",
    "email": "alice@example.com"
  },
  {
    "id": "img-2",
    "name": "image-2",
    "description": "Second image",
    "image_name": "otherapp",
    "image_registry": "registry.example.com",
    "latest_image_build_id": "ib-2",
    "current_state": "Building",
    "message": "in progress",
    "created_at": null,
    "username": "bob",
    "email": "bob@example.com"
  }
]
`

	imagesPrintNoTagResourcesJSON = `
[
  {
    "id": "img-1",
    "name": "image-1",
    "description": "First image",
    "image_name": "myapp",
    "image_registry": "registry.example.com",
    "controller_id": "",
    "latest_image_build_id": "ib-1",
    "current_state": "Ready",
    "message": "ok",
    "created_at": null,
    "username": "alice",
    "email": "alice@example.com"
  }
]
`

	imagesDescribeOneResourcesJSON = `
[
  {
    "id": "img-1",
    "name": "image-1",
    "description": "My Application Image",
    "image_name": "myapp",
    "image_registry": "registry.example.com",
    "image_tag": "v1.0",
    "controller_id": "ctrl-1",
    "latest_image_build_id": "ib-1",
    "current_state": "Ready",
    "message": "image ready",
    "created_by": "alice",
    "email": "alice@example.com",
    "created_at": "2024-01-02T03:04:05Z",
    "modified_at": "2024-01-02T04:05:06Z",
    "username": "alice"
  }
]
`

	imagesDescribeTwoResourcesJSON = `
[
  {
    "id": "img-1",
    "name": "image-1",
    "description": "My Application Image",
    "image_name": "myapp",
    "image_registry": "registry.example.com",
    "image_tag": "v1.0",
    "controller_id": "ctrl-1",
    "latest_image_build_id": "ib-1",
    "current_state": "Ready",
    "message": "image ready",
    "created_by": "alice",
    "email": "alice@example.com",
    "created_at": "2024-01-02T03:04:05Z",
    "modified_at": "2024-01-02T04:05:06Z",
    "username": "alice"
  },
  {
    "id": "img-2",
    "name": "image-2",
    "image_name": "otherapp",
    "image_registry": "registry.example.com",
    "controller_id": "ctrl-2",
    "current_state": "Building",
    "message": "in progress",
    "created_by": "bob",
    "email": "bob@example.com",
    "created_at": "2024-01-03T03:04:05Z",
    "modified_at": "2024-01-03T04:05:06Z",
    "username": "bob"
  }
]
`

	imagesDescribeWithLabelsResourcesJSON = `
[
  {
    "id": "img-1",
    "name": "image-1",
    "image_name": "myapp",
    "image_registry": "registry.example.com",
    "controller_id": "ctrl-1",
    "created_by": "alice",
    "email": "alice@example.com",
    "created_at": "2024-01-02T03:04:05Z",
    "modified_at": "2024-01-02T04:05:06Z",
    "username": "alice",
    "labels": {
      "app": "myapp",
      "version": "v1.0",
      "team": "platform"
    }
  }
]
`

	imagesDescribeWithPodStatusResourcesJSON = `
[
  {
    "id": "img-2",
    "name": "image-2",
    "description": "My Application Image",
    "image_name": "myapp",
    "image_registry": "registry.example.com",
    "image_tag": "v1.0",
    "controller_id": "ctrl-2",
    "latest_image_build_id": "ib-1",
    "current_state": "Ready",
    "message": "image ready",
    "created_by": "bob",
    "email": "alice@example.com",
    "created_at": "2024-01-02T03:04:05Z",
    "modified_at": "2024-01-02T04:05:06Z",
    "username": "alice",
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

	imagesDeleteSuccessResponseJSON = `["img-1", "img-2"]`

	expectedImagesEmptyTableNonAdmin = "" +
		"ID\tCONTROLLER_ID\tNAME\tDESCRIPTION\tIMAGE\tIMAGE_BUILD_ID\tSTATE\tMESSAGE\tAGE"

	expectedImagesEmptyTableAdmin = "" +
		"ID\tCONTROLLER_ID\tNAME\tUSERNAME\tDESCRIPTION\tIMAGE\tIMAGE_BUILD_ID\tSTATE\tMESSAGE\tAGE"

	expectedImagesOneRowTableNonAdmin = "" +
		"ID\tCONTROLLER_ID\tNAME\tDESCRIPTION\tIMAGE\tIMAGE_BUILD_ID\tSTATE\tMESSAGE\tAGE\n" +
		"img-1\tctrl-1\timage-1\tFirst image\tregistry.example.com/myapp:v1.0\tib-1\tReady\tok\t"

	expectedImagesOneRowTableAdmin = "" +
		"ID\tCONTROLLER_ID\tNAME\tUSERNAME\tDESCRIPTION\tIMAGE\tIMAGE_BUILD_ID\tSTATE\tMESSAGE\tAGE\n" +
		"img-1\tctrl-1\timage-1\talice\tFirst image\tregistry.example.com/myapp:v1.0\tib-1\tReady\tok\t"

	expectedImagesTwoRowsTableAdmin = "" +
		"ID\tCONTROLLER_ID\tNAME\tUSERNAME\tDESCRIPTION\tIMAGE\tIMAGE_BUILD_ID\tSTATE\tMESSAGE\tAGE\n" +
		"img-1\tctrl-1\timage-1\talice\tFirst image\tregistry.example.com/myapp:v1.0\tib-1\tReady\tok\t\n" +
		"img-2\t<unknown>\timage-2\tbob\tSecond image\tregistry.example.com/otherapp\tib-2\tBuilding\tin progress\t"

	expectedImagesOneRowTableNoControllerID = "" +
		"ID\tCONTROLLER_ID\tNAME\tDESCRIPTION\tIMAGE\tIMAGE_BUILD_ID\tSTATE\tMESSAGE\tAGE\n" +
		"img-1\t<unknown>\timage-1\tFirst image\tregistry.example.com/myapp\tib-1\tReady\tok\t"

	expectedDescribeImageOneDisplay = `ID:                      img-1
Controller ID:           ctrl-1
Name:                    image-1
Description:             My Application Image
Created By:              alice (alice@example.com)
Created At:              2024-01-02T03:04:05Z (2y126d ago)
Modified At:             2024-01-02T04:05:06Z (2y126d ago)
Image:                   registry.example.com/myapp:v1.0
Latest Image Build ID:   ib-1
State                    Ready
Message                  image ready
`

	expectedDescribeImageMultipleDisplay = `ID:                      img-1
Controller ID:           ctrl-1
Name:                    image-1
Description:             My Application Image
Created By:              alice (alice@example.com)
Created At:              2024-01-02T03:04:05Z (2y126d ago)
Modified At:             2024-01-02T04:05:06Z (2y126d ago)
Image:                   registry.example.com/myapp:v1.0
Latest Image Build ID:   ib-1
State                    Ready
Message                  image ready


ID:              img-2
Controller ID:   ctrl-2
Name:            image-2
Created By:      bob (bob@example.com)
Created At:      2024-01-03T03:04:05Z (2y125d ago)
Modified At:     2024-01-03T04:05:06Z (2y125d ago)
Image:           registry.example.com/otherapp
State            Building
Message          in progress
`

	expectedDescribeImageWithLabelsLabelsDisplay = `ID:              img-1
Controller ID:   ctrl-1
Name:            image-1
Created By:      alice (alice@example.com)
Created At:      2024-01-02T03:04:05Z (2y126d ago)
Modified At:     2024-01-02T04:05:06Z (2y126d ago)
Image:           registry.example.com/myapp
Labels:
  app:       myapp
  team:      platform
  version:   v1.0
`

	expectedDescribeImageWithPodStatusPodDisplay = `ID:                      img-2
Controller ID:           ctrl-2
Name:                    image-2
Description:             My Application Image
Created By:              alice (alice@example.com)
Created At:              2024-01-02T03:04:05Z (2y126d ago)
Modified At:             2024-01-02T04:05:06Z (2y126d ago)
Image:                   registry.example.com/myapp:v1.0
Latest Image Build ID:   ib-1
State                    Ready
Message                  image ready
Pod pod-1:
  Phase:            Running
  Message:          pod-ok
  Reason:           scheduled
  Nominated node:   node-a
  IPs:              10.0.0.1
`

	imagesBulkUpdateResponseJSON = `
[
  {
    "id": "img-1",
    "name": "img-one-updated"
  },
  {
    "id": "img-2",
    "name": "img-two-updated"
  }
]
`
	expectedDeleteImageRequestJSON     = `{"ids":["img-1","img-2"]}`
	expectedCreateImageRequestJSON     = `{"param":{"name":"image-1"}}`
	createImageOneResponseBodyJSON     = `{"id":"img-1","name":"image-1"}`
	expectedRebuildImageRequestJSON    = `{"id":"img-1"}`
	expectedBulkUpdateImageRequestJSON = `{"ids":["img-1","img-2"],"param":{"name":"updated-name"}}`
)

var (
	expectedResponseOneImage = types.Images{
		{
			ID:            "img-1",
			Name:          "image-1",
			ImageName:     "myapp",
			ImageRegistry: "registry.example.com",
			Description:   "My Application Image",
		},
	}

	expectedResponseTwoImages = types.Images{
		{
			ID:            "123e4567-e89b-12d3-a456-426614174000",
			Name:          "app-uuid",
			ImageName:     "app-uuid",
			ImageRegistry: "registry.uuid.com",
			ControllerID:  "ctrl-uuid",
		},
		{
			ID:            "img-controller",
			Name:          "app-controller",
			ImageName:     "app-controller",
			ImageRegistry: "registry.controller.com",
			ControllerID:  "controller-app-2",
		},
	}
)

func TestDeleteImage(t *testing.T) {
	tests := []deleteTestCase{
		{
			name: "delete image success",
			request: deleteRequest{
				ids: []string{"img-1", "img-2"},
			},
			mockEndpoints: getImagesDeleteEndpoints(
				expectedDeleteImageRequestJSON,
				http.StatusOK,
				imagesDeleteSuccessResponseJSON,
			),
			expected: deleteExpected{
				result: []string{"img-1", "img-2"},
			},
		},
		{
			name: "delete image non-200 from server",
			request: deleteRequest{
				ids: []string{"img-1", "img-2"},
			},
			mockEndpoints: getImagesDeleteEndpoints(
				expectedDeleteImageRequestJSON,
				http.StatusUnauthorized,
				`unauthorized`,
			),
			expected: deleteExpected{
				err: "POST ${URL}/postgrest/rpc/image_bulk_delete failed\n" +
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

			actual, err := workbench.BulkDeleteImages(context.Background(), svc, tt.request.ids)
			if tt.expected.err != "" {
				require.EqualError(t, err, strings.ReplaceAll(tt.expected.err, "${URL}", srv.URL))
				return
			}

			require.NoError(t, err)
			require.Equal(t, tt.expected.result, actual)
		})
	}
}

func TestCreateImage(t *testing.T) {
	tests := []createTestCase[types.Image]{
		{
			name: "create image success",
			request: createRequest[types.Image]{
				body: types.Image{
					Name: "image-1",
				},
			},
			mockEndpoints: getImagesCreateEndpoints(
				expectedCreateImageRequestJSON,
				http.StatusOK,
				createImageOneResponseBodyJSON,
			),
			expected: createExpected[types.Image]{
				result: types.Image{
					ID:   "img-1",
					Name: "image-1",
				},
			},
		},
		{
			name: "create image non-200 from server",
			request: createRequest[types.Image]{
				body: types.Image{
					Name: "image-1",
				},
			},
			mockEndpoints: getImagesCreateEndpoints(
				expectedCreateImageRequestJSON,
				http.StatusUnauthorized,
				`unauthorized`,
			),
			expected: createExpected[types.Image]{
				err: "POST ${URL}/postgrest/rpc/image_create failed\n" +
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

			actual, err := workbench.CreateImage(context.Background(), svc, tt.request.body)
			if tt.expected.err != "" {
				require.EqualError(t, err, strings.ReplaceAll(tt.expected.err, "${URL}", srv.URL))
				return
			}

			require.NoError(t, err)
			require.Equal(t, tt.expected.result, actual)
		})
	}
}

func TestRebuildImage(t *testing.T) {
	tests := []rebuildTestCase{
		{
			name:          "rebuild image success",
			request:       rebuildRequest{rebuildID: "img-1"},
			mockEndpoints: getImagesRebuildEndpoints(expectedRebuildImageRequestJSON, http.StatusOK, `true`),
			expected:      rebuildExpected{result: true},
		},
		{
			name:          "rebuild image non-200 from server",
			request:       rebuildRequest{rebuildID: "img-1"},
			mockEndpoints: getImagesRebuildEndpoints(expectedRebuildImageRequestJSON, http.StatusUnauthorized, `unauthorized`),
			expected: rebuildExpected{err: "POST ${URL}/postgrest/rpc/image_rebuild failed\n" +
				" Caused by:\n" +
				"\t* 401 unauthorized"},
		},
		{
			name:          "rebuild image invalid json response",
			request:       rebuildRequest{rebuildID: "img-1"},
			mockEndpoints: getImagesRebuildEndpoints(expectedRebuildImageRequestJSON, http.StatusOK, `[`),
			expected: rebuildExpected{err: "POST ${URL}/postgrest/rpc/image_rebuild failed\n" +
				" Caused by:\n" +
				"\t* unexpected EOF"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv, _ := stubs.NewStubHTTPServer(t, tt.mockEndpoints)
			defer srv.Close()
			svc := workbench.NewWorkbenchService(srv.URL, "access-token", "csrf-token")

			actual, err := workbench.RebuildImage(context.Background(), svc, tt.request.rebuildID)
			if tt.expected.err != "" {
				require.EqualError(t, err, strings.ReplaceAll(tt.expected.err, "${URL}", srv.URL))
				return
			}

			require.NoError(t, err)
			require.Equal(t, tt.expected.result, actual)
		})
	}
}

func TestBulkUpdateImages(t *testing.T) {
	tests := []bulkupdateTestCase[types.Image, types.Images]{
		{
			name: "bulk update images success",
			request: bulkupdateRequest[types.Image]{
				ids:  []string{"img-1", "img-2"},
				body: types.Image{Name: "updated-name"},
			},
			mockEndpoints: getImagesBulkUpdateEndpoints(
				expectedBulkUpdateImageRequestJSON,
				http.StatusOK,
				imagesBulkUpdateResponseJSON,
			),
			expected: bulkupdateExpected[types.Images]{
				result: types.Images{
					{ID: "img-1", Name: "img-one-updated"},
					{ID: "img-2", Name: "img-two-updated"},
				}},
		},
		{
			name:    "bulk update images non-200 from server",
			request: bulkupdateRequest[types.Image]{ids: []string{"img-1", "img-2"}, body: types.Image{Name: "updated-name"}},
			mockEndpoints: getImagesBulkUpdateEndpoints(
				expectedBulkUpdateImageRequestJSON,
				http.StatusUnauthorized,
				`unauthorized`,
			),
			expected: bulkupdateExpected[types.Images]{
				err: "POST ${URL}/postgrest/rpc/image_bulk_update failed\n" +
					" Caused by:\n" +
					"\t* 401 unauthorized",
			},
		},
		{
			name:    "bulk update images invalid json response",
			request: bulkupdateRequest[types.Image]{ids: []string{"img-1", "img-2"}, body: types.Image{Name: "updated-name"}},
			mockEndpoints: getImagesBulkUpdateEndpoints(
				expectedBulkUpdateImageRequestJSON,
				http.StatusOK,
				`[`,
			),
			expected: bulkupdateExpected[types.Images]{
				err: "POST ${URL}/postgrest/rpc/image_bulk_update failed\n" +
					" Caused by:\n" +
					"\t* unexpected EOF",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv, _ := stubs.NewStubHTTPServer(t, tt.mockEndpoints)
			defer srv.Close()
			svc := workbench.NewWorkbenchService(srv.URL, "access-token", "csrf-token")

			actual, err := workbench.BulkUpdateImages(context.Background(), svc, tt.request.ids, tt.request.body)
			if tt.expected.err != "" {
				require.EqualError(t, err, strings.ReplaceAll(tt.expected.err, "${URL}", srv.URL))
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.expected.result, actual)
		})
	}
}

func TestDescribeImages(t *testing.T) {
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
			name:             "describe images describe empty resources prints nothing",
			resourcesJSON:    imagesEmptyResourcesJSON,
			expectedDisplays: []string{""},
		},
		{
			name:             "describe images describe prints one image",
			resourcesJSON:    imagesDescribeOneResourcesJSON,
			expectedDisplays: []string{expectedDescribeImageOneDisplay},
		},
		{
			name:             "describe images describe prints two images",
			resourcesJSON:    imagesDescribeTwoResourcesJSON,
			expectedDisplays: []string{expectedDescribeImageMultipleDisplay},
		},
		{
			name:             "describe images describe prints image with labels",
			resourcesJSON:    imagesDescribeWithLabelsResourcesJSON,
			expectedDisplays: []string{expectedDescribeImageWithLabelsLabelsDisplay},
		},
		{
			name:             "describe images describe prints pod status details when present",
			resourcesJSON:    imagesDescribeWithPodStatusResourcesJSON,
			expectedDisplays: []string{expectedDescribeImageWithPodStatusPodDisplay},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var resources types.Images
			err := json.Unmarshal([]byte(tt.resourcesJSON), &resources)
			require.NoError(t, err)

			stubUI := &stubs.StubUI{}
			err = workbench.DescribeImages(stubUI, resources)
			require.NoError(t, err)
			require.Equal(t, tt.expectedDisplays, stubUI.Displays, strings.Join(stubUI.Displays, ""))
			require.Equal(t, tt.expectedWarns, stubUI.Warns, strings.Join(stubUI.Warns, ""))
		})
	}
}

func TestGetImages(t *testing.T) {
	tests := []getTestCase[types.Images]{
		{
			name:          "get images success without ids query",
			mockEndpoints: getImagesGetEndpoints("", http.StatusOK, imagesGetOneResponseJSON),
			expected:      getExpected[types.Images]{result: expectedResponseOneImage},
		},
		{
			name: "get images success with mixed uuid and controller ids query",
			request: getRequest{
				ids:  []string{"123e4567-e89b-12d3-a456-426614174000", "controller-app-2"},
				sort: "id.asc",
			},
			mockEndpoints: getImagesGetEndpoints(
				"or=(id.in.(123e4567-e89b-12d3-a456-426614174000),controller_id.in.(controller-app-2))&order=id.asc",
				http.StatusOK,
				imagesGetTwoResponseJSON,
			),
			expected: getExpected[types.Images]{
				result: expectedResponseTwoImages,
			},
		},
		{
			name:          "get images sort by controller_id",
			request:       getRequest{sort: "controller_id"},
			mockEndpoints: getImagesGetEndpoints("order=controller_id", http.StatusOK, imagesGetOneResponseJSON),
			expected:      getExpected[types.Images]{result: expectedResponseOneImage},
		},
		{
			name:          "get images sort by name",
			request:       getRequest{sort: "name"},
			mockEndpoints: getImagesGetEndpoints("order=name", http.StatusOK, imagesGetOneResponseJSON),
			expected:      getExpected[types.Images]{result: expectedResponseOneImage},
		},
		{
			name:          "get images sort by username",
			request:       getRequest{sort: "username"},
			mockEndpoints: getImagesGetEndpoints("order=username", http.StatusOK, imagesGetOneResponseJSON),
			expected:      getExpected[types.Images]{result: expectedResponseOneImage},
		},
		{
			name:          "get images sort by description",
			request:       getRequest{sort: "description"},
			mockEndpoints: getImagesGetEndpoints("order=description", http.StatusOK, imagesGetOneResponseJSON),
			expected:      getExpected[types.Images]{result: expectedResponseOneImage},
		},
		{
			name:    "get images sort by image",
			request: getRequest{sort: "image.asc"},
			mockEndpoints: getImagesGetEndpoints(
				"order=image_registry.asc,image_name.asc,image_tag.asc",
				http.StatusOK,
				imagesGetOneResponseJSON,
			),
			expected: getExpected[types.Images]{
				result: expectedResponseOneImage,
			},
		},
		{
			name:    "get images sort by image and age",
			request: getRequest{sort: "image.asc,age.desc"},
			mockEndpoints: getImagesGetEndpoints(
				"order=image_registry.asc,image_name.asc,image_tag.asc,created_at.desc",
				http.StatusOK,
				imagesGetOneResponseJSON,
			),
			expected: getExpected[types.Images]{
				result: expectedResponseOneImage,
			},
		},
		{
			name:          "get images sort by image_build_id",
			request:       getRequest{sort: "image_build_id"},
			mockEndpoints: getImagesGetEndpoints("order=latest_image_build_id", http.StatusOK, imagesGetOneResponseJSON),
			expected: getExpected[types.Images]{
				result: expectedResponseOneImage,
			},
		},
		{
			name:          "get images sort by state",
			request:       getRequest{sort: "state"},
			mockEndpoints: getImagesGetEndpoints("order=current_state", http.StatusOK, imagesGetOneResponseJSON),
			expected:      getExpected[types.Images]{result: expectedResponseOneImage},
		},
		{
			name:          "get images sort by message",
			request:       getRequest{sort: "message"},
			mockEndpoints: getImagesGetEndpoints("order=message", http.StatusOK, imagesGetOneResponseJSON),
			expected:      getExpected[types.Images]{result: expectedResponseOneImage},
		},
		{
			name:          "get images sort by age",
			request:       getRequest{sort: "age"},
			mockEndpoints: getImagesGetEndpoints("order=created_at", http.StatusOK, imagesGetOneResponseJSON),
			expected:      getExpected[types.Images]{result: expectedResponseOneImage},
		},
		{
			name:          "get images invalid json response",
			request:       getRequest{ids: []string{"id-a"}},
			mockEndpoints: getImagesGetEndpoints("controller_id=in.(id-a)", http.StatusOK, "["),
			expected: getExpected[types.Images]{
				err: "GET ${URL}/postgrest/images?controller_id=in.%28id-a%29 failed\n" +
					" Caused by:\n" +
					"\t* unexpected EOF"},
		},
		{
			name:          "get images non-200 from server",
			request:       getRequest{ids: []string{"id-a"}},
			mockEndpoints: getImagesGetEndpoints("controller_id=in.(id-a)", http.StatusUnauthorized, "unauthorized"),
			expected: getExpected[types.Images]{
				err: "GET ${URL}/postgrest/images?controller_id=in.%28id-a%29 failed\n" +
					" Caused by:\n" +
					"\t* 401 unauthorized"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv, _ := stubs.NewStubHTTPServer(t, tt.mockEndpoints)
			defer srv.Close()
			svc := workbench.NewWorkbenchService(srv.URL, "access-token", "csrf-token")

			actual, err := workbench.GetImages(context.Background(), svc, tt.request.ids, tt.request.sort)
			if tt.expected.err != "" {
				require.EqualError(t, err, strings.ReplaceAll(tt.expected.err, "${URL}", srv.URL))
				return
			}

			require.NoError(t, err)
			require.Equal(t, tt.expected.result, actual)
		})
	}
}

func TestPrintImages(t *testing.T) {
	tests := []struct {
		name               string
		resourcesJSON      string
		token              string
		expectedTableCount int
		expectedTable      string
		expectedPrintCalls int
	}{
		{
			name:               "print images empty resources with non-admin token",
			resourcesJSON:      imagesEmptyResourcesJSON,
			token:              "token-without-admin-role",
			expectedTableCount: 1,
			expectedTable:      expectedImagesEmptyTableNonAdmin,
			expectedPrintCalls: 1,
		},
		{
			name:               "print images empty resources with admin token",
			resourcesJSON:      imagesEmptyResourcesJSON,
			token:              createTokenWithRole(constants.WorkbenchAdminRole),
			expectedTableCount: 1,
			expectedTable:      expectedImagesEmptyTableAdmin,
			expectedPrintCalls: 1,
		},
		{
			name:               "print images prints one resource with tag and controller as non-admin",
			resourcesJSON:      imagesPrintOneResourcesJSON,
			token:              "token-without-admin-role",
			expectedTableCount: 1,
			expectedTable:      expectedImagesOneRowTableNonAdmin, expectedPrintCalls: 1,
		},
		{
			name:               "print images prints one resource with tag and controller as admin",
			resourcesJSON:      imagesPrintOneResourcesJSON,
			token:              createTokenWithRole(constants.WorkbenchAdminRole),
			expectedTableCount: 1,
			expectedTable:      expectedImagesOneRowTableAdmin,
			expectedPrintCalls: 1,
		},
		{
			name:               "print images prints multiple resources as admin",
			resourcesJSON:      imagesPrintAdminResourcesJSON,
			token:              createTokenWithRole(constants.WorkbenchAdminRole),
			expectedTableCount: 1,
			expectedTable:      expectedImagesTwoRowsTableAdmin,
			expectedPrintCalls: 1,
		},
		{
			name:               "print images prints resource without image tag, coalesces missing controller id",
			resourcesJSON:      imagesPrintNoTagResourcesJSON,
			token:              "token-without-admin-role",
			expectedTableCount: 1,
			expectedTable:      expectedImagesOneRowTableNoControllerID,
			expectedPrintCalls: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var resources types.Images
			err := json.Unmarshal([]byte(tt.resourcesJSON), &resources)
			require.NoError(t, err)

			u := &stubs.StubUI{}
			err = workbench.PrintImages(u, tt.token, resources)
			require.NoError(t, err)

			require.Len(t, u.Tables, tt.expectedTableCount)
			tbl := u.Tables[0]
			require.Equal(t, tt.expectedTable, tbl.String())
			require.Equal(t, tt.expectedPrintCalls, tbl.PrintCalls)
		})
	}
}

// createTokenWithRole creates a valid JWT token with the specified role for
// testing. This function simulates different admin/non-admin scenarios for
// testing role-based permissions.
// nolint:unparam
func createTokenWithRole(role string) string {
	roles := []string{"test-user"}
	if role == constants.WorkbenchAdminRole {
		roles = append(roles, constants.WorkbenchAdminRole)
	}
	payload := map[string]any{
		"sas": map[string]any{
			"roles": roles,
		},
	}
	b, _ := json.Marshal(payload) //nolint:errchkjson
	enc := base64.StdEncoding.EncodeToString(b)
	return "h." + enc + ".s"
}

func getImagesDeleteEndpoints(
	expectedBody string,
	responseStatus int,
	responseBody string,
) []stubs.StubHTTPEndpoint {
	return []stubs.StubHTTPEndpoint{
		{
			Request: stubs.StubHTTPRequest{
				Method:          http.MethodPost,
				URL:             postgrest.Endpoint(constants.WorkbenchEndpointRPCImageBulkDelete),
				ExpectedHeaders: authHeaders,
				ExpectedBody:    expectedBody,
			},
			Response: stubs.StubHTTPResponse{
				Status: responseStatus,
				Body:   responseBody,
			},
		},
	}
}

func getImagesCreateEndpoints(
	expectedBody string,
	responseStatus int,
	responseBody string,
) []stubs.StubHTTPEndpoint {
	return []stubs.StubHTTPEndpoint{
		{
			Request: stubs.StubHTTPRequest{
				Method:          http.MethodPost,
				URL:             postgrest.Endpoint(constants.WorkbenchEndpointRPCImageCreate),
				ExpectedHeaders: authHeaders,
				ExpectedBody:    expectedBody,
			},
			Response: stubs.StubHTTPResponse{
				Status: responseStatus,
				Body:   responseBody,
			},
		},
	}
}

func getImagesBulkUpdateEndpoints(
	expectedBody string,
	responseStatus int,
	responseBody string,
) []stubs.StubHTTPEndpoint {
	return []stubs.StubHTTPEndpoint{
		{
			Request: stubs.StubHTTPRequest{
				Method:          http.MethodPost,
				URL:             postgrest.Endpoint(constants.WorkbenchEndpointRPCImageBulkUpdate),
				ExpectedHeaders: authHeaders,
				ExpectedBody:    expectedBody,
			},
			Response: stubs.StubHTTPResponse{
				Status: responseStatus,
				Body:   responseBody,
			},
		},
	}
}

func getImagesGetEndpoints(
	expectedQuery string,
	responseStatus int,
	responseBody string,
) []stubs.StubHTTPEndpoint {
	return []stubs.StubHTTPEndpoint{
		{
			Request: stubs.StubHTTPRequest{
				Method:          http.MethodGet,
				URL:             postgrest.Endpoint(constants.WorkbenchEndpointImages),
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

func getImagesRebuildEndpoints(
	expectedBody string,
	responseStatus int,
	responseBody string,
) []stubs.StubHTTPEndpoint {
	return []stubs.StubHTTPEndpoint{
		{
			Request: stubs.StubHTTPRequest{
				Method:          http.MethodPost,
				URL:             postgrest.Endpoint(constants.WorkbenchEndpointRPCImageRebuild),
				ExpectedHeaders: authHeaders,
				ExpectedBody:    expectedBody,
			},
			Response: stubs.StubHTTPResponse{
				Status: responseStatus,
				Body:   responseBody,
			},
		},
	}
}
