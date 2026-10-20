<pre>
Prints a table of basic information about the Initializer resources identified by an optional list of resource ids. All resources printed if no ids provided.

Usage:
  wbctl get initializers [flags]

Aliases:
  initializers, initializer

Flags:
  -h, --help            help for initializers
  -o, --output string   The output format. One of: json, text. (default "text")
      --sort string     The sort column. Direction specified by appending '.asc' or '.desc' to the column name.

Global Flags:
  -m, --auth-mode string      Authentication mode. One of: auto, pkce, client-credentials, password, device. (default "device")
      --client-id string      The client-id to authenticate with the Workbench API ($WBCLIENTID).
  -e, --endpoint string       The Workbench API endpoint URL.
      --ignore-token-cache    Ignore token cache.
      --no-listen             Do not open a local port to listen for PKCE authentication.
      --organization string   The organization to authenticate with the Workbench API.
</pre>
