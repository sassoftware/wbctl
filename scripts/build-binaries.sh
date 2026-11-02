#!/usr/bin/env bash
# Copyright © 2026, SAS Institute Inc., Cary, NC, USA.  All Rights Reserved.
# SPDX-License-Identifier: Apache-2.0
#
# This script builds release binaries for different platforms and handles signing
# based on the specified SIGNING_MODE. The output is placed in the directory
# specified by the first argument (default: build/release).

set -euo pipefail

output=${1:-build/release}
go_command=${GO:-go}
signing_mode=${SIGNING_MODE:-unsigned}
case "$signing_mode" in
    unsigned|auto|keyless) ;;
    *) echo "Unsupported SIGNING_MODE: $signing_mode (expected unsigned, auto, or keyless)" >&2; exit 1 ;;
esac
if [[ "$signing_mode" != auto ]]; then
    MACOS_SIGN_IDENTITY=
    WINDOWS_SIGN_CERT_FILE=
fi
if [[ "$signing_mode" != unsigned ]]; then
    command -v cosign >/dev/null || {
        echo 'Signed releases require Cosign v3. Install it with: brew install cosign' >&2
        echo 'To build without signing, set SIGNING_MODE=unsigned.' >&2
        exit 1
    }
fi
command -v lipo >/dev/null || { echo 'Release builds require macOS with Xcode command line tools (lipo).' >&2; exit 1; }
if [[ -n ${MACOS_SIGN_IDENTITY:-} ]]; then
    command -v codesign >/dev/null
fi
if [[ -n ${WINDOWS_SIGN_CERT_FILE:-} ]]; then
    command -v osslsigncode >/dev/null
    : "${WINDOWS_SIGN_PASSWORD:?Set WINDOWS_SIGN_PASSWORD for the signing certificate}"
fi

mkdir -p "$output"
temporary=$(mktemp -d)
trap 'rm -rf "$temporary"' EXIT

build() {
    CGO_ENABLED=0 GOOS="$1" GOARCH="$2" "$go_command" build -trimpath -o "$3" ./cmd/wbctl
}

build linux amd64 "$output/wbctl-linux-amd64"
build windows amd64 "$output/wbctl-windows-amd64.exe"
build darwin amd64 "$temporary/wbctl-amd64"
build darwin arm64 "$temporary/wbctl-arm64"
lipo -create "$temporary/wbctl-amd64" "$temporary/wbctl-arm64" -output "$output/wbctl-darwin-universal"
lipo -verify_arch x86_64 "$output/wbctl-darwin-universal"
lipo -verify_arch arm64 "$output/wbctl-darwin-universal"

if [[ -n ${MACOS_SIGN_IDENTITY:-} ]]; then
    codesign --force --options runtime --timestamp --sign "$MACOS_SIGN_IDENTITY" "$output/wbctl-darwin-universal"
    codesign --verify --strict "$output/wbctl-darwin-universal"
fi
if [[ -n ${WINDOWS_SIGN_CERT_FILE:-} ]]; then
    osslsigncode sign -pkcs12 "$WINDOWS_SIGN_CERT_FILE" -readpass <(printf '%s' "$WINDOWS_SIGN_PASSWORD") \
        -h sha256 -ts "${WINDOWS_TIMESTAMP_URL:-http://timestamp.digicert.com}" \
        -in "$output/wbctl-windows-amd64.exe" -out "$temporary/wbctl-signed.exe"
    mv "$temporary/wbctl-signed.exe" "$output/wbctl-windows-amd64.exe"
fi

chmod +x "$output/wbctl-linux-amd64" "$output/wbctl-darwin-universal"
cd "$output"
shasum -a 256 wbctl-linux-amd64 wbctl-darwin-universal wbctl-windows-amd64.exe > checksums.txt
for asset in wbctl-linux-amd64 wbctl-darwin-universal wbctl-windows-amd64.exe checksums.txt; do
    rm -f "$asset.asc" "$asset.sigstore.json"
    if [[ "$signing_mode" == unsigned ]]; then
        continue
    fi
    if [[ "$asset" == wbctl-darwin-universal && -n ${MACOS_SIGN_IDENTITY:-} ]] ||
        [[ "$asset" == wbctl-windows-amd64.exe && -n ${WINDOWS_SIGN_CERT_FILE:-} ]]; then
        continue
    fi
    cosign sign-blob --yes --bundle "$asset.sigstore.json" "$asset"
done
