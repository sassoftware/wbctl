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
	snapshotsEmptyResourcesJSON = `[]`

	snapshotsGetOneResponseJSON = `
[
  {
    "id": "snap-1",
    "controller_id": "ctrl-1",
    "name": "snapshot-1",
    "username": "alice",
    "description": "first snapshot"
  }
]
`

	snapshotsGetMixedIDsResponseJSON = `
[
  {
    "id": "123e4567-e89b-12d3-a456-426614174000",
    "controller_id": "ctrl-uuid",
    "name": "snapshot-uuid",
    "username": "alice",
    "description": "uuid"
  },
  {
    "id": "snap-controller",
    "controller_id": "controller-snapshot-2",
    "name": "snapshot-controller",
    "username": "bob",
    "description": "controller"
  }
]
`

	snapshotsPrintOneResourcesJSON = `
[
  {
    "id": "snap-1",
    "controller_id": "",
    "name": "snapshot-1",
    "username": "alice",
    "description": "first snapshot",
    "created_at": null,
    "spec": {
      "source": {
        "persistentVolumeClaimName": "pvc-a"
      }
    },
    "volume_status": {
      "status": {
        "readyToUse": true,
        "restoreSize": "10Gi",
        "creationTime": "2026-05-08T00:00:00Z",
        "boundVolumeSnapshotContentName": "vsc-a"
      }
    }
  }
]
`

	snapshotsPrintTwoResourcesJSON = `
[
  {
    "id": "snap-1",
    "controller_id": "",
    "name": "snapshot-1",
    "username": "alice",
    "description": "first snapshot",
    "created_at": null,
    "spec": {
      "source": {
        "persistentVolumeClaimName": "pvc-a"
      }
    },
    "volume_status": {
      "status": {
        "readyToUse": true,
        "restoreSize": "10Gi",
        "creationTime": "2026-05-08T00:00:00Z",
        "boundVolumeSnapshotContentName": "vsc-a"
      }
    }
  },
  {
    "id": "snap-2",
    "controller_id": "ctrl-2",
    "name": "snapshot-2",
    "username": "bob",
    "description": "second snapshot",
    "created_at": null,
    "spec": {
      "source": {
        "persistentVolumeClaimName": "pvc-b"
      }
    },
    "volume_status": {
      "status": {
        "readyToUse": false,
        "creationTime": "2026-05-08T00:00:00Z",
        "boundVolumeSnapshotContentName": "vsc-b"
      }
    }
  }
]
`

	snapshotsDescribeOneResourcesJSON = `
[
  {
    "id": "snap-1",
    "controller_id": "ctrl-1",
    "name": "snapshot-1",
    "username": "alice",
    "email": "alice@example.com",
    "description": "first snapshot",
    "created_at": "2026-05-08T00:00:00Z",
    "modified_at": "2026-05-08T00:00:00Z",
    "message": "ok",
    "spec": {
      "source": {
        "persistentVolumeClaimName": "pvc-a"
      }
    },
    "properties": {
      "region": "us-east"
    },
    "volume_status": {
      "status": {
        "readyToUse": true,
        "restoreSize": "10Gi",
        "creationTime": "2026-05-08T00:00:00Z",
        "boundVolumeSnapshotContentName": "vsc-a"
      }
    }
  }
]
`

	snapshotsDescribeTwoResourcesJSON = `
[
{
    "id": "snap-1",
    "controller_id": "ctrl-1",
    "name": "snapshot-1",
    "username": "alice",
    "email": "alice@example.com",
    "description": "first snapshot",
    "created_at": "2026-05-08T00:00:00Z",
    "modified_at": "2026-05-08T00:00:00Z",
    "message": "ok",
    "spec": {
      "source": {
        "persistentVolumeClaimName": "pvc-a"
      }
    },
    "properties": {
      "region": "us-east"
    },
    "volume_status": {
      "status": {
        "readyToUse": true,
        "restoreSize": "10Gi",
        "creationTime": "2026-05-08T00:00:00Z",
        "boundVolumeSnapshotContentName": "vsc-a"
      }
    }
  },
  {
    "id": "snap-2",
    "controller_id": "ctrl-2",
    "name": "snapshot-2",
    "username": "bob",
    "email": "bob@example.com",
    "description": "second snapshot",
    "created_at": "2026-05-08T00:00:00Z",
    "modified_at": "2026-05-08T00:00:00Z",
    "message": "ok",
    "spec": {
      "source": {
        "persistentVolumeClaimName": "pvc-a"
      }
    },
    "properties": {
      "region": "us-east"
    },
    "volume_status": {
      "status": {
        "readyToUse": true,
        "restoreSize": "10Gi",
        "creationTime": "2026-05-08T00:00:00Z",
        "boundVolumeSnapshotContentName": "vsc-a"
      }
    }
  }
]
`

	snapshotsDescribeNilSpecResourcesJSON = `
[
  {
	"id": "snap-2",
	"controller_id": "ctrl-2",
	"name": "snapshot-2",
	"username": "bob",
	"email": "bob@example.com",
	"description": "no spec",
	"created_at": "2026-05-08T00:00:00Z",
	"modified_at": "2026-05-08T00:00:00Z"
  }
]
`

	snapshotsBulkUpdateResponseJSON = `
[
  {
    "id": "snap-1",
    "name": "snapshot-one-updated"
  },
  {
    "id": "snap-2",
    "name": "snapshot-two-updated"
  }
]
`

	expectedSnapshotsEmptyTable  = "ID\tCONTROLLER_ID\tNAME\tDESCRIPTION\tSOURCE\tSIZE\tREADY\tAGE"
	expectedSnapshotsOneRowTable = "ID\tCONTROLLER_ID\tNAME\tDESCRIPTION\tSOURCE\tSIZE\tREADY\tAGE\n" +
		"snap-1\t<unknown>\tsnapshot-1\tfirst snapshot\tpvc-a\t10Gi\ttrue\t"
	expectedSnapshotsTwoRowsTable = "ID\tCONTROLLER_ID\tNAME\tDESCRIPTION\tSOURCE\tSIZE\tREADY\tAGE\n" +
		"snap-1\t<unknown>\tsnapshot-1\tfirst snapshot\tpvc-a\t10Gi\ttrue\t\n" +
		"snap-2\tctrl-2\tsnapshot-2\tsecond snapshot\tpvc-b\t<unknown>\tfalse\t"
	expectedSnapshotsAdminEmptyTable  = "ID\tCONTROLLER_ID\tNAME\tUSERNAME\tDESCRIPTION\tSOURCE\tSIZE\tREADY\tAGE"
	expectedSnapshotsAdminOneRowTable = "ID\tCONTROLLER_ID\tNAME\tUSERNAME\tDESCRIPTION\tSOURCE\tSIZE\tREADY\tAGE\n" +
		"snap-1\t<unknown>\tsnapshot-1\talice\tfirst snapshot\tpvc-a\t10Gi\ttrue\t"

	expectedDescribeSnapshotOneDisplay = `ID:              snap-1
Controller ID:   ctrl-1
Name:            snapshot-1
Description:     first snapshot
Created By:      alice (alice@example.com)
Created At:      2026-05-08T00:00:00Z (0s ago)
Modified At:     2026-05-08T00:00:00Z (0s ago)
Spec:            {
                   "source": {
                     "persistentVolumeClaimName": "pvc-a"
                   }
                 }
Message:   ok
Properties:
  region:   us-east
Volume Status:
  Ready To Use:    true
  Restore Size:    10Gi
  Content Name:    vsc-a
  Creation Time:   2026-05-08T00:00:00Z (0s ago)
`
	expectedDescribeSnapshotTwoDisplay = `ID:              snap-1
Controller ID:   ctrl-1
Name:            snapshot-1
Description:     first snapshot
Created By:      alice (alice@example.com)
Created At:      2026-05-08T00:00:00Z (0s ago)
Modified At:     2026-05-08T00:00:00Z (0s ago)
Spec:            {
                   "source": {
                     "persistentVolumeClaimName": "pvc-a"
                   }
                 }
Message:   ok
Properties:
  region:   us-east
Volume Status:
  Ready To Use:    true
  Restore Size:    10Gi
  Content Name:    vsc-a
  Creation Time:   2026-05-08T00:00:00Z (0s ago)


ID:              snap-2
Controller ID:   ctrl-2
Name:            snapshot-2
Description:     second snapshot
Created By:      bob (bob@example.com)
Created At:      2026-05-08T00:00:00Z (0s ago)
Modified At:     2026-05-08T00:00:00Z (0s ago)
Spec:            {
                   "source": {
                     "persistentVolumeClaimName": "pvc-a"
                   }
                 }
Message:   ok
Properties:
  region:   us-east
Volume Status:
  Ready To Use:    true
  Restore Size:    10Gi
  Content Name:    vsc-a
  Creation Time:   2026-05-08T00:00:00Z (0s ago)
`
	expectedDescribeSnapshotNilSpecOneDisplay = `ID:              snap-2
Controller ID:   ctrl-2
Name:            snapshot-2
Description:     no spec
Created By:      bob (bob@example.com)
Created At:      2026-05-08T00:00:00Z (0s ago)
Modified At:     2026-05-08T00:00:00Z (0s ago)
Volume Status:
  Ready To Use:    false
  Restore Size:    <unknown>
  Content Name:    <unknown>
  Creation Time:   <unknown>
`
	expectedDeleteSnapshotRequestJSON     = `{"ids":["snap-1","snap-2"]}`
	createSnapshotOneResponseBodyJSON     = `{"id":"snap-1","name":"snapshot-1"}`
	expectedCreateSnapshotRequestJSON     = `{"param":{"name":"snapshot-1"}}`
	expectedBulkUpdateSnapshotRequestJSON = `{"ids":["snap-1","snap-2"],"param":{"name":"updated-name"}}`
)

