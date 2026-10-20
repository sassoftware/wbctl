<pre>
Generate the autocompletion script for the fish shell.

To load completions in your current shell session:

	wbctl completion fish | source

To load completions for every new session, execute once:

	wbctl completion fish > ~/.config/fish/completions/wbctl.fish

You will need to start a new shell for this setup to take effect.

Usage:
  wbctl completion fish [flags]

Flags:
  -h, --help              help for fish
      --no-descriptions   disable completion descriptions

Global Flags:
  -m, --auth-mode string      Authentication mode. One of: auto, pkce, client-credentials, password, device. (default "device")
      --client-id string      The client-id to authenticate with the Workbench API ($WBCLIENTID).
  -e, --endpoint string       The Workbench API endpoint URL.
      --ignore-token-cache    Ignore token cache.
      --no-listen             Do not open a local port to listen for PKCE authentication.
      --organization string   The organization to authenticate with the Workbench API.
</pre>
