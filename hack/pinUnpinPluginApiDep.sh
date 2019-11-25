#!/usr/bin/env bash
# Copyright 2019 The Kubernetes Authors.
# SPDX-License-Identifier: Apache-2.0
#
# To fix all plugin dependence to a particular
# released version of the kustomize API,
# run this from the repo root:
#
#   ./hack/pinUnpinPluginApiDep.sh pin v0.2.0
#
# To replace fixed dependence with
# dependence on local filesystem (HEAD)
# run this from the repo root:
#
#   ./hack/pinUnpinPluginApiDep.sh unPin
#
# All plugins, even plugins not written in Go,
# have a unit test written in Go that depends
# on a particular version of the api for a test
# harness.  The plugins written in Go, either
# as exec or Go-plugin style plugins,
# will likely depend directly on the kustomize
# API, and any number of other 3rd party packages.
#
# The Go plugins in the `builtin` directory
# are in practice converted to static libraries
# in the API, so should remain unpinned (dependent
# on HEAD). The other example plugins can be pinned
# or unpinned on a case by case basis, since
# they are just examples - but likely should
# remain unpinned too.  Nothing in the outside
# world should depend on these plugin modules,
# so there's no reason for them to be properly
# pinned.
#
# An external plugin author will obviously
# want to pin their plugin to some set of fixed
# dependencies, and the go.mod files for the
# plugins in this repo are just examples of how
# to do so.

set -o errexit
set -o nounset
set -o pipefail

function doUnPin {
  oldV=$(grep -m 1 sigs.k8s.io/kustomize/api go.mod | awk '{print $NF}')
  go mod edit -replace=sigs.k8s.io/kustomize/api@${oldV}=$1
  go mod tidy
}

function doPin {
  oldV=$(grep -m 1 sigs.k8s.io/kustomize/api go.mod | awk '{print $NF}')
  go mod edit -dropreplace=sigs.k8s.io/kustomize/api@${oldV}
  go mod edit -require=sigs.k8s.io/kustomize/api@$1
  go mod tidy
}

function forEachGoMod {
  for goMod in $(find $2 -name 'go.mod'); do
    d=$(dirname "${goMod}")
    echo "$1 $d"
    (cd $d; $1 $3)
  done
}

function unPin {
  forEachGoMod doUnPin ./plugin/builtin                 ../../../api
  forEachGoMod doUnPin ./plugin/someteam.example.com/v1 ../../../../api
}

function pin {
  forEachGoMod doPin ./plugin/builtin                 $version
  forEachGoMod doPin ./plugin/someteam.example.com/v1 $version
}

operation=$1
if [[ ("$operation" != "pin") && ("$operation" != "unPin") ]]; then
  echo "unknown operation $operation"
  exit 1
fi

version="unnecessary"
if [ "$operation" == "pin" ]; then
  if [ "$#" -le 1 ]; then
    echo "Specify version to pin, e.g. '$0 pin v0.2.0'"
    exit 1
  fi
  version=$2
fi

$operation