var (
	expectedResponseOneSnapshot = types.Snapshots{
		{
			ID:           "snap-1",
			ControllerID: "ctrl-1",
			Name:         "snapshot-1",
			Username:     "alice",
			Description:  "first snapshot",
		},
	}

	expectedResponseTwoSnapshots = types.Snapshots{
		{
			ID:           "123e4567-e89b-12d3-a456-426614174000",
			ControllerID: "ctrl-uuid",
			Name:         "snapshot-uuid",
			Username:     "alice",
			Description:  "uuid",
		},
		{
			ID:           "snap-controller",
			ControllerID: "controller-snapshot-2",
			Name:         "snapshot-controller",
			Username:     "bob",
			Description:  "controller",
		},
	}
)

func TestDeleteSnapshot(t *testing.T) {
	tests := []deleteTestCase{
		{
			name: "delete snapshot success",
			request: deleteRequest{
				ids: []string{"snap-1", "snap-2"},
			},
			mockEndpoints: getSnapshotDeleteEndpoints(
				expectedDeleteSnapshotRequestJSON,
				http.StatusOK,
				`["snap-1","snap-2"]`,
			),
			expected: deleteExpected{
				result: []string{"snap-1", "snap-2"},
			},
		},
		{
			name: "delete snapshot non-200 from server",
			request: deleteRequest{
				ids: []string{"snap-1", "snap-2"},
			},
			mockEndpoints: getSnapshotDeleteEndpoints(
				expectedDeleteSnapshotRequestJSON,
				http.StatusUnauthorized,
				`unauthorized`,
			),
			expected: deleteExpected{
				err: "POST ${URL}/postgrest/rpc/snapshot_bulk_delete failed\n" +
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

			actual, err := workbench.BulkDeleteSnapshots(context.Background(), svc, tt.request.ids)
			if tt.expected.err != "" {
				require.EqualError(t, err, strings.ReplaceAll(tt.expected.err, "${URL}", srv.URL))
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.expected.result, actual)
		})
	}
}

