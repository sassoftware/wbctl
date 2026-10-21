<pre>
Start the Job-Submission resources identified by a list of resource ids. At least one id is required.

Usage:
  wbctl start jobs [flags]

Aliases:
  jobs, job

Flags:
      --dataplane string      The ID of the dataplane where the job-submission will be created. If only one dataplane is available, this flag can be omitted.
  -d, --dry-run               Only print the object that would be sent, without sending it.
      --environment strings   Environment variable(s) to set in the job-submission's services, in KEY=VALUE format.
  -h, --help                  help for jobs
  -l, --label strings         Labels to apply to the job-submission, in key=value format.
      --nodepool string       The nodepool in which to run the job-submission.
      --volume strings        Volume(s) to attach to the job-submission.

Global Flags:
  -m, --auth-mode string      Authentication mode. One of: device, pkce. (default "device")
  -e, --endpoint string       The Workbench API endpoint URL.
      --ignore-token-cache    Ignore token cache.
      --no-listen             Do not open a local port to listen for PKCE authentication.
      --organization string   The organization to authenticate with the Workbench API.
</pre>
