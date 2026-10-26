# wbctl - Workbench Control

`wbctl` is a command line tool for interacting with a SAS Workbench
installation.

## Getting started

Install with

```shell
go install github.com/sassoftware/wbctl/cmd/wbctl
```

## Usage

<pre>
SAS Workbench Control

Usage:
  wbctl [command]

Available Commands:
  cancel      Cancel running Workbench resources
  completion  Generate the autocompletion script for the specified shell
  create      Create Workbench resources
  delete      Delete Workbench resources
  describe    Describe Workbench resources in detail
  edit        Edit Workbench resources
  get         Print Workbench resources
  help        Help about any command
  login       Authenticate and cache an access token
  logout      Remove cached access token
  logs        Fetch the logs of Workbench resources
  rebuild     Rebuild Workbench image
  settings    Manage user settings
  start       Start stopped Workbench resources
  stop        Stop running Workbench resources
  update      Update Workbench resources
  version     Print version information
  whoami      Display the authenticated user.

Flags:
  -m, --auth-mode string      Authentication mode. One of: device, pkce. (default "device")
  -e, --endpoint string       The Workbench API endpoint URL.
  -h, --help                  help for wbctl
      --ignore-token-cache    Ignore token cache.
      --no-listen             Do not open a local port to listen for PKCE authentication.
      --organization string   The organization to authenticate with the Workbench API.
  -v, --verbose               Log verbosity. Repeat option for increased levels of verbose output (-v, -vv, or -vvv).

Use "wbctl [command] --help" for more information about a command.
</pre>

Detailed usage can be found at [doc/wbctl.md](doc/wbctl.md).

## Configuration

For configuration, `wbctl` looks for a file named `config.json` in the
`$HOME/.wbctl` directory. A different config file can be specified by setting
the `WBCONFIG` environment variable. Any command line argument can be specified
in this file as a top level field. For example:

```
{
  "auth-mode": "pkce",
  "endpoint": "https://workbench.sas.com",
  "organization": "myorganization",
  "output": "json"
}
```

Note that verbose can be specified in the config file as silent, error, warn,
info, debug, or trace.

## Authentication

`wbctl` supports two OAuth authentication methods selected by `--auth-mode`:

1. **Device authorization flow**:
    - Set `--auth-mode device`
    - Open the displayed url in your browser to authorize the device with the
      displayed device code
    - Example: `wbctl --auth-mode device -e https://workbench.sas.com get workbenches`

2. **PKCE flow**:
   - Set `--auth-mode pkce`
   - Open the displayed url in your browser to authorize with your
     username/password
   - Example: `wbctl --auth-mode pkce -e https://workbench.sas.com get workbeches`

`--auth-mode device` is the default.

## Examples

### Get applications

```shell
> ./wbctl get applications
ID          NAME                                DESCRIPTION   SERVICES
vscode      VS Code - SAS and Python                          compute, vscode
python      JupyterLab - SAS and Python                       compute, python
notebook    Jupyter Notebook - SAS and Python                 compute, notebook
jupyter-r   JupyterLab - Python and R                         jupyter-r
eg          Enterprise Guide                                  enterprise-guide-compute
code-r      VS Code - SAS and R                               compute, code-r
```

### Get dataplanes

```shell
> ./wbctl get dataplanes
ID                                     NAME          DESCRIPTION   MESSAGE   AGE
4374c916-9c23-4b29-90e4-11bb8af349d3   my-dataplane  my-dataplane            27h
```

### Get events

``` shell
> ./wbctl get events
ID                                     TYPE     RESOURCE_TYPE   KIND                    ACTION    NOTE                   REASON                                                    AGE
d833ea40-5b4c-4a41-9b54-6bf3e077116f   Normal   workspace       Workspace               Created   Reconcile                                                                        4h49m
d833ea40-5b4c-4a41-9b54-6bf3e077116f   Normal   workspace       PersistentVolumeClaim             WaitForFirstConsumer   waiting for first consumer to be created before binding   4h49m
d833ea40-5b4c-4a41-9b54-6bf3e077116f   Normal   workspace       PersistentVolumeClaim             WaitForFirstConsumer   waiting for first consumer to be created before binding   4h49m
d833ea40-5b4c-4a41-9b54-6bf3e077116f   Normal   workspace       PersistentVolumeClaim             WaitForFirstConsumer   waiting for first consumer to be created before binding   4h49m
...
4f9766a6-29b2-49ad-a711-d170c1f23a6a   Normal   workspace       PersistentVolumeClaim             WaitForFirstConsumer   waiting for first consumer to be created before binding   27h
4f9766a6-29b2-49ad-a711-d170c1f23a6a   Normal   workspace       PersistentVolumeClaim             WaitForFirstConsumer   waiting for first consumer to be created before binding   27h
4f9766a6-29b2-49ad-a711-d170c1f23a6a   Normal   workspace       PersistentVolumeClaim             WaitForFirstConsumer   waiting for first consumer to be created before binding   27h
```

