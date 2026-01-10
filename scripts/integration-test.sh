#!/usr/bin/env bash
# Copyright © 2026, SAS Institute Inc., Cary, NC, USA.  All Rights Reserved.
# SPDX-License-Identifier: Apache-2.0
#
# integration-test.sh — driver script for the wbctl integration test suite.
#
# Usage:
#   WBCONFIG=/path/to/config.json ./scripts/integration-test.sh
#   ./scripts/integration-test.sh          # prompts for WBCONFIG if unset
#
# Optional:
#   WBUSER=<username> WBPASS=<password>    # enable the login test
#
# The driver:
#   1. Builds the wbctl binary at the repo root.
#   2. Validates (or prompts for) WBCONFIG.
#   3. Runs every test-*.sh in scripts/integration/, capturing pass/fail.
#   4. Prints a combined summary table.

set -euo pipefail
export WBCTL_IT_FORCE_COLOR=1
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
INT_DIR="$SCRIPT_DIR/integration"
LABEL="integration-test"
source "$INT_DIR/common.sh"

# ---------------------------------------------------------------------------
# Binary readiness step (auto-rebuild if missing, stale, or wrong format)
# ---------------------------------------------------------------------------
if ensure_wbctl_binary; then
    printf '[integration-test] %-53s %s\n' "wbctl binary ready" "${_GRN}PASS${_RST}"
else
    printf '[integration-test] %-53s %s\n' "wbctl binary ready" "${_RED}FAIL${_RST}"
    printf '%s\n' "$_SEP"
    printf 'Binary setup failed; aborting.\n'
    exit 1
fi

resolve_wbconfig

# ---------------------------------------------------------------------------
# Run each test script
# ---------------------------------------------------------------------------
SCRIPTS_PASS=0
SCRIPTS_FAIL=0

# Collect test scripts in a stable order.
# shellcheck disable=SC2012
TEST_SCRIPTS=$(ls "$INT_DIR"/test-*.sh 2>/dev/null | sort)

if [ -z "$TEST_SCRIPTS" ]; then
    printf 'Error: no test-*.sh scripts found in %s\n' "$INT_DIR" >&2
    exit 1
fi

for script in $TEST_SCRIPTS; do
    name="$(basename "$script" .sh)"
    printf '[%s] starting\n' "$name"
    if "$script"; then
        rc=0
    else
        rc=$?
    fi

    if [ "$rc" -eq 0 ]; then
        SCRIPTS_PASS=$((SCRIPTS_PASS + 1))
    else
        SCRIPTS_FAIL=$((SCRIPTS_FAIL + 1))
    fi
done

# ---------------------------------------------------------------------------
# Combined summary
# ---------------------------------------------------------------------------
printf '%s\n' "$_SEP"
printf 'Total scripts: %d  Passed: %d  Failed: %d\n' \
    $((SCRIPTS_PASS + SCRIPTS_FAIL)) "$SCRIPTS_PASS" "$SCRIPTS_FAIL"

if [ "$SCRIPTS_FAIL" -gt 0 ]; then
    exit 1
fi
