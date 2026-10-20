<pre>
Rebuilds one or many Workbench resources

Usage:
  wbctl rebuild [command]

Available Commands:
  <a href='wbctl rebuild images.md'>images</a>      Rebuild one or many Image resources

Flags:
  -h, --help   help for rebuild

Global Flags:
  -m, --auth-mode string      Authentication mode. One of: auto, pkce, client-credentials, password, device. (default "device")
      --client-id string      The client-id to authenticate with the Workbench API ($WBCLIENTID).
  -e, --endpoint string       The Workbench API endpoint URL.
      --ignore-token-cache    Ignore token cache.
      --no-listen             Do not open a local port to listen for PKCE authentication.
      --organization string   The organization to authenticate with the Workbench API.

Use "wbctl rebuild [command] --help" for more information about a command.
</pre>
