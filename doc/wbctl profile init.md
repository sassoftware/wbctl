<pre>
Creates or replaces a named profile in the config file, prompting for the profile name, endpoint, and organization

Usage:
  wbctl profile init [flags]

Flags:
  -h, --help   help for init

Global Flags:
  -m, --auth-mode string      Authentication mode. One of: device, pkce. (default "device")
  -e, --endpoint string       The Workbench API endpoint URL.
      --ignore-token-cache    Ignore token cache.
      --no-listen             Do not open a local port to listen for PKCE authentication.
      --organization string   The organization to authenticate with the Workbench API.
  -p, --profile string        The configuration profile to use. (default "Default")
  -v, --verbose               Log verbosity. Repeat option for increased levels of verbose output (-v, -vv, or -vvv).
</pre>
