### Release Workflow

Merging a `release: <tag>` commit to `main` retains the existing tag creation
flow on GitHub's `ubuntu-latest` runner. The workflow then calls
`make -f Makefile-ci.mk build-sign-binaries` on GitHub's hosted `macos-latest` runner and publishes
binaries, checksums, and both installers as
GitHub Release assets attached to that tag. Binaries are not committed to Git.

Hosted builds use `auto` signing by default and do not require signing secrets
unless native Apple or Windows signing is configured. The workflow calls
`make -f Makefile-ci.mk build-sign-binaries` to import
temporary native certificates through the same Make task used locally, build
and sign the binaries, and remove the certificates even on failure. Keyless signing uses Cosign and the workflow's
GitHub OIDC identity; the calling job must grant `id-token: write`.
Go is installed by `actions/setup-go`; the macOS image supplies Xcode command line
tools, Make, and GitHub CLI. This workflow does not run pull request code.

To override the hosted signing mode, configure this repository Actions setting:

| Setting | Purpose |
| --- | --- |
| Variable `SIGNING_MODE` | `auto` (default), `keyless` to ignore native certificates, or `unsigned` to disable signing. Replace any previous `gpg` setting with `keyless`. |

For native signing in `auto` mode, also configure:

| Setting | Purpose |
| --- | --- |
| Secret `MACOS_CERTIFICATE_P12` | Base64-encoded Developer ID Application PKCS#12 certificate with its private key. |
| Secret `MACOS_CERTIFICATE_PASSWORD` | Password for the Apple PKCS#12 certificate. |
| Variable `MACOS_SIGN_IDENTITY` | Developer ID Application identity matching the imported Apple certificate. |
| Secret `APPLE_ID` | Apple developer account email for notarization. |
| Secret `APPLE_APP_SPECIFIC_PASSWORD` | App-specific password for that Apple account. |
| Variable `APPLE_TEAM_ID` | Apple Developer team ID associated with the signing certificate. |
| Secret `WINDOWS_CERTIFICATE_P12` | Base64-encoded Windows code-signing PKCS#12 certificate with its private key. |
| Secret `WINDOWS_SIGN_PASSWORD` | Password for the Windows PKCS#12 certificate. |

With both certificates present, `auto` signs macOS and Windows natively and
Linux with keyless signing. Without native certificates, it signs all binaries
with keyless signing.
If only one native certificate is present, only that platform uses native signing.
Use `keyless` to force keyless signing regardless of configured certificates.
The checksum manifest is also keyless-signed in both signed modes. Cosign
publishes signing records to Sigstore's public transparency log.

After native Apple signing, `make -f Makefile-ci.mk notarize-macos`
submits a temporary ZIP of the macOS
binary using `xcrun notarytool` and waits up to 30 minutes for an `Accepted`
result before uploading or publishing assets. Missing notarization credentials,
rejection, or timeout fails the build. Keyless-only and unsigned builds skip this
step. The published raw executable is unchanged, so checksums remain valid.
Standalone Mach-O executables cannot have notarization tickets stapled; Gatekeeper
may need network access to retrieve Apple's ticket.

Signing or identity-token failures fail signed builds. The workflow installs
Cosign for signed builds and `osslsigncode` when needed. Sigstore bundles are
included in published release assets. Never store signing material in the
repository.

## Building Binaries

On a Mac with Go (the version in [go.mod](go.mod)), Make, and Xcode command line
tools installed:

```shell
make -f Makefile-ci.mk build-binaries
```

This writes Linux x64, Windows x64, and universal macOS (Intel + Apple Silicon)
binaries plus `checksums.txt` to `build/release`. All platforms are built on the
Mac; no Linux or Windows build host is required.

### Optional Signing

Local signing is off by default for `make -f Makefile-ci.mk build-binaries`.
Choose a signing mode through `SIGNING_MODE`:

| `SIGNING_MODE` | Behavior |
| --- | --- |
| `unsigned` (default) | No signatures, even if signing credentials are present. |
| `auto` | Apple signing for macOS and Authenticode for Windows when their certificates are available; keyless signing for Linux and for each platform without a native certificate. |
| `keyless` | Keyless signing for all binaries, ignoring Apple and Windows credentials. |

Both signed modes require Cosign v3 and also sign `checksums.txt`. Local keyless
signing opens an interactive OIDC login; hosted builds use GitHub's OIDC token.
For example:

```shell
SIGNING_MODE=auto make -f Makefile-ci.mk build-binaries
SIGNING_MODE=keyless make -f Makefile-ci.mk build-binaries
```

Configure credentials through these environment variables:

| Variable | Purpose |
| --- | --- |
| `MACOS_SIGN_IDENTITY` | Developer ID Application identity in the build user's unlocked macOS keychain, with its private key. |
| `WINDOWS_SIGN_CERT_FILE` | Absolute path to a code-signing PKCS#12 (`.pfx`/`.p12`) certificate and private key on the build Mac. Requires `osslsigncode` (install with `brew install osslsigncode`). |
| `WINDOWS_SIGN_PASSWORD` | Password for the Windows certificate. Supply through a secret, not source control or command history. |
| `WINDOWS_TIMESTAMP_URL` | Optional RFC 3161 timestamp service; defaults to `http://timestamp.digicert.com`. |

To test certificate import and cleanup as well, use the same task as the workflow:

```shell
make -f Makefile-ci.mk build-sign-binaries
```

This task defaults to `SIGNING_MODE=auto`. It accepts `MACOS_CERTIFICATE_P12`,
`MACOS_CERTIFICATE_PASSWORD`, and `WINDOWS_CERTIFICATE_P12` as environment variables
with the same base64-encoded values used by the workflow. Alternatively, use an
existing unlocked macOS keychain and/or `WINDOWS_SIGN_CERT_FILE`. Install Cosign v3
and, for Windows native signing, `osslsigncode` before running locally. Temporary
certificates are removed and the original keychain search list is restored on
success or failure.

The Mac binary is signed after combining architectures; checksums are generated
after all signing. Notarization is a separate task, used by the workflow
and available locally after native Apple signing:

```shell
make -f Makefile-ci.mk notarize-macos
make -f Makefile-ci.mk notarize-macos BINARY=/path/to/wbctl-darwin-universal
```

Export `APPLE_ID`, `APPLE_APP_SPECIFIC_PASSWORD`, and `APPLE_TEAM_ID` securely
before running it. Missing credentials, rejection, or timeout fails the task.
`BUILDDIR` overrides the build output directory and default notarization path.
Provision and unlock the signing
keychain securely before enabling signing; never commit certificates or passwords.

Keyless signing creates `.sigstore.json` bundles after native signing and
checksum generation. Natively signed binaries do not get an additional bundle.
An invalid supplied certificate or any signing failure fails the build; fallback
only applies to absent certificates. Keyless
signatures do not replace Apple or Windows code signing and do not suppress
Gatekeeper or SmartScreen warnings. Verify downloaded assets against the exact
publishing workflow identity, substituting the trusted repository owner and name:

```shell
identity='https://github.com/OWNER/REPOSITORY/.github/workflows/release.yaml@refs/heads/main'
cosign verify-blob --bundle checksums.txt.sigstore.json \
	--certificate-identity "$identity" \
	--certificate-oidc-issuer https://token.actions.githubusercontent.com checksums.txt
cosign verify-blob --bundle wbctl-linux-amd64.sigstore.json \
	--certificate-identity "$identity" \
	--certificate-oidc-issuer https://token.actions.githubusercontent.com wbctl-linux-amd64
```

The installers check SHA-256 checksums only; they do not verify Sigstore bundles.
