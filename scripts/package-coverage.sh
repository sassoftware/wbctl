#!/bin/sh
# Copyright © 2026, SAS Institute Inc., Cary, NC, USA.  All Rights Reserved.
# SPDX-License-Identifier: Apache-2.0
set -eu

profile=${1:-build/reports/coverage.out}

if [ ! -f "$profile" ]; then
  printf 'Error: coverage profile not found: %s\n' "$profile" >&2
  exit 1
fi

awk '
  BEGIN {
    # FS splits go coverage lines into: location, statement-count, hit-count.
    FS = "[[:space:]]+"
  }
  /^mode:[[:space:]]/ { next }
  NF < 3 { next }
  {
    # loc is "path/to/file.go:startLine.startCol,endLine.endCol".
    loc = $1
    # colon separates file path from range metadata.
    colon = index(loc, ":")
    if (colon == 0) next

    # file is the full source path.
    file = substr(loc, 1, colon - 1)
    # range is start/stop line.column span.
    range = substr(loc, colon + 1)
    # comma splits start and stop positions inside range.
    comma = index(range, ",")
    if (comma == 0) next

    # start/stop are "line.column".
    start = substr(range, 1, comma - 1)
    stop = substr(range, comma + 1)
    # dot1/dot2 find the line/column separator.
    dot1 = index(start, ".")
    dot2 = index(stop, ".")
    if (dot1 == 0 || dot2 == 0) next

    # pkg is the directory path used as the package grouping key.
    pkg = file
    sub("/[^/]*$", "", pkg)
    if (pkg == file) pkg = "."

    # sline/eline are numeric line bounds for the covered block.
    sline = substr(start, 1, dot1 - 1) + 0
    eline = substr(stop, 1, dot2 - 1) + 0
    # cnt is execution count for the whole block; >0 means covered.
    cnt = $3 + 0

    # Expand each block to per-line keys so overlapping blocks dedupe cleanly.
    for (ln = sline; ln <= eline; ln++) {
      # key uniquely identifies a source line across all records.
      key = file ":" ln
      # seen tracks discovered lines once; line_pkg maps each line key to package.
      if (!(key in seen)) {
        seen[key] = 1
        line_pkg[key] = pkg
      }
      # covered marks lines hit by any record with cnt > 0.
      if (cnt > 0) covered[key] = 1
    }
  }
  END {
    # k iterates line keys.
    # p is package for that key.
    # total[p] counts known lines per package.
    # cov[p] counts covered lines.
    for (k in seen) {
      p = line_pkg[k]
      total[p]++
      if (k in covered) cov[p]++
    }
    # maxlen stores widest package name for aligned output formatting.
    maxlen = 0
    for (p in total) {
      if (length(p) > maxlen) maxlen = length(p)
    }
    maxlen = maxlen
    # Emit helper records as: group^sort-key^display-text
    # group keeps header before data, sort-key orders packages, display-text is final row.
    printf "0^Package name^# %-" maxlen "s | %s\n", "Package name", "Coverage"
    for (p in total) {
      # c/t are covered/total lines for package p.
      # pct is computed percentage.
      c = cov[p] + 0
      t = total[p] + 0
      pct = (t > 0) ? (100 * c / t) : 0
      if (pct < 100) {
        printf "1^%s^- %-" maxlen "s | %6.2f%% (%d/%d lines)\n", p, p, pct, c, t
      } else {
        printf "1^%s^+ %-" maxlen "s | %6.2f%% (%d/%d lines)\n", p, p, pct, c, t
      }
    }
  }
' "$profile" |
  # awk emits records in this helper format: group^sort-key^display-text
  # - group: numeric bucket (0 for the header row, 1 for package rows)
  # - sort-key: package name used for alphabetical ordering within group 1
  # - display-text: final row content to print after helper fields are removed
  # sort args:
  # -t '^'   -> use '^' as the field delimiter
  # -k1,1n   -> primary key: field 1, numeric (0 before 1)
  # -k2,2    -> secondary key: field 2, lexicographic package name
  LC_ALL=C sort -t '^' -k1,1n -k2,2 |
  # cut args:
  # -d'^'    -> split fields using '^'
  # -f3-     -> keep field 3 through end (drop helper fields 1 and 2)
  cut -d'^' -f3-
