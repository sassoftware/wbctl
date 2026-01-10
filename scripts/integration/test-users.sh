#!/usr/bin/env bash
# Copyright © 2026, SAS Institute Inc., Cary, NC, USA.  All Rights Reserved.
# SPDX-License-Identifier: Apache-2.0
#
# test-users.sh — smoke tests for `wbctl get/describe users`.
# Can be run standalone or via integration-test.sh.

set -euo pipefail
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
LABEL="test-users"
source "$SCRIPT_DIR/common.sh"

require_wbconfig

# -- tests ------------------------------------------------------------------

run_test_table_has_headers_and_data "get users" users "ID" "NAME" "EMAIL"

run_test_json_object \
    "get users --output json returns JSON array" \
    wbctl get users --output json

# Describe the first available user — ID sourced from captured get output.
USER_ID="$_LAST_TABLE_ID"
if [ -n "$USER_ID" ]; then
    run_test_output_contains \
        "describe user contains $USER_ID" \
        "$USER_ID" \
        wbctl describe user "$USER_ID"
else
    printf '[%s] SKIP: no users found; skipping describe test.\n' "$LABEL"
fi

# -- summary ----------------------------------------------------------------
print_summary
