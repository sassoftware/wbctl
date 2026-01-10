#!/usr/bin/env bash
# Copyright © 2026, SAS Institute Inc., Cary, NC, USA.  All Rights Reserved.
# SPDX-License-Identifier: Apache-2.0
#
# common.sh — shared helpers for wbctl integration tests.
#
# Source this file from each test script; do not execute it directly.
#
# Required: SCRIPT_DIR must be set by the sourcing script to the absolute path
# of the directory containing that test script. It is only used for relative
# test/script paths, not repo-root discovery.
#
# Required: LABEL must be set by the sourcing script to the short name used in
# output lines (e.g. "test-volumes").

set -euo pipefail

# ---------------------------------------------------------------------------
# wbctl binary location.
# ---------------------------------------------------------------------------
REPO_ROOT="$(git rev-parse --show-toplevel)"
WBCTL_BIN="${REPO_ROOT}/build/wbctl"

# Keep wbctl's token-cache temp files on the same filesystem as the cache
# itself. Root's default /tmp -> /root rename can cross devices here.
WBCTL_TMPDIR="${TMPDIR:-${HOME:-/tmp}/.wbctl/tmp}"
mkdir -p "$WBCTL_TMPDIR"
export TMPDIR="$WBCTL_TMPDIR"

# wbctl — thin wrapper that resolve $WBCTL_BIN
wbctl() {
    "$WBCTL_BIN" "$@"
}

# ---------------------------------------------------------------------------
# State tracking
# ---------------------------------------------------------------------------
_PASS=0
_FAIL=0

# ---------------------------------------------------------------------------
# Colour helpers (only when stdout is a terminal)
# ---------------------------------------------------------------------------
_RED=''
_GRN=''
_RST=''
if [ "${WBCTL_IT_FORCE_COLOR:-0}" = "1" ] || [ -t 1 ]; then
    _RED=$(printf '\033[0;31m')
    _GRN=$(printf '\033[0;32m')
    _RST=$(printf '\033[0m')
fi

_SEP='────────────────────────────────────────────────────────────────────'

wait_with_notice() {
    local seconds="$1"
    shift
    local message="${*:-sleeping ${seconds}s}"
    printf '[%s] INFO: waiting %ss: %s\n' "${LABEL}" "$seconds" "$message"
    sleep "$seconds"
}

# ---------------------------------------------------------------------------
# ensure_wbctl_binary
#   Ensures wbctl binary exists, is runnable on current host, and is not stale
#   versus cmd/internal Go sources. Rebuilds when needed.
# ---------------------------------------------------------------------------
ensure_wbctl_binary() {
    local reason='' tmp_bin=''
    if [ ! -x "$WBCTL_BIN" ]; then
        reason="missing or not executable"
    elif ! "$WBCTL_BIN" --help >/dev/null 2>&1; then
        reason="not runnable (e.g. exec format error)"
    elif find "$REPO_ROOT/cmd" "$REPO_ROOT/internal" -type f -name '*.go' -newer "$WBCTL_BIN" -print -quit | grep -q .; then
        reason="out of date"
    fi

    if [ -z "$reason" ]; then
        return 0
    fi

    printf '[%s] rebuilding wbctl binary (%s) ...\n' "${LABEL:-integration}" "$reason"
    if (cd "$REPO_ROOT" && make build); then
        return 0
    fi
    rm -f "$WBCTL_BIN"

    printf '[%s] FAIL: unable to build wbctl binary at %s\n' "${LABEL:-integration}" "$WBCTL_BIN" >&2
    return 1
}

resolve_wbconfig() {
    if [ -z "${WBCONFIG:-}" ]; then
        if [ -t 0 ]; then
            printf 'WBCONFIG is not set. Enter path to wbctl config file: '
            read -r WBCONFIG
            export WBCONFIG
        else
            printf 'Error: WBCONFIG is not set and stdin is not a terminal.\n' >&2
            exit 1
        fi
    fi
    if [ ! -f "$WBCONFIG" ]; then
        printf 'Error: WBCONFIG file not found: %s\n' "$WBCONFIG" >&2
        exit 1
    fi
    export WBCONFIG
}

