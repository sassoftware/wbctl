<pre>
Creates a new Job-Submission resource with the options provided.

Usage:
  wbctl create job-submission [flags]

Aliases:
  job-submission, job-submissions

Flags:
      --dataplane string      The ID of the dataplane where the job-submission will be created. If only one dataplane is available, this flag can be omitted.
  -d, --dry-run               Only print the object that would be sent, without sending it.
      --environment strings   Environment variable(s) to set in the job-submission's services, in KEY=VALUE format.
  -f, --filename string       Path to a file containing the job-submission in JSON format. Use '-' to read from stdin.
  -h, --help                  help for job-submission
      --job-id string         The ID of the <a href='wbctl create job.md'>job</a> to run.
  -l, --label strings         Labels to apply to the job-submission, in key=value format.
      --nodepool string       The nodepool in which to run the job-submission.
      --volume strings        Volume(s) to attach to the job-submission.

Global Flags:
  -m, --auth-mode string      Authentication mode. One of: auto, pkce, client-credentials, password, device. (default "device")
      --client-id string      The client-id to authenticate with the Workbench API ($WBCLIENTID).
  -e, --endpoint string       The Workbench API endpoint URL.
      --ignore-token-cache    Ignore token cache.
      --no-listen             Do not open a local port to listen for PKCE authentication.
      --organization string   The organization to authenticate with the Workbench API.
</pre>
