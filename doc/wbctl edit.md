<pre>
Edit a Workbench resource using the default editor.

Usage:
  wbctl edit [command]

Available Commands:
  <a href='wbctl edit image.md'>image</a>       Edit a Image resource
  <a href='wbctl edit job.md'>job</a>         Edit a Job resource
  <a href='wbctl edit snapshot.md'>snapshot</a>    Edit a Snapshot resource
  <a href='wbctl edit volume.md'>volume</a>      Edit a Volume resource
  <a href='wbctl edit workbench.md'>workbench</a>   Edit a Workbench resource

Flags:
  -h, --help   help for edit

Global Flags:
  -m, --auth-mode string       Authentication mode. One of: auto, pkce, client-credentials, password, device. (default "auto")
      --client-id string       The client-id to authenticate with the Workbench API ($WBCLIENTID).
      --client-secret string   The client-secret to authenticate with the Workbench API ($WBCLIENTSECRET).
  -e, --endpoint string        The Workbench API endpoint URL.
      --ignore-token-cache     Ignore token cache.
      --no-listen              Do not open a local port to listen for PKCE authentication.
      --organization string    The organization to authenticate with the Workbench API.
  -p, --password string        The password to authenticate with the Workbench API ($WBPASSWORD).
  -u, --username string        The username to authenticate with the Workbench API ($WBUSERNAME).

Use "wbctl edit [command] --help" for more information about a command.
</pre>
