<pre>
Print the log of the Job-Submission resources identified a list of resource ids. At least one id is required.

Usage:
  wbctl logs job-submissions [flags]

Aliases:
  job-submissions, job-submission

Flags:
  -h, --help   help for job-submissions

Global Flags:
  -m, --auth-mode string      Authentication mode. One of: auto, pkce, client-credentials, password, device. (default "device")
      --client-id string      The client-id to authenticate with the Workbench API ($WBCLIENTID).
  -e, --endpoint string       The Workbench API endpoint URL.
      --ignore-token-cache    Ignore token cache.
      --no-listen             Do not open a local port to listen for PKCE authentication.
      --organization string   The organization to authenticate with the Workbench API.
</pre>