# ---------------------------------------------------------------------------
# require_wbconfig
#   Validates $WBCONFIG: prompts interactively if unset, aborts if the file
#   does not exist.
# ---------------------------------------------------------------------------
require_wbconfig() {
    ensure_wbctl_binary
    resolve_wbconfig
}

# ---------------------------------------------------------------------------
# run_test <label> <cmd> [args...]
#   Runs the given command. Records PASS if exit code is 0, FAIL otherwise.
#   Stderr is always visible (separate from captured stdout).
#   On failure: prints the command and stdout to stderr.
# ---------------------------------------------------------------------------
run_test() {
    local label="$1"
    shift
    local _out _rc _tmpdir _stdout _stderr
    _tmpdir=$(mktemp -d)
    trap "rm -rf '$_tmpdir'" RETURN
    "$@" >"$_tmpdir/stdout" 2>"$_tmpdir/stderr" && _rc=0 || _rc=$?
    _out=$(cat "$_tmpdir/stdout")
    cat "$_tmpdir/stderr" >&2
    if [ "$_rc" -eq 0 ]; then
        _PASS=$((_PASS + 1))
        printf '[%s] %-55s %s\n' "${LABEL}" "$label" "${_GRN}PASS${_RST}"
    else
        _FAIL=$((_FAIL + 1))
        printf '[%s] %-55s %s\n' "${LABEL}" "$label" "${_RED}FAIL${_RST}"
        printf '  command: %s\n' "$*" >&2
        printf '  stdout:  %s\n' "$_out" >&2
    fi
}

# ---------------------------------------------------------------------------
# run_test_output_contains <label> <pattern> <cmd> [args...]
#   Like run_test but also greps stdout for <pattern>. Fails if exit code != 0
#   OR if the pattern is not found. Stderr is always visible.
# ---------------------------------------------------------------------------
run_test_output_contains() {
    local label="$1"
    local pattern="$2"
    shift 2
    local _out _rc _tmpdir
    _tmpdir=$(mktemp -d)
    trap "rm -rf '$_tmpdir'" RETURN
    "$@" >"$_tmpdir/stdout" 2>"$_tmpdir/stderr" && _rc=0 || _rc=$?
    _out=$(cat "$_tmpdir/stdout")
    cat "$_tmpdir/stderr" >&2
    if [ "$_rc" -ne 0 ]; then
        _FAIL=$((_FAIL + 1))
        printf '[%s] %-55s %s\n' "${LABEL}" "$label" "${_RED}FAIL${_RST}"
        printf '  (exit %d) command: %s\n' "$_rc" "$*" >&2
        printf '  stdout: %s\n' "$_out" >&2
        return
    fi
    if grep -qF -- "$pattern" <<<"$_out"; then
        _PASS=$((_PASS + 1))
        printf '[%s] %-55s %s\n' "${LABEL}" "$label" "${_GRN}PASS${_RST}"
    else
        _FAIL=$((_FAIL + 1))
        printf '[%s] %-55s %s\n' "${LABEL}" "$label" "${_RED}FAIL${_RST}"
        printf '  pattern not found: %s\n' "$pattern" >&2
        printf '  stdout: %s\n' "$_out" >&2
    fi
}

