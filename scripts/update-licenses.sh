#!/bin/bash
# Copyright © 2026, SAS Institute Inc., Cary, NC, USA.  All Rights Reserved.
# SPDX-License-Identifier: Apache-2.0

set -eu -o pipefail

die() {
  msg="${1}"
  echo "${msg}"
  exit 1
}

licenses() {

  workspace_root="$(git rev-parse --show-toplevel)"
  cd "${workspace_root}" || die "unable to cd to root: ${workspace_root}"
  mkdir -p "${workspace_root}/licenses" || die "unable to create ${workspace_root}/licenses"
  index="${workspace_root}/licenses/licenses.md"
  echo "# Third-party licenses for this project" > "${index}"
  echo "Name | License" >> "${index}"
  echo "--- | ---" >> "${index}"
  modules="$(go list -m -f '{{if not (or .Indirect .Main)}}{{.Path}}@{{.Version}}{{end}}' all)"

  echo "$modules" | while read -r module; do
      echo "Downloading files for $module..."

      # Get local module directory
      src_dir="$(go mod download -json "${module}" | jq -r '.Dir')"

      # Look for license related files in the module directory
      file_names=("LICENSE" "NOTICE")
      for file in "${file_names[@]}"
      do
          src_path="$(find $src_dir -maxdepth 1 -name "${file}*")"
          if [ -f "$src_path" ]
          then
              # Calculate destination file name
              dst_name="${module}"                  # start with module name
              dst_name="${dst_name%@v*}"            # remove version from module
              dst_name="${dst_name//\//_}-${file}"  # replace all / with _
              # Copy file content to the licenses folder
              echo "    ...Copying licenses/${dst_name}"
              cp "${src_path}" "licenses/${dst_name}"
              # Append module information to the index file
              echo "[${module%@v*}](https://${module%@v*}) | [${dst_name}](${dst_name})" >> "${index}"
          fi
      done
  done
  # make sure all licenses are writeable
  chmod +w licenses/*
}

licenses
