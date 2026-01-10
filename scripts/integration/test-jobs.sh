#!/usr/bin/env bash
# Copyright © 2026, SAS Institute Inc., Cary, NC, USA.  All Rights Reserved.
# SPDX-License-Identifier: Apache-2.0
#
# test-jobs.sh — full job lifecycle test.
#
# Steps: create volume (prereq) → dry-run create job → create job →
#        get → describe → delete (via EXIT trap).
# Can be run standalone or via integration-test.sh.
#
# Prerequisites resolved at runtime:
#   storageclass ID — first result from `get storageclasses`
#   nodepool ID     — first result from `get nodepools`
#   service/app ID  — first service, falling back to first application

set -euo pipefail
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
LABEL="test-jobs"
source "$SCRIPT_DIR/common.sh"

require_wbconfig

# -- cleanup trap (registered before any resource is created) ---------------
# The trap body references variables by name; they expand at execution time,
# so it is safe to register the trap before the IDs are populated.
WORKSPACE_ID=""
JOB_ID=""
cleanup_jobs() {
    cleanup_delete job JOB_ID
    cleanup_delete volume WORKSPACE_ID
    JOB_ID=""
    WORKSPACE_ID=""
}
trap 'cleanup_jobs' EXIT

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

# -- create prerequisite volume ---------------------------------------------
WORKSPACE_ID=$(wbctl create volume \
    --name "it-${TS}-job-ws" \
    --storageclass "$SC_ID" \
    --size 1Gi) || WORKSPACE_ID=""

if [ -z "$WORKSPACE_ID" ]; then
    printf '[%s] FAIL: could not create prerequisite volume.\n' "$LABEL"; exit 1
fi

if [ -n "$SVC_ID" ]; then
    SVC_FLAG=(--service "$SVC_ID")
else
    SVC_FLAG=(--application "$APP_ID")
fi

# -- tests ------------------------------------------------------------------

run_test_output_contains \
    "dry-run create job contains name field" \
    '"name"' \
    wbctl create job \
        --dry-run \
        --name "it-${TS}-job-dry" \
        --nodepool "$NP_ID" \
        --command "echo hello" \
        --volume "$WORKSPACE_ID" \
        "${SVC_FLAG[@]}"

JOB_ID=$(wbctl create job \
    --name "it-${TS}-job" \
    --nodepool "$NP_ID" \
    --command "echo hello" \
    --volume "$WORKSPACE_ID" \
    "${SVC_FLAG[@]}") || JOB_ID=""

if [ -n "$JOB_ID" ]; then
    run_test \
        "create job exit 0" \
        true

    run_test_output_contains \
        "get jobs contains ID header" \
        "ID" \
        wbctl get jobs

    run_test_json_array \
        "get jobs --output json returns JSON array" \
        wbctl get jobs --output json

    run_test_output_contains \
        "get jobs contains $JOB_ID" \
        "$JOB_ID" \
        wbctl get jobs

    JOB_EDIT_NAME="it-${TS}-job-edit"
    run_test_edit_name_round_trip \
        "edit job $JOB_ID" \
        job \
        "$JOB_ID" \
        "$JOB_EDIT_NAME"

    run_test_output_contains \
        "describe job contains $JOB_ID" \
        "$JOB_ID" \
        wbctl describe job "$JOB_ID"

    JOB_UPDATE_NAME="it-${TS}-job-upd"
    run_test_output_contains \
        "dry-run update job $JOB_ID contains updated name" \
        "$JOB_UPDATE_NAME" \
        wbctl update job \
            --dry-run \
            --name "$JOB_UPDATE_NAME" \
            "$JOB_ID"

    run_test_output_contains \
        "update job $JOB_ID returns id" \
        "$JOB_ID" \
        wbctl update job \
            --name "$JOB_UPDATE_NAME" \
            "$JOB_ID"

    run_test_output_contains \
        "get job $JOB_ID returns updated name" \
        "$JOB_UPDATE_NAME" \
        wbctl get job \
            "$JOB_ID"

else
    _FAIL=$((_FAIL + 1))
    printf '[%s] %-55s %s\n' "$LABEL" "create job" "${_RED}FAIL${_RST}"
    printf '  create job returned empty ID\n' >&2
fi

# cleanup output first, then footer
cleanup_jobs

# -- summary ----------------------------------------------------------------
print_summary
