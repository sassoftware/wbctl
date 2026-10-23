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
  -m, --auth-mode string      Authentication mode. One of: device, pkce. (default "device")
  -e, --endpoint string       The Workbench API endpoint URL.
      --ignore-token-cache    Ignore token cache.
      --no-listen             Do not open a local port to listen for PKCE authentication.
      --organization string   The organization to authenticate with the Workbench API.
  -v, --verbose               Log verbosity. Repeat option for increased levels of verbose output (-v, -vv, or -vvv).

Use "wbctl update [command] --help" for more information about a command.
</pre>
