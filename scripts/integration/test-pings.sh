#!/usr/bin/env bash
# Copyright © 2026, SAS Institute Inc., Cary, NC, USA.  All Rights Reserved.
# SPDX-License-Identifier: Apache-2.0
#
# test-pings.sh — smoke tests for `wbctl get pings`.
# Can be run standalone or via integration-test.sh.

set -euo pipefail
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
LABEL="test-pings"
source "$SCRIPT_DIR/common.sh"

require_wbconfig

# -- tests ------------------------------------------------------------------

run_test_table_has_headers_and_optional_data "get pings" pings "DATAPLANE_ID" "STATE"

run_test_json_array \
    "get pings --output json returns JSON array" \
    wbctl get pings --output json

# -- summary ----------------------------------------------------------------
print_summary
