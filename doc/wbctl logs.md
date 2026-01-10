<pre>
Fetches the logs of one or many Workbench resources

Usage:
  wbctl logs [command]

Available Commands:
  <a href='wbctl logs image-builds.md'>image-builds</a>    Print the log of one or many Image-Build resources
  <a href='wbctl logs job-submissions.md'>job-submissions</a> Print the log of one or many Job-Submission resources

Flags:
  -h, --help   help for logs

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

Use "wbctl logs [command] --help" for more information about a command.
</pre>
