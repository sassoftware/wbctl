<pre>
Print a table of basic information about the Event resources identified by an optional list of resource ids. All resources printed if no ids provided.

Usage:
  wbctl get events [flags]

Aliases:
  events, event

Flags:
  -h, --help            help for events
  -o, --output string   The output format. One of: json, text. (default "text")
      --sort string     The sort column. Direction specified by appending '.asc' or '.desc' to the column name.

Global Flags:
  -m, --auth-mode string      Authentication mode. One of: device, pkce. (default "device")
  -e, --endpoint string       The Workbench API endpoint URL.
      --ignore-token-cache    Ignore token cache.
      --no-listen             Do not open a local port to listen for PKCE authentication.
      --organization string   The organization to authenticate with the Workbench API.
  -v, --verbose               Log verbosity. Repeat option for increased levels of verbose output (-v, -vv, or -vvv).
</pre>
