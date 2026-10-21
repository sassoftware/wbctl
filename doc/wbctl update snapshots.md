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
  -m, --auth-mode string      Authentication mode. One of: device, pkce. (default "device")
  -e, --endpoint string       The Workbench API endpoint URL.
      --ignore-token-cache    Ignore token cache.
      --no-listen             Do not open a local port to listen for PKCE authentication.
      --organization string   The organization to authenticate with the Workbench API.
</pre>
