<pre>
Creates Workbench resources.

Usage:
  wbctl create [command]

Available Commands:
  <a href='wbctl create image.md'>image</a>          Create a new Image resource
  <a href='wbctl create job.md'>job</a>            Create a new Job resource
  <a href='wbctl create job-submission.md'>job-submission</a> Create a new Job-Submission resource
  <a href='wbctl create snapshot.md'>snapshot</a>       Create a new Snapshot resource
  <a href='wbctl create volume.md'>volume</a>         Create a new Volume resource
  <a href='wbctl create workbench.md'>workbench</a>      Create a new Workbench resource

Flags:
  -h, --help   help for create

Global Flags:
  -m, --auth-mode string      Authentication mode. One of: auto, pkce, client-credentials, password, device. (default "device")
      --client-id string      The client-id to authenticate with the Workbench API ($WBCLIENTID).
  -e, --endpoint string       The Workbench API endpoint URL.
      --ignore-token-cache    Ignore token cache.
      --no-listen             Do not open a local port to listen for PKCE authentication.
      --organization string   The organization to authenticate with the Workbench API.

Use "wbctl create [command] --help" for more information about a command.
</pre>
