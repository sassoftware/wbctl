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

	"github.com/sas-institute-rnd-saas/wbctl/internal/auth0admin"
	"github.com/sas-institute-rnd-saas/wbctl/internal/constants"
	"github.com/sas-institute-rnd-saas/wbctl/internal/services/workbench"
	"github.com/sas-institute-rnd-saas/wbctl/internal/stubs"
	"github.com/sas-institute-rnd-saas/wbctl/internal/timeutil"
	"github.com/sas-institute-rnd-saas/wbctl/internal/types"
	"github.com/stretchr/testify/require"
)

const (
	usersEmptyResourcesJSON = `{}`

	usersGetOneResponseJSON = `
{
  "version": 2,
  "accept": "application/vnd.sas.auth0admin.user+json",
  "count": 1,
  "start": 0,
  "limit": 20,
  "name": "users",
  "links": [],
  "items": [
    {
      "version": 1,
      "id": "auth0|6a986b2efeca03ccb89eca7b",
      "databaseId": "u-1",
      "name": "alice",
      "email": "alice@example.com",
      "creationTimeStamp": "2025-05-08T00:00:00Z",
      "roles": [
        {
          "id": "r-1",
          "name": "member"
        }
      ],
      "userMetadata": {
        "theme": "dark"
      },
      "links": []
    }
  ]
}
`

	usersGetTwoResponseJSON = `
{
  "version": 2,
  "accept": "application/vnd.sas.auth0admin.user+json",
  "count": 2,
  "start": 0,
  "limit": 20,
  "name": "users",
  "links": [],
  "items": [
    {
      "version": 1,
      "id": "auth0|6a986b2efeca03ccb89eca7b",
      "databaseId": "id-a",
      "name": "alpha",
      "email": "a@example.com",
      "creationTimeStamp": "2025-05-08T00:00:00Z",
      "roles": [
        {
          "id": "r-1",
          "name": "member"
        }
      ],
      "links": []
    },
    {
      "version": 1,
      "id": "auth0|6a986b2a9fcd628b746cee66",
      "databaseId": "id-b",
      "name": "beta",
      "email": "b@example.com",
      "creationTimeStamp": "2025-05-08T00:00:00Z",
      "roles": [
        {
          "id": "r-2",
          "name": "admin"
        }
      ],
      "links": []
    }
  ]
}
`

	usersPrintOneResourcesJSON = `
{
  "version": 2,
  "accept": "application/vnd.sas.auth0admin.user+json",
  "count": 1,
  "start": 0,
  "limit": 20,
  "name": "users",
  "links": [],
  "items": [
    {
      "version": 1,
      "id": "auth0|6a986b2efeca03ccb89eca7b",
      "databaseId": "u-1",
      "name": "alice",
      "email": "alice@example.com",
      "creationTimeStamp": "2025-05-08T00:00:00Z",
      "roles": [
        {
          "id": "r-1",
          "name": "member"
        },
        {
          "id": "r-2",
          "name": "admin"
        }
      ],
      "userMetadata": {
        "a": "1",
        "b": "2"
      }
    }
  ]
}
`

	usersPrintOverflowRolesJSON = `
{
  "version": 2,
  "accept": "application/vnd.sas.auth0admin.user+json",
  "count": 1,
  "start": 0,
  "limit": 20,
  "name": "users",
  "links": [],
  "items": [
    {
      "version": 1,
      "id": "auth0|6a986b2efeca03ccb89eca7b",
      "databaseId": "u-2",
      "name": "bob",
      "email": "bob@example.com",
      "creationTimeStamp": "2025-05-08T00:00:00Z",
      "roles": [
        { "id": "r-1", "name": "r1" },
        { "id": "r-2", "name": "r2" },
        { "id": "r-3", "name": "r3" },
        { "id": "r-4", "name": "r4" },
        { "id": "r-5", "name": "r5" },
        { "id": "r-6", "name": "r6" }
      ],
      "userMetadata": {}
    }
  ]
}
`

	usersDescribeOneResourcesJSON = `
{
    "accept": "application/vnd.sas.auth0admin.user+json",
    "count": 1,
    "items": [
        {
            "creationTimeStamp": "2025-05-08T00:00:00Z",
            "databaseId": "u-1",
            "email": "alice@example.com",
            "id": "auth0|6aa3fb641649c92a814a12b0",
            "links": [],
            "name": "alice",
            "roles": [
                {
                    "id": "rol_CBru4D4ruLp6PN7t",
                    "name": "admin"
                },
                {
                    "id": "rol_c8teNEL6bBzCYU5y",
                    "name": "member"
                }
            ],
			"userMetadata": {
				"theme": "dark"
			},
            "version": 1
        }
    ],
    "limit": 20,
    "links": [],
    "name": "users",
    "start": 0,
    "version": 2
}
`

	usersDescribeTwoResourcesJSON = `
{
    "accept": "application/vnd.sas.auth0admin.user+json",
    "count": 2,
    "items": [
        {
            "creationTimeStamp": "2025-05-08T00:00:00Z",
            "databaseId": "u-1",
            "email": "alice@example.com",
            "id": "auth0|6aa3fb641649c92a814a12b0",
            "links": [],
            "name": "alice",
            "roles": [
                {
                    "id": "rol_CBru4D4ruLp6PN7t",
                    "name": "admin"
                },
                {
                    "id": "rol_c8teNEL6bBzCYU5y",
                    "name": "member"
                }
            ],
			"userMetadata": {
				"theme": "dark"
			},
            "version": 1
        },
        {
            "creationTimeStamp": "2025-05-08T00:00:00Z",
            "databaseId": "u-2",
            "email": "bob@example.com",
            "id": "auth0|6aa3fb641649c92a814a12b1",
            "links": [],
            "name": "bob",
            "roles": [
                {
                    "id": "rol_CBru4D4ruLp6PN7t",
                    "name": "admin"
                },
                {
                    "id": "rol_c8teNEL6bBzCYU5y",
                    "name": "member"
                }
            ],
			"userMetadata": {
				"theme": "light"
			},
            "version": 1
        }
    ],
    "limit": 20,
    "links": [],
    "name": "users",
    "start": 0,
    "version": 2
}
`

	expectedUsersEmptyTable  = "ID\tNAME\tEMAIL\tROLES\tMETADATA\tAGE"
	expectedUsersOneRowTable = "ID\tNAME\tEMAIL\tROLES\tMETADATA\tAGE\n" +
		"auth0|6a986b2efeca03ccb89eca7b\talice\talice@example.com\tadmin, member\ta=1, b=2\t0s"
	expectedUsersOverflowRolesTable = "ID\tNAME\tEMAIL\tROLES\tMETADATA\tAGE\n" +
		"auth0|6a986b2efeca03ccb89eca7b\tbob\tbob@example.com\tr1, r2, r3, r4, r5, ...\t\t0s"

	expectedDescribeUserOneDisplay = `ID:            auth0|6aa3fb641649c92a814a12b0
Database ID:   u-1
Created At:    2025-05-08T00:00:00Z (4d ago)
Username:      alice
Email:         alice@example.com
Roles:
  - admin
  - member
Metadata:
  theme:   dark
`
	expectedDescribeUsersTwoDisplay = `ID:            auth0|6aa3fb641649c92a814a12b0
Database ID:   u-1
Created At:    2025-05-08T00:00:00Z (4d ago)
Username:      alice
Email:         alice@example.com
Roles:
  - admin
  - member
Metadata:
  theme:   dark


ID:            auth0|6aa3fb641649c92a814a12b1
Database ID:   u-2
Created At:    2025-05-08T00:00:00Z (4d ago)
Username:      bob
Email:         bob@example.com
Roles:
  - admin
  - member
Metadata:
  theme:   light
`
)

