#!/usr/bin/env bash
# Copyright 2019 The Kubernetes Authors.
# SPDX-License-Identifier: Apache-2.0
#
# In general, pin modules to a specific version of the
# kustomize API before a release of that module, and
# unpin the module after the module release so that
# development proceeds against the API's HEAD.
#
# E.g. for the kustomize CLI module, do this before
# releasing the CLI:
#
#   ./hack/pinUnpin.sh pin kustomize v0.3.1
#
# where v0.3.1 is the most recently released version of
# the API, and do the following afterwards:
#
#   ./hack/pinUnpin.sh unPin kustomize

set -o nounset
set -o pipefail

if [ "$#" -lt 2 ]; then
  echo "usage:"
  echo "  ./hack/pinUnpin.sh pin kustomize v0.3.1"
  echo " or "
  echo "  ./hack/pinUnpin.sh unPin kustomize"
  exit 1
fi

operation=$1
if [[ ("$operation" != "pin") && ("$operation" != "unPin") ]]; then
  echo "unknown operation $operation"
  exit 1
fi

module=$2
if [ ! -d "$module" ]; then
  echo "directory $module doesn't exist"
  exit 1
fi

version="unnecessary"
if [ "$operation" == "pin" ]; then
  if [ "$#" -le 2 ]; then
    echo "Specify version to pin, e.g. '$0 $module pin v0.2.0'"
    exit 1
  fi
  version=$3
fi

function unPin {
  oldV=$(grep -m 1 sigs.k8s.io/kustomize/api go.mod | awk '{print $NF}')
  go mod edit -replace=sigs.k8s.io/kustomize/api@${oldV}=../api
  go mod tidy
}

function pin {
  oldV=$(grep -m 1 sigs.k8s.io/kustomize/api go.mod | awk '{print $NF}')
  go mod edit -dropreplace=sigs.k8s.io/kustomize/api@${oldV}
  go mod edit -require=sigs.k8s.io/kustomize/api@$version
  go mod tidy
}

pushd $module >& /dev/null
$operation
popd >& /dev/null
