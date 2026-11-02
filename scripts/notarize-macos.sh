#!/usr/bin/env bash
# Copyright © 2026, SAS Institute Inc., Cary, NC, USA.  All Rights Reserved.
# SPDX-License-Identifier: Apache-2.0
set -euo pipefail

: "${APPLE_ID:?Set APPLE_ID for notarization}"
: "${APPLE_APP_SPECIFIC_PASSWORD:?Set APPLE_APP_SPECIFIC_PASSWORD for notarization}"
: "${APPLE_TEAM_ID:?Set APPLE_TEAM_ID for notarization}"
binary=${1:-build/release/wbctl-darwin-universal}
codesign --verify --strict "$binary"
temporary=$(mktemp -d)
trap 'rm -rf "$temporary"' EXIT
archive="$temporary/notarization.zip"
result="$temporary/notarization.json"
ditto -c -k --keepParent "$binary" "$archive"
xcrun notarytool submit "$archive" \
    --apple-id "$APPLE_ID" --password "$APPLE_APP_SPECIFIC_PASSWORD" \
    --team-id "$APPLE_TEAM_ID" --wait --timeout 30m --output-format json > "$result"
cat "$result"
if [[ $(plutil -extract status raw -o - "$result") != Accepted ]]; then
    echo 'Apple did not accept the notarization submission.' >&2
    exit 1
fi
