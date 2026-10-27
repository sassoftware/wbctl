<pre>
Stops one or many Workbench resources

Usage:
  wbctl stop [command]

Available Commands:
  <a href='wbctl stop job-submission.md'>job-submission</a> Cancel one or many Job-Submission resources
  <a href='wbctl stop workbenches.md'>workbenches</a>    Stop one or many Workbench resources

Flags:
  -h, --help   help for stop

Global Flags:
  -m, --auth-mode string      Authentication mode. One of: device, pkce. (default "device")
  -e, --endpoint string       The Workbench API endpoint URL.
      --ignore-token-cache    Ignore token cache.
      --no-listen             Do not open a local port to listen for PKCE authentication.
      --organization string   The organization to authenticate with the Workbench API.
  -p, --profile string        The configuration profile to use. (default "Default")
  -v, --verbose               Log verbosity. Repeat option for increased levels of verbose output (-v, -vv, or -vvv).

Use "wbctl stop [command] --help" for more information about a command.
</pre>
