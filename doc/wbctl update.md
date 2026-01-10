<pre>
Updates one or many Workbench resources.

Usage:
  wbctl update [command]

Available Commands:
  <a href='wbctl update images.md'>images</a>      Update one or many Image resources
  <a href='wbctl update jobs.md'>jobs</a>        Update one or many Job resources
  <a href='wbctl update snapshots.md'>snapshots</a>   Update one or many Snapshot resources
  <a href='wbctl update volumes.md'>volumes</a>     Update one or many Volume resources
  <a href='wbctl update workbenches.md'>workbenches</a> Update one or many Workbench resources

Flags:
  -h, --help   help for update

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

Use "wbctl update [command] --help" for more information about a command.
</pre>