var (
	expectedResponseOneUser = types.Auth0UserList{
		Version: 2,
		Accept:  "application/vnd.sas.auth0admin.user+json",
		Count:   1,
		Start:   0,
		Limit:   20,
		Name:    "users",
		Items: []types.Auth0User{
			{
				Version:           1,
				ID:                "auth0|6a986b2efeca03ccb89eca7b",
				DatabaseID:        "u-1",
				Name:              "alice",
				Email:             "alice@example.com",
				CreationTimeStamp: time.Date(2025, time.May, 8, 0, 0, 0, 0, time.UTC),
				Roles:             []types.Role{{ID: "r-1", Name: "member"}},
				UserMetadata:      map[string]any{"theme": "dark"},
			},
		},
	}

	expectedResponseTwoUsers = types.Auth0UserList{
		Version: 2,
		Accept:  "application/vnd.sas.auth0admin.user+json",
		Count:   2,
		Start:   0,
		Limit:   20,
		Name:    "users",
		Items: []types.Auth0User{
			{
				Version:           1,
				ID:                "auth0|6a986b2efeca03ccb89eca7b",
				DatabaseID:        "id-a",
				Name:              "alpha",
				Email:             "a@example.com",
				CreationTimeStamp: time.Date(2025, time.May, 8, 0, 0, 0, 0, time.UTC),
				Roles:             []types.Role{{ID: "r-1", Name: "member"}},
			},
			{
				Version:           1,
				ID:                "auth0|6a986b2a9fcd628b746cee66",
				DatabaseID:        "id-b",
				Name:              "beta",
				Email:             "b@example.com",
				CreationTimeStamp: time.Date(2025, time.May, 8, 0, 0, 0, 0, time.UTC),
				Roles:             []types.Role{{ID: "r-2", Name: "admin"}},
			},
		},
	}
)