func TestCreateSnapshot(t *testing.T) {
	tests := []createTestCase[types.Snapshot]{
		{
			name: "create snapshot success",
			request: createRequest[types.Snapshot]{
				body: types.Snapshot{
					Name: "snapshot-1",
				},
			},
			mockEndpoints: getSnapshotCreateEndpoints(
				expectedCreateSnapshotRequestJSON,
				http.StatusOK,
				createSnapshotOneResponseBodyJSON,
			),
			expected: createExpected[types.Snapshot]{
				result: types.Snapshot{
					ID:   "snap-1",
					Name: "snapshot-1",
				},
			},
		},
		{
			name: "create snapshot non-200 from server",
			request: createRequest[types.Snapshot]{
				body: types.Snapshot{
					Name: "snapshot-1",
				},
			},
			mockEndpoints: getSnapshotCreateEndpoints(
				expectedCreateSnapshotRequestJSON,
				http.StatusUnauthorized,
				`unauthorized`,
			),
			expected: createExpected[types.Snapshot]{
				err: "POST ${URL}/postgrest/rpc/snapshot_create failed\n" +
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

			actual, err := workbench.CreateSnapshot(context.Background(), svc, tt.request.body)
			if tt.expected.err != "" {
				require.EqualError(t, err, strings.ReplaceAll(tt.expected.err, "${URL}", srv.URL))
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.expected.result, actual)
		})
	}
}

