<pre>
Authenticates with the Workbench API and stores the resulting token in ~/.wbctl/token.json so that subsequent commands do not require repeated authorization.

Usage:
  wbctl login [flags]

Flags:
  -h, --help   help for login

Global Flags:
  -m, --auth-mode string      Authentication mode. One of: device, pkce. (default "device")
  -e, --endpoint string       The Workbench API endpoint URL.
      --ignore-token-cache    Ignore token cache.
      --no-listen             Do not open a local port to listen for PKCE authentication.
      --organization string   The organization to authenticate with the Workbench API.
  -v, --verbose               Log verbosity. Repeat option for increased levels of verbose output (-v, -vv, or -vvv).
</pre>
