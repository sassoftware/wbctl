#!/usr/bin/env bash
# Copyright © 2026, SAS Institute Inc., Cary, NC, USA.  All Rights Reserved.
# SPDX-License-Identifier: Apache-2.0
#
# test-applications.sh — smoke tests for `wbctl get applications`.
# Can be run standalone or via integration-test.sh.
#
# Note: `describe` does not support the `applications` resource type; only
# `get` is exercised here.

set -euo pipefail
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
LABEL="test-applications"
source "$SCRIPT_DIR/common.sh"

require_wbconfig

# -- tests ------------------------------------------------------------------

run_test_table_has_headers_and_data "get applications" applications "ID" "NAME"

run_test_json_array \
    "get applications --output json returns JSON array" \
    wbctl get applications --output json

# -- summary ----------------------------------------------------------------
print_summary
