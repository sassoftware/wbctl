<pre>
Generate the autocompletion script for the bash shell.

This script depends on the 'bash-completion' package.
If it is not installed already, you can install it via your OS's package manager.

To load completions in your current shell session:

	source <(wbctl completion bash)

To load completions for every new session, execute once:

#### Linux:

	wbctl completion bash > /etc/bash_completion.d/wbctl

#### macOS:

	wbctl completion bash > $(brew --prefix)/etc/bash_completion.d/wbctl

You will need to start a new shell for this setup to take effect.

Usage:
  wbctl completion bash

Flags:
  -h, --help              help for bash
      --no-descriptions   disable completion descriptions

Global Flags:
  -m, --auth-mode string      Authentication mode. One of: device, pkce. (default "device")
  -e, --endpoint string       The Workbench API endpoint URL.
      --ignore-token-cache    Ignore token cache.
      --no-listen             Do not open a local port to listen for PKCE authentication.
      --organization string   The organization to authenticate with the Workbench API.
</pre>
