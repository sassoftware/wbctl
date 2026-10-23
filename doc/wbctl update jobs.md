<pre>
Updates all the Job resources identified by a list of ids with the options provided.

Usage:
  wbctl update jobs [flags]

Aliases:
  jobs, job

Flags:
  -a, --application string    The application to run in the job. Not required if at least one service is specified.
  -c, --command string        The command to run in the job.
      --context string        The context (volume, snapshot, or s3 bucket) for the job.
      --description string    The description of the job.
  -d, --dry-run               Only print the object that would be sent, without sending it.
      --environment strings   Environment variable(s) to set in the job's services, in KEY=VALUE format.
  -f, --filename string       Path to a file containing the job in JSON format. Use '-' to read from stdin.
  -h, --help                  help for jobs
  -l, --label strings         Labels to apply to the job, in key=value format.
  -n, --name string           The name of the job.
      --nodepool string       The nodepool in which to run the job.
      --schedule string       The cron schedule for the job. If not specified, the job will only run when a job-submission is explicitly created.
  -s, --service strings       The Service(s) to run in job. At least one service is required if application is not specified.
      --sub-path string       The subpath of the job. Only valid when context is specified.
      --volume strings        Volume(s) to attach to the job.

Global Flags:
  -m, --auth-mode string      Authentication mode. One of: device, pkce. (default "device")
  -e, --endpoint string       The Workbench API endpoint URL.
      --ignore-token-cache    Ignore token cache.
      --no-listen             Do not open a local port to listen for PKCE authentication.
      --organization string   The organization to authenticate with the Workbench API.
  -v, --verbose               Log verbosity. Repeat option for increased levels of verbose output (-v, -vv, or -vvv).
</pre>
