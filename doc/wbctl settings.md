<pre>
Displays or sets user settings.

Usage:
  wbctl settings [flags]

Aliases:
  settings, setting

Flags:
  -h, --help             help for settings
  -o, --output string    The output format. One of: json, text. (default "text")
      --set string       The name of a setting to set.
      --tenant           If true, manage tenant settings.
      --user-id string   The user's settings to fetch or set.
      --value string     The value of a setting to set.

Global Flags:
  -m, --auth-mode string      Authentication mode. One of: device, pkce. (default "device")
  -e, --endpoint string       The Workbench API endpoint URL.
      --ignore-token-cache    Ignore token cache.
      --no-listen             Do not open a local port to listen for PKCE authentication.
      --organization string   The organization to authenticate with the Workbench API.
  -v, --verbose               Log verbosity. Repeat option for increased levels of verbose output (-v, -vv, or -vvv).
</pre>