# ---------------------------------------------------------------------------
# run_test_json_type <label> <type> <cmd> [args...]
#   Like run_test, but also requires stdout to be valid JSON with the given
#   top-level type (e.g. array, object). Stderr is always visible.
# ---------------------------------------------------------------------------
run_test_json_type() {
    local label="$1"
    local expected_type="$2"
    shift 2
    local _out _rc _tmpdir
    _tmpdir=$(mktemp -d)
    trap "rm -rf '$_tmpdir'" RETURN
    "$@" >"$_tmpdir/stdout" 2>"$_tmpdir/stderr" && _rc=0 || _rc=$?
    _out=$(cat "$_tmpdir/stdout")
    cat "$_tmpdir/stderr" >&2
    if [ "$_rc" -ne 0 ]; then
        _FAIL=$((_FAIL + 1))
        printf '[%s] %-55s %s\n' "${LABEL}" "$label" "${_RED}FAIL${_RST}"
        printf '  (exit %d) command: %s\n' "$_rc" "$*" >&2
        printf '  stdout: %s\n' "$_out" >&2
        return
    fi
    if jq -e --arg type "$expected_type" 'type == $type' >/dev/null 2>&1 <<<"$_out"; then
        _PASS=$((_PASS + 1))
        printf '[%s] %-55s %s\n' "${LABEL}" "$label" "${_GRN}PASS${_RST}"
    else
        _FAIL=$((_FAIL + 1))
        printf '[%s] %-55s %s\n' "${LABEL}" "$label" "${_RED}FAIL${_RST}"
        printf '  output was not a JSON %s\n' "$expected_type" >&2
        printf '  stdout: %s\n' "$_out" >&2
    fi
}

# ---------------------------------------------------------------------------
# run_test_json_array <label> <cmd> [args...]
#   Like run_test, but also requires stdout to be valid JSON with array type.
#   Stderr is always visible.
# ---------------------------------------------------------------------------
run_test_json_array() {
    local label="$1"
    shift
    run_test_json_type "$label" "array" "$@"
}

# ---------------------------------------------------------------------------
# run_test_json_object <label> <cmd> [args...]
#   Like run_test, but also requires stdout to be valid JSON with object type.
#   Stderr is always visible.
# ---------------------------------------------------------------------------
run_test_json_object() {
    local label="$1"
    shift
    run_test_json_type "$label" "object" "$@"
}

# ---------------------------------------------------------------------------
# make_edit_name_editor
#   Creates a tiny editor script that rewrites `.name` in the edit temp file.
#   wbctl edit opens a temp JSON document, so a deterministic editor lets the
#   integration tests exercise the full edit flow without human interaction.
# ---------------------------------------------------------------------------
make_edit_name_editor() {
    local editor_script="$WBCTL_TMPDIR/wbctl-edit-name.sh"
    if [ ! -x "$editor_script" ]; then
        cat >"$editor_script" <<'EOF'
#!/usr/bin/env bash
set -euo pipefail

file="${1:?edit file path required}"
new_name="${WBCTL_EDIT_NAME:?WBCTL_EDIT_NAME required}"
tmp="${file}.tmp"

jq --arg name "$new_name" '.name = $name' "$file" >"$tmp"
mv "$tmp" "$file"
EOF
        chmod +x "$editor_script"
    fi
    printf '%s\n' "$editor_script"
}

# ---------------------------------------------------------------------------
# run_test_edit_name_round_trip <label> <resource> <id> <new_name>
#   Runs `wbctl edit <resource> <id>` with the deterministic editor above and
#   then verifies the resource reflects the new name when fetched again.
# ---------------------------------------------------------------------------
run_test_edit_name_round_trip() {
    local label="$1"
    local resource="$2"
    local resource_id="$3"
    local new_name="$4"

    edit_resource_with_name() {
        local editor
        editor=$(make_edit_name_editor)
        EDITOR="$editor" WBCTL_EDIT_NAME="$new_name" wbctl edit "$resource" "$resource_id"
    }

    run_test \
        "$label edit exit 0" \
        edit_resource_with_name

    run_test_output_contains \
        "$label get returns edited name" \
        "$new_name" \
        wbctl get "$resource" "$resource_id"
}

# ---------------------------------------------------------------------------
# _LAST_TABLE_ID — set by run_test_table_* to the first field of the first
# data row returned by the most recent table helper call, or empty string when
# no data rows were present.  Read this after calling either table helper to
# obtain an ID for follow-on describe checks without issuing a second get.
# ---------------------------------------------------------------------------
_LAST_TABLE_ID=''

