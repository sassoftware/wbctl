#!/usr/bin/env bash
# Copyright © 2026, SAS Institute Inc., Cary, NC, USA.  All Rights Reserved.
# SPDX-License-Identifier: Apache-2.0
#
# test-login.sh — smoke test for `wbctl login` using password-auth.
#
# Requires environment variables:
#   WBCONFIG  — path to wbctl config file (endpoint must be configured).
#   WBUSER    — username for password-auth flow.
#   WBPASS    — password for password-auth flow.
#
# Can be run standalone or via integration-test.sh.

set -euo pipefail
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
LABEL="test-login"
source "$SCRIPT_DIR/common.sh"

require_wbconfig

# Require credentials for non-interactive password auth.
if [ -z "${WBUSER:-}" ] || [ -z "${WBPASS:-}" ]; then
    printf '[%s] SKIP: WBUSER or WBPASS is not set; password-auth login test skipped.\n' "$LABEL"
    printf '[%s] Set both env vars to enable this test.\n' "$LABEL"
    print_summary
    exit 0
fi

# Export so the global --username / --password flag env lookups pick them up.
export WBUSER WBPASS

# -- tests ------------------------------------------------------------------

# Login via password-auth (non-interactive): --auth-mode password forces
# username+password path; $WBUSER / $WBPASS satisfy the flag env lookups.
run_test \
    "login with password-auth succeeds" \
    wbctl --auth-mode password login

# -- summary ----------------------------------------------------------------
print_summary
