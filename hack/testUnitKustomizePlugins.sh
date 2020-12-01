#!/usr/bin/env bash
#
# Copyright 2019 The Kubernetes Authors.
# SPDX-License-Identifier: Apache-2.0

# Run this script with no arguments from the repo root
# to test all the plugins.

# Want this to keep going even if one test fails,
# to see how many pass, so do not errexit.
set -o nounset
# set -o errexit
set -o pipefail

rcAccumulator=0

# All hack scripts should run from top level.
. hack/shellHelpers.sh

function runTest {
  local file=$1
  local code=0
  if grep -q "// +build notravis" "$file"; then
    if onLinuxAndNotOnRemoteCI; then
      go test -v -tags=notravis $file
      code=$?
    else
      # TODO: make work for non-linux
      echo "Not on linux or on remote CI; skipping $file"
    fi
  else
    go test -v $file
    code=$?
  fi
  rcAccumulator=$((rcAccumulator || $code))
  if [ $code -ne 0 ]; then
    echo "Failure in $d"
  fi
}

function scanDir {
  pushd $1 >& /dev/null
  echo "Testing $1"
  for t in $(find . -name '*_test.go'); do
    runTest $t
  done
  popd >& /dev/null
}

if onLinuxAndNotOnRemoteCI; then
  # Some of these tests have special deps.
  make $(go env GOPATH)/bin/helmV2
  make $(go env GOPATH)/bin/helmV3
  make $(go env GOPATH)/bin/helm
  make $(go env GOPATH)/bin/kubeval
fi

for goMod in $(find ./plugin -name 'go.mod' -not -path "./plugin/untested/*"); do
  d=$(dirname "${goMod}")
  scanDir $d
done

if [ $rcAccumulator -ne 0 ]; then
  echo "FAILURE; exit code $rcAccumulator"
  exit 1
fi


