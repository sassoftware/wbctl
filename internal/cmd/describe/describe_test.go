// Copyright © 2026, SAS Institute Inc., Cary, NC, USA.  All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package describe_test

import (
	"context"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/sassoftware/wbctl/internal/auth0admin"
	"github.com/sassoftware/wbctl/internal/cmd"
	"github.com/sassoftware/wbctl/internal/cmd/describe"
	"github.com/sassoftware/wbctl/internal/constants"
	"github.com/sassoftware/wbctl/internal/postgrest"
	"github.com/sassoftware/wbctl/internal/stubs"
	"github.com/sassoftware/wbctl/internal/timeutil"
	"github.com/spf13/afero"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
)

const (
	describeConflictsByPluralTypeExpectedResponseBody = `[{
				"id":"conflict-10",
				"controller_id":"controller-10",
				"name":"Conflict Ten",
				"type":"directory",
				"conflict_id":"conflict-20",
				"conflict_controller_id":"controller-20",
				"conflict_name":"Conflict Twenty",
				"conflict_type":"directory",
				"workspace_id":"workspace-10",
				"workspace_controller_id":"workspace-controller-10",
				"workspace_name":"Workspace Ten"
			},
			{
				"id":"conflict-11",
				"controller_id":"controller-11",
				"name":"Conflict Eleven",
				"type":"directory",
				"conflict_id":"conflict-22",
				"conflict_controller_id":"controller-22",
				"conflict_name":"Conflict TwentyTwo",
				"conflict_type":"directory",
				"workspace_id":"workspace-11",
				"workspace_controller_id":"workspace-controller-11",
				"workspace_name":"Workspace Eleven"
			}]`
	describeGenericResourceBySingularTypeExpectedResponseBody = `[{
				"id":"resource-1",
				"created_at":"2025-01-01T00:00:00Z",
				"modified_at":"2025-01-01T00:00:00Z"
			}]`
	describeGenericResourceByPluralTypeExpectedResponseBody = `[{
				"id":"resource-1",
				"created_at":"2025-01-01T00:00:00Z",
				"modified_at":"2025-01-01T00:00:00Z"
			}]`
	describeUsersBySingularTypeExpectedResponseBody = `
{
  "version": 2,
  "accept": "application/vnd.sas.auth0admin.user+json",
  "count": 6,
  "start": 0,
  "limit": 20,
  "name": "users",
  "links": [
    {
      "method": "GET",
      "rel": "collection",
      "href": "/auth0Admin/users",
      "uri": "/auth0Admin/users",
      "type": "application/vnd.sas.collection",
      "itemType": "application/vnd.sas.auth0admin.user+json"
    },
    {
      "method": "GET",
      "rel": "self",
      "href": "/auth0Admin/users?start=0&limit=20",
      "uri": "/auth0Admin/users?start=0&limit=20",
      "type": "application/vnd.sas.collection",
      "itemType": "application/vnd.sas.auth0admin.user+json"
    }
  ],
  "items": [
    {
      "version": 1,
      "id": "auth0|6a986b2efeca03ccb89eca7b",
      "databaseId": "7d40724d-2cbf-50f0-8482-2b11afa9603c",
      "name": "pupadmin1 Test User",
      "email": "pupadmin1@sas-test-mickey2.com",
      "creationTimeStamp": "2026-09-02T18:30:07.021Z",
      "roles": [
        {
          "id": "rol_CBru4D4ruLp6PN7t",
          "name": "workbench_admin"
        },
        {
          "id": "rol_c8teNEL6bBzCYU5y",
          "name": "org_admin"
        }
      ],
      "links": [
        {
          "method": "GET",
          "rel": "up",
          "href": "/auth0Admin/users",
          "uri": "/auth0Admin/users",
          "type": "application/vnd.sas.collection+json"
        },
        {
          "method": "GET",
          "rel": "self",
          "href": "/auth0Admin/users/auth0|6a986b2efeca03ccb89eca7b",
          "uri": "/auth0Admin/users/auth0|6a986b2efeca03ccb89eca7b",
          "type": "application/vnd.sas.auth0admin.user+json"
        },
        {
          "method": "PATCH",
          "rel": "patch",
          "href": "/auth0Admin/users/auth0|6a986b2efeca03ccb89eca7b",
          "uri": "/auth0Admin/users/auth0|6a986b2efeca03ccb89eca7b",
          "type": "application/vnd.sas.auth0admin.user+json"
        },
        {
          "method": "DELETE",
          "rel": "delete",
          "href": "/auth0Admin/users/auth0|6a986b2efeca03ccb89eca7b",
          "uri": "/auth0Admin/users/auth0|6a986b2efeca03ccb89eca7b"
        }
      ]
    },
    {
      "version": 1,
      "id": "auth0|6a986b2a9fcd628b746cee66",
      "databaseId": "7a056a78-d94a-5eed-9a6f-a077016b13a1",
      "name": "pupadmin2 Test User",
      "email": "pupadmin2@sas-test-mickey2.com",
      "creationTimeStamp": "2026-09-02T18:30:02.939Z",
      "roles": [
        {
          "id": "rol_c8teNEL6bBzCYU5y",
          "name": "org_admin"
        },
        {
          "id": "rol_CBru4D4ruLp6PN7t",
          "name": "workbench_admin"
        }
      ],
      "links": [
        {
          "method": "GET",
          "rel": "up",
          "href": "/auth0Admin/users",
          "uri": "/auth0Admin/users",
          "type": "application/vnd.sas.collection+json"
        },
        {
          "method": "GET",
          "rel": "self",
          "href": "/auth0Admin/users/auth0|6a986b2a9fcd628b746cee66",
          "uri": "/auth0Admin/users/auth0|6a986b2a9fcd628b746cee66",
          "type": "application/vnd.sas.auth0admin.user+json"
        },
        {
          "method": "PATCH",
          "rel": "patch",
          "href": "/auth0Admin/users/auth0|6a986b2a9fcd628b746cee66",
          "uri": "/auth0Admin/users/auth0|6a986b2a9fcd628b746cee66",
          "type": "application/vnd.sas.auth0admin.user+json"
        },
        {
          "method": "DELETE",
          "rel": "delete",
          "href": "/auth0Admin/users/auth0|6a986b2a9fcd628b746cee66",
          "uri": "/auth0Admin/users/auth0|6a986b2a9fcd628b746cee66"
        }
      ]
    },
    {
      "version": 1,
      "id": "auth0|6a986b2c04816dc0574018dc",
      "databaseId": "6e6a0a0b-270c-5a68-9bc4-f904bb5deee9",
      "name": "pupadmin3 Test User",
      "email": "pupadmin3@sas-test-mickey2.com",
      "creationTimeStamp": "2026-09-02T18:30:04.96Z",
      "roles": [
        {
          "id": "rol_CBru4D4ruLp6PN7t",
          "name": "workbench_admin"
        },
        {
          "id": "rol_c8teNEL6bBzCYU5y",
          "name": "org_admin"
        }
      ],
      "links": [
        {
          "method": "GET",
          "rel": "up",
          "href": "/auth0Admin/users",
          "uri": "/auth0Admin/users",
          "type": "application/vnd.sas.collection+json"
        },
        {
          "method": "GET",
          "rel": "self",
          "href": "/auth0Admin/users/auth0|6a986b2c04816dc0574018dc",
          "uri": "/auth0Admin/users/auth0|6a986b2c04816dc0574018dc",
          "type": "application/vnd.sas.auth0admin.user+json"
        },
        {
          "method": "PATCH",
          "rel": "patch",
          "href": "/auth0Admin/users/auth0|6a986b2c04816dc0574018dc",
          "uri": "/auth0Admin/users/auth0|6a986b2c04816dc0574018dc",
          "type": "application/vnd.sas.auth0admin.user+json"
        },
        {
          "method": "DELETE",
          "rel": "delete",
          "href": "/auth0Admin/users/auth0|6a986b2c04816dc0574018dc",
          "uri": "/auth0Admin/users/auth0|6a986b2c04816dc0574018dc"
        }
      ]
    },
    {
      "version": 1,
      "id": "auth0|6a986b2a526605e8b186210b",
      "databaseId": "e60cb636-0aa5-586b-9441-0d24ae0e69b7",
      "name": "pupusr1 Test User",
      "email": "pupusr1@sas-test-mickey2.com",
      "creationTimeStamp": "2026-09-02T18:30:02.922Z",
      "roles": [
        {
          "id": "rol_G2Xo84BaiLvnBLvD",
          "name": "workbench_user"
        }
      ],
      "links": [
        {
          "method": "GET",
          "rel": "up",
          "href": "/auth0Admin/users",
          "uri": "/auth0Admin/users",
          "type": "application/vnd.sas.collection+json"
        },
        {
          "method": "GET",
          "rel": "self",
          "href": "/auth0Admin/users/auth0|6a986b2a526605e8b186210b",
          "uri": "/auth0Admin/users/auth0|6a986b2a526605e8b186210b",
          "type": "application/vnd.sas.auth0admin.user+json"
        },
        {
          "method": "PATCH",
          "rel": "patch",
          "href": "/auth0Admin/users/auth0|6a986b2a526605e8b186210b",
          "uri": "/auth0Admin/users/auth0|6a986b2a526605e8b186210b",
          "type": "application/vnd.sas.auth0admin.user+json"
        },
        {
          "method": "DELETE",
          "rel": "delete",
          "href": "/auth0Admin/users/auth0|6a986b2a526605e8b186210b",
          "uri": "/auth0Admin/users/auth0|6a986b2a526605e8b186210b"
        }
      ]
    },
    {
      "version": 1,
      "id": "auth0|6a986b2c1f42bc857feb9738",
      "databaseId": "6c1df094-36d4-5fc1-88aa-6131c4408bd5",
      "name": "pupusr2 Test User",
      "email": "pupusr2@sas-test-mickey2.com",
      "creationTimeStamp": "2026-09-02T18:30:04.904Z",
      "roles": [
        {
          "id": "rol_G2Xo84BaiLvnBLvD",
          "name": "workbench_user"
        }
      ],
      "links": [
        {
          "method": "GET",
          "rel": "up",
          "href": "/auth0Admin/users",
          "uri": "/auth0Admin/users",
          "type": "application/vnd.sas.collection+json"
        },
        {
          "method": "GET",
          "rel": "self",
          "href": "/auth0Admin/users/auth0|6a986b2c1f42bc857feb9738",
          "uri": "/auth0Admin/users/auth0|6a986b2c1f42bc857feb9738",
          "type": "application/vnd.sas.auth0admin.user+json"
        },
        {
          "method": "PATCH",
          "rel": "patch",
          "href": "/auth0Admin/users/auth0|6a986b2c1f42bc857feb9738",
          "uri": "/auth0Admin/users/auth0|6a986b2c1f42bc857feb9738",
          "type": "application/vnd.sas.auth0admin.user+json"
        },
        {
          "method": "DELETE",
          "rel": "delete",
          "href": "/auth0Admin/users/auth0|6a986b2c1f42bc857feb9738",
          "uri": "/auth0Admin/users/auth0|6a986b2c1f42bc857feb9738"
        }
      ]
    },
    {
      "version": 1,
      "id": "auth0|6a986b2e9fcd628b746cee6f",
      "databaseId": "1dcd256b-3deb-50a1-839b-2d102b9bf1d5",
      "name": "pupusr3 Test User",
      "email": "pupusr3@sas-test-mickey2.com",
      "creationTimeStamp": "2026-09-02T18:30:07.02Z",
      "roles": [
        {
          "id": "rol_G2Xo84BaiLvnBLvD",
          "name": "workbench_user"
        }
      ],
      "links": [
        {
          "method": "GET",
          "rel": "up",
          "href": "/auth0Admin/users",
          "uri": "/auth0Admin/users",
          "type": "application/vnd.sas.collection+json"
        },
        {
          "method": "GET",
          "rel": "self",
          "href": "/auth0Admin/users/auth0|6a986b2e9fcd628b746cee6f",
          "uri": "/auth0Admin/users/auth0|6a986b2e9fcd628b746cee6f",
          "type": "application/vnd.sas.auth0admin.user+json"
        },
        {
          "method": "PATCH",
          "rel": "patch",
          "href": "/auth0Admin/users/auth0|6a986b2e9fcd628b746cee6f",
          "uri": "/auth0Admin/users/auth0|6a986b2e9fcd628b746cee6f",
          "type": "application/vnd.sas.auth0admin.user+json"
        },
        {
          "method": "DELETE",
          "rel": "delete",
          "href": "/auth0Admin/users/auth0|6a986b2e9fcd628b746cee6f",
          "uri": "/auth0Admin/users/auth0|6a986b2e9fcd628b746cee6f"
        }
      ]
    }
  ]
}
`
	describeConflictExpectedDisplay = `ID:                        resource-1
Controller ID:             
Name:                      
Type:                      
Conflict ID:               
Conflict Controller ID:    
Conflict Name:             
Conflict Type:             
Workspace ID:              
Workspace Controller ID:   
Workspace Name:            
`
	describeConflictsPluralExpectedDisplay = `ID:                        conflict-10
Controller ID:             controller-10
Name:                      Conflict Ten
Type:                      directory
Conflict ID:               conflict-20
Conflict Controller ID:    controller-20
Conflict Name:             Conflict Twenty
Conflict Type:             directory
Workspace ID:              workspace-10
Workspace Controller ID:   workspace-controller-10
Workspace Name:            Workspace Ten


ID:                        conflict-11
Controller ID:             controller-11
Name:                      Conflict Eleven
Type:                      directory
Conflict ID:               conflict-22
Conflict Controller ID:    controller-22
Conflict Name:             Conflict TwentyTwo
Conflict Type:             directory
Workspace ID:              workspace-11
Workspace Controller ID:   workspace-controller-11
Workspace Name:            Workspace Eleven
`
	describeDataplaneExpectedDisplay = `ID:            resource-1
Name:          
Created By:    
Created At:    2025-01-01T00:00:00Z (127d ago)
Modified At:   2025-01-01T00:00:00Z (127d ago)
State:         
`
	describeImageBuildExpectedDisplay = `ID:              resource-1
Controller ID:   
Name:            
Created By:      
Created At:      2025-01-01T00:00:00Z (127d ago)
Image ID:        
`
	describeImageExpectedDisplay = `ID:              resource-1
Controller ID:   
Name:            
Created By:       ()
Created At:      2025-01-01T00:00:00Z (127d ago)
Modified At:     2025-01-01T00:00:00Z (127d ago)
Image:           /
`
	describeJobExpectedDisplay = `ID:            resource-1
Name:          
Created By:     ()
Created At:    2025-01-01T00:00:00Z (127d ago)
Modified At:   2025-01-01T00:00:00Z (127d ago)
Nodepool:      
`
	describeJobSubmissionExpectedDisplay = `ID:              resource-1
Controller ID:   
Job Name:        
Job ID:          
URL:             
Log URL:         
Created By:       ()
Created At:      2025-01-01T00:00:00Z (127d ago)
Modified At:     2025-01-01T00:00:00Z (127d ago)
Runtime:         0.000000
Scheduled:       false
State:           
`
	describeSnapshotExpectedDisplay = `ID:              resource-1
Controller ID:   
Name:            
Description:     
Created By:       ()
Created At:      2025-01-01T00:00:00Z (127d ago)
Modified At:     2025-01-01T00:00:00Z (127d ago)
Volume Status:
  Ready To Use:    false
  Restore Size:    <unknown>
  Content Name:    <unknown>
  Creation Time:   <unknown>
`
	describeUserExpectedDisplay = `ID:            auth0|6a986b2efeca03ccb89eca7b
Database ID:   7d40724d-2cbf-50f0-8482-2b11afa9603c
Created At:    2026-09-02T18:30:07Z (<invalid> ago)
Username:      pupadmin1 Test User
Email:         pupadmin1@sas-test-mickey2.com
Roles:
  - org_admin
  - workbench_admin


ID:            auth0|6a986b2a9fcd628b746cee66
Database ID:   7a056a78-d94a-5eed-9a6f-a077016b13a1
Created At:    2026-09-02T18:30:02Z (<invalid> ago)
Username:      pupadmin2 Test User
Email:         pupadmin2@sas-test-mickey2.com
Roles:
  - org_admin
  - workbench_admin


ID:            auth0|6a986b2c04816dc0574018dc
Database ID:   6e6a0a0b-270c-5a68-9bc4-f904bb5deee9
Created At:    2026-09-02T18:30:04Z (<invalid> ago)
Username:      pupadmin3 Test User
Email:         pupadmin3@sas-test-mickey2.com
Roles:
  - org_admin
  - workbench_admin


ID:            auth0|6a986b2a526605e8b186210b
Database ID:   e60cb636-0aa5-586b-9441-0d24ae0e69b7
Created At:    2026-09-02T18:30:02Z (<invalid> ago)
Username:      pupusr1 Test User
Email:         pupusr1@sas-test-mickey2.com
Roles:
  - workbench_user


ID:            auth0|6a986b2c1f42bc857feb9738
Database ID:   6c1df094-36d4-5fc1-88aa-6131c4408bd5
Created At:    2026-09-02T18:30:04Z (<invalid> ago)
Username:      pupusr2 Test User
Email:         pupusr2@sas-test-mickey2.com
Roles:
  - workbench_user


ID:            auth0|6a986b2e9fcd628b746cee6f
Database ID:   1dcd256b-3deb-50a1-839b-2d102b9bf1d5
Created At:    2026-09-02T18:30:07Z (<invalid> ago)
Username:      pupusr3 Test User
Email:         pupusr3@sas-test-mickey2.com
Roles:
  - workbench_user
`
	describeWorkbenchExpectedDisplay = `ID:              resource-1
Controller ID:   
Name:            
URL:             
Created By:       ()
Created At:      2025-01-01T00:00:00Z (127d ago)
Modified At:     2025-01-01T00:00:00Z (127d ago)
Nodepool:        
Desired State:   
Current State:   
`
	describeWorkspaceExpectedDisplay = `ID:              resource-1
Controller ID:   
Name:            
Storage Type:    
Created By:       ()
Created At:      2025-01-01T00:00:00Z (127d ago)
Modified At:     2025-01-01T00:00:00Z (127d ago)
State:           
Phase:           
`
	expectedInvalidEndpointErr = "failed to get device token" +
		"\n Caused by:" +
		"\n\t* parse \"a\\nb\": net/url: invalid control character in URL"
)

