#!/usr/bin/env bash
# Copyright 2019 The Kubernetes Authors.
# SPDX-License-Identifier: Apache-2.0
#
# To fix all plugin dependence to a particular
# released version of the kustomize API,
# run this from the repo root:
#
#   ./hack/pinUnpinPluginApiDep.sh pin api v0.2.0
#
# To replace fixed dependence with
# dependence on local filesystem (HEAD)
# run this from the repo root:
#
#   ./hack/pinUnpinPluginApiDep.sh unPin api
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
# so there's no reason for them to be pinned.

set -o errexit
set -o nounset
#set -o pipefail

function doUnPin {
  oldV=$(grep -m 1 sigs.k8s.io/kustomize/${module} go.mod | awk '{print $NF}')
  if [ ! -z $oldV ]; then
    go mod edit -replace=sigs.k8s.io/kustomize/${module}@${oldV}=$1
  fi
  go mod tidy
}

function doPin {
  oldV=$(grep -m 1 sigs.k8s.io/kustomize/${module} go.mod | awk '{print $NF}')
  if [ ! -z $oldV ]; then
    go mod edit -dropreplace=sigs.k8s.io/kustomize/${module}@${oldV}
    go mod edit -require=sigs.k8s.io/kustomize/${module}@$1
  fi
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
  echo "Unpinning $module"
  forEachGoMod doUnPin ./plugin/builtin                 ../../../${module}
  forEachGoMod doUnPin ./plugin/someteam.example.com/v1 ../../../../${module}
}

function pin {
  echo "Pinning $module to $version"
  forEachGoMod doPin ./plugin/builtin                 ${version}
  forEachGoMod doPin ./plugin/someteam.example.com/v1 ${version}
}

if [ "$#" -eq 0 ]; then
  echo "Pin or unpin plugins, e.g."
  echo " "
  echo "  ./hack/pinUnpinPluginApiDep.sh pin api v0.2.0"
  echo " "
  echo "  ./hack/pinUnpinPluginApiDep.sh unPin api"
  echo " "
  exit 1
fi

operation=$1
if [[ ("$operation" != "pin") && ("$operation" != "unPin") ]]; then
  echo "unknown operation $operation"
  exit 1
fi

module=$2
if [[ ("$module" != "api") && ("$module" != "kyaml") ]]; then
  echo "unknown module $module"
  exit 1
fi

version="unnecessary"
if [ "$operation" == "pin" ]; then
  if [ "$#" -le 2 ]; then
    echo "Specify version to pin, e.g. '$0 pin v0.2.0'"
    exit 1
  fi
  version=$3
fi

$operation
