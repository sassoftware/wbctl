<pre>
Updates all the Workbench resources identified by a list of ids with the options provided.

Usage:
  wbctl update workbenches [flags]

Aliases:
  workbenches, workbench

Flags:
  -a, --application string     The application to run in the workbench. Not required if at least one service is specified.
      --desired-state string   The desired state of the workbench (e.g., RUNNING, STOPPED).
  -d, --dry-run                Only print the object that would be sent, without sending it.
  -f, --filename string        Path to a file containing the workbench in JSON format. Use '-' to read from stdin.
  -h, --help                   help for workbenches
  -l, --label strings          Labels to apply to the workbench, in key=value format.
  -n, --name string            The name of the workbench.
      --nodepool string        The nodepool in which to run the workbench.
  -s, --service strings        The Service(s) to run in workbench. At least one service is required if application is not specified.
      --volume strings         Volume(s) to attach to the workbench.

Global Flags:
  -m, --auth-mode string      Authentication mode. One of: device, pkce. (default "device")
  -e, --endpoint string       The Workbench API endpoint URL.
      --ignore-token-cache    Ignore token cache.
      --no-listen             Do not open a local port to listen for PKCE authentication.
      --organization string   The organization to authenticate with the Workbench API.
</pre>
