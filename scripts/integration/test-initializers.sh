#!/usr/bin/env bash
# Copyright © 2026, SAS Institute Inc., Cary, NC, USA.  All Rights Reserved.
# SPDX-License-Identifier: Apache-2.0
#
# test-initializers.sh — smoke tests for `wbctl get initializers`.
# Can be run standalone or via integration-test.sh.
#
# Note: `describe` does not support the `initializers` resource type.

set -euo pipefail
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
LABEL="test-initializers"
source "$SCRIPT_DIR/common.sh"

require_wbconfig

# -- tests ------------------------------------------------------------------

run_test_table_has_headers_and_data "get initializers" initializers "ID" "NAME"

run_test_json_array \
    "get initializers --output json returns JSON array" \
    wbctl get initializers --output json

# -- summary ----------------------------------------------------------------
print_summary
