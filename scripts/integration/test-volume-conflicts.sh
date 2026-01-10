#!/usr/bin/env bash
# Copyright © 2026, SAS Institute Inc., Cary, NC, USA.  All Rights Reserved.
# SPDX-License-Identifier: Apache-2.0
#
# test-volume-conflicts.sh — smoke tests for `wbctl get/describe
# volume-conflicts`.
# Can be run standalone or via integration-test.sh.
#
# volume conflicts only exist when two volumes are simultaneously bound
# to the same backing volume. The test always verifies the `get` command
# succeeds; `describe` is only exercised when an ID is actually returned.

set -euo pipefail
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
LABEL="test-volume-conflicts"
source "$SCRIPT_DIR/common.sh"

require_wbconfig

# -- tests ------------------------------------------------------------------

run_test_table_has_headers_and_optional_data "get volume-conflicts" volume-conflicts "ID"

run_test_json_array \
    "get volume-conflicts --output json returns JSON array" \
    wbctl get volume-conflicts --output json

# Only attempt describe when there is at least one conflict to describe.
# CONFLICT_ID is sourced from captured get output; no second get needed.
CONFLICT_ID="$_LAST_TABLE_ID"
if [ -n "$CONFLICT_ID" ]; then
    run_test_output_contains \
        "describe volume-conflicts contains $CONFLICT_ID" \
        "$CONFLICT_ID" \
        wbctl describe volume-conflicts "$CONFLICT_ID"
else
    printf '[%s] INFO: no volume-conflicts found; describe test skipped.\n' "$LABEL"
fi

# -- summary ----------------------------------------------------------------
print_summary
