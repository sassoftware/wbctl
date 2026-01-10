#!/usr/bin/env bash
# Copyright © 2026, SAS Institute Inc., Cary, NC, USA.  All Rights Reserved.
# SPDX-License-Identifier: Apache-2.0
#
# test-workbenches.sh — full workbench lifecycle test.
#
# Steps: create volume (prereq) → dry-run create workbench per app →
#        create workbenches → start in batch → wait for running → stop/start →
#        delete (via EXIT trap).
# Can be run standalone or via integration-test.sh.
#
# Prerequisites resolved at runtime:
#   storageclass ID — first result from `get storageclasses`
#   nodepool ID     — first result from `get nodepools`
#   applications    — all rows from `get applications --output json`

set -euo pipefail
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
LABEL="test-workbenches"
source "$SCRIPT_DIR/common.sh"

require_wbconfig

# -- cleanup trap (registered before any resource is created) ---------------
WORKSPACE_ID=""
ORIGINAL_MAX_WORKBENCHES=""
WORKBENCH_IDS=()

cleanup_workbenches() {
    local workbench_id
    local workbench_id_ref
    if [ "${#WORKBENCH_IDS[@]}" -gt 0 ]; then
        for workbench_id in "${WORKBENCH_IDS[@]}"; do
            if [ -n "$workbench_id" ]; then
                workbench_id_ref="$workbench_id"
                cleanup_delete workbench workbench_id_ref
            fi
        done
    fi
    WORKBENCH_IDS=()
    cleanup_delete volume WORKSPACE_ID
    WORKSPACE_ID=""
    if [ -n "$ORIGINAL_MAX_WORKBENCHES" ]; then
        cleanup_cmd wbctl settings --set max_workbenches_per_user --value "$ORIGINAL_MAX_WORKBENCHES"
        ORIGINAL_MAX_WORKBENCHES=""
    fi
}

trap '
    cleanup_workbenches
' EXIT

# -- resolve prerequisites --------------------------------------------------
SC_ID=$(capture_id storageclasses)
if [ -z "$SC_ID" ]; then
    printf '[%s] FAIL: no storageclasses available.\n' "$LABEL"
    exit 1
fi

NP_ID=$(capture_id nodepools)
if [ -z "$NP_ID" ]; then
    printf '[%s] FAIL: no nodepools available.\n' "$LABEL"
    exit 1
fi

TS=$(date +%s)

# -- create prerequisite volume ---------------------------------------------
WORKSPACE_ID=$(wbctl create volume \
    --name "it-${TS}-wb-ws" \
    --storageclass "$SC_ID" \
    --size 1Gi) || WORKSPACE_ID=""

if [ -z "$WORKSPACE_ID" ]; then
    printf '[%s] FAIL: could not create prerequisite volume.\n' "$LABEL"
    exit 1
fi

app_records=$(
    wbctl get applications --output json | jq -r '.[] | .id'
) || app_records=""

run_test_json_array \
    "get applications --output json returns JSON array" \
    wbctl get applications --output json

if [ -z "$app_records" ]; then
    printf '[%s] FAIL: no applications available.\n' "$LABEL"
    exit 1
fi

APP_RECORDS=()
while IFS= read -r app_record; do
    [ -n "$app_record" ] || continue
    APP_RECORDS+=("$app_record")
done <<EOF
$app_records
EOF

