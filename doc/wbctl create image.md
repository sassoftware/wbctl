<pre>
Creates a new Image resource with the options provided.

Usage:
  wbctl create image [flags]

Aliases:
  image, images

Flags:
      --context string       The context (volume, snapshot, or s3 bucket) for the image.
      --dataplane string     The ID of the dataplane where the image will be created. If only one dataplane is available, this flag can be omitted.
      --description string   The description of the image.
  -d, --dry-run              Only print the object that would be sent, without sending it.
  -f, --filename string      Path to a file containing the image in JSON format. Use '-' to read from stdin.
  -h, --help                 help for image
  -i, --image string         The image to create. (registry/repository:tag)
  -l, --label strings        Labels to apply to the image, in key=value format.
  -n, --name string          The name of the image.
  -s, --spec string          The image spec to build (file path or literal string).
      --sub-path string      The subpath of the image. Only valid when context is specified.

Global Flags:
  -m, --auth-mode string      Authentication mode. One of: device, pkce. (default "device")
  -e, --endpoint string       The Workbench API endpoint URL.
      --ignore-token-cache    Ignore token cache.
      --no-listen             Do not open a local port to listen for PKCE authentication.
      --organization string   The organization to authenticate with the Workbench API.
</pre>