# ---------------------------------------------------------------------------
# run_test_table_has_headers_and_data <base_label> <resource> <header> [...]
#   Runs `wbctl get <resource>` exactly once.  For each required header,
#   records an individual PASS/FAIL.  Then records a PASS/FAIL asserting that
#   at least one data row (non-empty first field on line 2+) is present.
#   Sets _LAST_TABLE_ID to the first field of the first data row (or '').
#   Stderr is always visible.
# ---------------------------------------------------------------------------
run_test_table_has_headers_and_data() {
    local base_label="$1"
    local resource="$2"
    shift 2
    local _out _rc _header _first_id _tmpdir
    _LAST_TABLE_ID=''
    _tmpdir=$(mktemp -d)
    trap "rm -rf '$_tmpdir'" RETURN
    wbctl get "$resource" >"$_tmpdir/stdout" 2>"$_tmpdir/stderr" && _rc=0 || _rc=$?
    _out=$(cat "$_tmpdir/stdout")
    cat "$_tmpdir/stderr" >&2
    if [ "$_rc" -ne 0 ]; then
        _FAIL=$((_FAIL + 1))
        printf '[%s] %-55s %s\n' "${LABEL}" "$base_label" "${_RED}FAIL${_RST}"
        printf '  (exit %d) wbctl get %s\n' "$_rc" "$resource" >&2
        printf '  stdout: %s\n' "$_out" >&2
        return
    fi
    for _header in "$@"; do
        if grep -qF -- "$_header" <<<"$_out"; then
            _PASS=$((_PASS + 1))
            printf '[%s] %-55s %s\n' "${LABEL}" "$base_label contains $_header header" "${_GRN}PASS${_RST}"
        else
            _FAIL=$((_FAIL + 1))
            printf '[%s] %-55s %s\n' "${LABEL}" "$base_label contains $_header header" "${_RED}FAIL${_RST}"
            printf '  header not found: %s\n' "$_header" >&2
            printf '  stdout: %s\n' "$_out" >&2
        fi
    done
    _first_id=$(awk 'NR==2 && $1!="" {print $1; exit}' <<<"$_out")
    if [ -n "$_first_id" ]; then
        _LAST_TABLE_ID="$_first_id"
        _PASS=$((_PASS + 1))
        printf '[%s] %-55s %s\n' "${LABEL}" "$base_label has data rows" "${_GRN}PASS${_RST}"
    else
        _FAIL=$((_FAIL + 1))
        printf '[%s] %-55s %s\n' "${LABEL}" "$base_label has data rows" "${_RED}FAIL${_RST}"
        printf '  no data rows returned by: wbctl get %s\n' "$resource" >&2
    fi
}

# ---------------------------------------------------------------------------
# run_test_table_has_headers_and_optional_data <base_label> <resource> <hdr> …
#   Like run_test_table_has_headers_and_data but does NOT fail when no data
#   rows are present (intended for volume-conflicts only).
#   Sets _LAST_TABLE_ID to the first field of the first data row (or '').
#   Stderr is always visible.
# ---------------------------------------------------------------------------
run_test_table_has_headers_and_optional_data() {
    local base_label="$1"
    local resource="$2"
    shift 2
    local _out _rc _header _first_id _tmpdir
    _LAST_TABLE_ID=''
    _tmpdir=$(mktemp -d)
    trap "rm -rf '$_tmpdir'" RETURN
    wbctl get "$resource" >"$_tmpdir/stdout" 2>"$_tmpdir/stderr" && _rc=0 || _rc=$?
    _out=$(cat "$_tmpdir/stdout")
    cat "$_tmpdir/stderr" >&2
    if [ "$_rc" -ne 0 ]; then
        _FAIL=$((_FAIL + 1))
        printf '[%s] %-55s %s\n' "${LABEL}" "$base_label" "${_RED}FAIL${_RST}"
        printf '  (exit %d) wbctl get %s\n' "$_rc" "$resource" >&2
        printf '  stdout: %s\n' "$_out" >&2
        return
    fi
    for _header in "$@"; do
        if grep -qF -- "$_header" <<<"$_out"; then
            _PASS=$((_PASS + 1))
            printf '[%s] %-55s %s\n' "${LABEL}" "$base_label contains $_header header" "${_GRN}PASS${_RST}"
        else
            _FAIL=$((_FAIL + 1))
            printf '[%s] %-55s %s\n' "${LABEL}" "$base_label contains $_header header" "${_RED}FAIL${_RST}"
            printf '  header not found: %s\n' "$_header" >&2
            printf '  stdout: %s\n' "$_out" >&2
        fi
    done
    _first_id=$(awk 'NR==2 && $1!="" {print $1; exit}' <<<"$_out")
    _LAST_TABLE_ID="$_first_id"
}

