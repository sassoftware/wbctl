#!/usr/bin/env bash
# Copyright © 2026, SAS Institute Inc., Cary, NC, USA.  All Rights Reserved.
# SPDX-License-Identifier: Apache-2.0
#
# test-dataplanes.sh — smoke tests for `wbctl get/describe dataplanes`.
# Can be run standalone or via integration-test.sh.

set -euo pipefail
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
LABEL="test-dataplanes"
source "$SCRIPT_DIR/common.sh"

require_wbconfig

# -- tests ------------------------------------------------------------------

run_test_table_has_headers_and_data "get dataplanes" dataplanes "ID" "NAME"

run_test_json_array \
    "get dataplanes --output json returns JSON array" \
    wbctl get dataplanes --output json

# Describe the first available dataplane — ID sourced from captured get output.
DP_ID="$_LAST_TABLE_ID"
if [ -n "$DP_ID" ]; then
    run_test_output_contains \
        "describe dataplane contains $DP_ID" \
        "$DP_ID" \
        wbctl describe dataplane "$DP_ID"
else
    printf '[%s] SKIP: no dataplanes found; skipping describe test.\n' "$LABEL"
fi

# -- summary ----------------------------------------------------------------
print_summary
