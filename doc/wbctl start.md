<pre>
Starts one or many Workbench resources

Usage:
  wbctl start [command]

Available Commands:
  <a href='wbctl start jobs.md'>jobs</a>        Start one or many Job-Submission resources
  <a href='wbctl start workbenches.md'>workbenches</a> Start one or many Workbench resources

Flags:
  -h, --help   help for start

Global Flags:
  -m, --auth-mode string      Authentication mode. One of: device, pkce. (default "device")
  -e, --endpoint string       The Workbench API endpoint URL.
      --ignore-token-cache    Ignore token cache.
      --no-listen             Do not open a local port to listen for PKCE authentication.
      --organization string   The organization to authenticate with the Workbench API.
  -v, --verbose               Log verbosity. Repeat option for increased levels of verbose output (-v, -vv, or -vvv).

Use "wbctl start [command] --help" for more information about a command.
</pre>