func TestBulkUpdateSnapshots(t *testing.T) {
	tests := []bulkupdateTestCase[types.Snapshot, types.Snapshots]{
		{
			name: "bulk update snapshots success",
			request: bulkupdateRequest[types.Snapshot]{
				ids:  []string{"snap-1", "snap-2"},
				body: types.Snapshot{Name: "updated-name"},
			},
			mockEndpoints: getSnapshotBulkUpdateEndpoints(
				expectedBulkUpdateSnapshotRequestJSON,
				http.StatusOK,
				snapshotsBulkUpdateResponseJSON,
			),
			expected: bulkupdateExpected[types.Snapshots]{
				result: types.Snapshots{
					{ID: "snap-1", Name: "snapshot-one-updated"},
					{ID: "snap-2", Name: "snapshot-two-updated"},
				},
			},
		},
		{
			name: "bulk update snapshots non-200 from server",
			request: bulkupdateRequest[types.Snapshot]{
				ids:  []string{"snap-1", "snap-2"},
				body: types.Snapshot{Name: "updated-name"},
			},
			mockEndpoints: getSnapshotBulkUpdateEndpoints(
				expectedBulkUpdateSnapshotRequestJSON,
				http.StatusUnauthorized,
				`unauthorized`,
			),
			expected: bulkupdateExpected[types.Snapshots]{
				err: "POST ${URL}/postgrest/rpc/snapshot_bulk_update failed\n" +
					" Caused by:\n" +
					"\t* 401 unauthorized",
			},
		},
		{
			name: "bulk update snapshots invalid json response",
			request: bulkupdateRequest[types.Snapshot]{
				ids:  []string{"snap-1", "snap-2"},
				body: types.Snapshot{Name: "updated-name"},
			},
			mockEndpoints: getSnapshotBulkUpdateEndpoints(
				expectedBulkUpdateSnapshotRequestJSON,
				http.StatusOK,
				`[`,
			),
			expected: bulkupdateExpected[types.Snapshots]{
				err: "POST ${URL}/postgrest/rpc/snapshot_bulk_update failed\n" +
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

			actual, err := workbench.BulkUpdateSnapshots(context.Background(), svc, tt.request.ids, tt.request.body)
			if tt.expected.err != "" {
				require.EqualError(t, err, strings.ReplaceAll(tt.expected.err, "${URL}", srv.URL))
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.expected.result, actual)
		})
	}
}