### Get Workspace initializers

```shell
> ./wbctl get initializers
ID          NAME                         TYPE
home-init   Home directory initializer
```

### Get nodepools

```shell
> ./wbctl get nodepools
ID       NAME                  DESCRIPTION
C0M0     Minimal Resources     Micro System.
C2M8     2 cores, 8GiB RAM     Small System.
C4M16    4 cores, 16GiB RAM    Medium System.
C8M32    8 cores, 32GiB RAM    Large System.
C16M64   16 cores, 64GiB RAM   Extra Large System.
```

### Get pings

```shell
> ./wbctl get pings
DATAPLANE ID                           REQUEST AT                    RESPONSE AT                   RESPONSE TIME   STATE       MESSAGE   AGE
4374c916-9c23-4b29-90e4-11bb8af349d3   2026-01-21T21:20:00.692936Z   2026-01-21T21:20:00.806563Z   113.627ms       processed             2m45s
```

### Get services

```shell
> ./wbctl get services
ID         NAME                 TYPE
compute    compute              backend
notebook   Jupyter Notebook     frontend
python     Jupyter Lab          frontend
vscode     Visual Studio Code   frontend
```

### Get storageclasses

```shell
> ./wbctl get storageclasses
ID          NAME   DESCRIPTION         DEFAULT
ebs-wb-sc   EBS    EBS based storage   true
wb-efs-sc   EFS    EFS based storage
```

### Get users

```shell
> ./wbctl get users
ID                                     USERNAME                                               ROLES                        METADATA   AGE
22518a4b-b711-5a4a-bd91-e6bbc93c85d0   admin1 User (admin1@workbench.sas.com)                 org_admin, workbench_admin              42h
f54fdb49-3ce3-5d67-b1ed-9e26bd2a1dfd   admin2 User (admin2@workbench.sas.com)                 org_admin, workbench_admin              42h
3c9395a5-52ad-5e4e-944f-6ad81ab2ba92   admin3 User (admin3@workbench.sas.com)                 org_admin, workbench_admin              42h
128570b9-f47c-5cd0-98fb-e9704c000d10   user1 User (usr1@workbench.sas.com)                    workbench_user                          42h
f6e2918c-5704-589b-ae57-ce0ee569a6e9   user2 User (usr2@workbench.sas.com)                    workbench_user                          42h
61f2a449-107c-56e5-9052-90ca7d9aa5d5   user3 User (usr3@workbench.sas.com)                    workbench_user                          42h
```

### Describe user

```shell
> ./wbctl describe user 22518a4b-b711-5a4a-bd91-e6bbc93c85d0
ID:           22518a4b-b711-5a4a-bd91-e6bbc93c85d0
Created At:   2026-09-28T17:55:48Z (42h ago)
Username:     admin1 Test User
Email:        admin1@workbench.sas.com
Roles:
  - org_admin
  - workbench_admin
```

### Create Volume

```shell
> ./wbctl create volume --name demo-workspace --description Demo_Volume --size 1Gi --storageclass ebs-wb-sc
e437a8b9-d9c0-49bd-842b-d61876e5f575
```

### Get Volumes

```shell
> ./wbctl get volumes
ID                                     CONTROLLER_ID                             NAME             SIZE   HOME DIR   CLASS       STATE     MESSAGE   AGE
e437a8b9-d9c0-49bd-842b-d61876e5f575   sas-workbench-20mgwyaf1wrmyjfwi8ec8oz1s   demo-workspace   1Gi               ebs-wb-sc   CREATED             1s
d833ea40-5b4c-4a41-9b54-6bf3e077116f   sas-workbench-83zvx8fs7s4acskm0yjheezn4   myfolder         5Gi               ebs-wb-sc   CREATED             4h49m
```

