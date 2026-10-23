<pre>
Updates all the Image resources identified by a list of ids with the options provided.

Usage:
  wbctl update images [flags]

Aliases:
  images, image

Flags:
      --context string       The context (volume, snapshot, or s3 bucket) for the image.
      --description string   The description of the image.
  -d, --dry-run              Only print the object that would be sent, without sending it.
  -f, --filename string      Path to a file containing the image in JSON format. Use '-' to read from stdin.
  -h, --help                 help for images
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
  -v, --verbose               Log verbosity. Repeat option for increased levels of verbose output (-v, -vv, or -vvv).
</pre>
