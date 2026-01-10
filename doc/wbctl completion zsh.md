<pre>
Generate the autocompletion script for the zsh shell.

If shell completion is not already enabled in your environment you will need
to enable it.  You can execute the following once:

	echo "autoload -U compinit; compinit" >> ~/.zshrc

To load completions in your current shell session:

	source <(wbctl completion zsh)

To load completions for every new session, execute once:

#### Linux:

	wbctl completion zsh > "${fpath[1]}/_wbctl"

#### macOS:

	wbctl completion zsh > $(brew --prefix)/share/zsh/site-functions/_wbctl

You will need to start a new shell for this setup to take effect.

Usage:
  wbctl completion zsh [flags]

Flags:
  -h, --help              help for zsh
      --no-descriptions   disable completion descriptions

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