func TestGetSnapshots(t *testing.T) {
	tests := []getTestCase[types.Snapshots]{
		{
			name:    "get snapshots success without ids query",
			request: getRequest{},
			mockEndpoints: getSnapshotsGetEndpoints(
				"",
				http.StatusOK,
				snapshotsGetOneResponseJSON,
			),
			expected: getExpected[types.Snapshots]{
				result: expectedResponseOneSnapshot,
			},
		},
		{
			name: "get snapshots success with mixed uuid and controller ids query",
			request: getRequest{
				ids:  []string{"123e4567-e89b-12d3-a456-426614174000", "controller-snapshot-2"},
				sort: "id.asc",
			},
			mockEndpoints: getSnapshotsGetEndpoints(
				"or=(id.in.(123e4567-e89b-12d3-a456-426614174000),"+
					"controller_id.in.(controller-snapshot-2))&order=id.asc",
				http.StatusOK,
				snapshotsGetMixedIDsResponseJSON,
			),
			expected: getExpected[types.Snapshots]{
				result: expectedResponseTwoSnapshots,
			},
		},
		{
			name: "get snapshots sort by controller_id",
			request: getRequest{
				sort: "controller_id",
			},
			mockEndpoints: getSnapshotsGetEndpoints(
				"order=controller_id",
				http.StatusOK,
				snapshotsGetOneResponseJSON,
			),
			expected: getExpected[types.Snapshots]{
				result: expectedResponseOneSnapshot,
			},
		},
		{
			name: "get snapshots sort by name",
			request: getRequest{
				sort: "name",
			},
			mockEndpoints: getSnapshotsGetEndpoints(
				"order=name",
				http.StatusOK,
				snapshotsGetOneResponseJSON,
			),
			expected: getExpected[types.Snapshots]{
				result: expectedResponseOneSnapshot,
			},
		},
		{
			name: "get snapshots sort by username",
			request: getRequest{
				sort: "username",
			},
			mockEndpoints: getSnapshotsGetEndpoints(
				"order=username",
				http.StatusOK,
				snapshotsGetOneResponseJSON,
			),
			expected: getExpected[types.Snapshots]{
				result: expectedResponseOneSnapshot,
			},
		},
		{
			name: "get snapshots sort by description",
			request: getRequest{
				sort: "description",
			},
			mockEndpoints: getSnapshotsGetEndpoints(
				"order=description",
				http.StatusOK,
				snapshotsGetOneResponseJSON,
			),
			expected: getExpected[types.Snapshots]{
				result: expectedResponseOneSnapshot,
			},
		},
		{
			name: "get snapshots sort by source.asc",
			request: getRequest{
				sort: "source.asc",
			},
			mockEndpoints: getSnapshotsGetEndpoints(
				"order=spec->source->persistentVolumeClaimName.asc",
				http.StatusOK,
				snapshotsGetOneResponseJSON,
			),
			expected: getExpected[types.Snapshots]{
				result: expectedResponseOneSnapshot,
			},
		},
		{
			name: "get snapshots sort by size.desc",
			request: getRequest{
				sort: "size.desc",
			},
			mockEndpoints: getSnapshotsGetEndpoints(
				"order=volume_status->status->restoreSize.desc",
				http.StatusOK,
				snapshotsGetOneResponseJSON,
			),
			expected: getExpected[types.Snapshots]{
				result: expectedResponseOneSnapshot,
			},
		},
		{
			name: "get snapshots sort by ready.asc",
			request: getRequest{
				sort: "ready.asc",
			},
			mockEndpoints: getSnapshotsGetEndpoints(
				"order=volume_status->status->readyToUse.asc",
				http.StatusOK,
				snapshotsGetOneResponseJSON,
			),
			expected: getExpected[types.Snapshots]{
				result: expectedResponseOneSnapshot,
			},
		},
		{
			name: "get snapshots sort by age",
			request: getRequest{
				sort: "age",
			},
			mockEndpoints: getSnapshotsGetEndpoints(
				"order=created_at",
				http.StatusOK,
				snapshotsGetOneResponseJSON,
			),
			expected: getExpected[types.Snapshots]{
				result: expectedResponseOneSnapshot,
			},
		},
		{
			name: "get snapshots sort by age, source, size, and ready",
			request: getRequest{
				sort: "age.desc,source.asc,size,ready.desc",
			},
			mockEndpoints: getSnapshotsGetEndpoints(
				"order=created_at.desc,spec->source->persistentVolumeClaimName.asc,"+
					"volume_status->status->restoreSize,volume_status->status->readyToUse.desc",
				http.StatusOK,
				snapshotsGetOneResponseJSON,
			),
			expected: getExpected[types.Snapshots]{
				result: expectedResponseOneSnapshot,
			},
		},
		{
			name: "get snapshots invalid json response",
			request: getRequest{
				ids: []string{"id-a"},
			},
			mockEndpoints: getSnapshotsGetEndpoints(
				"controller_id=in.(id-a)",
				http.StatusOK,
				`[`,
			),
			expected: getExpected[types.Snapshots]{
				err: "GET %s/postgrest/snapshots?controller_id=in.%%28id-a%%29 failed\n" +
					" Caused by:\n" +
					"\t* unexpected EOF",
			},
		},
		{
			name: "get snapshots non-200 from server",
			request: getRequest{
				ids: []string{"id-a"},
			},
			mockEndpoints: getSnapshotsGetEndpoints(
				"controller_id=in.(id-a)",
				http.StatusUnauthorized,
				`unauthorized`,
			),
			expected: getExpected[types.Snapshots]{
				err: "GET %s/postgrest/snapshots?controller_id=in.%%28id-a%%29 failed\n" +
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

			actual, err := workbench.GetSnapshots(context.Background(), svc, tt.request.ids, tt.request.sort)
			if tt.expected.err != "" {
				require.EqualError(t, err, fmt.Sprintf(tt.expected.err, srv.URL))
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.expected.result, actual)
		})
	}
}

