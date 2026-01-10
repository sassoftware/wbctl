<pre>
Generate the autocompletion script for wbctl for the specified shell.
See each sub-command's help for details on how to use the generated script.

Usage:
  wbctl completion [command]

Available Commands:
  <a href='wbctl completion bash.md'>bash</a>        Generate the autocompletion script for bash
  <a href='wbctl completion fish.md'>fish</a>        Generate the autocompletion script for fish
  <a href='wbctl completion powershell.md'>powershell</a>  Generate the autocompletion script for powershell
  <a href='wbctl completion zsh.md'>zsh</a>         Generate the autocompletion script for zsh

Flags:
  -h, --help   help for completion

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

Use "wbctl completion [command] --help" for more information about a command.
</pre>