### Describe Volume

```shell
> ./wbctl describe volume e437a8b9-d9c0-49bd-842b-d61876e5f575
ID:              e437a8b9-d9c0-49bd-842b-d61876e5f575
Controller ID:   sas-workbench-20mgwyaf1wrmyjfwi8ec8oz1s
Name:            demo-volume
Description:     Demo_Volume
Storage Type:    PersistentVolumeClaim
Created By:      admin1 Test User (admin1@workbench.sas.com)
Created At:      2026-01-21T21:22:53.367045Z
Modified At:     2026-01-21T21:22:53.367045Z
State:           CREATED
Phase:           Pending
Access Modes:    ReadWriteOnce
Storage Class:   ebs-wb-sc
Size:            1Gi
```

### Create Snapshot

```shell
> ./wbctl create snapshot --volume e437a8b9-d9c0-49bd-842b-d61876e5f575 --name demo-snapshot --description Demo_Snapshot
f7b75b0f-6bad-4f76-a142-1c6d2fcd5a38
```

### Get Snapshots

```shell
> ./wbctl get snapshots
ID                                     CONTROLLER_ID                             NAME            DESCRIPTION     SOURCE                                    SIZE        READY   AGE
f7b75b0f-6bad-4f76-a142-1c6d2fcd5a38   sas-workbench-576upa4jdkz9ehwqk0hc7cetp   demo-snapshot   Demo_Snapshot   sas-workbench-20mgwyaf1wrmyjfwi8ec8oz1s   <unknown>   false   1s
```

### Describe Snapshot

```shell
> ./wbctl describe snapshot f7b75b0f-6bad-4f76-a142-1c6d2fcd5a38
ID:              f7b75b0f-6bad-4f76-a142-1c6d2fcd5a38
Controller ID:   sas-workbench-576upa4jdkz9ehwqk0hc7cetp
Name:            demo-snapshot
Description:     Demo_Snapshot
Created By:      admin1 Test User (admin1@workbench.sas.com)
Created At:      2026-01-21T21:22:57.866422Z
Modified At:     2026-01-21T21:22:57.866422Z
Spec:            {
                   "source": {
                     "persistentVolumeClaimName": "sas-workbench-20mgwyaf1wrmyjfwi8ec8oz1s"
                   }
                 }
Volume Status:
  Ready To Use:    false
  Restore Size:    <nil>
  Content Name:
  Creation Time:   0001-01-01T00:00:00Z
```

### Create Workbench

```shell
> ./wbctl create workbench --name demo-workbench --application vscode --volume e437a8b9-d9c0-49bd-842b-d61876e5f575 --nodepool C0M0
dd73e68a-f3f2-4cb4-b386-6adb9153a8a2
```

### Get Workbenches

```shell
> ./wbctl get workbenches
ID                                     CONTROLLER_ID                             NAME             APPLICATION   SERVICES          URL                                                                                          STATE      MESSAGE   AGE
dd73e68a-f3f2-4cb4-b386-6adb9153a8a2   sas-workbench-bwiprc05q4q5wocvzupxc9bto   demo-workbench                 compute, vscode   https://wbdev-data-plane.rnd-dev.workbench.sas.com/sas-workbench-bwiprc05q4q5wocvzupxc9bto   STARTING             1s
```

### Describe Workbench

