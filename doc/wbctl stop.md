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
  -m, --auth-mode string       Authentication mode. One of: auto, pkce, client-credentials, password, device. (default "auto")
      --client-id string       The client-id to authenticate with the Workbench API ($WBCLIENTID).
      --client-secret string   The client-secret to authenticate with the Workbench API ($WBCLIENTSECRET).
  -e, --endpoint string        The Workbench API endpoint URL.
      --ignore-token-cache     Ignore token cache.
      --no-listen              Do not open a local port to listen for PKCE authentication.
      --organization string    The organization to authenticate with the Workbench API.
  -p, --password string        The password to authenticate with the Workbench API ($WBPASSWORD).
  -u, --username string        The username to authenticate with the Workbench API ($WBUSERNAME).

Use "wbctl stop [command] --help" for more information about a command.
</pre>
