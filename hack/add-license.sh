#!/usr/bin/env bash
# Copyright 2019 The Kubernetes Authors.
# SPDX-License-Identifier: Apache-2.0

set -x
set -e
set -o pipefail
set -o nounset

if [[ -z "${1-}" ]] ; then
  echo "Usage: $0 <mode>"
  echo "Example: $0 check"
  exit 1
fi

if [[ $1 == "check" || $1 == "run" ]]; then
  mode=$1
else
  echo "Error: mode must be check or run"
  exit 1
fi

args=(
  -y 2022
  -c "The Kubernetes Authors."
  -f LICENSE_TEMPLATE
  -ignore "kyaml/internal/forked/github.com/**/*"
  -ignore "site/**/*"
  -ignore "**/*.md"
  -ignore "**/*.json"
  -ignore "**/*.yml"
  -ignore "**/*.yaml"
  -ignore "**/*.xml"
  -v
)
if [[ $mode == "check" ]]; then
  args+=(-check)
  if  ! addlicense "${args[@]}" .  ; then
    set +x
    echo -e "\n------------------------------------------------------------------------"
    echo "Error: license missing in one or more files. Run \`$0 run\` to update them."
    exit 1
  fi
  exit 0
fi

addlicense "${args[@]}" .