```shell
> ./wbctl describe workbench dd73e68a-f3f2-4cb4-b386-6adb9153a8a2
ID:              dd73e68a-f3f2-4cb4-b386-6adb9153a8a2
Controller ID:   sas-workbench-bwiprc05q4q5wocvzupxc9bto
Name:            demo-workbench
URL:             https://wbdev-data-plane.rnd-dev.workbench.sas.com/sas-workbench-bwiprc05q4q5wocvzupxc9bto
Created By:      admin1 Test User (admin1@workbench.sas.com)
Created At:      2026-01-21T21:23:04.047201Z
Modified At:     2026-01-21T21:23:04.047201Z
Nodepool:        C0M0
Desired State:   RUNNING
Current State:   STARTING
Services:
  0:
    ID:   compute
    Environment:
      DEFAULT_MOUNTNAME:   demo-workspace
  1:
    ID:   vscode
    Environment:
      DEFAULT_MOUNTNAME:   demo-workspace
Workspaces:
  0:
    ID:           sas-workbench-20mgwyaf1wrmyjfwi8ec8oz1s
    Mount name:   demo-workspace
Pod sas-workbench-bwiprc05q4q5wocvzupxc9bto-67d656749c-h7t4s:
  Phase:   Pending
  Conditions:
    PodReadyToStartContainers   False
    Initialized                 False
    Ready                       False
    ContainersReady             False
    PodScheduled                True
  Init Containers:
    sas-config-init:
      Image:           cr.sas.com/cdp-release-x64_oci_linux_2-docker-latest/sas-config-init:1.29.0-20260106.1767737031242
      Ready:           false
      Started:         false
      Restart Count:   0
      State:           Waiting
        Reason:   PodInitializing
      Volume Mounts:
        /config
        /opt/sas/viya/config/var
        /tmp
    job-task-binaries:
      Image:           cr.sas.com/cdp-release-x64_oci_linux_2-docker-latest/sas-workbench-job-runner:0.4.0-20260120.1768929420499
      Ready:           false
      Started:         false
      Restart Count:   0
      State:           Waiting
        Reason:   PodInitializing
      Volume Mounts:
        /opt/sas/workbench/job-task-binaries
  Containers:
    compute:
      Image:           cr.sas.com/cdp-release-x64_oci_linux_2-docker-latest/sas-programming-environment:1.81.6-20260120.1768892559391
      Ready:           false
      Started:         false
      Restart Count:   0
      State:           Waiting
        Reason:   PodInitializing
      Volume Mounts:
        /cas/license
        /config
        /config/autoexec_usermods.sas
        /config/sasv9_limits.template.cfg
        /etc/group (RO)
        /etc/passwd (RO)
        /etc/sas-workbench-job-log-aggregator-config (RO)
        /opt/sas/viya/config/var
        /opt/sas/viya/home/bin/compsrv_start.sh
        /opt/sas/viya/home/commonfiles
        /opt/sas/viya/home/share/refdata/qkb
        /opt/sas/workbench/job-task-binaries
        /rdutil
        /sashelp
        /sasuser
        /tmp
        /workspaces/demo-workspace
    vscode:
      Image:           cr.sas.com/cdpdefault-snapshot-x64_oci_linux_2-docker-latest/workbench-environment-codeserver-mla-scipy:0.4.1-20260120.1768880485459
      Ready:           false
      Started:         false
      Restart Count:   0
      State:           Waiting
        Reason:   PodInitializing
      Volume Mounts:
        /cas/license
        /config
        /etc/group (RO)
        /etc/passwd (RO)
        /etc/sas-workbench-job-log-aggregator-config (RO)
        /home/sas
        /opt/sas/viya/home/commonfiles
        /opt/sas/workbench/job-task-binaries
        /tmp
        /workspaces/demo-workspace

```

### Delete Workbench

```shell
> ./wbctl delete workbench dd73e68a-f3f2-4cb4-b386-6adb9153a8a2
dd73e68a-f3f2-4cb4-b386-6adb9153a8a2
```

### Login (cache an access token)

Authenticate once and cache the token so subsequent commands do not require
repeated authorization:

```shell
# PKCE grant (presents login url)
> ./wbctl --auth-mode pkce -e https://workbench.sas.com login
```

### Create Job

```shell
> ./wbctl create job --name demo-job --application vscode --volume e437a8b9-d9c0-49bd-842b-d61876e5f575 --nodepool C0M0 --command "echo hi"
282a2478-b7de-45e6-999d-79f27d4b53d3
```

### Get Jobs

```shell
> ./wbctl get jobs
ID                                     NAME       APPLICATION   SERVICES          AGE
282a2478-b7de-45e6-999d-79f27d4b53d3   demo-job                 compute, vscode   1s
```

### Describe Job

```shell
> ./wbctl describe job 282a2478-b7de-45e6-999d-79f27d4b53d3
ID:            282a2478-b7de-45e6-999d-79f27d4b53d3
Name:          demo-job
Created By:    admin1 Test User (admin1@workbench.sas.com)
Created At:    2026-01-21T21:23:10.071225Z
Modified At:   2026-01-21T21:23:10.071225Z
Nodepool:      C0M0
Spec:          {
                 "tasks": {
                   "default": {
                     "cmds": [
                       "echo hi"
                     ]
                   }
                 },
                 "version": "3"
               }
Services:
  0:
    ID:   compute
    Environment:
      DEFAULT_MOUNTNAME:   demo-workspace
  1:
    ID:   vscode
    Environment:
      DEFAULT_MOUNTNAME:   demo-workspace
Volumes:
  0:
    ID:           sas-workbench-20mgwyaf1wrmyjfwi8ec8oz1s
    Mount name:   demo-workspace
Properties:
  isActive:   true
```

