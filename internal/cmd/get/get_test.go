// Copyright © 2026, SAS Institute Inc., Cary, NC, USA.  All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package get_test

import (
	"context"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/sassoftware/wbctl/internal/auth0admin"
	"github.com/sassoftware/wbctl/internal/cmd"
	"github.com/sassoftware/wbctl/internal/cmd/get"
	"github.com/sassoftware/wbctl/internal/constants"
	"github.com/sassoftware/wbctl/internal/postgrest"
	"github.com/sassoftware/wbctl/internal/stubs"
	"github.com/sassoftware/wbctl/internal/timeutil"
	"github.com/spf13/afero"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
)

const (
	getExpectedDisplayApplicationJSON = `[
    {
        "id": "resource-1"
    }
]`
	getExpectedDisplayPluralApplicationJSON = `[
    {
        "id": "resource-1"
    },
    {
        "id": "resource-2"
    }
]`
	getExpectedDisplayDataplaneJSON = `[
    {
        "id": "resource-1",
        "name": "Resource One"
    }
]`
	getExpectedDisplayEventJSON = `[
    {
        "id": "resource-1"
    }
]`
	getExpectedDisplayIDJSON = `[
    {
        "id": "resource-1"
    }
]`
	getExpectedDiaplayUsersJSON = `{
    "version": 2,
    "accept": "application/vnd.sas.auth0admin.user+json",
    "count": 6,
    "start": 0,
    "limit": 20,
    "name": "users",
    "items": [
        {
            "version": 1,
            "id": "auth0|6a986b2efeca03ccb89eca7b",
            "databaseId": "7d40724d-2cbf-50f0-8482-2b11afa9603c",
            "name": "pupadmin1 Test User",
            "email": "pupadmin1@sas-test-mickey2.com",
            "creationTimeStamp": "2025-05-08T00:00:00Z",
            "roles": [
                {
                    "id": "rol_CBru4D4ruLp6PN7t",
                    "name": "workbench_admin"
                },
                {
                    "id": "rol_c8teNEL6bBzCYU5y",
                    "name": "org_admin"
                }
            ]
        },
        {
            "version": 1,
            "id": "auth0|6a986b2a9fcd628b746cee66",
            "databaseId": "7a056a78-d94a-5eed-9a6f-a077016b13a1",
            "name": "pupadmin2 Test User",
            "email": "pupadmin2@sas-test-mickey2.com",
            "creationTimeStamp": "2025-05-08T00:00:00Z",
            "roles": [
                {
                    "id": "rol_c8teNEL6bBzCYU5y",
                    "name": "org_admin"
                },
                {
                    "id": "rol_CBru4D4ruLp6PN7t",
                    "name": "workbench_admin"
                }
            ]
        },
        {
            "version": 1,
            "id": "auth0|6a986b2c04816dc0574018dc",
            "databaseId": "6e6a0a0b-270c-5a68-9bc4-f904bb5deee9",
            "name": "pupadmin3 Test User",
            "email": "pupadmin3@sas-test-mickey2.com",
            "creationTimeStamp": "2025-05-08T00:00:00Z",
            "roles": [
                {
                    "id": "rol_CBru4D4ruLp6PN7t",
                    "name": "workbench_admin"
                },
                {
                    "id": "rol_c8teNEL6bBzCYU5y",
                    "name": "org_admin"
                }
            ]
        },
        {
            "version": 1,
            "id": "auth0|6a986b2a526605e8b186210b",
            "databaseId": "e60cb636-0aa5-586b-9441-0d24ae0e69b7",
            "name": "pupusr1 Test User",
            "email": "pupusr1@sas-test-mickey2.com",
            "creationTimeStamp": "2025-05-08T00:00:00Z",
            "roles": [
                {
                    "id": "rol_G2Xo84BaiLvnBLvD",
                    "name": "workbench_user"
                }
            ]
        },
        {
            "version": 1,
            "id": "auth0|6a986b2c1f42bc857feb9738",
            "databaseId": "6c1df094-36d4-5fc1-88aa-6131c4408bd5",
            "name": "pupusr2 Test User",
            "email": "pupusr2@sas-test-mickey2.com",
            "creationTimeStamp": "2025-05-08T00:00:00Z",
            "roles": [
                {
                    "id": "rol_G2Xo84BaiLvnBLvD",
                    "name": "workbench_user"
                }
            ]
        },
        {
            "version": 1,
            "id": "auth0|6a986b2e9fcd628b746cee6f",
            "databaseId": "1dcd256b-3deb-50a1-839b-2d102b9bf1d5",
            "name": "pupusr3 Test User",
            "email": "pupusr3@sas-test-mickey2.com",
            "creationTimeStamp": "2025-05-08T00:00:00Z",
            "roles": [
                {
                    "id": "rol_G2Xo84BaiLvnBLvD",
                    "name": "workbench_user"
                }
            ]
        }
    ]
}`
	getExpectedDisplayInitializerJSON = `[
    {
        "id": "resource-1",
        "name": "Resource One"
    }
]`
	getExpectedDisplayPingJSON = `[
    {}
]`
	getExpectedDisplaySingularConflictJSON = `[
    {
        "conflict_id": "conflict-2",
        "conflict_name": "Conflict Two",
        "conflict_type": "file",
        "controller_id": "controller-1",
        "id": "conflict-1",
        "name": "Conflict One",
        "type": "file",
        "workspace_id": "workspace-1",
        "workspace_name": "Workspace One"
    }
]`
	getExpectedDisplaySnapshotJSON = `[
    {
        "id": "resource-1",
        "name": "Resource One"
    }
]`
	getExpectedDisplayStorageclassJSON = `[
    {
        "id": "resource-1"
    }
]`
	getExpectedDisplayWorkspaceJSON = `[
    {
        "id": "resource-1",
        "name": "Resource One",
        "pvc_status": {
            "status": {}
        }
    }
]`
	getExpectedResponsePluralApplication = `
[
  {
    "id": "app-1",
    "display_name": "RStudio",
    "description": "R app",
    "services": [
      "svc-a",
      "svc-b"
    ]
  },
  {
    "id": "app-2",
    "display_name": "VSCode",
    "description": "VS Code Ap",
    "services": [
      "svc-c",
      "svc-d"
    ]
  }
]
`
	getExpectedResponsePluralDataplane = `
[
  {
    "id": "dp-1",
    "name": "Dataplane 1",
    "description": "first",
    "message": "ok",
    "created_at": null
  },
  {
    "id": "dp-2",
    "name": "Dataplane 2",
    "description": "second",
    "message": "ok",
    "created_at": null
  }
]
`
	getExpectedResponsePluralEvent = `
[
  {
    "id": "evt-1",
    "type": "INFO",
    "resource_type": "workbench",
    "kind": "audit",
    "action": "create",
    "note": "line1\nline2",
    "reason": "manual",
    "created_at": null
  },
  {
    "id": "evt-2",
    "type": "INFO",
    "resource_type": "workbench",
    "kind": "audit",
    "action": "create",
    "note": "line3\nline4",
    "reason": "manual",
    "created_at": null
  }
]
`
	getExpectedResponsePluralImageBuild = `
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
    "controller_id": "",
    "image_id": "img-2",
    "created_at": null
  }
]
`
	getExpectedResponsePluralImage = `
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
    "message": "ok",
    "created_at": null
  },
  {
    "id": "img-2",
    "name": "image-2",
    "description": "My Application Image",
    "image_name": "myapp",
    "image_registry": "registry.example.com",
    "image_tag": "v2.0",
    "controller_id": "ctrl-2",
    "latest_image_build_id": "ib-2",
    "current_state": "Ready",
    "message": "ok",
    "created_at": null
  }
]
`
	getExpectedResponsePluralInitializer = `
[
  {
    "id": "init-1",
    "name": "git-clone"
  },
  {
    "id": "init-2",
    "name": "git-clone"
  }
]
`
	getExpectedResponsePluralJobSubmission = `
[
  {
    "id": "js-1",
    "controller_id": "",
    "job_name": "daily-job",
    "log_url": "https://logs/1",
    "current_state": "Succeeded",
    "message": "done",
    "runtime": 12.34,
    "created_at": null
  },
  {
    "id": "js-2",
    "controller_id": "",
    "job_name": "daily-job",
    "log_url": "https://logs/2",
    "current_state": "Succeeded",
    "message": "done",
    "runtime": 22.34,
    "created_at": null
  }
]
`
	getExpectedResponsePluralJob = `
[
  {
    "id": "job-1",
    "name": "daily-job",
    "application_id": "app-1",
    "services": [
      {
        "id": "svc-a"
      },
      {
        "id": "svc-b"
      }
    ],
    "created_at": null
  },
  {
    "id": "job-2",
    "name": "daily-job",
    "application_id": "app-2",
    "services": [
      {
        "id": "svc-a"
      },
      {
        "id": "svc-b"
      }
    ],
    "created_at": null
  }
]
`
	getExpectedResponsePluralNodepool = `
[
  {
    "id": "np-1",
    "display_name": "Nodepool A",
    "description": "first"
  },
  {
    "id": "np-2",
    "display_name": "Nodepool B",
    "description": "second"
  }
]
`
	getExpectedResponsePluralPing = `
[
  {
    "dataplane_id": "dp-a",
    "request_at": "2025-05-08T00:00:00Z",
    "status_updated_at": "2025-05-08T00:00:01Z",
    "status_state": "ok",
    "status_message": "fast"
  },
  {
    "dataplane_id": "dp-b",
    "request_at": "2025-05-08T00:00:00Z",
    "status_updated_at": "2025-05-08T00:00:01Z",
    "status_state": "ok",
    "status_message": "fast"
  }
]
`
	getExpectedResponsePluralService = `
[
  {
    "id": "svc-1",
    "name": "code-server",
    "type": "workbench"
  },
  {
    "id": "svc-2",
    "name": "code-server",
    "type": "workbench"
  }
]
`
	getExpectedResponsePluralSnapshot = `
[
  {
    "id": "snap-1",
    "name": "snapshot-1",
    "description": "first snapshot",
    "spec": {
      "source": {}
    },
    "volume_status": {
      "status": {}
    },
    "created_at": null
  },
  {
    "id": "snap-2",
    "name": "snapshot-2",
    "description": "first snapshot",
    "spec": {
      "source": {}
    },
    "volume_status": {
      "status": {}
    },
    "created_at": null
  }
]
`
	getExpectedResponsePluralStorageclass = `
[
  {
    "id": "sc-1",
    "display_name": "standard",
    "description": "default class",
    "properties": {
      "default": "true"
    }
  },
  {
    "id": "sc-2",
    "display_name": "standard",
    "description": "default class",
    "properties": {
      "default": "true"
    }
  }
]
	`
	getExpectedResponsePluralUser = `
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
      "creationTimeStamp": "2025-05-08T00:00:00Z",
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
      "creationTimeStamp": "2025-05-08T00:00:00Z",
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
      "creationTimeStamp": "2025-05-08T00:00:00Z",
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
      "creationTimeStamp": "2025-05-08T00:00:00Z",
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
      "creationTimeStamp": "2025-05-08T00:00:00Z",
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
      "creationTimeStamp": "2025-05-08T00:00:00Z",
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
	getExpectedResponsePluralWorkbench = `
[
  {
    "id": "wb-1",
    "controller_id": "",
    "name": "wb-one",
    "application_id": "app-1",
    "services": [
      {
        "id": "svc-a"
      },
      {
        "id": "svc-b"
      }
    ],
    "url": "https://wb/1",
    "current_state": "Running",
    "message": "ok",
    "created_at": null
  },
  {
    "id": "wb-2",
    "controller_id": "",
    "name": "wb-one",
    "application_id": "app-2",
    "services": [
      {
        "id": "svc-a"
      },
      {
        "id": "svc-b"
      }
    ],
    "url": "https://wb/2",
    "current_state": "Running",
    "message": "ok",
    "created_at": null
  }
]
	`
	getExpectedResponsePluralWorkspace = `
[
  {
    "id": "ws-1",
    "controller_id": "",
    "name": "workspace-1",
    "storage_size_g": 10,
    "storage_spec": {
      "storageClassName": "standard"
    },
    "properties": {
      "HOME_DIRECTORY": "true"
    },
    "state": "Running",
    "message": "ok",
    "created_at": null
  },
  {
    "id": "ws-2",
    "controller_id": "",
    "name": "workspace-2",
    "storage_size_g": 20,
    "storage_spec": {
      "storageClassName": "standard"
    },
    "properties": {
      "HOME_DIRECTORY": "true"
    },
    "state": "Running",
    "message": "ok",
    "created_at": null
  }
]
	`
	getExpectedResponseGenericResource = `
[
  {
    "id": "resource-1",
    "name": "Resource One"
  }
]
`
	getExpectedResponseUsers = `
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
      "creationTimeStamp": "2025-05-08T00:00:00Z",
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
      "creationTimeStamp": "2025-05-08T00:00:00Z",
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
      "creationTimeStamp": "2025-05-08T00:00:00Z",
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
      "creationTimeStamp": "2025-05-08T00:00:00Z",
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
      "creationTimeStamp": "2025-05-08T00:00:00Z",
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
      "creationTimeStamp": "2025-05-08T00:00:00Z",
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

	getExpectedPluralResponseGenericResource = `
[
  {
    "id": "resource-1",
    "name": "Resource One"
  },
  {
    "id": "resource-2",
    "name": "Resource Two"
  }
]
`
	getExpectedResponsePluralConflicts = `
[
  {
    "id": "conflict-11",
    "name": "Conflict Eleven",
    "type": "directory",
    "conflict_id": "conflict-22",
    "conflict_name": "Conflict TwentyTwo",
    "conflict_type": "directory",
    "workspace_id": "workspace-11",
    "workspace_name": "Workspace Eleven"
  },
  {
    "id": "conflict-12",
    "name": "Conflict Twelve",
    "type": "directory",
    "conflict_id": "conflict-22",
    "conflict_name": "Conflict TwentyTwo",
    "conflict_type": "directory",
    "workspace_id": "workspace-12",
    "workspace_name": "Workspace Twelve"
  }
]
`
	getExpectedResponseSingularConflict = `
[
  {
    "id": "conflict-1",
    "controller_id": "controller-1",
    "name": "Conflict One",
    "type": "file",
    "conflict_id": "conflict-2",
    "conflict_name": "Conflict Two",
    "conflict_type": "file",
    "workspace_id": "workspace-1",
    "workspace_name": "Workspace One"
  }
]
`

	getExpectedTablePluralApplication = "ID\tNAME\tDESCRIPTION\tSERVICES\n" +
		"app-1\tRStudio\tR app\tsvc-a, svc-b\n" +
		"app-2\tVSCode\tVS Code Ap\tsvc-c, svc-d"
	getExpectedTablePluralConflict = "ID\tNAME\tTYPE\tCONFLICT_ID\tCONFLICT_NAME\tCONFLICT_TYPE\t" +
		"VOLUME_ID\tVOLUME_NAME\n" +
		"conflict-11\tConflict Eleven\tdirectory\tconflict-22\tConflict TwentyTwo\tdirectory\t" +
		"workspace-11\tWorkspace Eleven\n" +
		"conflict-12\tConflict Twelve\tdirectory\tconflict-22\tConflict TwentyTwo\tdirectory\tworkspace-12\tWorkspace Twelve"
	getExpectedTablePluralDataplane = "ID\tNAME\tDESCRIPTION\tMESSAGE\tAGE\n" +
		"dp-1\tDataplane 1\tfirst\tok\t\n" +
		"dp-2\tDataplane 2\tsecond\tok\t"
	getExpectedTablePluralEvent = "ID\tTYPE\tRESOURCE_TYPE\tKIND\tACTION\tNOTE\tREASON\tAGE\n" +
		"evt-1\tINFO\tworkbench\taudit\tcreate\tline1 line2\tmanual\t\n" +
		"evt-2\tINFO\tworkbench\taudit\tcreate\tline3 line4\tmanual\t"
	getExpectedTablePluralImage = "ID\tCONTROLLER_ID\tNAME\tDESCRIPTION\tIMAGE\tIMAGE_BUILD_ID\tSTATE\tMESSAGE\tAGE\n" +
		"img-1\tctrl-1\timage-1\tMy Application Image\tregistry.example.com/myapp:v1.0\tib-1\tReady\tok\t\n" +
		"img-2\tctrl-2\timage-2\tMy Application Image\tregistry.example.com/myapp:v2.0\tib-2\tReady\tok\t"
	getExpectedTablePluralImageBuild = "ID\tCONTROLLER_ID\tNAME\tIMAGE_ID\tAGE\n" +
		"ib-1\t<unknown>\tbuild-1\timg-1\t\n" +
		"ib-2\t<unknown>\tbuild-2\timg-2\t"
	getExpectedTablePluralInitializer = "ID\tNAME\tTYPE\n" +
		"init-1\tgit-clone\n" +
		"init-2\tgit-clone"
	getExpectedTablePluralJob = "ID\tNAME\tAPPLICATION_ID\tSERVICES\tAGE\n" +
		"job-1\tdaily-job\tapp-1\tsvc-a, svc-b\t\n" +
		"job-2\tdaily-job\tapp-2\tsvc-a, svc-b\t"
	getExpectedTablePluralJobSubmission = "ID\tCONTROLLER_ID\tJOB_NAME\tLOG_URL\tSTATE\tMESSAGE\tRUNTIME\tAGE\n" +
		"js-1\t<unknown>\tdaily-job\thttps://logs/1\tSucceeded\tdone\t12.34s\t\n" +
		"js-2\t<unknown>\tdaily-job\thttps://logs/2\tSucceeded\tdone\t22.34s\t"
	getExpectedTablePluralNodepool = "ID\tNAME\tDESCRIPTION\n" +
		"np-1\tNodepool A\tfirst\n" +
		"np-2\tNodepool B\tsecond"
	getExpectedTablePluralPing = "DATAPLANE_ID\tREQUEST_AT\tRESPONSE_AT\tRESPONSE_TIME\tSTATE\tMESSAGE\tAGE\n" +
		"dp-a\t2025-05-08T00:00:00Z\t2025-05-08T00:00:01Z\t1s\tok\tfast\t0s\n" +
		"dp-b\t2025-05-08T00:00:00Z\t2025-05-08T00:00:01Z\t1s\tok\tfast\t0s"
	getExpectedTablePluralService = "ID\tNAME\tTYPE\n" +
		"svc-1\tcode-server\tworkbench\n" +
		"svc-2\tcode-server\tworkbench"
	getExpectedTablePluralSnapshot = "ID\tCONTROLLER_ID\tNAME\tDESCRIPTION\tSOURCE\tSIZE\tREADY\tAGE\n" +
		"snap-1\t<unknown>\tsnapshot-1\tfirst snapshot\t\t<unknown>\tfalse\t\n" +
		"snap-2\t<unknown>\tsnapshot-2\tfirst snapshot\t\t<unknown>\tfalse\t"
	getExpectedTablePluralStorageclass = "ID\tNAME\tDESCRIPTION\tDEFAULT\n" +
		"sc-1\tstandard\tdefault class\ttrue\n" +
		"sc-2\tstandard\tdefault class\ttrue"
	// nolint:lll
	getExpectedTablePluralUser = "ID\tNAME\tEMAIL\tROLES\tMETADATA\tAGE\n" +
		"auth0|6a986b2efeca03ccb89eca7b\tpupadmin1 Test User\tpupadmin1@sas-test-mickey2.com\torg_admin, workbench_admin\t\t0s\n" +
		"auth0|6a986b2a9fcd628b746cee66\tpupadmin2 Test User\tpupadmin2@sas-test-mickey2.com\torg_admin, workbench_admin\t\t0s\n" +
		"auth0|6a986b2c04816dc0574018dc\tpupadmin3 Test User\tpupadmin3@sas-test-mickey2.com\torg_admin, workbench_admin\t\t0s\n" +
		"auth0|6a986b2a526605e8b186210b\tpupusr1 Test User\tpupusr1@sas-test-mickey2.com\tworkbench_user\t\t0s\n" +
		"auth0|6a986b2c1f42bc857feb9738\tpupusr2 Test User\tpupusr2@sas-test-mickey2.com\tworkbench_user\t\t0s\n" +
		"auth0|6a986b2e9fcd628b746cee6f\tpupusr3 Test User\tpupusr3@sas-test-mickey2.com\tworkbench_user\t\t0s"

	getExpectedTablePluralWorkbench = "ID\tCONTROLLER_ID\tNAME\tAPPLICATION_ID\tSERVICES\tURL\tSTATE\tMESSAGE\tAGE\n" +
		"wb-1\t<unknown>\twb-one\tapp-1\tsvc-a, svc-b\thttps://wb/1\tRunning\tok\t\n" +
		"wb-2\t<unknown>\twb-one\tapp-2\tsvc-a, svc-b\thttps://wb/2\tRunning\tok\t"
	getExpectedTablePluralWorkspace = "ID\tCONTROLLER_ID\tNAME\tSIZE\tHOME_DIR\tSTORAGECLASSNAME\tSTATE\tMESSAGE\tAGE\n" +
		"ws-1\t<unknown>\tworkspace-1\t10Gi\ttrue\tstandard\tRunning\tok\t\n" +
		"ws-2\t<unknown>\tworkspace-2\t20Gi\ttrue\tstandard\tRunning\tok\t"
	expectedInvalidEndpointErr = "failed to get device token" +
		"\n Caused by:" +
		"\n\t* parse \"a\\nb\": net/url: invalid control character in URL"
	expectedInvalidResourceErr = "" +
		"GET ${URL}/postgrest/workspaces failed" +
		"\n Caused by:" +
		"\n\t* invalid character 'i' looking for beginning of value"
)