func TestGetUsers(t *testing.T) {
	tests := []getTestCase[types.Auth0UserList]{
		{
			name:    "get users success without ids query",
			request: getRequest{},
			mockEndpoints: getUsersGetEndpoints(
				"",
				http.StatusOK,
				usersGetOneResponseJSON,
			),
			expected: getExpected[types.Auth0UserList]{
				result: expectedResponseOneUser,
			},
		},
		{
			name: "get users success with ids query",
			request: getRequest{
				ids:  []string{"id-a", "id-b"},
				sort: "id.desc",
			},
			mockEndpoints: getUsersGetEndpoints(
				"id=id-a&id=id-b&sortBy=id:descending",
				http.StatusOK,
				usersGetTwoResponseJSON,
			),
			expected: getExpected[types.Auth0UserList]{
				result: expectedResponseTwoUsers,
			},
		},
		{
			name: "get users sort by name",
			request: getRequest{
				sort: "name",
			},
			mockEndpoints: getUsersGetEndpoints(
				"sortBy=name:ascending",
				http.StatusOK,
				usersGetOneResponseJSON,
			),
			expected: getExpected[types.Auth0UserList]{
				result: expectedResponseOneUser,
			},
		},
		{
			name: "get users sort by age",
			request: getRequest{
				sort: "age.desc",
			},
			mockEndpoints: getUsersGetEndpoints(
				"sortBy=creationTimeStamp:descending",
				http.StatusOK,
				usersGetOneResponseJSON,
			),
			expected: getExpected[types.Auth0UserList]{
				result: expectedResponseOneUser,
			},
		},
		{
			name: "get users sort by age and name",
			request: getRequest{
				sort: "age.desc,name",
			},
			expected: getExpected[types.Auth0UserList]{
				err: "only one sort field allowed at a time",
			},
		},
		{
			name: "get users sort by roles",
			request: getRequest{
				sort: "roles",
			},
			expected: getExpected[types.Auth0UserList]{
				err: "unsupported sort keys: roles",
			},
		},
		{
			name: "get users sort by metadata",
			request: getRequest{
				sort: "metadata",
			},
			expected: getExpected[types.Auth0UserList]{
				err: "unsupported sort keys: metadata",
			},
		},
		{
			name: "get users sort by roles and metadata",
			request: getRequest{
				sort: "roles,metadata",
			},
			expected: getExpected[types.Auth0UserList]{
				err: "unsupported sort keys: roles, metadata",
			},
		},
		{
			name: "get users invalid json response",
			request: getRequest{
				ids: []string{"id-a"},
			},
			mockEndpoints: getUsersGetEndpoints(
				"id=id-a",
				http.StatusOK,
				`[`,
			),
			expected: getExpected[types.Auth0UserList]{
				err: "GET %s/auth0Admin/users?id=id-a failed\n" +
					" Caused by:\n" +
					"\t* unexpected EOF",
			},
		},
		{
			name: "get users non-200 from server",
			request: getRequest{
				ids: []string{"id-a"},
			},
			mockEndpoints: getUsersGetEndpoints(
				"id=id-a",
				http.StatusUnauthorized,
				`unauthorized`,
			),
			expected: getExpected[types.Auth0UserList]{
				err: "GET %s/auth0Admin/users?id=id-a failed\n" +
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

			actual, err := workbench.GetUsers(context.Background(), svc, tt.request.ids, tt.request.sort)
			if tt.expected.err != "" {
				expectedErr := tt.expected.err
				if strings.Contains(expectedErr, "%s") {
					expectedErr = fmt.Sprintf(expectedErr, srv.URL)
				}
				require.EqualError(t, err, expectedErr)
				return
			}

			require.NoError(t, err)
			require.Equal(t, tt.expected.result, actual)
		})
	}
}