### Create Job Submission

```shell
> ./wbctl create job-submission --job-id 282a2478-b7de-45e6-999d-79f27d4b53d3
f6c9221c-546c-4351-badd-7b3080ea83d4
```

### Get Job Submissions

```shell
> ./wbctl get job-submissions
ID                                     CONTROLLER_ID                       JOB        LOG URL                                                                                                                                                                                                                                                STATE      MESSAGE   RUNTIME   AGE
f6c9221c-546c-4351-badd-7b3080ea83d4   sas-wbj-1nnlkcd6ioyamys10picgz3ok   demo-job   https://wbdev-data-plane.rnd-dev.workbench.sas.com/logs/4374c916-9c23-4b29-90e4-11bb8af349d3/4eca09fa-5fbf-4785-9ead-26ed162f4041/9bf8b2fb-e6d0-4ada-bb5d-6a073ae02448/282a2478-b7de-45e6-999d-79f27d4b53d3/f6c9221c-546c-4351-badd-7b3080ea83d4.log   STARTING             0.00s     2s
```

### Describe Job Submission

```shell
> ./wbctl describe job-submission f6c9221c-546c-4351-badd-7b3080ea83d4
ID:              f6c9221c-546c-4351-badd-7b3080ea83d4
Controller ID:   sas-wbj-1nnlkcd6ioyamys10picgz3ok
Job Name:        demo-job
Job ID:          282a2478-b7de-45e6-999d-79f27d4b53d3
URL:             https://wbdev-data-plane.rnd-dev.workbench.sas.com/sas-wbj-1nnlkcd6ioyamys10picgz3ok
Log URL:         https://wbdev-data-plane.rnd-dev.workbench.sas.com/logs/4374c916-9c23-4b29-90e4-11bb8af349d3/4eca09fa-5fbf-4785-9ead-26ed162f4041/9bf8b2fb-e6d0-4ada-bb5d-6a073ae02448/282a2478-b7de-45e6-999d-79f27d4b53d3/f6c9221c-546c-4351-badd-7b3080ea83d4.log
Created By:      admin1 Test User (admin1@workbench.sas.com)
Created At:      2026-01-21T21:23:16.490329Z
Modified At:     2026-01-21T21:23:16.490329Z
Runtime:         0.000000
Nodepool:        C0M0
Scheduled:       false
State:           STARTING
Pod sas-wbj-1nnlkcd6ioyamys10picgz3ok-wxfbg:
  Phase:   Pending
  IPs:     10.2.33.218
  Conditions:
    PodReadyToStartContainers   True
    Initialized                 True
    Ready                       False
    ContainersReady             False
    PodScheduled                True
  Init Containers:
    sas-config-init:
      Image:           cr.sas.com/cdp-release-x64_oci_linux_2-docker-latest/sas-config-init:1.29.0-20260106.1767737031242
      Ready:           true
      Started:         false
      Restart Count:   0
      State:           Terminated
        Reason:        Completed
        Exit Code:     0
        Started At:    2026-01-21T16:23:19-05:00
        Finished At:   2026-01-21T16:23:20-05:00
      Volume Mounts:
        /config
        /opt/sas/viya/config/var
        /tmp
      Linux User:
        UID:               1001
        GID:               1001
        Additional GIDs:   1001
    job-task-binaries:
      Image:           cr.sas.com/cdp-release-x64_oci_linux_2-docker-latest/sas-workbench-job-runner:0.4.0-20260120.1768929420499
      Ready:           true
      Started:         false
      Restart Count:   0
      State:           Terminated
        Reason:        Completed
        Exit Code:     0
        Started At:    2026-01-21T16:23:21-05:00
        Finished At:   2026-01-21T16:23:21-05:00
      Volume Mounts:
        /opt/sas/workbench/job-task-binaries
      Linux User:
        UID:               1001
        GID:               1001
        Additional GIDs:   1001
    compute:
      Image:           cr.sas.com/cdp-release-x64_oci_linux_2-docker-latest/sas-programming-environment:1.81.6-20260120.1768892559391
      Ready:           true
      Started:         true
      Restart Count:   0
      State:           Running
        Started At:   2026-01-21T16:23:22-05:00
      Limits:
        CPU:      1
        Memory:   4Gi
      Requests:
        CPU:      50m
        Memory:   100Mi
      Allocated:
        CPU:      50m
        Memory:   100Mi
      Volume Mounts:
        /cas/license
        /config
        /config/autoexec_usermods.sas
        /config/sasv9_limits.template.cfg
        /etc/group (RO)
        /etc/passwd (RO)
        /etc/sas-workbench-job-log-aggregator-config (RO)
        /opt/sas/viya/config/var
        /opt/sas/viya/home/bin/compsrv_start.sh
        /opt/sas/viya/home/commonfiles
        /opt/sas/viya/home/share/refdata/qkb
        /opt/sas/workbench/job-task-binaries
        /rdutil
        /sashelp
        /sasuser
        /tmp
        /workspaces/demo-workspace
      Linux User:
        UID:               1001
        GID:               1001
        Additional GIDs:   1001
  Containers:
    vscode:
      Image:           cr.sas.com/cdpdefault-snapshot-x64_oci_linux_2-docker-latest/workbench-environment-codeserver-mla-scipy:0.4.1-20260120.1768880485459
      Ready:           false
      Started:         false
      Restart Count:   0
      State:           Waiting
        Reason:   PodInitializing
      Volume Mounts:
        /cas/license
        /config
        /context
        /etc/group (RO)
        /etc/passwd (RO)
        /etc/sas-workbench-job-log-aggregator-config (RO)
        /home/sas
        /opt/sas/viya/home/commonfiles
        /opt/sas/workbench/job-task-binaries
        /tmp
        /workspaces/demo-workspace
```