func TestPrintSnapshots(t *testing.T) {
	tests := []struct {
		name               string
		resourcesJSON      string
		token              string
		expectedTableCount int
		expectedTable      string
		expectedPrintCalls int
	}{
		{
			name:               "print snapshots empty resources with non-admin token",
			resourcesJSON:      snapshotsEmptyResourcesJSON,
			token:              "token-without-admin-role",
			expectedTableCount: 1,
			expectedTable:      expectedSnapshotsEmptyTable,
			expectedPrintCalls: 1,
		},
		{
			name:               "print snapshots empty resources with admin token",
			resourcesJSON:      snapshotsEmptyResourcesJSON,
			token:              createTokenWithRole(constants.WorkbenchAdminRole),
			expectedTableCount: 1,
			expectedTable:      expectedSnapshotsAdminEmptyTable,
			expectedPrintCalls: 1,
		},
		{
			name:               "print snapshots prints one resource with non-admin token",
			resourcesJSON:      snapshotsPrintOneResourcesJSON,
			token:              "token-without-admin-role",
			expectedTableCount: 1,
			expectedTable:      expectedSnapshotsOneRowTable,
			expectedPrintCalls: 1,
		},
		{
			name:               "print snapshots prints one resource with admin token",
			resourcesJSON:      snapshotsPrintOneResourcesJSON,
			token:              createTokenWithRole(constants.WorkbenchAdminRole),
			expectedTableCount: 1,
			expectedTable:      expectedSnapshotsAdminOneRowTable,
			expectedPrintCalls: 1,
		},
		{
			name:               "print snapshots prints two resources with non-admin token",
			resourcesJSON:      snapshotsPrintTwoResourcesJSON,
			token:              "token-without-admin-role",
			expectedTableCount: 1,
			expectedTable:      expectedSnapshotsTwoRowsTable,
			expectedPrintCalls: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var resources types.Snapshots
			err := json.Unmarshal([]byte(tt.resourcesJSON), &resources)
			require.NoError(t, err)
			u := &stubs.StubUI{}
			err = workbench.PrintSnapshots(u, tt.token, resources)
			require.NoError(t, err)
			require.Len(t, u.Tables, tt.expectedTableCount)
			tbl := u.Tables[0]
			require.Equal(t, tt.expectedTable, tbl.String())
			require.Equal(t, tt.expectedPrintCalls, tbl.PrintCalls)
		})
	}
}

