<pre>
Deletes one or many Workbench resources.

Usage:
  wbctl delete [command]

Available Commands:
  <a href='wbctl delete images.md'>images</a>          Delete one or many Image resources
  <a href='wbctl delete job-submissions.md'>job-submissions</a> Delete one or many Job-Submission resources
  <a href='wbctl delete jobs.md'>jobs</a>            Delete one or many Job resources
  <a href='wbctl delete snapshots.md'>snapshots</a>       Delete one or many Snapshot resources
  <a href='wbctl delete volumes.md'>volumes</a>         Delete one or many Volume resources
  <a href='wbctl delete workbenches.md'>workbenches</a>     Delete one or many Workbench resources

Flags:
  -h, --help   help for delete

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

Use "wbctl delete [command] --help" for more information about a command.
</pre>
