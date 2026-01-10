#!/usr/bin/env bash
# Copyright © 2026, SAS Institute Inc., Cary, NC, USA.  All Rights Reserved.
# SPDX-License-Identifier: Apache-2.0
#
# test-api.sh — smoke tests for the `wbctl api` command.
# Can be run standalone or via integration-test.sh.

set -euo pipefail
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
LABEL="test-api"
source "$SCRIPT_DIR/common.sh"

require_wbconfig

# -- tests ------------------------------------------------------------------

# No path arg: prints the OpenAPI route table (PATH / VERB / DESCRIPTION cols).
run_test_output_contains \
    "api (no path) prints route table with PATH header" \
    "PATH" \
    wbctl api

run_test_json_object \
    "api --output json returns JSON object" \
    wbctl api --output json

# GET /dataplanes via raw API: returns a JSON array.
run_test_json_array \
    "api /dataplanes returns JSON array" \
    wbctl api /dataplanes

# -- summary ----------------------------------------------------------------
print_summary
