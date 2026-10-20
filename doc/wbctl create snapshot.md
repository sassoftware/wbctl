<pre>
Creates a new Snapshot resource with the options provided.

Usage:
  wbctl create snapshot [flags]

Aliases:
  snapshot, snapshots

Flags:
      --dataplane string     The ID of the dataplane where the snapshot will be created. If only one dataplane is available, this flag can be omitted.
      --description string   The description of the snapshot.
  -d, --dry-run              Only print the object that would be sent, without sending it.
  -f, --filename string      Path to a file containing the snapshot in JSON format. Use '-' to read from stdin.
  -h, --help                 help for snapshot
  -n, --name string          The name of the snapshot.
      --volume string        Volume to snapshot.

Global Flags:
  -m, --auth-mode string      Authentication mode. One of: auto, pkce, client-credentials, password, device. (default "device")
      --client-id string      The client-id to authenticate with the Workbench API ($WBCLIENTID).
  -e, --endpoint string       The Workbench API endpoint URL.
      --ignore-token-cache    Ignore token cache.
      --no-listen             Do not open a local port to listen for PKCE authentication.
      --organization string   The organization to authenticate with the Workbench API.
</pre>
