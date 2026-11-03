#!/usr/bin/env bash
# Copyright © 2026, SAS Institute Inc., Cary, NC, USA.  All Rights Reserved.
# SPDX-License-Identifier: Apache-2.0
set -euo pipefail

export SIGNING_MODE=${SIGNING_MODE:-auto}
case "$SIGNING_MODE" in
    unsigned|auto|keyless) ;;
    *) echo "Invalid signing mode: $SIGNING_MODE" >&2; exit 1 ;;
esac

umask 077
temporary=$(mktemp -d)
cleanup() {
    exit_code=$?
    trap - EXIT
    if [[ -f "$temporary/keychains" ]]; then
        keychains=()
        while IFS= read -r existing; do
            existing=${existing#*\"}
            keychains+=("${existing%\"*}")
        done < "$temporary/keychains"
        security list-keychains -d user -s "${keychains[@]}" || exit_code=1
    fi
    if [[ -f "$temporary/signing.keychain-db" ]]; then
        security delete-keychain "$temporary/signing.keychain-db" || exit_code=1
    fi
    rm -rf "$temporary"
    exit "$exit_code"
}
trap cleanup EXIT

if [[ "$SIGNING_MODE" == auto ]]; then
    if [[ -n ${MACOS_CERTIFICATE_P12:-} ]]; then
        : "${MACOS_SIGN_IDENTITY:?Set MACOS_SIGN_IDENTITY to the Developer ID Application identity}"
        certificate="$temporary/apple.p12"
        keychain="$temporary/signing.keychain-db"
        keychain_password=$(openssl rand -hex 32)
        printf '%s' "$MACOS_CERTIFICATE_P12" | base64 --decode > "$certificate"
        security list-keychains -d user > "$temporary/keychains"
        security create-keychain -p "$keychain_password" "$keychain"
        security set-keychain-settings -lut 21600 "$keychain"
        security unlock-keychain -p "$keychain_password" "$keychain"
        security import "$certificate" -P "${MACOS_CERTIFICATE_PASSWORD:-}" -A -t cert -f pkcs12 -k "$keychain"
        security set-key-partition-list -S apple-tool:,apple:,codesign: -s -k "$keychain_password" "$keychain"
        keychains=()
        while IFS= read -r existing; do
            existing=${existing#*\"}
            keychains+=("${existing%\"*}")
        done < "$temporary/keychains"
        security list-keychains -d user -s "$keychain" "${keychains[@]}"
    fi
    if [[ -n ${WINDOWS_CERTIFICATE_P12:-} ]]; then
        printf '%s' "$WINDOWS_CERTIFICATE_P12" | base64 --decode > "$temporary/windows.p12"
        export WINDOWS_SIGN_CERT_FILE="$temporary/windows.p12"
    fi
fi

bash "$(dirname "$0")/build-binaries.sh" "${1:-build/release}"
