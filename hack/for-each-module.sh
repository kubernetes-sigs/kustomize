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

seen=""
kustomize_root=$(pwd | sed 's|kustomize/.*|kustomize/|')

# verify all modules pass validation
for i in $(find . -name go.mod -not -path "./site/*"); do
  pushd .
  cd $(dirname $i);

  set +x
  dir=$(pwd)
  module="${dir#$kustomize_root}"
  echo -e "\n----------------------------------------------------------"
  echo "Running command in $module"
  echo -e "----------------------------------------------------------"
  set -x

  bash -c "$cmd"
  seen+="  - $module\n"
  popd
done

set +x
echo -e "\n\n----------------------------------------------------------"
echo -e "SUCCESS: Ran '$cmd' on the following modules:"
echo -e "$seen"
