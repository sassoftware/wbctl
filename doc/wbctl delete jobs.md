<pre>
Deletes the Job resources identified by a list of resource ids. At least one id is required.

Usage:
  wbctl delete jobs [flags]

Aliases:
  jobs, job

Flags:
  -h, --help   help for jobs

Global Flags:
  -m, --auth-mode string      Authentication mode. One of: device, pkce. (default "device")
  -e, --endpoint string       The Workbench API endpoint URL.
      --ignore-token-cache    Ignore token cache.
      --no-listen             Do not open a local port to listen for PKCE authentication.
      --organization string   The organization to authenticate with the Workbench API.
  -p, --profile string        The configuration profile to use. (default "Default")
  -v, --verbose               Log verbosity. Repeat option for increased levels of verbose output (-v, -vv, or -vvv).
</pre>
