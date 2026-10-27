<pre>
Rebuilds one or many Workbench resources

Usage:
  wbctl rebuild [command]

Available Commands:
  <a href='wbctl rebuild images.md'>images</a>      Rebuild one or many Image resources

Flags:
  -h, --help   help for rebuild

Global Flags:
  -m, --auth-mode string      Authentication mode. One of: device, pkce. (default "device")
  -e, --endpoint string       The Workbench API endpoint URL.
      --ignore-token-cache    Ignore token cache.
      --no-listen             Do not open a local port to listen for PKCE authentication.
      --organization string   The organization to authenticate with the Workbench API.
  -p, --profile string        The configuration profile to use. (default "Default")
  -v, --verbose               Log verbosity. Repeat option for increased levels of verbose output (-v, -vv, or -vvv).

Use "wbctl rebuild [command] --help" for more information about a command.
</pre>