ORIGINAL_MAX_WORKBENCHES=$(wbctl settings --output json | jq -r '.max_workbenches_per_user')
TARGET_MAX_WORKBENCHES=$((${ORIGINAL_MAX_WORKBENCHES//\"/} + ${#APP_RECORDS[@]}))

run_test_output_contains \
    "set max_workbenches_per_user to $TARGET_MAX_WORKBENCHES returns JSON boolean" \
    max_workbenches_per_user \
    wbctl settings --set max_workbenches_per_user --value "$TARGET_MAX_WORKBENCHES"

wait_for_running() {
    local workbench_id="$1"
    local workbench_name="$2"
    local max_attempts="${3:-24}"
    local attempt=1
    local describe_out describe_rc

    while [ "$attempt" -le "$max_attempts" ]; do
        describe_out=$(wbctl describe workbench "$workbench_id" 2>&1) && describe_rc=0 || describe_rc=$?
        if [ "$describe_rc" -eq 0 ] && printf '%s\n' "$describe_out" | grep -qiE '^Current State:[[:space:]]*Running'; then
            _PASS=$((_PASS + 1))
            printf '[%s] %-55s %s\n' "$LABEL" "describe workbench $workbench_name running" "${_GRN}PASS${_RST}"
            return 0
        fi

        if [ "$attempt" -eq "$max_attempts" ]; then
            _FAIL=$((_FAIL + 1))
            printf '[%s] %-55s %s\n' "$LABEL" "describe workbench $workbench_name running" "${_RED}FAIL${_RST}"
            printf '  workbench %s did not reach RUNNING state\n' "$workbench_id" >&2
            printf '  output: %s\n' "$describe_out" >&2
            return 1
        fi

        wait_with_notice 5 "workbench $workbench_name not running yet (attempt $attempt/$max_attempts)"
        attempt=$((attempt + 1))
    done
}

BATCH_SIZE="${#APP_RECORDS[@]}"

for ((batch_start=0; batch_start<${#APP_RECORDS[@]}; batch_start+=BATCH_SIZE)); do
    WORKBENCH_IDS=()
    batch_end=$((batch_start + BATCH_SIZE))
    if [ "$batch_end" -gt "${#APP_RECORDS[@]}" ]; then
        batch_end="${#APP_RECORDS[@]}"
    fi

    for ((i=batch_start; i<batch_end; i++)); do
        APP_ID="${APP_RECORDS[$i]}"
        [ -n "$APP_ID" ] || continue

        WORKBENCH_NAME="it-${TS}-wb-${APP_ID}"
        WORKBENCH_DRY_NAME="${WORKBENCH_NAME}-dry"

        run_test_output_contains \
            "dry-run create workbench $APP_ID contains name field" \
            '"name"' \
            wbctl create workbench \
                --dry-run \
                --name "$WORKBENCH_DRY_NAME" \
                --nodepool "$NP_ID" \
                --volume "$WORKSPACE_ID" \
                --application "$APP_ID"

        if WORKBENCH_OUT=$(wbctl create workbench \
            --name "$WORKBENCH_NAME" \
            --nodepool "$NP_ID" \
            --volume "$WORKSPACE_ID" \
            --desired-state STOPPED \
            --application "$APP_ID" 2>&1); then
            WORKBENCH_RC=0
        else
            WORKBENCH_RC=$?
        fi

        if [ "$WORKBENCH_RC" -eq 0 ]; then
            WORKBENCH_ID=$(printf '%s\n' "$WORKBENCH_OUT" | awk 'NF {id=$1} END {print id}')
        else
            WORKBENCH_ID=""
        fi

        if [ -z "$WORKBENCH_ID" ]; then
            _FAIL=$((_FAIL + 1))
            printf '[%s] %-55s %s\n' "$LABEL" "create workbench $APP_ID" "${_RED}FAIL${_RST}"
            printf '  create workbench returned empty ID (exit %d)\n' "$WORKBENCH_RC" >&2
            printf '  output: %s\n' "$WORKBENCH_OUT" >&2
            continue
        fi

        WORKBENCH_IDS+=("$WORKBENCH_ID")
        run_test \
            "create workbench $APP_ID exit 0" \
            true
        run_test_output_contains \
            "create workbench $APP_ID returned id" \
            "$WORKBENCH_ID" \
            printf '%s\n' "$WORKBENCH_OUT"

        WORKBENCH_UPDATE_NAME="${WORKBENCH_NAME}-upd"
        run_test_output_contains \
            "dry-run update workbench $WORKBENCH_ID contains updated name" \
            "$WORKBENCH_UPDATE_NAME" \
            wbctl update workbench \
                --dry-run \
                --name "$WORKBENCH_UPDATE_NAME" \
                "$WORKBENCH_ID"

        run_test_output_contains \
            "update workbench $WORKBENCH_ID returns id" \
            "$WORKBENCH_ID" \
            wbctl update workbench \
                --name "$WORKBENCH_UPDATE_NAME" \
                "$WORKBENCH_ID"

    run_test_output_contains \
        "get workbench $WORKBENCH_ID returns updated name" \
        "$WORKBENCH_UPDATE_NAME" \
        wbctl get workbench \
            "$WORKBENCH_ID"

    done

    if [ "${#WORKBENCH_IDS[@]}" -eq 0 ]; then
        continue
    fi

    run_test_output_contains \
        "get workbenches contains ID header" \
        "ID" \
        wbctl get workbenches

    WORKBENCH_EDIT_NAME="it-${TS}-wb-edit"
    run_test_edit_name_round_trip \
        "edit workbench $WORKBENCH_ID" \
        workbench \
        "$WORKBENCH_ID" \
        "$WORKBENCH_EDIT_NAME"

    for WORKBENCH_ID in "${WORKBENCH_IDS[@]}"; do
        run_test \
            "start workbench $WORKBENCH_ID exit 0" \
            wbctl start workbench "$WORKBENCH_ID"
    done

    for WORKBENCH_ID in "${WORKBENCH_IDS[@]}"; do
        wait_for_running "$WORKBENCH_ID" "workbench $WORKBENCH_ID"
    done

    for WORKBENCH_ID in "${WORKBENCH_IDS[@]}"; do
        run_test \
            "stop workbench $WORKBENCH_ID exit 0" \
            wbctl stop workbench "$WORKBENCH_ID"
        run_test \
            "start workbench $WORKBENCH_ID exit 0" \
            wbctl start workbench "$WORKBENCH_ID"
    done

done

# cleanup output first, then footer
cleanup_workbenches

# -- summary ----------------------------------------------------------------
print_summary
