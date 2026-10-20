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
  -m, --auth-mode string      Authentication mode. One of: auto, pkce, client-credentials, password, device. (default "device")
      --client-id string      The client-id to authenticate with the Workbench API ($WBCLIENTID).
  -e, --endpoint string       The Workbench API endpoint URL.
      --ignore-token-cache    Ignore token cache.
      --no-listen             Do not open a local port to listen for PKCE authentication.
      --organization string   The organization to authenticate with the Workbench API.

Use "wbctl stop [command] --help" for more information about a command.
</pre>