type getInput struct {
	endpoint     string
	outputValue  string
	resourceIDs  []string
	resourceType string
}

type getExpected struct {
	err          string
	displays     []string
	warns        []string
	getCalls     int
	tableStrings []string
}

type getTestCase struct {
	name          string
	input         getInput
	mockEndpoints []stubs.StubHTTPEndpoint
	expected      getExpected
}

func TestGetCommand(t *testing.T) {
	fixedNow := time.Date(2025, time.May, 8, 0, 0, 0, 0, time.UTC)
	timeutil.SetNow(func() time.Time { return fixedNow })
	t.Cleanup(func() { timeutil.SetNow(time.Now) })

	tests := []getTestCase{
		// ==== JSON SINGLE CASES
		{
			name: "json output for singular " + constants.VolumeConflictType + " resource",
			input: getInput{
				outputValue:  constants.JSONOutputValue,
				resourceType: constants.VolumeConflictType,
				resourceIDs:  []string{"conflict-1", "conflict-2"},
			},
			mockEndpoints: getGetEndpoints(
				postgrest.Endpoint(constants.WorkbenchEndpointWorkspaceConflicts),
				stubs.StubHTTPPartialGetConfig{
					ExpectedQuery:  "id=in.(conflict-1,conflict-2)",
					ResponseStatus: http.StatusOK,
					ResponseBody:   getExpectedResponseSingularConflict,
				},
			),
			expected: getExpected{
				getCalls: 1,
				displays: []string{stubs.ExpectedDeviceAuthDisplay, getExpectedDisplaySingularConflictJSON},
			},
		},
		{
			name: "json output for singular " + constants.ApplicationType + " resource",
			input: getInput{
				outputValue:  constants.JSONOutputValue,
				resourceType: constants.ApplicationType,
			},
			mockEndpoints: getGetEndpoints(
				postgrest.Endpoint(constants.WorkbenchEndpointDataplaneWorkbenchApplications),
				stubs.StubHTTPPartialGetConfig{
					ResponseStatus: http.StatusOK,
					ResponseBody:   getExpectedResponseGenericResource,
				},
			),
			expected: getExpected{
				displays: []string{stubs.ExpectedDeviceAuthDisplay, getExpectedDisplayApplicationJSON},
				getCalls: 1,
			},
		},
		{
			name: "json output for singular " + constants.DataplaneType + " resource",
			input: getInput{
				outputValue:  constants.JSONOutputValue,
				resourceType: constants.DataplaneType,
			},
			mockEndpoints: getGetEndpoints(
				postgrest.Endpoint(constants.WorkbenchEndpointDataplanes),
				stubs.StubHTTPPartialGetConfig{
					ResponseStatus: http.StatusOK,
					ResponseBody:   getExpectedResponseGenericResource,
				},
			),
			expected: getExpected{
				displays: []string{stubs.ExpectedDeviceAuthDisplay, getExpectedDisplayDataplaneJSON},
				getCalls: 1,
			},
		},
		{
			name: "json output for singular " + constants.EventType + " resource",
			input: getInput{
				outputValue:  constants.JSONOutputValue,
				resourceType: constants.EventType,
			},
			mockEndpoints: getGetEndpoints(
				postgrest.Endpoint(constants.WorkbenchEndpointEvents),
				stubs.StubHTTPPartialGetConfig{
					ResponseStatus: http.StatusOK,
					ResponseBody:   getExpectedResponseGenericResource,
				},
			),
			expected: getExpected{
				getCalls: 1,
				displays: []string{stubs.ExpectedDeviceAuthDisplay, getExpectedDisplayEventJSON},
			},
		},
		{
			name: "json output for singular " + constants.ImageBuildType + " resource",
			input: getInput{
				outputValue:  constants.JSONOutputValue,
				resourceType: constants.ImageBuildType,
			},
			mockEndpoints: getGetEndpoints(
				postgrest.Endpoint(constants.WorkbenchEndpointImageBuilds),
				stubs.StubHTTPPartialGetConfig{
					ResponseStatus: http.StatusOK,
					ResponseBody:   getExpectedResponseGenericResource,
				},
			),
			expected: getExpected{
				getCalls: 1,
				displays: []string{stubs.ExpectedDeviceAuthDisplay, getExpectedDisplayDataplaneJSON},
			},
		},
		{
			name: "json output for singular " + constants.ImageType + " resource",
			input: getInput{
				outputValue:  constants.JSONOutputValue,
				resourceType: constants.ImageType,
			},
			mockEndpoints: getGetEndpoints(
				postgrest.Endpoint(constants.WorkbenchEndpointImages),
				stubs.StubHTTPPartialGetConfig{
					ResponseStatus: http.StatusOK,
					ResponseBody:   getExpectedResponseGenericResource,
				},
			),
			expected: getExpected{
				getCalls: 1,
				displays: []string{stubs.ExpectedDeviceAuthDisplay, getExpectedDisplayDataplaneJSON},
			},
		},
		{
			name: "json output for singular " + constants.InitializerType + " resource",
			input: getInput{
				outputValue:  constants.JSONOutputValue,
				resourceType: constants.InitializerType,
			},
			mockEndpoints: getGetEndpoints(
				postgrest.Endpoint(constants.WorkbenchEndpointDataplaneWorkspaceServices),
				stubs.StubHTTPPartialGetConfig{
					ResponseStatus: http.StatusOK,
					ResponseBody:   getExpectedResponseGenericResource,
				},
			),
			expected: getExpected{
				getCalls: 1,
				displays: []string{stubs.ExpectedDeviceAuthDisplay, getExpectedDisplayInitializerJSON},
			},
		},
		{
			name: "json output for singular " + constants.JobSubmissionType + " resource",
			input: getInput{
				outputValue:  constants.JSONOutputValue,
				resourceType: constants.JobSubmissionType,
			},
			mockEndpoints: getGetEndpoints(
				postgrest.Endpoint(constants.WorkbenchEndpointJobRuns),
				stubs.StubHTTPPartialGetConfig{
					ResponseStatus: http.StatusOK,
					ResponseBody:   getExpectedResponseGenericResource,
				},
			),
			expected: getExpected{
				getCalls: 1,
				displays: []string{stubs.ExpectedDeviceAuthDisplay, getExpectedDisplayIDJSON},
			},
		},
		{
			name: "json output for singular " + constants.JobType + " resource",
			input: getInput{
				outputValue:  constants.JSONOutputValue,
				resourceType: constants.JobType,
			},
			mockEndpoints: getGetEndpoints(
				postgrest.Endpoint(constants.WorkbenchEndpointJobs),
				stubs.StubHTTPPartialGetConfig{
					ResponseStatus: http.StatusOK,
					ResponseBody:   getExpectedResponseGenericResource,
				},
			),
			expected: getExpected{
				getCalls: 1,
				displays: []string{stubs.ExpectedDeviceAuthDisplay, getExpectedDisplayDataplaneJSON},
			},
		},
		{
			name: "json output for singular " + constants.NodepoolType + " resource",
			input: getInput{
				outputValue:  constants.JSONOutputValue,
				resourceType: constants.NodepoolType,
			},
			mockEndpoints: getGetEndpoints(
				postgrest.Endpoint(constants.WorkbenchEndpointDataplaneNodepools),
				stubs.StubHTTPPartialGetConfig{
					ResponseStatus: http.StatusOK,
					ResponseBody:   getExpectedResponseGenericResource,
				},
			),
			expected: getExpected{
				getCalls: 1,
				displays: []string{stubs.ExpectedDeviceAuthDisplay, getExpectedDisplayIDJSON},
			},
		},
		{
			name: "json output for singular " + constants.PingType + " resource",
			input: getInput{
				outputValue:  constants.JSONOutputValue,
				resourceType: constants.PingType,
			},
			mockEndpoints: getGetEndpoints(
				postgrest.Endpoint(constants.WorkbenchEndpointPings),
				stubs.StubHTTPPartialGetConfig{
					ResponseStatus: http.StatusOK,
					ResponseBody:   getExpectedResponseGenericResource,
				},
			),
			expected: getExpected{
				getCalls: 1,
				displays: []string{stubs.ExpectedDeviceAuthDisplay, getExpectedDisplayPingJSON},
			},
		},
		{
			name: "json output for singular " + constants.ServiceType + " resource",
			input: getInput{
				outputValue:  constants.JSONOutputValue,
				resourceType: constants.ServiceType,
			},
			mockEndpoints: getGetEndpoints(
				postgrest.Endpoint(constants.WorkbenchEndpointDataplaneWorkbenchServices),
				stubs.StubHTTPPartialGetConfig{
					ResponseStatus: http.StatusOK,
					ResponseBody:   getExpectedResponseGenericResource,
				},
			),
			expected: getExpected{
				getCalls: 1,
				displays: []string{stubs.ExpectedDeviceAuthDisplay, getExpectedDisplayDataplaneJSON},
			},
		},
		{
			name: "json output for singular " + constants.SnapshotType + " resource",
			input: getInput{
				outputValue:  constants.JSONOutputValue,
				resourceType: constants.SnapshotType,
			},
			mockEndpoints: getGetEndpoints(
				postgrest.Endpoint(constants.WorkbenchEndpointSnapshots),
				stubs.StubHTTPPartialGetConfig{
					ResponseStatus: http.StatusOK,
					ResponseBody:   getExpectedResponseGenericResource,
				},
			),
			expected: getExpected{
				getCalls: 1,
				displays: []string{stubs.ExpectedDeviceAuthDisplay, getExpectedDisplaySnapshotJSON},
			},
		},
		{
			name: "json output for singular " + constants.StorageclassType + " resource",
			input: getInput{
				outputValue:  constants.JSONOutputValue,
				resourceType: constants.StorageclassType,
			},
			mockEndpoints: getGetEndpoints(
				postgrest.Endpoint(constants.WorkbenchEndpointDataplaneStorageclasses),
				stubs.StubHTTPPartialGetConfig{
					ResponseStatus: http.StatusOK,
					ResponseBody:   getExpectedResponseGenericResource,
				},
			),
			expected: getExpected{
				getCalls: 1,
				displays: []string{stubs.ExpectedDeviceAuthDisplay, getExpectedDisplayStorageclassJSON},
			},
		},
		{
			name: "json output for singular " + constants.UserType + " resource",
			input: getInput{
				outputValue:  constants.JSONOutputValue,
				resourceType: constants.UserType,
			},
			mockEndpoints: getGetEndpoints(
				auth0admin.Endpoint(constants.WorkbenchEndpointUsers),
				stubs.StubHTTPPartialGetConfig{
					ResponseStatus: http.StatusOK,
					ResponseBody:   getExpectedResponseUsers,
				},
			),
			expected: getExpected{
				getCalls: 1,
				displays: []string{stubs.ExpectedDeviceAuthDisplay, getExpectedDiaplayUsersJSON},
			},
		},
		{
			name: "json output for singular " + constants.WorkbenchType + " resource",
			input: getInput{
				outputValue:  constants.JSONOutputValue,
				resourceType: constants.WorkbenchType,
			},
			mockEndpoints: getGetEndpoints(
				postgrest.Endpoint(constants.WorkbenchEndpointWorkbenches),
				stubs.StubHTTPPartialGetConfig{
					ResponseStatus: http.StatusOK,
					ResponseBody:   getExpectedResponseGenericResource,
				},
			),
			expected: getExpected{
				getCalls: 1,
				displays: []string{stubs.ExpectedDeviceAuthDisplay, getExpectedDisplayDataplaneJSON},
			},
		},
		{
			name: "json output for singular " + constants.VolumeType + " resource",
			input: getInput{
				outputValue:  constants.JSONOutputValue,
				resourceType: constants.VolumeType,
			},
			mockEndpoints: getGetEndpoints(
				postgrest.Endpoint(constants.WorkbenchEndpointWorkspaces),
				stubs.StubHTTPPartialGetConfig{
					ResponseStatus: http.StatusOK,
					ResponseBody:   getExpectedResponseGenericResource,
				},
			),
			expected: getExpected{
				getCalls: 1,
				displays: []string{stubs.ExpectedDeviceAuthDisplay, getExpectedDisplayWorkspaceJSON},
			},
		},

		// ==== JSON PLURAL CASES
		{
			name: "json output for plural " + constants.ApplicationsType + " resource",
			input: getInput{
				outputValue:  constants.JSONOutputValue,
				resourceType: constants.ApplicationsType,
			},
			mockEndpoints: getGetEndpoints(
				postgrest.Endpoint(constants.WorkbenchEndpointDataplaneWorkbenchApplications),
				stubs.StubHTTPPartialGetConfig{
					ResponseStatus: http.StatusOK,
					ResponseBody:   getExpectedPluralResponseGenericResource,
				},
			),
			expected: getExpected{
				getCalls: 1,
				displays: []string{stubs.ExpectedDeviceAuthDisplay, getExpectedDisplayPluralApplicationJSON},
			},
		},
		{
			name: "json output for plural " + constants.DataplanesType + " resource",
			input: getInput{
				outputValue:  constants.JSONOutputValue,
				resourceType: constants.DataplanesType,
			},
			mockEndpoints: getGetEndpoints(
				postgrest.Endpoint(constants.WorkbenchEndpointDataplanes),
				stubs.StubHTTPPartialGetConfig{
					ResponseStatus: http.StatusOK,
					ResponseBody:   getExpectedResponseGenericResource,
				},
			),
			expected: getExpected{
				getCalls: 1,
				displays: []string{stubs.ExpectedDeviceAuthDisplay, getExpectedDisplayDataplaneJSON},
			},
		},
		{
			name: "json output for plural " + constants.EventsType + " resource",
			input: getInput{
				outputValue:  constants.JSONOutputValue,
				resourceType: constants.EventsType,
			},
			mockEndpoints: getGetEndpoints(
				postgrest.Endpoint(constants.WorkbenchEndpointEvents),
				stubs.StubHTTPPartialGetConfig{
					ResponseStatus: http.StatusOK,
					ResponseBody:   getExpectedResponseGenericResource,
				},
			),
			expected: getExpected{
				getCalls: 1,
				displays: []string{stubs.ExpectedDeviceAuthDisplay, getExpectedDisplayEventJSON},
			},
		},
		{
			name: "json output for plural " + constants.ImageBuildsType + " resource",
			input: getInput{
				outputValue:  constants.JSONOutputValue,
				resourceType: constants.ImageBuildsType,
			},
			mockEndpoints: getGetEndpoints(
				postgrest.Endpoint(constants.WorkbenchEndpointImageBuilds),
				stubs.StubHTTPPartialGetConfig{
					ResponseStatus: http.StatusOK,
					ResponseBody:   getExpectedResponseGenericResource,
				},
			),
			expected: getExpected{
				getCalls: 1,
				displays: []string{stubs.ExpectedDeviceAuthDisplay, getExpectedDisplayDataplaneJSON},
			},
		},
		{
			name: "json output for plural " + constants.ImagesType + " resource",
			input: getInput{
				outputValue:  constants.JSONOutputValue,
				resourceType: constants.ImagesType,
			},
			mockEndpoints: getGetEndpoints(
				postgrest.Endpoint(constants.WorkbenchEndpointImages),
				stubs.StubHTTPPartialGetConfig{
					ResponseStatus: http.StatusOK,
					ResponseBody:   getExpectedResponseGenericResource,
				},
			),
			expected: getExpected{
				getCalls: 1,
				displays: []string{stubs.ExpectedDeviceAuthDisplay, getExpectedDisplayDataplaneJSON},
			},
		},
		{
			name: "json output for plural " + constants.InitializersType + " resource",
			input: getInput{
				outputValue:  constants.JSONOutputValue,
				resourceType: constants.InitializersType,
			},
			mockEndpoints: getGetEndpoints(
				postgrest.Endpoint(constants.WorkbenchEndpointDataplaneWorkspaceServices),
				stubs.StubHTTPPartialGetConfig{
					ResponseStatus: http.StatusOK,
					ResponseBody:   getExpectedResponseGenericResource,
				},
			),
			expected: getExpected{
				getCalls: 1,
				displays: []string{stubs.ExpectedDeviceAuthDisplay, getExpectedDisplayInitializerJSON},
			},
		},
		{
			name: "json output for plural " + constants.JobSubmissionsType + " resource",
			input: getInput{
				outputValue:  constants.JSONOutputValue,
				resourceType: constants.JobSubmissionsType,
			},
			mockEndpoints: getGetEndpoints(
				postgrest.Endpoint(constants.WorkbenchEndpointJobRuns),
				stubs.StubHTTPPartialGetConfig{
					ResponseStatus: http.StatusOK,
					ResponseBody:   getExpectedResponseGenericResource,
				},
			),
			expected: getExpected{
				getCalls: 1,
				displays: []string{stubs.ExpectedDeviceAuthDisplay, getExpectedDisplayIDJSON},
			},
		},
		{
			name: "json output for plural " + constants.JobsType + " resource",
			input: getInput{
				outputValue:  constants.JSONOutputValue,
				resourceType: constants.JobsType,
			},
			mockEndpoints: getGetEndpoints(
				postgrest.Endpoint(constants.WorkbenchEndpointJobs),
				stubs.StubHTTPPartialGetConfig{
					ResponseStatus: http.StatusOK,
					ResponseBody:   getExpectedResponseGenericResource,
				},
			),
			expected: getExpected{
				getCalls: 1,
				displays: []string{stubs.ExpectedDeviceAuthDisplay, getExpectedDisplayDataplaneJSON},
			},
		},
		{
			name: "json output for plural " + constants.NodepoolsType + " resource",
			input: getInput{
				outputValue:  constants.JSONOutputValue,
				resourceType: constants.NodepoolsType,
			},
			mockEndpoints: getGetEndpoints(
				postgrest.Endpoint(constants.WorkbenchEndpointDataplaneNodepools),
				stubs.StubHTTPPartialGetConfig{
					ResponseStatus: http.StatusOK,
					ResponseBody:   getExpectedResponseGenericResource,
				},
			),
			expected: getExpected{
				getCalls: 1,
				displays: []string{stubs.ExpectedDeviceAuthDisplay, getExpectedDisplayIDJSON},
			},
		},
		{
			name: "json output for plural " + constants.PingsType + " resource",
			input: getInput{
				outputValue:  constants.JSONOutputValue,
				resourceType: constants.PingsType,
			},
			mockEndpoints: getGetEndpoints(
				postgrest.Endpoint(constants.WorkbenchEndpointPings),
				stubs.StubHTTPPartialGetConfig{
					ResponseStatus: http.StatusOK,
					ResponseBody:   getExpectedResponseGenericResource,
				},
			),
			expected: getExpected{
				getCalls: 1,
				displays: []string{stubs.ExpectedDeviceAuthDisplay, getExpectedDisplayPingJSON},
			},
		},
		{
			name: "json output for plural " + constants.ServicesType + " resource",
			input: getInput{
				outputValue:  constants.JSONOutputValue,
				resourceType: constants.ServicesType,
			},
			mockEndpoints: getGetEndpoints(
				postgrest.Endpoint(constants.WorkbenchEndpointDataplaneWorkbenchServices),
				stubs.StubHTTPPartialGetConfig{
					ResponseStatus: http.StatusOK,
					ResponseBody:   getExpectedResponseGenericResource,
				},
			),
			expected: getExpected{
				getCalls: 1,
				displays: []string{stubs.ExpectedDeviceAuthDisplay, getExpectedDisplayDataplaneJSON},
			},
		},
		{
			name: "json output for plural " + constants.SnapshotsType + " resource",
			input: getInput{
				outputValue:  constants.JSONOutputValue,
				resourceType: constants.SnapshotsType,
			},
			mockEndpoints: getGetEndpoints(
				postgrest.Endpoint(constants.WorkbenchEndpointSnapshots),
				stubs.StubHTTPPartialGetConfig{
					ResponseStatus: http.StatusOK,
					ResponseBody:   getExpectedResponseGenericResource,
				},
			),
			expected: getExpected{
				getCalls: 1,
				displays: []string{stubs.ExpectedDeviceAuthDisplay, getExpectedDisplaySnapshotJSON},
			},
		},
		{
			name: "json output for plural " + constants.StorageclassesType + " resource",
			input: getInput{
				outputValue:  constants.JSONOutputValue,
				resourceType: constants.StorageclassesType,
			},
			mockEndpoints: getGetEndpoints(
				postgrest.Endpoint(constants.WorkbenchEndpointDataplaneStorageclasses),
				stubs.StubHTTPPartialGetConfig{
					ResponseStatus: http.StatusOK,
					ResponseBody:   getExpectedResponseGenericResource,
				},
			),
			expected: getExpected{
				getCalls: 1,
				displays: []string{stubs.ExpectedDeviceAuthDisplay, getExpectedDisplayStorageclassJSON},
			},
		},
		{
			name: "json output for plural " + constants.UsersType + " resource",
			input: getInput{
				outputValue:  constants.JSONOutputValue,
				resourceType: constants.UsersType,
			},
			mockEndpoints: getGetEndpoints(
				auth0admin.Endpoint(constants.WorkbenchEndpointUsers),
				stubs.StubHTTPPartialGetConfig{
					ResponseStatus: http.StatusOK,
					ResponseBody:   getExpectedResponseUsers,
				},
			),
			expected: getExpected{
				getCalls: 1,
				displays: []string{stubs.ExpectedDeviceAuthDisplay, getExpectedDiaplayUsersJSON},
			},
		},
		{
			name: "json output for plural " + constants.WorkbenchesType + " resource",
			input: getInput{
				outputValue:  constants.JSONOutputValue,
				resourceType: constants.WorkbenchesType,
			},
			mockEndpoints: getGetEndpoints(
				postgrest.Endpoint(constants.WorkbenchEndpointWorkbenches),
				stubs.StubHTTPPartialGetConfig{
					ResponseStatus: http.StatusOK,
					ResponseBody:   getExpectedResponseGenericResource,
				},
			),
			expected: getExpected{
				getCalls: 1,
				displays: []string{stubs.ExpectedDeviceAuthDisplay, getExpectedDisplayDataplaneJSON},
			},
		},
		{
			name: "json output for plural " + constants.VolumesType + " resource",
			input: getInput{
				outputValue:  constants.JSONOutputValue,
				resourceType: constants.VolumesType,
			},
			mockEndpoints: getGetEndpoints(
				postgrest.Endpoint(constants.WorkbenchEndpointWorkspaces),
				stubs.StubHTTPPartialGetConfig{
					ResponseStatus: http.StatusOK,
					ResponseBody:   getExpectedResponseGenericResource,
				},
			),
			expected: getExpected{
				getCalls: 1,
				displays: []string{stubs.ExpectedDeviceAuthDisplay, getExpectedDisplayWorkspaceJSON},
			},
		},

		// ==== TABLE PLURAL CASES
		{
			name: "table output for plural " + constants.VolumeConflictsType + " resource",
			input: getInput{
				resourceType: constants.VolumeConflictsType,
			},
			mockEndpoints: getGetEndpoints(
				postgrest.Endpoint(constants.WorkbenchEndpointWorkspaceConflicts),
				stubs.StubHTTPPartialGetConfig{
					ResponseStatus: http.StatusOK,
					ResponseBody:   getExpectedResponsePluralConflicts,
				},
			),
			expected: getExpected{
				getCalls:     1,
				tableStrings: []string{getExpectedTablePluralConflict},
				displays:     []string{stubs.ExpectedDeviceAuthDisplay},
			},
		},
		{
			name: "table output for plural " + constants.ApplicationsType + " resource",
			input: getInput{
				resourceType: constants.ApplicationsType,
			},
			mockEndpoints: getGetEndpoints(
				postgrest.Endpoint(constants.WorkbenchEndpointDataplaneWorkbenchApplications),
				stubs.StubHTTPPartialGetConfig{
					ResponseStatus: http.StatusOK,
					ResponseBody:   getExpectedResponsePluralApplication,
				},
			),
			expected: getExpected{
				getCalls:     1,
				tableStrings: []string{getExpectedTablePluralApplication},
				displays:     []string{stubs.ExpectedDeviceAuthDisplay},
			},
		},
		{
			name: "table output for plural " + constants.DataplanesType + " resource",
			input: getInput{
				resourceType: constants.DataplanesType,
			},
			mockEndpoints: getGetEndpoints(
				postgrest.Endpoint(constants.WorkbenchEndpointDataplanes),
				stubs.StubHTTPPartialGetConfig{
					ResponseStatus: http.StatusOK,
					ResponseBody:   getExpectedResponsePluralDataplane,
				},
			),
			expected: getExpected{
				getCalls:     1,
				tableStrings: []string{getExpectedTablePluralDataplane},
				displays:     []string{stubs.ExpectedDeviceAuthDisplay},
			},
		},
		{
			name: "table output for plural " + constants.EventsType + " resource",
			input: getInput{
				resourceType: constants.EventsType,
			},
			mockEndpoints: getGetEndpoints(
				postgrest.Endpoint(constants.WorkbenchEndpointEvents),
				stubs.StubHTTPPartialGetConfig{
					ResponseStatus: http.StatusOK,
					ResponseBody:   getExpectedResponsePluralEvent,
				},
			),
			expected: getExpected{
				getCalls:     1,
				tableStrings: []string{getExpectedTablePluralEvent},
				displays:     []string{stubs.ExpectedDeviceAuthDisplay},
			},
		},
		{
			name: "table output for plural " + constants.ImageBuildsType + " resource",
			input: getInput{
				resourceType: constants.ImageBuildsType,
			},
			mockEndpoints: getGetEndpoints(
				postgrest.Endpoint(constants.WorkbenchEndpointImageBuilds),
				stubs.StubHTTPPartialGetConfig{
					ResponseStatus: http.StatusOK,
					ResponseBody:   getExpectedResponsePluralImageBuild,
				},
			),
			expected: getExpected{
				getCalls:     1,
				tableStrings: []string{getExpectedTablePluralImageBuild},
				displays:     []string{stubs.ExpectedDeviceAuthDisplay},
			},
		},
		{
			name: "table output for plural " + constants.ImagesType + " resource",
			input: getInput{
				resourceType: constants.ImagesType,
			},
			mockEndpoints: getGetEndpoints(
				postgrest.Endpoint(constants.WorkbenchEndpointImages),
				stubs.StubHTTPPartialGetConfig{
					ResponseStatus: http.StatusOK,
					ResponseBody:   getExpectedResponsePluralImage,
				},
			),
			expected: getExpected{
				getCalls:     1,
				tableStrings: []string{getExpectedTablePluralImage},
				displays:     []string{stubs.ExpectedDeviceAuthDisplay},
			},
		},
		{
			name: "table output for plural " + constants.InitializersType + " resource",
			input: getInput{
				resourceType: constants.InitializersType,
			},
			mockEndpoints: getGetEndpoints(
				postgrest.Endpoint(constants.WorkbenchEndpointDataplaneWorkspaceServices),
				stubs.StubHTTPPartialGetConfig{
					ResponseStatus: http.StatusOK,
					ResponseBody:   getExpectedResponsePluralInitializer,
				},
			),
			expected: getExpected{
				getCalls:     1,
				tableStrings: []string{getExpectedTablePluralInitializer},
				displays:     []string{stubs.ExpectedDeviceAuthDisplay},
			},
		},
		{
			name: "table output for plural " + constants.JobSubmissionsType + " resource",
			input: getInput{
				resourceType: constants.JobSubmissionsType,
			},
			mockEndpoints: getGetEndpoints(
				postgrest.Endpoint(constants.WorkbenchEndpointJobRuns),
				stubs.StubHTTPPartialGetConfig{
					ResponseStatus: http.StatusOK,
					ResponseBody:   getExpectedResponsePluralJobSubmission,
				},
			),
			expected: getExpected{
				getCalls:     1,
				tableStrings: []string{getExpectedTablePluralJobSubmission},
				displays:     []string{stubs.ExpectedDeviceAuthDisplay},
			},
		},
		{
			name: "table output for plural " + constants.JobsType + " resource",
			input: getInput{
				resourceType: constants.JobsType,
			},
			mockEndpoints: getGetEndpoints(
				postgrest.Endpoint(constants.WorkbenchEndpointJobs),
				stubs.StubHTTPPartialGetConfig{
					ResponseStatus: http.StatusOK,
					ResponseBody:   getExpectedResponsePluralJob,
				},
			),
			expected: getExpected{
				getCalls:     1,
				tableStrings: []string{getExpectedTablePluralJob},
				displays:     []string{stubs.ExpectedDeviceAuthDisplay},
			},
		},
		{
			name: "table output for plural " + constants.NodepoolsType + " resource",
			input: getInput{
				resourceType: constants.NodepoolsType,
			},
			mockEndpoints: getGetEndpoints(
				postgrest.Endpoint(constants.WorkbenchEndpointDataplaneNodepools),
				stubs.StubHTTPPartialGetConfig{
					ResponseStatus: http.StatusOK,
					ResponseBody:   getExpectedResponsePluralNodepool,
				},
			),
			expected: getExpected{
				getCalls:     1,
				tableStrings: []string{getExpectedTablePluralNodepool},
				displays:     []string{stubs.ExpectedDeviceAuthDisplay},
			},
		},
		{
			name: "table output for plural " + constants.PingsType + " resource",
			input: getInput{
				resourceType: constants.PingsType,
			},
			mockEndpoints: getGetEndpoints(
				postgrest.Endpoint(constants.WorkbenchEndpointPings),
				stubs.StubHTTPPartialGetConfig{
					ResponseStatus: http.StatusOK,
					ResponseBody:   getExpectedResponsePluralPing,
				},
			),
			expected: getExpected{
				getCalls:     1,
				tableStrings: []string{getExpectedTablePluralPing},
				displays:     []string{stubs.ExpectedDeviceAuthDisplay},
			},
		},
		{
			name: "table output for plural " + constants.ServicesType + " resource",
			input: getInput{
				resourceType: constants.ServicesType,
			},
			mockEndpoints: getGetEndpoints(
				postgrest.Endpoint(constants.WorkbenchEndpointDataplaneWorkbenchServices),
				stubs.StubHTTPPartialGetConfig{
					ResponseStatus: http.StatusOK,
					ResponseBody:   getExpectedResponsePluralService,
				},
			),
			expected: getExpected{
				getCalls:     1,
				tableStrings: []string{getExpectedTablePluralService},
				displays:     []string{stubs.ExpectedDeviceAuthDisplay},
			},
		},
		{
			name: "table output for plural " + constants.SnapshotsType + " resource",
			input: getInput{
				resourceType: constants.SnapshotsType,
			},
			mockEndpoints: getGetEndpoints(
				postgrest.Endpoint(constants.WorkbenchEndpointSnapshots),
				stubs.StubHTTPPartialGetConfig{
					ResponseStatus: http.StatusOK,
					ResponseBody:   getExpectedResponsePluralSnapshot,
				},
			),
			expected: getExpected{
				getCalls:     1,
				tableStrings: []string{getExpectedTablePluralSnapshot},
				displays:     []string{stubs.ExpectedDeviceAuthDisplay},
			},
		},
		{
			name: "table output for plural " + constants.StorageclassesType + " resource",
			input: getInput{
				resourceType: constants.StorageclassesType,
			},
			mockEndpoints: getGetEndpoints(
				postgrest.Endpoint(constants.WorkbenchEndpointDataplaneStorageclasses),
				stubs.StubHTTPPartialGetConfig{
					ResponseStatus: http.StatusOK,
					ResponseBody:   getExpectedResponsePluralStorageclass,
				},
			),
			expected: getExpected{
				getCalls:     1,
				tableStrings: []string{getExpectedTablePluralStorageclass},
				displays:     []string{stubs.ExpectedDeviceAuthDisplay},
			},
		},
		{
			name: "table output for plural " + constants.UsersType + " resource",
			input: getInput{
				resourceType: constants.UsersType,
			},
			mockEndpoints: getGetEndpoints(
				auth0admin.Endpoint(constants.WorkbenchEndpointUsers),
				stubs.StubHTTPPartialGetConfig{
					ResponseStatus: http.StatusOK,
					ResponseBody:   getExpectedResponsePluralUser,
				},
			),
			expected: getExpected{
				getCalls:     1,
				tableStrings: []string{getExpectedTablePluralUser},
				displays:     []string{stubs.ExpectedDeviceAuthDisplay},
			},
		},
		{
			name: "table output for plural " + constants.WorkbenchesType + " resource",
			input: getInput{
				resourceType: constants.WorkbenchesType,
			},
			mockEndpoints: getGetEndpoints(
				postgrest.Endpoint(constants.WorkbenchEndpointWorkbenches),
				stubs.StubHTTPPartialGetConfig{
					ResponseStatus: http.StatusOK,
					ResponseBody:   getExpectedResponsePluralWorkbench,
				},
			),
			expected: getExpected{
				getCalls:     1,
				tableStrings: []string{getExpectedTablePluralWorkbench},
				displays:     []string{stubs.ExpectedDeviceAuthDisplay},
			},
		},
		{
			name: "table output for plural " + constants.VolumesType + " resource",
			input: getInput{
				resourceType: constants.VolumesType,
			},
			mockEndpoints: getGetEndpoints(
				postgrest.Endpoint(constants.WorkbenchEndpointWorkspaces),
				stubs.StubHTTPPartialGetConfig{
					ResponseStatus: http.StatusOK,
					ResponseBody:   getExpectedResponsePluralWorkspace,
				},
			),
			expected: getExpected{
				getCalls:     1,
				tableStrings: []string{getExpectedTablePluralWorkspace},
				displays:     []string{stubs.ExpectedDeviceAuthDisplay},
			},
		},
		{
			name: "invalid endpoint",
			input: getInput{
				endpoint:     "a\nb",
				resourceType: constants.VolumesType,
			},
			expected: getExpected{
				err: expectedInvalidEndpointErr,
			},
		},
		{
			name: "invalid resource response",
			input: getInput{
				resourceType: constants.VolumesType,
			},
			mockEndpoints: getGetEndpoints(
				postgrest.Endpoint(constants.WorkbenchEndpointWorkspaces),
				stubs.StubHTTPPartialGetConfig{
					ResponseStatus: http.StatusOK,
					ResponseBody:   `[invalid`,
				},
			),
			expected: getExpected{
				getCalls: 1,
				err:      expectedInvalidResourceErr,
				displays: []string{stubs.ExpectedDeviceAuthDisplay},
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
			mockCmd := get.Command(context.Background(), afero.NewMemMapFs(), stubUI)
			cmd.AddRootFlags(mockCmd)
			mockCmd.SetArgs(getGetArgs(tt))
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

			require.Len(t, stubUI.Tables, len(tt.expected.tableStrings))
			for i, expected := range tt.expected.tableStrings {
				actual := stubUI.Tables[i].String()
				require.Equal(t, expected, actual, actual)
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
			require.Equal(t, tt.expected.warns, stubUI.Warns, strings.Join(stubUI.Warns, "\n"))
			expectedCallEndpointPrefix := "GET" + constants.PostgrestAPIEndpoint
			if tt.input.resourceType == constants.UsersType || tt.input.resourceType == constants.UserType {
				expectedCallEndpointPrefix = "GET" + auth0admin.Endpoint(constants.WorkbenchEndpointUsers)
			}
			require.Equal(
				t,
				tt.expected.getCalls,
				hitCounts.CountPrefix(expectedCallEndpointPrefix),
				"get calls",
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
			name: "returns sorted supported get resource types",
			expectedTypes: []string{
				constants.ApplicationsType,
				constants.DataplanesType,
				constants.EventsType,
				constants.ImagesType,
				constants.ImageBuildsType,
				constants.InitializersType,
				constants.JobsType,
				constants.JobSubmissionsType,
				constants.NodepoolsType,
				constants.PingsType,
				constants.ServicesType,
				constants.SnapshotsType,
				constants.StorageclassesType,
				constants.UsersType,
				constants.VolumesType,
				constants.VolumeConflictsType,
				constants.WorkbenchesType,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := get.SupportedResourceTypes()
			require.Equal(t, tt.expectedTypes, actual, actual)
		})
	}
}

func getGetArgs(
	tt getTestCase,
) []string {
	args := []string{}
	args = append(args, tt.input.resourceType)
	args = append(args, "--"+constants.AuthModeFlag, constants.AuthModeDevice)
	if tt.input.endpoint != "" {
		args = append(args, "--"+constants.EndpointFlag, tt.input.endpoint)
	}
	if tt.input.outputValue != "" {
		args = append(args, "--"+constants.OutputFlag, tt.input.outputValue)
	}
	args = append(args, tt.input.resourceIDs...)
	return args
}

func getGetEndpoints(
	getURL string,
	config stubs.StubHTTPPartialGetConfig,
) []stubs.StubHTTPEndpoint {
	endpoints := stubs.GetStubAuthEndpoints()
	endpoints = append(endpoints,
		stubs.StubHTTPEndpoint{
			Request: stubs.StubHTTPRequest{
				Method:          http.MethodGet,
				URL:             getURL,
				ExpectedHeaders: stubs.StubAuthHeaders,
				ExpectedQuery:   config.ExpectedQuery,
			},
			Response: stubs.StubHTTPResponse{
				Status: config.ResponseStatus,
				Body:   config.ResponseBody,
			},
		},
	)
	return endpoints
}
