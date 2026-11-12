<pre>
Manages named profiles stored in the config file

Usage:
  wbctl profile [command]

Aliases:
  profile, prof

Available Commands:
  <a href='wbctl profile init.md'>init</a>        Create or replace a profile
  <a href='wbctl profile list.md'>list</a>        List the available profiles
  <a href='wbctl profile show.md'>show</a>        Show current profile

Flags:
  -h, --help   help for profile

Global Flags:
  -m, --auth-mode string      Authentication mode. One of: device, pkce. (default "device")
  -e, --endpoint string       The Workbench API endpoint URL.
      --ignore-token-cache    Ignore token cache.
      --no-listen             Do not open a local port to listen for PKCE authentication.
      --organization string   The organization to authenticate with the Workbench API.
  -p, --profile string        The configuration profile to use. (default "Default")
  -v, --verbose               Log verbosity. Repeat option for increased levels of verbose output (-v, -vv, or -vvv).

Use "wbctl profile [command] --help" for more information about a command.
</pre>
