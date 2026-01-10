#!/usr/bin/env bash
# Copyright © 2026, SAS Institute Inc., Cary, NC, USA.  All Rights Reserved.
# SPDX-License-Identifier: Apache-2.0
#
# test-volumes.sh — full volume lifecycle test.
#
# Steps: dry-run create → create → get → describe → delete (via EXIT trap).
# Can be run standalone or via integration-test.sh.
#
# Prerequisites resolved at runtime:
#   storageclass ID — first result from `get storageclasses`

set -euo pipefail
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
LABEL="test-volumes"
source "$SCRIPT_DIR/common.sh"

require_wbconfig

# -- cleanup trap (registered before any resource is created) ---------------
WORKSPACE_ID=""
cleanup_volume() {
    cleanup_delete volume WORKSPACE_ID
    WORKSPACE_ID=""
}
trap 'cleanup_volume' EXIT

# -- resolve prerequisites --------------------------------------------------
SC_ID=$(capture_id storageclasses)
if [ -z "$SC_ID" ]; then
    printf '[%s] FAIL: no storageclasses available; cannot run volume tests.\n' "$LABEL"
    exit 1
fi

TS=$(date +%s)

# -- tests ------------------------------------------------------------------

run_test_output_contains \
    "dry-run create volume contains name field" \
    '"name"' \
    wbctl create volume \
        --dry-run \
        --name "it-${TS}-ws-dry" \
        --storageclass "$SC_ID" \
        --size 1Gi

WORKSPACE_ID=$(wbctl create volume \
    --name "it-${TS}-ws" \
    --storageclass "$SC_ID" \
    --size 1Gi) || WORKSPACE_ID=""

if [ -n "$WORKSPACE_ID" ]; then
    run_test \
        "create volume exit 0" \
        true   # command already succeeded above; record the pass

    run_test_output_contains \
        "get volumes contains ID header" \
        "ID" \
        wbctl get volumes

    run_test_json_array \
        "get volumes --output json returns JSON array" \
        wbctl get volumes --output json

    run_test_output_contains \
        "get volumes contains $WORKSPACE_ID" \
        "$WORKSPACE_ID" \
        wbctl get volumes

    WORKSPACE_EDIT_NAME="it-${TS}-ws-edit"
    run_test_edit_name_round_trip \
        "edit volume $WORKSPACE_ID" \
        volume \
        "$WORKSPACE_ID" \
        "$WORKSPACE_EDIT_NAME"

    run_test_output_contains \
        "describe volume contains $WORKSPACE_ID" \
        "$WORKSPACE_ID" \
        wbctl describe volume "$WORKSPACE_ID"

    WORKSPACE_UPDATE_NAME="it-${TS}-ws-upd"
    run_test_output_contains \
        "dry-run update volume $WORKSPACE_ID contains updated name" \
        "$WORKSPACE_UPDATE_NAME" \
        wbctl update volume \
            --dry-run \
            --name "$WORKSPACE_UPDATE_NAME" \
            "$WORKSPACE_ID"

    run_test_output_contains \
        "update volume $WORKSPACE_ID returns id" \
        "$WORKSPACE_ID" \
        wbctl update volume \
            --name "$WORKSPACE_UPDATE_NAME" \
            "$WORKSPACE_ID"

    run_test_output_contains \
        "get volume $WORKSPACE_ID returns updated name" \
        "$WORKSPACE_UPDATE_NAME" \
        wbctl get volume \
            "$WORKSPACE_ID"

else
    _FAIL=$((_FAIL + 1))
    printf '[%s] %-55s %s\n' "$LABEL" "create volume" "${_RED}FAIL${_RST}"
    printf '  create volume returned empty ID\n' >&2
fi

# cleanup output first, then footer
cleanup_volume

# -- summary ----------------------------------------------------------------
print_summary