func TestDescribeSnapshots(t *testing.T) {
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
			name:             "describe snapshots describe empty resources prints nothing",
			resourcesJSON:    snapshotsEmptyResourcesJSON,
			expectedDisplays: []string{""},
		},
		{
			name:             "describe snapshots describe prints snapshot with sections",
			resourcesJSON:    snapshotsDescribeOneResourcesJSON,
			expectedDisplays: []string{expectedDescribeSnapshotOneDisplay},
		},
		{
			name:             "describe snapshots describe prints snapshots with sections",
			resourcesJSON:    snapshotsDescribeTwoResourcesJSON,
			expectedDisplays: []string{expectedDescribeSnapshotTwoDisplay},
		},
		{
			name:             "describe snapshots describe prints snapshot without spec",
			resourcesJSON:    snapshotsDescribeNilSpecResourcesJSON,
			expectedDisplays: []string{expectedDescribeSnapshotNilSpecOneDisplay},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var resources types.Snapshots
			err := json.Unmarshal([]byte(tt.resourcesJSON), &resources)
			require.NoError(t, err)
			stubUI := &stubs.StubUI{}
			err = workbench.DescribeSnapshots(stubUI, resources)
			require.NoError(t, err)
			require.Equal(t, tt.expectedDisplays, stubUI.Displays, strings.Join(stubUI.Displays, ""))
			require.Equal(t, tt.expectedWarns, stubUI.Warns, strings.Join(stubUI.Warns, ""))
		})
	}
}

func getSnapshotDeleteEndpoints(
	expectedBody string,
	responseStatus int,
	responseBody string,
) []stubs.StubHTTPEndpoint {
	return []stubs.StubHTTPEndpoint{
		{
			Request: stubs.StubHTTPRequest{
				Method:          http.MethodPost,
				URL:             postgrest.Endpoint(constants.WorkbenchEndpointRPCSnapshotBulkDelete),
				ExpectedHeaders: authHeaders,
				ExpectedBody:    expectedBody,
			},
			Response: stubs.StubHTTPResponse{
				Body:   responseBody,
				Status: responseStatus,
			},
		},
	}
}

func getSnapshotCreateEndpoints(
	expectedBody string,
	responseStatus int,
	responseBody string,
) []stubs.StubHTTPEndpoint {
	return []stubs.StubHTTPEndpoint{
		{
			Request: stubs.StubHTTPRequest{
				Method:          http.MethodPost,
				URL:             postgrest.Endpoint(constants.WorkbenchEndpointRPCSnapshotCreate),
				ExpectedHeaders: authHeaders,
				ExpectedBody:    expectedBody,
			},
			Response: stubs.StubHTTPResponse{
				Body:   responseBody,
				Status: responseStatus,
			},
		},
	}
}

func getSnapshotBulkUpdateEndpoints(
	expectedBody string,
	responseStatus int,
	responseBody string,
) []stubs.StubHTTPEndpoint {
	return []stubs.StubHTTPEndpoint{
		{
			Request: stubs.StubHTTPRequest{
				Method:          http.MethodPost,
				URL:             postgrest.Endpoint(constants.WorkbenchEndpointRPCSnapshotBulkUpdate),
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

func getSnapshotsGetEndpoints(
	exectedQuery string,
	responseStatus int,
	responseBody string,
) []stubs.StubHTTPEndpoint {
	return []stubs.StubHTTPEndpoint{
		{
			Request: stubs.StubHTTPRequest{
				Method:          http.MethodGet,
				URL:             postgrest.Endpoint(constants.WorkbenchEndpointSnapshots),
				ExpectedHeaders: authHeaders,
				ExpectedQuery:   exectedQuery,
			},
			Response: stubs.StubHTTPResponse{
				Status: responseStatus,
				Body:   responseBody,
			},
		},
	}
}