### Get Job Submission status in JSON

```shell
> ./wbctl --output json get job-submissions ${jobsubmissionID} | \
    jq '.[]|{
      id: .id,
      current_state: .current_state, pod_status_phases: (.pod_status[]|.status.phase),
      container_statuses: [.pod_status[]|.status.containerStatuses[]|
        {
         name: (.name),
         state: (.state.terminated.reason)
        }
      ],
      initcontainer_statuses: [.pod_status[]|.status.initContainerStatuses[]|
        {
         name: (.name),
         state: (.state.terminated.reason)
        }
      ]
    }'
{
  "id": "f6c9221c-546c-4351-badd-7b3080ea83d4",
  "current_state": "RUNNING",
  "pod_status_phases": "Running",
  "container_statuses": [
    {
      "name": "vscode",
      "state": null
    }
  ],
  "initcontainer_statuses": [
    {
      "name": "sas-config-init",
      "state": "Completed"
    },
    {
      "name": "job-task-binaries",
      "state": "Completed"
    },
    {
      "name": "compute",
      "state": null
    }
  ]
}
```

### Get Job Submission logs

```shell
> ./wbctl logs job-submission f6c9221c-546c-4351-badd-7b3080ea83d4
2026-01-21T21:23:29.603528749Z: sas.workbench.job.run f6c9221c-546c-4351-badd-7b3080ea83d4 started
2026-01-21T21:23:29.603719772Z: INFO: Copying contents of '/opt/sas/viya/config/home/sas' to '/home/sas'...
2026-01-21T21:23:30.220679234Z: /workspaces/demo-workspace is empty - copying the README that points the user to the Workbench examples in GitHub.
2026-01-21T21:23:30.251523407Z: INFO: Using local en-US (en_US.UTF-8).
2026-01-21T21:23:30.364254936Z: task: [default] echo hi
2026-01-21T21:23:30.372109601Z: hi
2026-01-21T21:23:30.372145351Z: sas.workbench.job.run f6c9221c-546c-4351-badd-7b3080ea83d4 completed
```

### Create Workspace From Snapshot

```shell
> ./wbctl create volume --name restore --snapshot f7b75b0f-6bad-4f76-a142-1c6d2fcd5a38
6a51cb99-5ed0-4e91-80f5-d981e11a3b59
```

## Contributing

Maintainers are not currently accepting patches and contributions to this
project.

## License

This project is licensed under the Apache License 2.0. See [LICENSE](LICENSE).

The dependencies of this project include third-party software, each under their
own license. They are listed in [licenses/licenses.md](./licenses/licenses.md).