type describeInput struct {
	endpoint     string
	resourceIDs  []string
	resourceType string
}

type describeExpected struct {
	err           string
	displays      []string
	warns         []string
	describeCalls int
}

type describeTestCase struct {
	name          string
	input         describeInput
	mockEndpoints []stubs.StubHTTPEndpoint
	expected      describeExpected
}

func TestDescribeCommand(t *testing.T) {
	fixedNow := time.Date(2025, time.May, 8, 0, 0, 0, 0, time.UTC)
	timeutil.SetNow(func() time.Time { return fixedNow })
	t.Cleanup(func() { timeutil.SetNow(time.Now) })

	tests := []describeTestCase{
		// ==== DESCRIBE SINGLE CASES
		{
			name: "describe " + constants.DataplaneType + " by singular type",
			input: describeInput{
				resourceType: constants.DataplaneType,
				resourceIDs:  []string{"resource-1"},
			},
			mockEndpoints: getDescribeEndpoints(
				postgrest.Endpoint(constants.WorkbenchEndpointDataplanes),
				stubs.StubHTTPPartialGetConfig{
					ExpectedQuery:  "id=in.(resource-1)",
					ResponseStatus: http.StatusOK,
					ResponseBody:   describeGenericResourceBySingularTypeExpectedResponseBody,
				},
			),
			expected: describeExpected{
				describeCalls: 1,
				displays:      []string{stubs.ExpectedDeviceAuthDisplay, describeDataplaneExpectedDisplay},
			},
		},
		{
			name: "describe " + constants.ImageBuildType + " by singular type",
			input: describeInput{
				resourceType: constants.ImageBuildType,
				resourceIDs:  []string{"resource-1"},
			},
			mockEndpoints: getDescribeEndpoints(
				postgrest.Endpoint(constants.WorkbenchEndpointImageBuilds),
				stubs.StubHTTPPartialGetConfig{
					ExpectedQuery:  "controller_id=in.(resource-1)",
					ResponseStatus: http.StatusOK,
					ResponseBody:   describeGenericResourceBySingularTypeExpectedResponseBody,
				},
			),
			expected: describeExpected{
				describeCalls: 1,
				displays:      []string{stubs.ExpectedDeviceAuthDisplay, describeImageBuildExpectedDisplay},
			},
		},
		{
			name: "describe " + constants.ImageType + " by singular type",
			input: describeInput{
				resourceType: constants.ImageType,
				resourceIDs:  []string{"resource-1"},
			},
			mockEndpoints: getDescribeEndpoints(
				postgrest.Endpoint(constants.WorkbenchEndpointImages),
				stubs.StubHTTPPartialGetConfig{
					ExpectedQuery:  "controller_id=in.(resource-1)",
					ResponseStatus: http.StatusOK,
					ResponseBody:   describeGenericResourceBySingularTypeExpectedResponseBody,
				},
			),
			expected: describeExpected{
				describeCalls: 1,
				displays:      []string{stubs.ExpectedDeviceAuthDisplay, describeImageExpectedDisplay},
			},
		},
		{
			name: "describe " + constants.JobSubmissionType + " by singular type",
			input: describeInput{
				resourceType: constants.JobSubmissionType,
				resourceIDs:  []string{"resource-1"},
			},
			mockEndpoints: getDescribeEndpoints(
				postgrest.Endpoint(constants.WorkbenchEndpointJobRuns),
				stubs.StubHTTPPartialGetConfig{
					ExpectedQuery:  "controller_id=in.(resource-1)",
					ResponseStatus: http.StatusOK,
					ResponseBody:   describeGenericResourceBySingularTypeExpectedResponseBody,
				},
			),
			expected: describeExpected{
				describeCalls: 1,
				displays:      []string{stubs.ExpectedDeviceAuthDisplay, describeJobSubmissionExpectedDisplay},
			},
		},
		{
			name: "describe " + constants.JobType + " by singular type",
			input: describeInput{
				resourceType: constants.JobType,
				resourceIDs:  []string{"resource-1"},
			},
			mockEndpoints: getDescribeEndpoints(
				postgrest.Endpoint(constants.WorkbenchEndpointJobs),
				stubs.StubHTTPPartialGetConfig{
					ExpectedQuery:  "id=in.(resource-1)",
					ResponseStatus: http.StatusOK,
					ResponseBody:   describeGenericResourceBySingularTypeExpectedResponseBody,
				},
			),
			expected: describeExpected{
				describeCalls: 1,
				displays:      []string{stubs.ExpectedDeviceAuthDisplay, describeJobExpectedDisplay},
			},
		},
		{
			name: "describe " + constants.SnapshotType + " by singular type",
			input: describeInput{
				resourceType: constants.SnapshotType,
				resourceIDs:  []string{"resource-1"},
			},
			mockEndpoints: getDescribeEndpoints(
				postgrest.Endpoint(constants.WorkbenchEndpointSnapshots),
				stubs.StubHTTPPartialGetConfig{
					ExpectedQuery:  "controller_id=in.(resource-1)",
					ResponseStatus: http.StatusOK,
					ResponseBody:   describeGenericResourceBySingularTypeExpectedResponseBody,
				},
			),
			expected: describeExpected{
				describeCalls: 1,
				displays:      []string{stubs.ExpectedDeviceAuthDisplay, describeSnapshotExpectedDisplay},
			},
		},
		{
			name: "describe " + constants.UserType + " by singular type",
			input: describeInput{
				resourceType: constants.UserType,
				resourceIDs:  []string{"resource-1"},
			},
			mockEndpoints: getDescribeEndpoints(
				auth0admin.Endpoint(constants.WorkbenchEndpointUsers),
				stubs.StubHTTPPartialGetConfig{
					ExpectedQuery:  "id=resource-1",
					ResponseStatus: http.StatusOK,
					ResponseBody:   describeUsersBySingularTypeExpectedResponseBody,
				},
			),
			expected: describeExpected{
				describeCalls: 1,
				displays:      []string{stubs.ExpectedDeviceAuthDisplay, describeUserExpectedDisplay},
			},
		},
		{
			name: "describe " + constants.WorkbenchType + " by singular type",
			input: describeInput{
				resourceType: constants.WorkbenchType,
				resourceIDs:  []string{"resource-1"},
			},
			mockEndpoints: getDescribeEndpoints(
				postgrest.Endpoint(constants.WorkbenchEndpointWorkbenches),
				stubs.StubHTTPPartialGetConfig{
					ExpectedQuery:  "controller_id=in.(resource-1)",
					ResponseStatus: http.StatusOK,
					ResponseBody:   describeGenericResourceBySingularTypeExpectedResponseBody,
				},
			),
			expected: describeExpected{
				describeCalls: 1,
				displays:      []string{stubs.ExpectedDeviceAuthDisplay, describeWorkbenchExpectedDisplay},
			},
		},
		{
			name: "describe " + constants.VolumeType + " by singular type",
			input: describeInput{
				resourceType: constants.VolumeType,
				resourceIDs:  []string{"resource-1"},
			},
			mockEndpoints: getDescribeEndpoints(
				postgrest.Endpoint(constants.WorkbenchEndpointWorkspaces),
				stubs.StubHTTPPartialGetConfig{
					ExpectedQuery:  "controller_id=in.(resource-1)",
					ResponseStatus: http.StatusOK,
					ResponseBody:   describeGenericResourceBySingularTypeExpectedResponseBody,
				},
			),
			expected: describeExpected{
				describeCalls: 1,
				displays:      []string{stubs.ExpectedDeviceAuthDisplay, describeWorkspaceExpectedDisplay},
			},
		},
		{
			name: "describe " + constants.VolumeConflictType + " by singular type",
			input: describeInput{
				resourceType: constants.VolumeConflictType,
				resourceIDs:  []string{"conflict-1"},
			},
			mockEndpoints: getDescribeEndpoints(
				postgrest.Endpoint(constants.WorkbenchEndpointWorkspaceConflicts),
				stubs.StubHTTPPartialGetConfig{
					ExpectedQuery:  "id=in.(conflict-1)",
					ResponseStatus: http.StatusOK,
					ResponseBody:   describeGenericResourceBySingularTypeExpectedResponseBody,
				},
			),
			expected: describeExpected{
				describeCalls: 1,
				displays:      []string{stubs.ExpectedDeviceAuthDisplay, describeConflictExpectedDisplay},
			},
		},

		// ==== DESCRIBE PLURAL CASES
		{
			name: "describe " + constants.DataplanesType + " by plural type",
			input: describeInput{
				resourceType: constants.DataplanesType,
			},
			mockEndpoints: getDescribeEndpoints(
				postgrest.Endpoint(constants.WorkbenchEndpointDataplanes),
				stubs.StubHTTPPartialGetConfig{
					ResponseStatus: http.StatusOK,
					ResponseBody:   describeGenericResourceByPluralTypeExpectedResponseBody,
				},
			),
			expected: describeExpected{
				describeCalls: 1,
				displays:      []string{stubs.ExpectedDeviceAuthDisplay, describeDataplaneExpectedDisplay},
			},
		},
		{
			name: "describe " + constants.ImageBuildsType + " by plural type",
			input: describeInput{
				resourceType: constants.ImageBuildsType,
			},
			mockEndpoints: getDescribeEndpoints(
				postgrest.Endpoint(constants.WorkbenchEndpointImageBuilds),
				stubs.StubHTTPPartialGetConfig{
					ResponseStatus: http.StatusOK,
					ResponseBody:   describeGenericResourceByPluralTypeExpectedResponseBody,
				},
			),
			expected: describeExpected{
				describeCalls: 1,
				displays:      []string{stubs.ExpectedDeviceAuthDisplay, describeImageBuildExpectedDisplay},
			},
		},
		{
			name: "describe " + constants.ImagesType + " by plural type",
			input: describeInput{
				resourceType: constants.ImagesType,
			},
			mockEndpoints: getDescribeEndpoints(
				postgrest.Endpoint(constants.WorkbenchEndpointImages),
				stubs.StubHTTPPartialGetConfig{
					ResponseStatus: http.StatusOK,
					ResponseBody:   describeGenericResourceByPluralTypeExpectedResponseBody,
				},
			),
			expected: describeExpected{
				describeCalls: 1,
				displays:      []string{stubs.ExpectedDeviceAuthDisplay, describeImageExpectedDisplay},
			},
		},
		{
			name: "describe " + constants.JobSubmissionsType + " by plural type",
			input: describeInput{
				resourceType: constants.JobSubmissionsType,
			},
			mockEndpoints: getDescribeEndpoints(
				postgrest.Endpoint(constants.WorkbenchEndpointJobRuns),
				stubs.StubHTTPPartialGetConfig{
					ResponseStatus: http.StatusOK,
					ResponseBody:   describeGenericResourceByPluralTypeExpectedResponseBody,
				},
			),
			expected: describeExpected{
				describeCalls: 1,
				displays:      []string{stubs.ExpectedDeviceAuthDisplay, describeJobSubmissionExpectedDisplay},
			},
		},
		{
			name: "describe " + constants.JobsType + " by plural type",
			input: describeInput{
				resourceType: constants.JobsType,
			},
			mockEndpoints: getDescribeEndpoints(
				postgrest.Endpoint(constants.WorkbenchEndpointJobs),
				stubs.StubHTTPPartialGetConfig{
					ResponseStatus: http.StatusOK,
					ResponseBody:   describeGenericResourceByPluralTypeExpectedResponseBody,
				},
			),
			expected: describeExpected{
				describeCalls: 1,
				displays:      []string{stubs.ExpectedDeviceAuthDisplay, describeJobExpectedDisplay},
			},
		},
		{
			name: "describe " + constants.SnapshotsType + " by plural type",
			input: describeInput{
				resourceType: constants.SnapshotsType,
			},
			mockEndpoints: getDescribeEndpoints(
				postgrest.Endpoint(constants.WorkbenchEndpointSnapshots),
				stubs.StubHTTPPartialGetConfig{
					ResponseStatus: http.StatusOK,
					ResponseBody:   describeGenericResourceByPluralTypeExpectedResponseBody,
				},
			),
			expected: describeExpected{
				describeCalls: 1,
				displays:      []string{stubs.ExpectedDeviceAuthDisplay, describeSnapshotExpectedDisplay},
			},
		},
		{
			name: "describe " + constants.UsersType + " by plural type",
			input: describeInput{
				resourceType: constants.UsersType,
			},
			mockEndpoints: getDescribeEndpoints(
				auth0admin.Endpoint(constants.WorkbenchEndpointUsers),
				stubs.StubHTTPPartialGetConfig{
					ResponseStatus: http.StatusOK,
					ResponseBody:   describeUsersBySingularTypeExpectedResponseBody,
				},
			),
			expected: describeExpected{
				describeCalls: 1,
				displays:      []string{stubs.ExpectedDeviceAuthDisplay, describeUserExpectedDisplay},
			},
		},
		{
			name: "describe " + constants.WorkbenchesType + " by plural type",
			input: describeInput{
				resourceType: constants.WorkbenchesType,
			},
			mockEndpoints: getDescribeEndpoints(
				postgrest.Endpoint(constants.WorkbenchEndpointWorkbenches),
				stubs.StubHTTPPartialGetConfig{
					ResponseStatus: http.StatusOK,
					ResponseBody:   describeGenericResourceByPluralTypeExpectedResponseBody,
				},
			),
			expected: describeExpected{
				describeCalls: 1,
				displays:      []string{stubs.ExpectedDeviceAuthDisplay, describeWorkbenchExpectedDisplay},
			},
		},
		{
			name: "describe " + constants.VolumesType + " by plural type",
			input: describeInput{
				resourceType: constants.VolumesType,
			},
			mockEndpoints: getDescribeEndpoints(
				postgrest.Endpoint(constants.WorkbenchEndpointWorkspaces),
				stubs.StubHTTPPartialGetConfig{
					ResponseStatus: http.StatusOK,
					ResponseBody:   describeGenericResourceByPluralTypeExpectedResponseBody,
				},
			),
			expected: describeExpected{
				describeCalls: 1,
				displays:      []string{stubs.ExpectedDeviceAuthDisplay, describeWorkspaceExpectedDisplay},
			},
		},
		{
			name: "describe " + constants.VolumeConflictsType + " by plural type",
			input: describeInput{
				resourceType: constants.VolumeConflictsType,
			},
			mockEndpoints: getDescribeEndpoints(
				postgrest.Endpoint(constants.WorkbenchEndpointWorkspaceConflicts),
				stubs.StubHTTPPartialGetConfig{
					ResponseStatus: http.StatusOK,
					ResponseBody:   describeConflictsByPluralTypeExpectedResponseBody,
				},
			),
			expected: describeExpected{
				describeCalls: 1,
				displays:      []string{stubs.ExpectedDeviceAuthDisplay, describeConflictsPluralExpectedDisplay},
			},
		},
		{
			name: "invalid endpoint",
			input: describeInput{
				resourceType: constants.VolumeType,
				resourceIDs:  []string{"resource-1"},
				endpoint:     "a\nb",
			},
			expected: describeExpected{
				err: expectedInvalidEndpointErr,
			},
		},
		{
			name: "invalid dataplanes",
			input: describeInput{
				resourceType: constants.DataplanesType,
			},
			mockEndpoints: getDescribeEndpoints(
				postgrest.Endpoint(constants.WorkbenchEndpointDataplanes),
				stubs.StubHTTPPartialGetConfig{
					ResponseStatus: http.StatusOK,
					ResponseBody:   `[invalid`,
				},
			),
			expected: describeExpected{
				displays:      []string{stubs.ExpectedDeviceAuthDisplay},
				describeCalls: 1,
				err: "" +
					"GET ${URL}/postgrest/dataplanes failed" +
					"\n Caused by:" +
					"\n\t* invalid character 'i' looking for beginning of value",
			},
		},
		{
			name: "invalid workspaces",
			input: describeInput{
				resourceType: constants.VolumesType,
			},
			mockEndpoints: getDescribeEndpoints(
				postgrest.Endpoint(constants.WorkbenchEndpointWorkspaces),
				stubs.StubHTTPPartialGetConfig{
					ResponseStatus: http.StatusOK,
					ResponseBody:   `[invalid`,
				},
			),
			expected: describeExpected{
				displays:      []string{stubs.ExpectedDeviceAuthDisplay},
				describeCalls: 1,
				err: "" +
					"GET ${URL}/postgrest/workspaces failed" +
					"\n Caused by:" +
					"\n\t* invalid character 'i' looking for beginning of value",
			},
		},
		{
			name: "invalid workbenches",
			input: describeInput{
				resourceType: constants.WorkbenchesType,
			},
			mockEndpoints: getDescribeEndpoints(
				postgrest.Endpoint(constants.WorkbenchEndpointWorkbenches),
				stubs.StubHTTPPartialGetConfig{
					ResponseStatus: http.StatusOK,
					ResponseBody:   `[invalid`,
				},
			),
			expected: describeExpected{
				displays:      []string{stubs.ExpectedDeviceAuthDisplay},
				describeCalls: 1,
				err: "" +
					"GET ${URL}/postgrest/workbenches failed" +
					"\n Caused by:" +
					"\n\t* invalid character 'i' looking for beginning of value",
			},
		},
		{
			name: "invalid images",
			input: describeInput{
				resourceType: constants.ImagesType,
			},
			mockEndpoints: getDescribeEndpoints(
				postgrest.Endpoint(constants.WorkbenchEndpointImages),
				stubs.StubHTTPPartialGetConfig{
					ResponseStatus: http.StatusOK,
					ResponseBody:   `[invalid`,
				},
			),
			expected: describeExpected{
				displays:      []string{stubs.ExpectedDeviceAuthDisplay},
				describeCalls: 1,
				err: "" +
					"GET ${URL}/postgrest/images failed" +
					"\n Caused by:" +
					"\n\t* invalid character 'i' looking for beginning of value",
			},
		},
		{
			name: "invalid image builds",
			input: describeInput{
				resourceType: constants.ImageBuildsType,
			},
			mockEndpoints: getDescribeEndpoints(
				postgrest.Endpoint(constants.WorkbenchEndpointImageBuilds),
				stubs.StubHTTPPartialGetConfig{
					ResponseStatus: http.StatusOK,
					ResponseBody:   `[invalid`,
				},
			),
			expected: describeExpected{
				displays:      []string{stubs.ExpectedDeviceAuthDisplay},
				describeCalls: 1,
				err: "" +
					"GET ${URL}/postgrest/image_builds failed" +
					"\n Caused by:" +
					"\n\t* invalid character 'i' looking for beginning of value",
			},
		},
		{
			name: "invalid jobs",
			input: describeInput{
				resourceType: constants.JobsType,
			},
			mockEndpoints: getDescribeEndpoints(
				postgrest.Endpoint(constants.WorkbenchEndpointJobs),
				stubs.StubHTTPPartialGetConfig{
					ResponseStatus: http.StatusOK,
					ResponseBody:   `[invalid`,
				},
			),
			expected: describeExpected{
				displays:      []string{stubs.ExpectedDeviceAuthDisplay},
				describeCalls: 1,
				err: "" +
					"GET ${URL}/postgrest/jobs failed" +
					"\n Caused by:" +
					"\n\t* invalid character 'i' looking for beginning of value",
			},
		},
		{
			name: "invalid job submissions",
			input: describeInput{
				resourceType: constants.JobSubmissionType,
			},
			mockEndpoints: getDescribeEndpoints(
				postgrest.Endpoint(constants.WorkbenchEndpointJobRuns),
				stubs.StubHTTPPartialGetConfig{
					ResponseStatus: http.StatusOK,
					ResponseBody:   `[invalid`,
				},
			),
			expected: describeExpected{
				displays:      []string{stubs.ExpectedDeviceAuthDisplay},
				describeCalls: 1,
				err: "" +
					"GET ${URL}/postgrest/job_runs failed" +
					"\n Caused by:" +
					"\n\t* invalid character 'i' looking for beginning of value",
			},
		},
		{
			name: "invalid snapshots",
			input: describeInput{
				resourceType: constants.SnapshotsType,
			},
			mockEndpoints: getDescribeEndpoints(
				postgrest.Endpoint(constants.WorkbenchEndpointSnapshots),
				stubs.StubHTTPPartialGetConfig{
					ResponseStatus: http.StatusOK,
					ResponseBody:   `[invalid`,
				},
			),
			expected: describeExpected{
				displays:      []string{stubs.ExpectedDeviceAuthDisplay},
				describeCalls: 1,
				err: "" +
					"GET ${URL}/postgrest/snapshots failed" +
					"\n Caused by:" +
					"\n\t* invalid character 'i' looking for beginning of value",
			},
		},
		{
			name: "invalid users",
			input: describeInput{
				resourceType: constants.UsersType,
			},
			mockEndpoints: getDescribeEndpoints(
				auth0admin.Endpoint(constants.WorkbenchEndpointUsers),
				stubs.StubHTTPPartialGetConfig{
					ResponseStatus: http.StatusOK,
					ResponseBody:   `[invalid`,
				},
			),
			expected: describeExpected{
				displays:      []string{stubs.ExpectedDeviceAuthDisplay},
				describeCalls: 1,
				err: "" +
					"GET ${URL}/auth0Admin/users failed" +
					"\n Caused by:" +
					"\n\t* invalid character 'i' looking for beginning of value",
			},
		},
		{
			name: "invalid conflicts",
			input: describeInput{
				resourceType: constants.VolumeConflictsType,
			},
			mockEndpoints: getDescribeEndpoints(
				postgrest.Endpoint(constants.WorkbenchEndpointWorkspaceConflicts),
				stubs.StubHTTPPartialGetConfig{
					ResponseStatus: http.StatusOK,
					ResponseBody:   `[invalid`,
				},
			),
			expected: describeExpected{
				displays:      []string{stubs.ExpectedDeviceAuthDisplay},
				describeCalls: 1,
				err: "" +
					"GET ${URL}/postgrest/workspace_conflicts failed" +
					"\n Caused by:" +
					"\n\t* invalid character 'i' looking for beginning of value",
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
			mockCmd := describe.Command(context.Background(), afero.NewMemMapFs(), stubUI)
			cmd.AddRootFlags(mockCmd)
			mockCmd.SetArgs(getDescribeArgs(tt))
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
			expectedCallEndpointPrefix := "GET" + constants.PostgrestAPIEndpoint
			if tt.input.resourceType == constants.UsersType || tt.input.resourceType == constants.UserType {
				expectedCallEndpointPrefix = "GET" + auth0admin.Endpoint(constants.WorkbenchEndpointUsers)
			}
			require.Equal(
				t,
				tt.expected.describeCalls,
				hitCounts.CountPrefix(expectedCallEndpointPrefix),
				"describe calls: %v", hitCounts,
			)
		})
	}
}

