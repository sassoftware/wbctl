<pre>
Updates all the Snapshot resources identified by a list of ids with the options provided.

Usage:
  wbctl update snapshots [flags]

Aliases:
  snapshots, snapshot

Flags:
      --description string   The description of the snapshot.
  -d, --dry-run              Only print the object that would be sent, without sending it.
  -f, --filename string      Path to a file containing the snapshot in JSON format. Use '-' to read from stdin.
  -h, --help                 help for snapshots
  -n, --name string          The name of the snapshot.
      --volume string        Volume to snapshot.

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
