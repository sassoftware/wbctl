<pre>
Generate the autocompletion script for powershell.

To load completions in your current shell session:

	wbctl completion powershell | Out-String | Invoke-Expression

To load completions for every new session, add the output of the above command
to your powershell profile.

Usage:
  wbctl completion powershell [flags]

Flags:
  -h, --help              help for powershell
      --no-descriptions   disable completion descriptions

Global Flags:
  -m, --auth-mode string      Authentication mode. One of: device, pkce. (default "device")
  -e, --endpoint string       The Workbench API endpoint URL.
      --ignore-token-cache    Ignore token cache.
      --no-listen             Do not open a local port to listen for PKCE authentication.
      --organization string   The organization to authenticate with the Workbench API.
</pre>
