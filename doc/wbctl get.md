<pre>
Prints a table of basic information about one or many Workbench resources.

Usage:
  wbctl get [command]

Available Commands:
  <a href='wbctl get applications.md'>applications</a>     Get one or many Application resources
  <a href='wbctl get dataplanes.md'>dataplanes</a>       Get one or many Dataplane resources
  <a href='wbctl get events.md'>events</a>           Get one or many Event resources
  <a href='wbctl get image-builds.md'>image-builds</a>     Get one or many Image-Build resources
  <a href='wbctl get images.md'>images</a>           Get one or many Image resources
  <a href='wbctl get initializers.md'>initializers</a>     Get one or many Initializer resources
  <a href='wbctl get job-submissions.md'>job-submissions</a>  Get one or many Job-Submission resources
  <a href='wbctl get jobs.md'>jobs</a>             Get one or many Job resources
  <a href='wbctl get nodepools.md'>nodepools</a>        Get one or many Nodepool resources
  <a href='wbctl get pings.md'>pings</a>            Get one or many Ping resources
  <a href='wbctl get services.md'>services</a>         Get one or many Service resources
  <a href='wbctl get snapshots.md'>snapshots</a>        Get one or many Snapshot resources
  <a href='wbctl get storageclasses.md'>storageclasses</a>   Get one or many Storageclass resources
  <a href='wbctl get users.md'>users</a>            Get one or many User resources
  <a href='wbctl get volume-conflicts.md'>volume-conflicts</a> Get one or many Volume-Conflict resources
  <a href='wbctl get volumes.md'>volumes</a>          Get one or many Volume resources
  <a href='wbctl get workbenches.md'>workbenches</a>      Get one or many Workbench resources

Flags:
  -h, --help   help for get

Global Flags:
  -m, --auth-mode string      Authentication mode. One of: device, pkce. (default "device")
  -e, --endpoint string       The Workbench API endpoint URL.
      --ignore-token-cache    Ignore token cache.
      --no-listen             Do not open a local port to listen for PKCE authentication.
      --organization string   The organization to authenticate with the Workbench API.

Use "wbctl get [command] --help" for more information about a command.
</pre>
