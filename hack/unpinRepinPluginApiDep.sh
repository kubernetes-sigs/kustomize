#!/usr/bin/env bash
# Copyright 2019 The Kubernetes Authors.
# SPDX-License-Identifier: Apache-2.0

# This script unpins or repins the Go kustomize API dependence
# for all the plugins in the repo.

# Run this from repo root, e.g.
#
#   ./hack/unpinRepinPluginApiDep.sh unPin v0.1.1
#

set -o errexit
set -o nounset
set -o pipefail

operation=$1
version=$2

if [[ ("$operation" != "unPin") && ("$operation" != "rePin") ]]; then
  echo "unknown operation $operation"
  exit 1
fi

function addReplace {
  go mod edit -replace=sigs.k8s.io/kustomize/api@${version}=$1
  go mod tidy
}

function dropReplace {
  go mod edit -dropreplace=sigs.k8s.io/kustomize/api@${version}
  go mod tidy
}

function forEachGoMod {
  for goMod in $(find $2 -name 'go.mod'); do
    d=$(dirname "${goMod}")
    echo $d
    (cd $d; $1 $3 )
  done
}

function unPin {
  forEachGoMod addReplace ./plugin/builtin              ../../../api
  forEachGoMod addReplace ./plugin/someteam.example.com ../../../../api
}

function rePin {
  forEachGoMod dropReplace ./plugin/builtin
  forEachGoMod dropReplace ./plugin/someteam.example.com
}

$operation
