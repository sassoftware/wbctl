#!/usr/bin/env bash
# Copyright © 2026, SAS Institute Inc., Cary, NC, USA.  All Rights Reserved.
# SPDX-License-Identifier: Apache-2.0
#
# test-snapshots.sh — full snapshot lifecycle test.
#
# A snapshot can only be created from a volume that is *bound*, meaning a
# running workbench is currently using it. This script therefore:
#   1. Creates a volume (prereq).
#   2. Creates a workbench bound to that volume (prereq).
#   3. Creates a snapshot of the bound volume.
#   4. Restores a new volume from the snapshot.
#   5. Cleans up in reverse order: workbench → snapshot → restored volume
#      → original volume.
#
# Can be run standalone or via integration-test.sh.
#
# Prerequisites resolved at runtime:
#   storageclass ID — first result from `get storageclasses`
#   nodepool ID     — first result from `get nodepools`
#   service/app ID  — first service, falling back to first application

set -euo pipefail
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
LABEL="test-snapshots"
source "$SCRIPT_DIR/common.sh"

require_wbconfig

# -- cleanup trap (registered before any resource is created) ---------------
WORKSPACE_ID=""
WORKBENCH_ID=""
SNAPSHOT_ID=""
RESTORE_ID=""
cleanup_snapshots() {
    cleanup_delete workbench WORKBENCH_ID
    cleanup_delete snapshot  SNAPSHOT_ID
    cleanup_delete volume RESTORE_ID
    cleanup_delete volume WORKSPACE_ID
    WORKBENCH_ID=""
    SNAPSHOT_ID=""
    RESTORE_ID=""
    WORKSPACE_ID=""
}
trap 'cleanup_snapshots' EXIT

# -- resolve prerequisites --------------------------------------------------
SC_ID=$(capture_id storageclasses)
if [ -z "$SC_ID" ]; then
    printf '[%s] FAIL: no storageclasses available.\n' "$LABEL"; exit 1
fi

NP_ID=$(capture_id nodepools)
if [ -z "$NP_ID" ]; then
    printf '[%s] FAIL: no nodepools available.\n' "$LABEL"; exit 1
fi

SVC_ID=$(capture_id services)
APP_ID=""
if [ -z "$SVC_ID" ]; then
    APP_ID=$(capture_id applications)
fi
if [ -z "$SVC_ID" ] && [ -z "$APP_ID" ]; then
    printf '[%s] FAIL: no services or applications available.\n' "$LABEL"; exit 1
fi

TS=$(date +%s)

if [ -n "$SVC_ID" ]; then
    SVC_FLAG=(--service "$SVC_ID")
else
    SVC_FLAG=(--application "$APP_ID")
fi

extract_uuid() {
    local out="$1"
    printf '%s\n' "$out" | grep -Eo '[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}' | tail -n1
}

# -- create prerequisite volume ---------------------------------------------
WORKSPACE_OUT=$(wbctl create volume \
    --name "it-${TS}-snap-ws" \
    --storageclass "$SC_ID" \
    --size 1Gi 2>&1) && WORKSPACE_RC=0 || WORKSPACE_RC=$?
WORKSPACE_ID=""
if [ "$WORKSPACE_RC" -eq 0 ]; then
    WORKSPACE_ID=$(extract_uuid "$WORKSPACE_OUT")
fi

if [ -z "$WORKSPACE_ID" ]; then
    printf '%s\n' "$WORKSPACE_OUT" >&2
    printf '[%s] FAIL: could not create prerequisite volume.\n' "$LABEL"; exit 1
fi

# -- tests ------------------------------------------------------------------

run_test_output_contains \
    "dry-run create snapshot contains name field" \
    '"name"' \
    wbctl create snapshot \
        --dry-run \
        --name "it-${TS}-snap-dry" \
        --volume "$WORKSPACE_ID"

