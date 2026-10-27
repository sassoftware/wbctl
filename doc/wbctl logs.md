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
  -m, --auth-mode string      Authentication mode. One of: device, pkce. (default "device")
  -e, --endpoint string       The Workbench API endpoint URL.
      --ignore-token-cache    Ignore token cache.
      --no-listen             Do not open a local port to listen for PKCE authentication.
      --organization string   The organization to authenticate with the Workbench API.
  -p, --profile string        The configuration profile to use. (default "Default")
  -v, --verbose               Log verbosity. Repeat option for increased levels of verbose output (-v, -vv, or -vvv).

Use "wbctl logs [command] --help" for more information about a command.
</pre>
