#!/bin/bash
# Copyright © 2026, SAS Institute Inc., Cary, NC, USA.  All Rights Reserved.
# SPDX-License-Identifier: Apache-2.0

set -euo pipefail

copyright() {
  year="$(date +%Y)"
  copyright="Copyright © $year, SAS Institute Inc., Cary, NC, USA.  All Rights Reserved."
  spdx="SPDX-License-Identifier: Apache-2.0"
  go_copyright="// $copyright"
  go_spdx="// $spdx"
  bat_copyright="rem $copyright"
  bat_spdx="rem $spdx"
  sh_copyright="# $copyright"
  sh_spdx="# $spdx"

  find . \( -type f -name '*.go' -o -type f -name '*.bat' -o -type f -name '*.sh' \) -print0 |
  while IFS= read -r -d '' file; do
    [[ ! -s "$file" ]] && continue
    tmp="$(mktemp)"
    case $(uname -s) in
      Darwin) mode="$(stat -f '%Lp' "$file")" ;;
      *)      mode="$(stat -c '%a' "$file")" ;;
    esac

    case "$file" in
      *.go)
        grep -Fqx "$go_copyright" "$file" && continue
        {
          printf '%s\n' "$go_copyright"
          printf '%s\n\n' "$go_spdx"
          cat "$file"
        } >"$tmp"
        ;;
      *.bat)
        grep -Fqx "$bat_copyright" "$file" && continue
        {
          printf '%s\n' "$bat_copyright"
          printf '%s\n\n' "$bat_spdx"
          cat "$file"
        } >"$tmp"
        ;;
      *.sh)
        grep -Fqx "$sh_copyright" "$file" && continue

        if IFS= read -r first <"$file" && [[ "$first" == '#!'* ]]; then
          {
            printf '%s\n' "$first"
            printf '%s\n' "$sh_copyright"
            printf '%s\n' "$sh_spdx"
            tail -n +2 "$file"
          } >"$tmp"
        else
          {
            printf '%s\n' "$sh_copyright"
            printf '%s\n' "$sh_spdx"
            cat "$file"
          } >"$tmp"
        fi
        ;;
    esac
    echo "Updating $file"
    chmod "$mode" "$tmp"
    mv "$tmp" "$file"
  done
}

cd $(git rev-parse --show-toplevel) || exit 1
copyright
