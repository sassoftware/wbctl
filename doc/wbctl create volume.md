<pre>
Creates a new Volume resource with the options provided.

Usage:
  wbctl create volume [flags]

Aliases:
  volume, volumes, workspace, workspaces

Flags:
      --dataplane string      The ID of the dataplane where the volume will be created. If only one dataplane is available, this flag can be omitted.
      --description string    The description of the volume.
  -d, --dry-run               Only print the object that would be sent, without sending it.
  -f, --filename string       Path to a file containing the volume in JSON format. Use '-' to read from stdin.
  -h, --help                  help for volume
      --home-dir              If true, make this volume a home directory.
      --initializer strings   The initializers(s) to run on the worksapce after it is created.
  -n, --name string           The name of the volume.
      --size string           The size of the volume.
      --snapshot string       The ID of the <a href='wbctl create snapshot.md'>snapshot</a> from which to create the volume. If specified, storage-class and size are ignored.
      --storageclass string   The storage class to use for the volume.

Global Flags:
  -m, --auth-mode string      Authentication mode. One of: auto, pkce, client-credentials, password, device. (default "device")
      --client-id string      The client-id to authenticate with the Workbench API ($WBCLIENTID).
  -e, --endpoint string       The Workbench API endpoint URL.
      --ignore-token-cache    Ignore token cache.
      --no-listen             Do not open a local port to listen for PKCE authentication.
      --organization string   The organization to authenticate with the Workbench API.
</pre>
