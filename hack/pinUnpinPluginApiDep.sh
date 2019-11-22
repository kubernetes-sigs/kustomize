#!/usr/bin/env bash
# Copyright 2019 The Kubernetes Authors.
# SPDX-License-Identifier: Apache-2.0

# This script unpins or repins the Go kustomize API dependence
# for all the plugins in the repo.

# Run from repo root, e.g.
#
#   ./hack/pinUnpinPluginApiDep.sh pin v0.2.0
#
# or
#
#   ./hack/pinUnpinPluginApiDep.sh unPin
#

set -o errexit
set -o nounset
set -o pipefail

operation=$1
if [[ ("$operation" != "pin") && ("$operation" != "unPin") ]]; then
  echo "unknown operation $operation"
  exit 1
fi

version="unused"
if [ "$operation" == "pin" ]; then
  if [ "$#" -le 1 ]; then
    echo "must specify version to pin"
    exit 1
  else
    version=$2
  fi
fi

function doUnPin {
  # TODO fix bug where there's only one required module
  # (need $3, not $2).
  oldV=$(grep -m 1 sigs.k8s.io/kustomize/api go.mod | awk '{print $2}')
  go mod edit -replace=sigs.k8s.io/kustomize/api@${oldV}=$1
  go mod tidy
}

function doPin {
  oldV=$(grep -m 1 sigs.k8s.io/kustomize/api go.mod | awk '{print $2}')
  go mod edit -dropreplace=sigs.k8s.io/kustomize/api@${oldV}
  go mod edit -dropreplace=sigs.k8s.io/kustomize/api@v0.1.1
  go mod edit -require=sigs.k8s.io/kustomize/api@$1
  go mod tidy
}

function forEachGoMod {
  for goMod in $(find $2 -name 'go.mod'); do
    d=$(dirname "${goMod}")
    echo $d
    (cd $d; $1 $3)
  done
}

function unPin {
  forEachGoMod doUnPin ./plugin/builtin              ../../../api
#  forEachGoMod doUnPin ./plugin/someteam.example.com ../../../../api
}

function pin {
  forEachGoMod doPin ./plugin/builtin              $version
#  forEachGoMod doPin ./plugin/someteam.example.com $version
}

$operation
