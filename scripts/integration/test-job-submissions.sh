#!/usr/bin/env bash
# Copyright © 2026, SAS Institute Inc., Cary, NC, USA.  All Rights Reserved.
# SPDX-License-Identifier: Apache-2.0
#
# test-job-submissions.sh — full job-submission lifecycle test.
#
# Steps: create volume + job (prereqs) → dry-run create job-submission →
#        create job-submission → get → describe → stream logs (5s) →
#        cancel → delete (via EXIT trap).
# Can be run standalone or via integration-test.sh.
#
# Prerequisites resolved at runtime:
#   storageclass ID — first result from `get storageclasses`
#   nodepool ID     — first result from `get nodepools`
#   service/app ID  — first service, falling back to first application

set -euo pipefail
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
LABEL="test-job-submissions"
source "$SCRIPT_DIR/common.sh"

require_wbconfig

# -- cleanup trap (registered before any resource is created) ---------------
WORKSPACE_ID=""
JOB_ID=""
JOB_SUB_ID=""
cleanup_job_submission() {
    cleanup_delete job-submission JOB_SUB_ID
    cleanup_delete job JOB_ID
    cleanup_delete volume WORKSPACE_ID
    JOB_SUB_ID=""
    JOB_ID=""
    WORKSPACE_ID=""
}
trap 'cleanup_job_submission' EXIT

# -- resolve prerequisites --------------------------------------------------
SC_ID=$(capture_id storageclasses)
if [ -z "$SC_ID" ]; then
    printf '[%s] %s: no storageclasses available.\n' "$LABEL" "${_RED}FAIL${_RST}"; exit 1
fi

NP_ID=$(capture_id nodepools)
if [ -z "$NP_ID" ]; then
    printf '[%s] %s: no nodepools available.\n' "$LABEL" "${_RED}FAIL${_RST}"; exit 1
fi

SVC_ID=$(capture_id services)
APP_ID=""
if [ -z "$SVC_ID" ]; then
    APP_ID=$(capture_id applications)
fi
if [ -z "$SVC_ID" ] && [ -z "$APP_ID" ]; then
    printf '[%s] %s: no services or applications available.\n' "$LABEL" "${_RED}FAIL${_RST}"; exit 1
fi

TS=$(date +%s)

# -- create prerequisite volume ---------------------------------------------
WORKSPACE_ID=$(wbctl create volume \
    --name "it-${TS}-jsub-ws" \
    --storageclass "$SC_ID" \
    --size 1Gi) || WORKSPACE_ID=""

if [ -z "$WORKSPACE_ID" ]; then
    printf '[%s] %s: could not create prerequisite volume.\n' "$LABEL" "${_RED}FAIL${_RST}"; exit 1
fi

if [ -n "$SVC_ID" ]; then
    SVC_FLAG=(--service "$SVC_ID")
else
    SVC_FLAG=(--application "$APP_ID")
fi

# -- create prerequisite job ------------------------------------------------
JOB_ID=$(wbctl create job \
    --name "it-${TS}-jsub-job" \
    --nodepool "$NP_ID" \
    --command "echo hello" \
    --volume "$WORKSPACE_ID" \
    "${SVC_FLAG[@]}") || JOB_ID=""

if [ -z "$JOB_ID" ]; then
    printf '[%s] %s: could not create prerequisite job.\n' "$LABEL" "${_RED}FAIL${_RST}"; exit 1
fi

# -- tests ------------------------------------------------------------------

run_test_output_contains \
    "dry-run create job-submission contains job_id field" \
    '"job_id"' \
    wbctl create job-submission \
        --dry-run \
        --job-id "$JOB_ID"

JOB_SUB_ID=$(wbctl create job-submission \
    --job-id "$JOB_ID") || JOB_SUB_ID=""

if [ -n "$JOB_SUB_ID" ]; then
    run_test \
        "create job-submission exit 0" \
        true

    run_test_output_contains \
        "get job-submissions contains ID header" \
        "ID" \
        wbctl get job-submissions

    run_test_json_array \
        "get job-submissions --output json returns JSON array" \
        wbctl get job-submissions --output json

    run_test_output_contains \
        "get job-submissions contains $JOB_SUB_ID" \
        "$JOB_SUB_ID" \
        wbctl get job-submissions

    run_test_output_contains \
        "describe job-submission contains $JOB_SUB_ID" \
        "$JOB_SUB_ID" \
        wbctl describe job-submission "$JOB_SUB_ID"

    wait_for_succeeded() {
        local attempt max_attempts=24 sleep_seconds=5 out
        for attempt in $(seq 1 "$max_attempts"); do
            out=$(wbctl describe job-submission "$JOB_SUB_ID" 2>&1) || out=""
            if grep -Eq '^State:[[:space:]]+SUCCEEDED' <<<"$out"; then
                return 0
            fi
            wait_with_notice "$sleep_seconds" \
                "job-submission not succeeded yet (attempt ${attempt}/${max_attempts}); polling state"
        done
        printf '%s\n' "$out" >&2
        return 1
    }

    if wait_for_succeeded; then
        _PASS=$((_PASS + 1))
        printf '[%s] %-55s %s\n' "$LABEL" "wait for job-submission success" "${_GRN}PASS${_RST}"
    else
        _FAIL=$((_FAIL + 1))
        printf '[%s] %-55s %s\n' "$LABEL" "wait for job-submission success" "${_RED}FAIL${_RST}"
    fi

    JOB_LOGS=$(wbctl logs job-submissions "$JOB_SUB_ID" 2>&1) || JOB_LOGS=""
    if grep -qF "hello" <<<"$JOB_LOGS"; then
        _PASS=$((_PASS + 1))
        printf '[%s] %-55s %s\n' "$LABEL" "job-submission logs contain expected output" "${_GRN}PASS${_RST}"
    else
        _FAIL=$((_FAIL + 1))
        printf '[%s] %-55s %s\n' "$LABEL" "job-submission logs contain expected output" "${_RED}FAIL${_RST}"
        printf '  expected log output not found\n' >&2
        printf '  output: %s\n' "$JOB_LOGS" >&2
    fi

    run_test \
        "cancel job-submission exit 0" \
        wbctl cancel job-submission "$JOB_SUB_ID"
else
    _FAIL=$((_FAIL + 1))
    printf '[%s] %-55s %s\n' "$LABEL" "create job-submission" "${_RED}FAIL${_RST}"
    printf '  create job-submission returned empty ID\n' >&2
fi

# cleanup output first, then footer
cleanup_job_submission

# -- summary ----------------------------------------------------------------
print_summary