# ---------------------------------------------------------------------------
# capture_id <resource>
#   Runs `wbctl get <resource>` and returns the first column of the first data
#   row (row 2 after the header). Returns empty string when a successful table
#   has no data rows. Propagates `wbctl get` failures so callers see backend
#   errors instead of silent empties. Prefer reading _LAST_TABLE_ID after
#   run_test_table_* to avoid an extra get.
# ---------------------------------------------------------------------------
capture_id() {
    local resource="$1"
    local _out
    if ! _out=$(wbctl get "$resource" 2>&1); then
        printf '%s\n' "$_out" >&2
        return 1
    fi
    awk '
        /^ID[[:space:]]/ {in_table=1; next}
        in_table && NF > 0 {print $1; exit}
    ' <<<"$_out"
}

# ---------------------------------------------------------------------------
# cleanup_delete <type> <id_var_name>
#   Deletes the resource identified by the variable named <id_var_name> if
#   that variable is non-empty. Always returns 0 so cleanup traps never abort.
# ---------------------------------------------------------------------------
cleanup_delete() {
    local type="$1"
    local id_var="$2"
    local id="${!id_var:-}"
    local out rc
    if [ -n "$id" ]; then
        printf '[%s] cleanup: delete %s %s\n' "${LABEL}" "$type" "$id"
        local _out _rc _tmpdir _stdout _stderr
        _tmpdir=$(mktemp -d)
        wbctl delete "$type" "$id" >"$_tmpdir/stdout" 2>"$_tmpdir/stderr" && _rc=0 || _rc=$?
        _out=$(cat "$_tmpdir/stdout")
        cat "$_tmpdir/stderr" >&2
        if [ "$_rc" -ne 0 ]; then
            printf '[%s] WARN: cleanup delete %s %s failed (exit %d): %s\n' \
                "${LABEL}" "$type" "$id" "$_rc" "$_out" >&2
        elif [ "$_out" != "$id" ]; then
            if [ -z "$_out" ]; then
                printf '[%s] WARN: cleanup delete %s %s produced no output\n' \
                    "${LABEL}" "$type" "$id" >&2
            else
                printf '[%s] WARN: cleanup delete %s %s returned unexpected output: %s\n' \
                    "${LABEL}" "$type" "$id" "$_out" >&2
            fi
        fi
    fi
}

# ---------------------------------------------------------------------------
# cleanup_cmd <args>
#   Runs the cleanup command. Always returns 0 so cleanup traps never abort.
# ---------------------------------------------------------------------------
cleanup_cmd() {
    local out rc
    printf '[%s] cleanup: %s\n' "${LABEL}" "$*"
    local _out _rc _tmpdir _stdout _stderr
    _tmpdir=$(mktemp -d)
    "$@" >"$_tmpdir/stdout" 2>"$_tmpdir/stderr" && _rc=0 || _rc=$?
    _out=$(cat "$_tmpdir/stdout")
    cat "$_tmpdir/stderr" >&2
    if [ "$_rc" -ne 0 ]; then
        printf '[%s] WARN: cleanup cmd %s failed (exit %d): %s\n' \
            "${LABEL}" "$*" "$_rc" "$_out" >&2
    fi
}


# ---------------------------------------------------------------------------
# print_summary
#   Prints the PASS/FAIL footer. Exits 1 when any test failed, 0 otherwise.
# ---------------------------------------------------------------------------
print_summary() {
    printf '%s\n' "$_SEP"
    printf 'PASS: %d  FAIL: %d\n' "$_PASS" "$_FAIL"
    if [ "$_FAIL" -gt 0 ]; then
        exit 1
    fi
}
