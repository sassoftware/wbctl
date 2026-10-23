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
  -m, --auth-mode string      Authentication mode. One of: device, pkce. (default "device")
  -e, --endpoint string       The Workbench API endpoint URL.
      --ignore-token-cache    Ignore token cache.
      --no-listen             Do not open a local port to listen for PKCE authentication.
      --organization string   The organization to authenticate with the Workbench API.
  -v, --verbose               Log verbosity. Repeat option for increased levels of verbose output (-v, -vv, or -vvv).

Use "wbctl delete [command] --help" for more information about a command.
</pre>
