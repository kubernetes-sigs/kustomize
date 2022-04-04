#!/usr/bin/env bash
# Copyright 2019 The Kubernetes Authors.
# SPDX-License-Identifier: Apache-2.0

set -x
set -e
set -o pipefail
set -o nounset

if [[ -z "${1-}" ]] ; then
  echo "Usage: $0 <cmd>"
  echo "Example: $0 lint"
  exit 1
fi

cmd=$1

seen=()
# Hack scripts must be run from the root of the repository.
KUSTOMIZE_ROOT=$(pwd)
export KUSTOMIZE_ROOT

# verify all modules pass validation
for i in $(find . -name go.mod -not -path "./site/*"); do
  pushd .
  cd $(dirname $i);

  set +x
  dir=$(pwd)
  module="${dir#$KUSTOMIZE_ROOT}"
  echo -e "\n----------------------------------------------------------"
  echo "Running command in $module"
  echo -e "----------------------------------------------------------"
  set -x

  bash -c "$cmd"
  seen+=("$module")
  popd
done

set +x
echo -e "\n\n----------------------------------------------------------"
echo -e "SUCCESS: Ran '$cmd' on the following modules:"
printf "  - %s\n" "${seen[@]}"

EXPECTED_MODULE_COUNT=44
if [[ "${#seen[@]}" -ne $EXPECTED_MODULE_COUNT ]]; then
  echo
  echo "SANITY CHECK FAILURE: Expected to see $EXPECTED_MODULE_COUNT modules, but saw ${#seen[@]}"
  exit 1
fi