func TestPrintUsers(t *testing.T) {
	fixedNow := time.Date(2025, time.May, 8, 0, 0, 0, 0, time.UTC)
	timeutil.SetNow(func() time.Time { return fixedNow })
	t.Cleanup(func() { timeutil.SetNow(time.Now) })

	tests := []struct {
		name               string
		resourcesJSON      string
		expectedTableCount int
		expectedTable      string
		expectedPrintCalls int
	}{
		{
			name:               "print users empty resources",
			resourcesJSON:      usersEmptyResourcesJSON,
			expectedTableCount: 1,
			expectedTable:      expectedUsersEmptyTable,
			expectedPrintCalls: 1,
		},
		{
			name:               "print users prints resources",
			resourcesJSON:      usersPrintOneResourcesJSON,
			expectedTableCount: 1,
			expectedTable:      expectedUsersOneRowTable,
			expectedPrintCalls: 1,
		},
		{
			name:               "print users prints roles overflow marker",
			resourcesJSON:      usersPrintOverflowRolesJSON,
			expectedTableCount: 1,
			expectedTable:      expectedUsersOverflowRolesTable,
			expectedPrintCalls: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var resources types.Auth0UserList
			err := json.Unmarshal([]byte(tt.resourcesJSON), &resources)
			require.NoError(t, err)

			u := &stubs.StubUI{}
			err = workbench.PrintUsers(u, "ignored-token", resources)
			require.NoError(t, err)

			require.Len(t, u.Tables, tt.expectedTableCount)
			tbl := u.Tables[0]
			require.Equal(t, tt.expectedTable, tbl.String())
			require.Equal(t, tt.expectedPrintCalls, tbl.PrintCalls)
		})
	}
}

func TestDescribeUsers(t *testing.T) {
	tests := []struct {
		name             string
		resourcesJSON    string
		expectedDisplays []string
		expectedWarns    []string
	}{
		{
			name:             "describe users describe empty resources prints nothing",
			resourcesJSON:    usersEmptyResourcesJSON,
			expectedDisplays: []string{""},
		},
		{
			name:             "describe users describe prints user with roles and settings",
			resourcesJSON:    usersDescribeOneResourcesJSON,
			expectedDisplays: []string{expectedDescribeUserOneDisplay},
		},
		{
			name:             "describe users describe prints two users with roles and settings",
			resourcesJSON:    usersDescribeTwoResourcesJSON,
			expectedDisplays: []string{expectedDescribeUsersTwoDisplay},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fixedNow := time.Date(2025, time.May, 12, 0, 0, 0, 0, time.UTC)
			timeutil.SetNow(func() time.Time { return fixedNow })
			t.Cleanup(func() { timeutil.SetNow(time.Now) })
			var resources types.Auth0UserList
			err := json.Unmarshal([]byte(tt.resourcesJSON), &resources)
			require.NoError(t, err)

			stubUI := &stubs.StubUI{}
			err = workbench.DescribeUsers(stubUI, resources)
			require.NoError(t, err)
			require.Equal(t, tt.expectedDisplays, stubUI.Displays, strings.Join(stubUI.Displays, ""))
			require.Equal(t, tt.expectedWarns, stubUI.Warns, strings.Join(stubUI.Warns, ""))
		})
	}
}

func getUsersGetEndpoints(
	expectedQuery string,
	responseStatus int,
	responseBody string,
) []stubs.StubHTTPEndpoint {
	return []stubs.StubHTTPEndpoint{
		{
			Request: stubs.StubHTTPRequest{
				Method:          http.MethodGet,
				URL:             auth0admin.Endpoint(constants.WorkbenchEndpointUsers),
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
