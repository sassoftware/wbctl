<pre>
Prints a table of basic information about the Ping resources identified by an optional list of resource ids. All resources printed if no ids provided.

Usage:
  wbctl get pings [flags]

Aliases:
  pings, ping

Flags:
  -h, --help            help for pings
  -o, --output string   The output format. One of: json, text. (default "text")
      --sort string     The sort column. Direction specified by appending '.asc' or '.desc' to the column name.

Global Flags:
  -m, --auth-mode string       Authentication mode. One of: auto, pkce, client-credentials, password, device. (default "auto")
      --client-id string       The client-id to authenticate with the Workbench API ($WBCLIENTID).
      --client-secret string   The client-secret to authenticate with the Workbench API ($WBCLIENTSECRET).
  -e, --endpoint string        The Workbench API endpoint URL.
      --ignore-token-cache     Ignore token cache.
      --no-listen              Do not open a local port to listen for PKCE authentication.
      --organization string    The organization to authenticate with the Workbench API.
  -p, --password string        The password to authenticate with the Workbench API ($WBPASSWORD).
  -u, --username string        The username to authenticate with the Workbench API ($WBUSERNAME).
</pre>
