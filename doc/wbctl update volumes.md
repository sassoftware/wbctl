<pre>
Updates all the Volume resources identified by a list of ids with the options provided.

Usage:
  wbctl update volumes [flags]

Aliases:
  volumes, volume, workspace, workspaces

Flags:
      --description string    The description of the volume.
  -d, --dry-run               Only print the object that would be sent, without sending it.
  -f, --filename string       Path to a file containing the volume in JSON format. Use '-' to read from stdin.
  -h, --help                  help for volumes
      --home-dir              If true, make this volume a home directory.
      --initializer strings   The initializers(s) to run on the worksapce after it is created.
  -n, --name string           The name of the volume.
      --size string           The size of the volume.

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
</pre>
