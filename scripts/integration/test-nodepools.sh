#!/usr/bin/env bash
# Copyright © 2026, SAS Institute Inc., Cary, NC, USA.  All Rights Reserved.
# SPDX-License-Identifier: Apache-2.0
#
# test-nodepools.sh — smoke tests for `wbctl get nodepools`.
# Can be run standalone or via integration-test.sh.
#
# Note: `describe` does not support the `nodepools` resource type.

set -euo pipefail
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
LABEL="test-nodepools"
source "$SCRIPT_DIR/common.sh"

require_wbconfig

# -- tests ------------------------------------------------------------------

run_test_table_has_headers_and_data "get nodepools" nodepools "ID" "NAME"

run_test_json_array \
    "get nodepools --output json returns JSON array" \
    wbctl get nodepools --output json

# -- summary ----------------------------------------------------------------
print_summary