# -- create prerequisite workbench (binds the volume) -----------------------
WORKBENCH_OUT=$(wbctl create workbench \
    --name "it-${TS}-snap-wb" \
    --nodepool "$NP_ID" \
    --volume "$WORKSPACE_ID" \
    --desired-state RUNNING \
    "${SVC_FLAG[@]}" 2>&1) && WORKBENCH_RC=0 || WORKBENCH_RC=$?
WORKBENCH_ID=""
if [ "$WORKBENCH_RC" -eq 0 ]; then
    WORKBENCH_ID=$(extract_uuid "$WORKBENCH_OUT")
fi

if [ -z "$WORKBENCH_ID" ]; then
    printf '%s\n' "$WORKBENCH_OUT" >&2
    printf '[%s] FAIL: could not create prerequisite workbench.\n' "$LABEL"; exit 1
fi

wait_for_volume_available() {
    local attempt max_attempts=24 sleep_seconds=5 out
    for attempt in $(seq 1 "$max_attempts"); do
        out=$(wbctl describe workbench "$WORKBENCH_ID" 2>&1) || out=""
        if printf '%s' "$out" | grep -Eq 'Current State:[[:space:]]+RUNNING'; then
            return 0
        fi
        wait_with_notice "$sleep_seconds" \
            "volume not ready yet (attempt ${attempt}/${max_attempts}); polling workbench state"
    done
    printf '%s\n' "$out" >&2
    return 1
}

create_snapshot_after_wait() {
    local attempt max_attempts=12 sleep_seconds=5 out rc
    SNAPSHOT_ID=""
    for attempt in $(seq 1 "$max_attempts"); do
        out=$(wbctl create snapshot \
            --name "it-${TS}-snap" \
            --volume "$WORKSPACE_ID" 2>&1) && rc=0 || rc=$?
        if [ "$rc" -eq 0 ]; then
            SNAPSHOT_ID=$(extract_uuid "$out")
        fi
        if [ -n "$SNAPSHOT_ID" ]; then
            return 0
        fi
        wait_with_notice "$sleep_seconds" \
            "snapshot create not ready yet (attempt ${attempt}/${max_attempts}); retrying"
    done
    printf '%s\n' "$out" >&2
    return 1
}

if wait_for_volume_available; then
    _PASS=$((_PASS + 1))
    printf '[%s] %-55s %s\n' "$LABEL" \
        "wait for volume availability (workbench RUNNING)" \
        "${_GRN}PASS${_RST}"
else
    _FAIL=$((_FAIL + 1))
    printf '[%s] %-55s %s\n' "$LABEL" \
        "wait for volume availability (workbench RUNNING)" \
        "${_RED}FAIL${_RST}"
fi

if create_snapshot_after_wait; then
    _PASS=$((_PASS + 1))
    printf '[%s] %-55s %s\n' "$LABEL" \
        "create snapshot after volume becomes available" \
        "${_GRN}PASS${_RST}"
else
    _FAIL=$((_FAIL + 1))
    printf '[%s] %-55s %s\n' "$LABEL" \
        "create snapshot after volume becomes available" \
        "${_RED}FAIL${_RST}"
fi

