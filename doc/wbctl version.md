<pre>
Print version information of the wbctl binary.

Usage:
  wbctl version [flags]

Flags:
  -h, --help            help for version
  -o, --output string   The output format. One of: json, text. (default "text")

Global Flags:
  -m, --auth-mode string      Authentication mode. One of: device, pkce. (default "device")
  -e, --endpoint string       The Workbench API endpoint URL.
      --ignore-token-cache    Ignore token cache.
      --no-listen             Do not open a local port to listen for PKCE authentication.
      --organization string   The organization to authenticate with the Workbench API.
  -p, --profile string        The configuration <a href='wbctl profile.md'>profile</a> to use. (default "Default")
  -v, --verbose               Log verbosity. Repeat option for increased levels of verbose output (-v, -vv, or -vvv).
</pre>
