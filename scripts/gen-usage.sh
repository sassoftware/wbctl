#!/usr/bin/env bash
# Copyright © 2026, SAS Institute Inc., Cary, NC, USA.  All Rights Reserved.
# SPDX-License-Identifier: Apache-2.0
#
# This script generates the usage documentation for the command passed as the
# first argument into the ./doc directory.
set -eufo pipefail

usage() {
  local file="$1"
  shift
  local depth="$1"
  shift
  local cmd="$@"
  local helptext=$($cmd --help)
  echo Generating usage doc for: $cmd
  local subcommands=()
  while read -r subcommand
  do
    subcommands+=("$subcommand")
  done < <($cmd --help | awk -F' ' 'BEGIN{cmds=0} /Available Commands:/{cmds=1;getline} /^$/{cmds=0} {if(cmds) print $1}')
  if [ ${#subcommands[@]} -gt 0 ]
  then
    local tmpfile=$(mktemp)
    trap 'rm -f "$tmpfile"' RETURN
    local subcommandA
    for subcommandA in "${subcommands[@]}"
    do
      if [ "$subcommandA" = "help" ]
      then
        continue
      fi
      subcommandA="${subcommandA%%,*}"
      local base="${cmd%%/ }"
      base=${base##*/}
      local subfile="${base} ${subcommandA}"
      usage "${subfile}" $((depth+1)) ${cmd} ${subcommandA}
      echo "s_\([^a-z0-9-]\)${subcommandA}\([^a-z0-9-]\)_\1<a href='${base} ${subcommandA}.md'>${subcommandA}</a>\2_" >> "${tmpfile}"
    done
  fi
  echo "<pre>" > "doc/${file}.md"
  sed -f "${tmpfile}" <<< "${helptext}" >> "doc/${file}.md"
  echo "</pre>" >> "doc/${file}.md"
}

cd $(git rev-parse --show-toplevel) || exit 1
usage $(basename $1) 0 "$@"