func TestSupportedResourceTypes(t *testing.T) {
	tests := []struct {
		name          string
		expectedTypes []string
	}{
		{
			name: "returns sorted supported describe resource types",
			expectedTypes: []string{
				constants.DataplanesType,
				constants.ImagesType,
				constants.ImageBuildsType,
				constants.JobsType,
				constants.JobSubmissionsType,
				constants.SnapshotsType,
				constants.UsersType,
				constants.VolumesType,
				constants.VolumeConflictsType,
				constants.WorkbenchesType,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := describe.SupportedResourceTypes()
			require.Equal(t, tt.expectedTypes, actual)
		})
	}
}

func getDescribeArgs(
	tt describeTestCase,
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

func getDescribeEndpoints(
	urlUnderTest string,
	config ...stubs.StubHTTPPartialGetConfig,
) []stubs.StubHTTPEndpoint {
	endpoints := stubs.GetStubAuthEndpoints()
	for _, cfg := range config {
		endpoints = append(endpoints,
			stubs.StubHTTPEndpoint{
				Request: stubs.StubHTTPRequest{
					Method:          http.MethodGet,
					URL:             urlUnderTest,
					ExpectedHeaders: stubs.StubAuthHeaders,
					ExpectedQuery:   cfg.ExpectedQuery,
				},
				Response: stubs.StubHTTPResponse{
					Status: cfg.ResponseStatus,
					Body:   cfg.ResponseBody,
				},
			})
	}
	return endpoints
}
