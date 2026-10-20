<pre>
Deletes the Volume resources identified by a list of resource ids. At least one id is required.

Usage:
  wbctl delete volumes [flags]

Aliases:
  volumes, volume, workspace, workspaces

Flags:
  -h, --help   help for volumes

Global Flags:
  -m, --auth-mode string      Authentication mode. One of: auto, pkce, client-credentials, password, device. (default "device")
      --client-id string      The client-id to authenticate with the Workbench API ($WBCLIENTID).
  -e, --endpoint string       The Workbench API endpoint URL.
      --ignore-token-cache    Ignore token cache.
      --no-listen             Do not open a local port to listen for PKCE authentication.
      --organization string   The organization to authenticate with the Workbench API.
</pre>