if [ -n "$SNAPSHOT_ID" ]; then
    run_test_output_contains \
        "get snapshots contains ID header" \
        "ID" \
        wbctl get snapshots

    run_test_json_array \
        "get snapshots --output json returns JSON array" \
        wbctl get snapshots --output json

    run_test_output_contains \
        "get snapshots contains $SNAPSHOT_ID" \
        "$SNAPSHOT_ID" \
        wbctl get snapshots

    SNAPSHOT_EDIT_NAME="it-${TS}-snap-edit"
    run_test_edit_name_round_trip \
        "edit snapshot $SNAPSHOT_ID" \
        snapshot \
        "$SNAPSHOT_ID" \
        "$SNAPSHOT_EDIT_NAME"

    run_test_output_contains \
        "describe snapshot contains $SNAPSHOT_ID" \
        "$SNAPSHOT_ID" \
        wbctl describe snapshot "$SNAPSHOT_ID"

    SNAPSHOT_UPDATE_NAME="it-${TS}-snap-upd"
    run_test_output_contains \
        "dry-run update snapshot $SNAPSHOT_ID contains updated name" \
        "$SNAPSHOT_UPDATE_NAME" \
        wbctl update snapshot \
            --dry-run \
            --name "$SNAPSHOT_UPDATE_NAME" \
            "$SNAPSHOT_ID"

    run_test_output_contains \
        "update snapshot $SNAPSHOT_ID returns id" \
        "$SNAPSHOT_ID" \
        wbctl update snapshot \
            --name "$SNAPSHOT_UPDATE_NAME" \
            "$SNAPSHOT_ID"

    run_test_output_contains \
        "get snapshot $SNAPSHOT_ID returns updated name" \
        "$SNAPSHOT_UPDATE_NAME" \
        wbctl get snapshot \
            "$SNAPSHOT_ID"

    # Immediate restore attempt should fail while snapshot volume status is not
    # yet ready.
    RESTORE_IMMEDIATE_OUT=$(wbctl create volume \
        --name "it-${TS}-snap-restore-immediate" \
        --snapshot "$SNAPSHOT_ID" 2>&1) && RESTORE_IMMEDIATE_RC=0 || RESTORE_IMMEDIATE_RC=$?
    if [ "$RESTORE_IMMEDIATE_RC" -ne 0 ] && \
        printf '%s' "$RESTORE_IMMEDIATE_OUT" | grep -qF "snapshot not ready: volume status missing or not ready"; then
        _PASS=$((_PASS + 1))
        printf '[%s] %-55s %s\n' "$LABEL" \
            "immediate restore emits snapshot-not-ready error" \
            "${_GRN}PASS${_RST}"
    else
        _FAIL=$((_FAIL + 1))
        printf '[%s] %-55s %s\n' "$LABEL" \
            "immediate restore emits snapshot-not-ready error" \
            "${_RED}FAIL${_RST}"
        printf '  expected restore failure containing snapshot-not-ready error\n' >&2
        printf '  output: %s\n' "$RESTORE_IMMEDIATE_OUT" >&2
    fi

    restore_volume_after_wait() {
        local attempt max_attempts=24 sleep_seconds=5 out rc
        RESTORE_ID=""
        for attempt in $(seq 1 "$max_attempts"); do
            out=$(wbctl create volume \
                --name "it-${TS}-snap-restore" \
                --snapshot "$SNAPSHOT_ID" 2>&1) && rc=0 || rc=$?
            if [ "$rc" -eq 0 ]; then
                RESTORE_ID=$(extract_uuid "$out")
                if [ -n "$RESTORE_ID" ]; then
                    return 0
                fi
            fi
            wait_with_notice "$sleep_seconds" \
                "restore volume from snapshot not ready yet (attempt ${attempt}/${max_attempts}); retrying"
        done
        printf '%s\n' "$out" >&2
        return 1
    }

    if restore_volume_after_wait; then
        _PASS=$((_PASS + 1))
        printf '[%s] %-55s %s\n' "$LABEL" \
            "restore volume succeeds after waiting for snapshot readiness" \
            "${_GRN}PASS${_RST}"
        run_test_output_contains \
            "describe restored volume contains $RESTORE_ID" \
            "$RESTORE_ID" \
            wbctl describe volume "$RESTORE_ID"
    else
        _FAIL=$((_FAIL + 1))
        printf '[%s] %-55s %s\n' "$LABEL" \
            "restore volume succeeds after waiting for snapshot readiness" \
            "${_RED}FAIL${_RST}"
    fi
else
    _FAIL=$((_FAIL + 1))
    printf '[%s] %-55s %s\n' "$LABEL" "create snapshot" "${_RED}FAIL${_RST}"
    printf '  create snapshot returned empty ID\n' >&2
fi

# cleanup output first, then footer
cleanup_snapshots

# -- summary ----------------------------------------------------------------
print_summary
