#!/usr/bin/env bash
# Copyright 2019 The Kubernetes Authors.
# SPDX-License-Identifier: Apache-2.0

set -euo pipefail

if [[ -z "${1-}" ]] ; then
  echo "Usage: $0 <cmd>"
  echo "Example: $0 lint"
  exit 1
fi
cmd=$1

# Sanity check. Update if you add a Go module.
EXPECTED_MODULE_COUNT=${EXPECTED_MODULE_COUNT:-49}

# Skip paths. Comma delimited. Must start with `./` to match find results.
SKIP_MODULE_PATHS="${SKIP_MODULE_PATHS:-}"

# Always skip vendored Hugo themes, since we don't control their Makefiles.
if [[ -n "${SKIP_MODULE_PATHS}" ]]; then
  SKIP_MODULE_PATHS+=","
fi
SKIP_MODULE_PATHS+='./site/themes/*'

seen=()
skipped=()

# Hack scripts must be run from the root of the repository.
KUSTOMIZE_ROOT=$(pwd)
export KUSTOMIZE_ROOT

# Ignore vendored Hugo themes, since we don't control their Makefiles.
FIND_FILTER_ARGS=()

# Loop through the SKIP_MODULE_PATHS and add any modules to `skipped`.
while IFS='' read -r filter_path; do
  FIND_FILTER_ARGS+=(-not -path "${filter_path}")
   echo "Skipped Path: $filter_path"

  # Record skipped modules
  for i in $(find . -name go.mod -path "${filter_path}"); do
    dir=$(cd "$(dirname "$i")" && pwd)
    module="${dir#"$KUSTOMIZE_ROOT"}"
    skipped+=("$module")
  done
done <<< "$(echo "${SKIP_MODULE_PATHS}" | tr ',' '\n')"

if [[ ${#skipped[@]} -gt 0 ]]; then
  echo -e "Skipped the following modules:"
  printf "  - %s\n" "${skipped[@]}"
fi

# Loop through the not-skipped modules and run the specified command on each
for i in $(find . -name go.mod "${FIND_FILTER_ARGS[@]}"); do
  pushd . > /dev/null
  cd $(dirname "$i");

  dir=$(pwd)
  module="${dir#"$KUSTOMIZE_ROOT"}"
  echo -e "\n----------------------------------------------------------"
  echo "Running command in $module"
  echo -e "----------------------------------------------------------"
  echo "$cmd"
  bash -c "$cmd"
  seen+=("$module")
  popd > /dev/null
done

echo -e "\n\n----------------------------------------------------------"
echo -e "SUCCESS: Ran '$cmd' on the following modules:"
printf "  - %s\n" "${seen[@]}"

FOUND_MODULE_COUNT="$((${#seen[@]} + ${#skipped[@]}))"
if [[ $FOUND_MODULE_COUNT -ne $EXPECTED_MODULE_COUNT ]]; then
  echo
  echo "SANITY CHECK FAILURE: Expected to see & skip $EXPECTED_MODULE_COUNT modules, but saw ${#seen[@]} and skipped ${#skipped[@]}"
  exit 1
fi
