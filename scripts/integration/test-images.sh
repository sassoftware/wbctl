#!/usr/bin/env bash
# Copyright © 2026, SAS Institute Inc., Cary, NC, USA.  All Rights Reserved.
# SPDX-License-Identifier: Apache-2.0
#
# test-images.sh — full image lifecycle test.
#
# Steps: dry-run create image → create image → get → describe → rebuild →
#        get image-builds → describe image-build → stream build logs (5s) →
#        delete image (via EXIT trap).
# Can be run standalone or via integration-test.sh.
#
# The image spec uses an inline JSON string (`--spec`) so no external context
# is required beyond a reachable API server.

set -euo pipefail
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
LABEL="test-images"
source "$SCRIPT_DIR/common.sh"

require_wbconfig

# -- cleanup trap (registered before any resource is created) ---------------
IMAGE_ID=""
cleanup_image() {
    cleanup_delete image IMAGE_ID
    IMAGE_ID=""
}
trap 'cleanup_image' EXIT

TS=$(date +%s)

# Minimal image spec used for all image tests.
SPEC='{"version":"3"}'

# -- tests ------------------------------------------------------------------

run_test_output_contains \
    "dry-run create image contains name field" \
    '"name"' \
    wbctl create image \
        --dry-run \
        --name "it-${TS}-img-dry" \
        --spec "$SPEC"

IMAGE_ID=$(wbctl create image \
    --name "it-${TS}-img" \
    --spec "$SPEC") || IMAGE_ID=""

if [ -n "$IMAGE_ID" ]; then
    run_test \
        "create image exit 0" \
        true

    run_test_output_contains \
        "get images contains ID header" \
        "ID" \
        wbctl get images

    run_test_json_array \
        "get images --output json returns JSON array" \
        wbctl get images --output json

    run_test_output_contains \
        "get images contains $IMAGE_ID" \
        "$IMAGE_ID" \
        wbctl get images

    IMAGE_EDIT_NAME="it-${TS}-img-edit"
    run_test_edit_name_round_trip \
        "edit image $IMAGE_ID" \
        image \
        "$IMAGE_ID" \
        "$IMAGE_EDIT_NAME"

    run_test_output_contains \
        "describe image contains $IMAGE_ID" \
        "$IMAGE_ID" \
        wbctl describe image "$IMAGE_ID"

    IMAGE_UPDATE_NAME="it-${TS}-img-upd"
    run_test_output_contains \
        "dry-run update image $IMAGE_ID contains updated name" \
        "$IMAGE_UPDATE_NAME" \
        wbctl update image \
            --dry-run \
            --name "$IMAGE_UPDATE_NAME" \
            "$IMAGE_ID"

    run_test_output_contains \
        "update image $IMAGE_ID returns id" \
        "$IMAGE_ID" \
        wbctl update image \
            --name "$IMAGE_UPDATE_NAME" \
            "$IMAGE_ID"

    run_test_output_contains \
        "get image $IMAGE_ID returns updated name" \
        "$IMAGE_UPDATE_NAME" \
        wbctl get image \
            "$IMAGE_ID"

    run_test \
        "rebuild image exit 0" \
        wbctl rebuild image "$IMAGE_ID"

    # image-builds are created by the rebuild; give the API a moment to record it.
    wait_with_notice 2 "waiting for image-build record after rebuild"

    run_test_output_contains \
        "get image-builds contains ID header" \
        "ID" \
        wbctl get image-builds

    run_test_json_array \
        "get image-builds --output json returns JSON array" \
        wbctl get image-builds --output json

    BUILD_ID=$(capture_id image-builds)
    if [ -n "$BUILD_ID" ]; then
        run_test_output_contains \
            "describe image-build contains $BUILD_ID" \
            "$BUILD_ID" \
            wbctl describe image-build "$BUILD_ID"

        # Stream build logs for up to 5 s; timeout is treated as success.
        printf '[%s] INFO: waiting up to 5s while streaming image-build logs\n' "$LABEL"
        stream_image_build_logs() {
            timeout 5 "$WBCTL_BIN" logs image-builds "$BUILD_ID"
            rc=$?
            [ "$rc" -eq 0 ] || [ "$rc" -eq 124 ]
        }
        run_test \
            "stream log image-build (5s timeout ok)" \
            stream_image_build_logs
    else
        printf '[%s] INFO: no image-builds found after rebuild; build log/describe skipped.\n' "$LABEL"
    fi
else
    _FAIL=$((_FAIL + 1))
    printf '[%s] %-55s %s\n' "$LABEL" "create image" "${_RED}FAIL${_RST}"
    printf '  create image returned empty ID\n' >&2
fi

# cleanup output first, then footer
cleanup_image

# -- summary ----------------------------------------------------------------
print_summary
