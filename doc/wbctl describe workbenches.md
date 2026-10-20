<pre>
Prints detailed information about the Workbench resources identified by an optional list of resource ids. All resources printed if no ids provided.

Usage:
  wbctl describe workbenches [flags]

Aliases:
  workbenches, workbench

Flags:
  -h, --help   help for workbenches

Global Flags:
  -m, --auth-mode string      Authentication mode. One of: auto, pkce, client-credentials, password, device. (default "device")
      --client-id string      The client-id to authenticate with the Workbench API ($WBCLIENTID).
  -e, --endpoint string       The Workbench API endpoint URL.
      --ignore-token-cache    Ignore token cache.
      --no-listen             Do not open a local port to listen for PKCE authentication.
      --organization string   The organization to authenticate with the Workbench API.
</pre>
