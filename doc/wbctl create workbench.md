<pre>
Creates a new Workbench resource with the options provided.

Usage:
  wbctl create workbench [flags]

Aliases:
  workbench, workbenches

Flags:
  -a, --application string     The application to run in the workbench. Not required if at least one service is specified.
      --dataplane string       The ID of the dataplane where the workbench will be created. If only one dataplane is available, this flag can be omitted.
      --desired-state string   The desired state of the workbench (e.g., RUNNING, STOPPED).
  -d, --dry-run                Only print the object that would be sent, without sending it.
  -f, --filename string        Path to a file containing the workbench in JSON format. Use '-' to read from stdin.
  -h, --help                   help for workbench
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
  -v, --verbose               Log verbosity. Repeat option for increased levels of verbose output (-v, -vv, or -vvv).
</pre>
