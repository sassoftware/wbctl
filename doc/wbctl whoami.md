<pre>
Loads any cached identity token and displays information about the authenticated user.

Usage:
  wbctl whoami [flags]

Flags:
  -h, --help            help for whoami
  -o, --output string   The output format. One of: json, text. (default "text")

Global Flags:
  -m, --auth-mode string      Authentication mode. One of: device, pkce. (default "device")
  -e, --endpoint string       The Workbench API endpoint URL.
      --ignore-token-cache    Ignore token cache.
      --no-listen             Do not open a local port to listen for PKCE authentication.
      --organization string   The organization to authenticate with the Workbench API.
  -p, --profile string        The configuration profile to use. (default "Default")
  -v, --verbose               Log verbosity. Repeat option for increased levels of verbose output (-v, -vv, or -vvv).
</pre>
