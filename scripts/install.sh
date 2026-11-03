#!/bin/bash
# Copyright © 2026, SAS Institute Inc., Cary, NC, USA.  All Rights Reserved.
# SPDX-License-Identifier: Apache-2.0
#
# This script installs the wbctl binary for the current platform, verifying checksums.
# It supports macOS Intel/Apple Silicon and Linux x64 platforms.
set -euo pipefail

version=${1:-latest}
repository=${WBCTL_REPOSITORY:-sassoftware/wbctl}
install_dir=${WBCTL_INSTALL_DIR:-$HOME/.local/bin}

case "$(uname -s)/$(uname -m)" in
    Darwin/x86_64|Darwin/arm64) asset=wbctl-darwin-universal ;;
    Linux/x86_64|Linux/amd64) asset=wbctl-linux-amd64 ;;
    *) echo 'Supported platforms: macOS Intel/Apple Silicon and Linux x64.' >&2; exit 1 ;;
esac

base_url="https://github.com/$repository/releases"
if [[ $version == latest ]]; then
    release_url=$(curl --fail --silent --show-error --location --output /dev/null \
        --write-out '%{url_effective}' "$base_url/latest")
    version=${release_url##*/}
fi
if [[ ! $version =~ ^[a-zA-Z0-9][a-zA-Z0-9._+-]*$ || $version == latest ]]; then
    echo 'Expected a release tag, or latest.' >&2
    exit 1
fi

temporary=$(mktemp -d)
trap 'rm -rf "$temporary"' EXIT
for filename in "$asset" checksums.txt; do
    curl --fail --silent --show-error --location \
        "$base_url/download/$version/$filename" --output "$temporary/$filename"
done

expected=$(awk -v asset="$asset" '$2 == asset {print $1}' "$temporary/checksums.txt")
if command -v sha256sum >/dev/null; then
    actual=$(sha256sum "$temporary/$asset" | awk '{print $1}')
else
    actual=$(shasum -a 256 "$temporary/$asset" | awk '{print $1}')
fi
if [[ ! $expected =~ ^[a-fA-F0-9]{64}$ || $actual != "$expected" ]]; then
    echo 'Checksum verification failed; nothing was installed.' >&2
    exit 1
fi

mkdir -p "$install_dir"
install -m 755 "$temporary/$asset" "$install_dir/wbctl"
printf 'Installed wbctl %s to %s/wbctl\n' "$version" "$install_dir"
case ":$PATH:" in
    *":$install_dir:"*) ;;
    *) printf 'Add this directory to your PATH: %s\n' "$install_dir" ;;
esac
