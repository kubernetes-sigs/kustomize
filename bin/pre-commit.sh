#!/bin/bash
set -e

# Make sure, we run in the root of the repo and
# therefore run the tests on all packages
base_dir="$( cd "$(dirname "$0")/.." && pwd )"
cd "$base_dir" || {
  echo "Cannot cd to '$base_dir'. Aborting." >&2
  exit 1
}

rc=0

function runFunc {
  local name=$1
  local result="SUCCESS"
  printf "============== begin %s\n" "$name"
  $name
  local code=$?
  rc=$((rc || $code))
  if [ $code -ne 0 ]; then
    result="FAILURE"
  fi
  printf "============== end %s : %s code=%d\n\n\n" "$name" "$result" $code
}

function testGoLangCILint {
  golangci-lint run ./...
}

function testGoTest {
  go test -v ./...
}

# These tests require the helm program, and at the moment
# we're not asking travis to install helm.
function testNoTravisGoTest {
  go test -v sigs.k8s.io/kustomize/pkg/target \
      -run TestChartInflatorPlugin -tags=notravis
  go test -v sigs.k8s.io/kustomize/plugin/... \
    -run TestChartInflator -tags=notravis
  mdrip --mode test --label helmtest README.md ./examples/chart.md
}

function testExamples {
  mdrip --mode test --label test README.md ./examples
}

# Use of GOPATH is optional if go modules are
# used.  This script tries to work for people who
# don't have GOPATH set, and work for travis.
#
# Upon entry, travis has GOPATH set, and used it
# to install mdrip and the like.
#
# Use GOPATH to define XDG_CONFIG_HOME, then unset
# GOPATH so that go.mod is unambiguously honored.
echo "GOPATH=$GOPATH"

if [ -z ${GOPATH+x} ]; then
  echo GOPATH is unset
  tmp=$HOME/gopath
  if [ -d "$tmp" ]; then
    oldGoPath=$tmp
  else
    tmp=$HOME/go
    if [ -d "$tmp" ]; then
      oldGoPath=$tmp
    fi
  fi
else
  oldGoPath=$GOPATH
  unset GOPATH
fi
echo "oldGoPath=$oldGoPath"
export XDG_CONFIG_HOME=$oldGoPath/src/sigs.k8s.io
echo "XDG_CONFIG_HOME=$XDG_CONFIG_HOME"
if [ ! -d "$XDG_CONFIG_HOME" ]; then
  echo "$XDG_CONFIG_HOME is not a directory."
	exit 1
fi

# Until go v1.13, set this explicitly.
export GO111MODULE=on

echo "HOME=$HOME"
echo "GOPATH=$GOPATH"
echo "GO111MODULE=$GO111MODULE"
echo pwd=`pwd`
echo " "
echo "Working..."

runFunc testGoLangCILint
runFunc testGoTest

if [ -z ${TRAVIS+x} ]; then
  echo Not on travis, so running the notravis tests
  runFunc testNoTravisGoTest
fi

PATH=$HOME/go/bin:$PATH
runFunc testExamples

if [ $rc -eq 0 ]; then
  echo "SUCCESS!"
else
  echo "FAILURE; exit code $rc"
fi

exit $rc
